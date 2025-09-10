package sync

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/callowaysutton/servercon/api"
)

// APIKeyNode represents a node in the LRU linked list
type APIKeyNode struct {
	key      string
	prev     *APIKeyNode
	next     *APIKeyNode
	lastUsed time.Time
}

// LRUAPIKeyManager manages API keys using LRU algorithm
type LRUAPIKeyManager struct {
	keys     map[string]*APIKeyNode
	head     *APIKeyNode
	tail     *APIKeyNode
	mutex    sync.RWMutex
	capacity int
}

// NewLRUAPIKeyManager creates a new LRU API key manager
func NewLRUAPIKeyManager(apiKeys []string) *LRUAPIKeyManager {
	if len(apiKeys) == 0 {
		return nil
	}

	manager := &LRUAPIKeyManager{
		keys:     make(map[string]*APIKeyNode),
		capacity: len(apiKeys),
	}

	// Create dummy head and tail nodes
	manager.head = &APIKeyNode{}
	manager.tail = &APIKeyNode{}
	manager.head.next = manager.tail
	manager.tail.prev = manager.head

	// Add all API keys to the LRU cache
	for _, key := range apiKeys {
		if key != "" {
			manager.addKey(key)
		}
	}

	return manager
}

// addKey adds a new API key to the LRU cache
func (lru *LRUAPIKeyManager) addKey(key string) {
	node := &APIKeyNode{
		key:      key,
		lastUsed: time.Now(),
	}

	lru.keys[key] = node
	lru.addToHead(node)
}

// addToHead adds a node right after the head
func (lru *LRUAPIKeyManager) addToHead(node *APIKeyNode) {
	node.prev = lru.head
	node.next = lru.head.next
	lru.head.next.prev = node
	lru.head.next = node
}

// removeNode removes a node from the linked list
func (lru *LRUAPIKeyManager) removeNode(node *APIKeyNode) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

// moveToHead moves a node to the head (mark as recently used)
func (lru *LRUAPIKeyManager) moveToHead(node *APIKeyNode) {
	lru.removeNode(node)
	lru.addToHead(node)
}

// GetLeastRecentlyUsedKey returns the least recently used API key
func (lru *LRUAPIKeyManager) GetLeastRecentlyUsedKey() string {
	if lru == nil {
		return ""
	}

	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if lru.tail.prev == lru.head {
		// No keys available
		return ""
	}

	// Get the least recently used key (tail's previous)
	lruNode := lru.tail.prev

	// Move it to head (mark as recently used)
	lru.moveToHead(lruNode)
	lruNode.lastUsed = time.Now()

	log.Printf("Selected LRU API key ending in ...%s (last used: %s)",
		lruNode.key[len(lruNode.key)-8:],
		lruNode.lastUsed.Format("15:04:05"))

	return lruNode.key
}

// GetAllKeysStatus returns the status of all keys for debugging
func (lru *LRUAPIKeyManager) GetAllKeysStatus() []string {
	if lru == nil {
		return nil
	}

	lru.mutex.RLock()
	defer lru.mutex.RUnlock()

	var status []string
	current := lru.head.next
	position := 1

	for current != lru.tail {
		status = append(status, fmt.Sprintf("Pos %d: ...%s (last used: %s)",
			position,
			current.key[len(current.key)-8:],
			current.lastUsed.Format("15:04:05")))
		current = current.next
		position++
	}

	return status
}

// SyncRequest represents a queued API request
type SyncRequest struct {
	URL        string
	Type       string // "servers", "os-types", "apps", "server-detail", "snapshots", "root-password", "console-credentials"
	Method     string // HTTP method: "GET", "POST", etc.
	Body       []byte // Request body for POST requests
	ServerID   int    // Only used for server-specific requests
	RetryCount int
	MaxRetries int
}

type SyncService struct {
	db            *sql.DB
	managementURL string
	lruKeyManager *LRUAPIKeyManager
	queue         chan SyncRequest
	stopChan      chan struct{}
	client        *http.Client
	serverIDs     []int
	serverIDsMux  sync.RWMutex
	requestCycle  []func() SyncRequest
	cycleIndex    int
	cycleMux      sync.Mutex
}

func NewSyncService(db *sql.DB, syncInterval time.Duration) *SyncService {
	// Get API keys from environment
	// Fetch all environment variables containing "SERVER_MANAGEMENT_KEY"
	var apiKeys []string
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 && strings.HasPrefix(parts[0], "SERVER_MANAGEMENT_KEY") {
			if parts[1] != "" {
				apiKeys = append(apiKeys, parts[1])
			}
		}
	}

	service := &SyncService{
		db:            db,
		managementURL: os.Getenv("SERVER_MANAGEMENT_URL"),
		lruKeyManager: NewLRUAPIKeyManager(apiKeys),
		queue:         make(chan SyncRequest, 100), // Smaller buffer since we're cycling
		stopChan:      make(chan struct{}),
		client:        &http.Client{Timeout: 30 * time.Second},
		serverIDs:     make([]int, 0),
		requestCycle:  make([]func() SyncRequest, 0),
		cycleIndex:    0,
	}

	service.initializeRequestCycle()
	return service
}

func (s *SyncService) Start() {
	if s.lruKeyManager == nil || s.managementURL == "" {
		log.Println("Warning: SERVER_MANAGEMENT_KEY or SERVER_MANAGEMENT_URL not set, sync service disabled")
		return
	}

	log.Println("Starting continuous sync service with LRU API key balancing...")

	// Log initial API key status
	keyStatus := s.lruKeyManager.GetAllKeysStatus()
	log.Printf("Loaded %d API keys:", len(keyStatus))
	for _, status := range keyStatus {
		log.Printf("  %s", status)
	}

	// Start the single combined processor (3 request per minute)
	go s.processNextRequest()

	// Initialize server IDs first
	go s.initializeServerIDs()
}

func (s *SyncService) Stop() {
	close(s.stopChan)
}

// GetAPIKeyStatus returns the current status of all API keys for debugging
func (s *SyncService) GetAPIKeyStatus() []string {
	if s.lruKeyManager == nil {
		return []string{"LRU API Key Manager not initialized"}
	}
	return s.lruKeyManager.GetAllKeysStatus()
}

// ManualSync triggers a manual sync of all request types
func (s *SyncService) ManualSync() error {
	log.Println("Manual sync triggered")

	// Trigger a complete cycle of all requests
	s.cycleMux.Lock()
	requestsToProcess := make([]SyncRequest, 0, len(s.requestCycle))
	for _, reqFunc := range s.requestCycle {
		requestsToProcess = append(requestsToProcess, reqFunc())
	}
	s.cycleMux.Unlock()

	// Process all requests
	for _, req := range requestsToProcess {
		s.processRequest(req)
	}

	log.Printf("Manual sync completed: processed %d requests", len(requestsToProcess))
	return nil
}

func (s *SyncService) processNextRequest() {
	ticker := time.NewTicker(time.Duration(60/s.lruKeyManager.capacity) * time.Second) // 1 request per minute per API key
	statusTicker := time.NewTicker(5 * time.Minute)                                    // Log status every 5 minutes
	defer ticker.Stop()
	defer statusTicker.Stop()

	requestCount := 0

	for {
		select {
		case <-s.stopChan:
			return
		case <-statusTicker.C:
			// Log API key status periodically
			keyStatus := s.GetAPIKeyStatus()
			log.Printf("API Key LRU Status after %d requests:", requestCount)
			for _, status := range keyStatus {
				log.Printf("  %s", status)
			}
		case <-ticker.C:
			s.cycleMux.Lock()
			if len(s.requestCycle) == 0 {
				s.cycleMux.Unlock()
				continue // No requests in cycle yet, skip
			}

			// Get the next request in the cycle
			req := s.requestCycle[s.cycleIndex]()

			// Move to next request in cycle
			s.cycleIndex = (s.cycleIndex + 1) % len(s.requestCycle)
			s.cycleMux.Unlock()

			// Process the request immediately
			s.processRequest(req)
			requestCount++
		}
	}
}

func (s *SyncService) initializeRequestCycle() {
	s.cycleMux.Lock()
	defer s.cycleMux.Unlock()

	s.requestCycle = []func() SyncRequest{
		// Servers endpoint with items_per_page in body
		func() SyncRequest {
			body := []byte(`{"items_per_page": "1000"}`)
			return SyncRequest{
				URL:        s.managementURL + "/servers",
				Type:       "servers",
				Method:     "GET",
				Body:       body,
				MaxRetries: 3,
			}
		},
		func() SyncRequest {
			return SyncRequest{
				URL:        s.managementURL + "/operating-systems",
				Type:       "os-types",
				Method:     "GET",
				MaxRetries: 3,
			}
		},
		func() SyncRequest {
			return SyncRequest{
				URL:        s.managementURL + "/apps",
				Type:       "apps",
				Method:     "GET",
				MaxRetries: 3,
			}
		},
	}
}

func (s *SyncService) initializeServerIDs() {
	// Make initial request to get server IDs
	body := []byte(`{"items_per_page": "1000"}`)
	req := SyncRequest{
		URL:        s.managementURL + "/servers",
		Type:       "servers",
		Method:     "GET",
		Body:       body,
		MaxRetries: 3,
	}
	s.processRequest(req)

	// Rebuild the cycle with server-specific requests
	s.rebuildRequestCycle()
}

func (s *SyncService) rebuildRequestCycle() {
	s.serverIDsMux.RLock()
	serverIDs := make([]int, len(s.serverIDs))
	copy(serverIDs, s.serverIDs)
	s.serverIDsMux.RUnlock()

	s.cycleMux.Lock()
	defer s.cycleMux.Unlock()

	// Reset cycle
	s.requestCycle = []func() SyncRequest{
		// Servers endpoint with items_per_page in body
		func() SyncRequest {
			body := []byte(`{"items_per_page": "1000"}`)
			return SyncRequest{
				URL:        s.managementURL + "/servers",
				Type:       "servers",
				Method:     "GET",
				Body:       body,
				MaxRetries: 3,
			}
		},
		func() SyncRequest {
			return SyncRequest{
				URL:        s.managementURL + "/operating-systems",
				Type:       "os-types",
				Method:     "GET",
				MaxRetries: 3,
			}
		},
		// func() SyncRequest {
		// 	return SyncRequest{
		// 		URL:        s.managementURL + "/apps",
		// 		Type:       "apps",
		// 		Method:     "GET",
		// 		MaxRetries: 3,
		// 	}
		// },
	}

	// Add server-specific requests
	for _, serverID := range serverIDs {
		// Capture serverID in closure
		id := serverID

		s.requestCycle = append(s.requestCycle,
			func() SyncRequest {
				return SyncRequest{
					URL:        s.managementURL + "/servers/" + strconv.Itoa(id),
					Type:       "server-detail",
					Method:     "GET",
					ServerID:   id,
					MaxRetries: 3,
				}
			},
			// func() SyncRequest {
			// 	return SyncRequest{
			// 		URL:        s.managementURL + "/servers/" + strconv.Itoa(id) + "/snapshots",
			// 		Type:       "snapshots",
			// 		Method:     "GET",
			// 		ServerID:   id,
			// 		MaxRetries: 3,
			// 	}
			// },
			func() SyncRequest {
				return SyncRequest{
					URL:        s.managementURL + "/servers/" + strconv.Itoa(id) + "/get-root-password",
					Type:       "root-password",
					Method:     "GET",
					ServerID:   id,
					MaxRetries: 3,
				}
			},
			func() SyncRequest {
				return SyncRequest{
					URL:        s.managementURL + "/servers/" + strconv.Itoa(id) + "/get-console-credentials",
					Type:       "console-credentials",
					Method:     "GET",
					ServerID:   id,
					MaxRetries: 3,
				}
			},
		)
	}

	// Reset cycle index to start from beginning
	s.cycleIndex = 0

	log.Printf("Request cycle rebuilt with %d total requests (base: 3, server-specific: %d)",
		len(s.requestCycle), len(serverIDs)*4)
}

func (s *SyncService) processRequest(req SyncRequest) {
	log.Printf("Processing: %s (%s)", req.URL, req.Type)

	var httpReq *http.Request
	var err error

	if len(req.Body) > 0 {
		httpReq, err = http.NewRequest(req.Method, req.URL, bytes.NewBuffer(req.Body))
	} else {
		httpReq, err = http.NewRequest(req.Method, req.URL, nil)
	}

	if err != nil {
		log.Printf("Error creating request for %s: %v", req.URL, err)
		s.retryRequest(req)
		return
	}

	// Get the least recently used API key
	apiKey := s.lruKeyManager.GetLeastRecentlyUsedKey()
	if apiKey == "" {
		log.Printf("No API key available for request %s", req.URL)
		s.retryRequest(req)
		return
	}

	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		log.Printf("Error making request to %s: %v", req.URL, err)
		s.retryRequest(req)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Don't retry on 401 (Unauthorized) or 404 (Not Found) - these are likely permanent failures
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusNotFound {
			log.Printf("API returned permanent error %s for %s - skipping retries", resp.Status, req.URL)
			return
		}

		log.Printf("API returned status %s for %s", resp.Status, req.URL)
		s.retryRequest(req)
		return
	}

	switch req.Type {
	case "servers":
		s.processServersResponse(resp)
	case "os-types":
		s.processOSTypesResponse(resp)
	case "apps":
		s.processAppsResponse(resp)
	case "server-detail":
		s.processServerDetailResponse(resp, req.ServerID)
	case "snapshots":
		s.processSnapshotsResponse(resp, req.ServerID)
	case "root-password":
		s.processRootPasswordResponse(resp, req.ServerID)
	case "console-credentials":
		s.processConsoleCredentialsResponse(resp, req.ServerID)
	default:
		log.Printf("Unknown request type: %s", req.Type)
	}
}

func (s *SyncService) retryRequest(req SyncRequest) {
	if req.RetryCount < req.MaxRetries {
		req.RetryCount++
		log.Printf("Retrying request %s (attempt %d/%d) - will retry on next cycle", req.URL, req.RetryCount, req.MaxRetries)

		// Instead of immediate retry, we'll add it back to the cycle at the current position
		// This ensures it gets processed in the next normal cycle, maintaining 1-second spacing
		s.cycleMux.Lock()
		if len(s.requestCycle) > 0 {
			// Insert the retry request at the current cycle position
			// Create a closure that captures the current request
			retryFunc := func() SyncRequest {
				return req
			}

			// Insert at current position so it gets processed next
			s.requestCycle = append(s.requestCycle[:s.cycleIndex],
				append([]func() SyncRequest{retryFunc}, s.requestCycle[s.cycleIndex:]...)...)
		}
		s.cycleMux.Unlock()
	} else {
		log.Printf("Max retries exceeded for %s", req.URL)
	}
}

func (s *SyncService) processServersResponse(resp *http.Response) {
	var servers []api.Server
	if err := json.NewDecoder(resp.Body).Decode(&servers); err != nil {
		log.Printf("Error parsing servers response: %v", err)
		return
	}

	ctx := context.Background()

	// Update servers cache
	if _, err := s.db.ExecContext(ctx, "TRUNCATE TABLE servers_cache"); err != nil {
		log.Printf("Error clearing servers_cache: %v", err)
		return
	}

	serverIDs := make([]int, 0, len(servers))
	for _, server := range servers {
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO servers_cache (id, domain, reg_date, billing_cycle, next_due_date, domain_status, last_updated) 
			VALUES (?, ?, ?, ?, ?, ?, now())
		`, server.ID, server.Domain, server.RegDate, server.BillingCycle, server.NextDueDate, server.DomainStatus)

		if err != nil {
			log.Printf("Error inserting server %d: %v", server.ID, err)
		} else {
			if server.DomainStatus == "Active" && server.Domain != "" {
				serverIDs = append(serverIDs, server.ID)
			}

		}
	}

	// Update the known server IDs
	s.serverIDsMux.Lock()
	oldCount := len(s.serverIDs)
	s.serverIDs = serverIDs
	s.serverIDsMux.Unlock()

	// Rebuild request cycle if server list changed
	if len(serverIDs) != oldCount {
		log.Printf("Server count changed from %d to %d, rebuilding request cycle", oldCount, len(serverIDs))
		s.rebuildRequestCycle()
	}

	log.Printf("Updated servers cache with %d servers", len(servers))
}

func (s *SyncService) processOSTypesResponse(resp *http.Response) {
	var osTypes []api.OsType
	if err := json.NewDecoder(resp.Body).Decode(&osTypes); err != nil {
		log.Printf("Error parsing OS types response: %v", err)
		return
	}

	ctx := context.Background()
	if _, err := s.db.ExecContext(ctx, "TRUNCATE TABLE os_types_cache"); err != nil {
		log.Printf("Error clearing os_types_cache: %v", err)
		return
	}

	for _, osType := range osTypes {
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO os_types_cache (id, name, last_updated) 
			VALUES (?, ?, now())
		`, osType.ID, osType.Name)

		if err != nil {
			log.Printf("Error inserting OS type %d: %v", osType.ID, err)
		}
	}

	log.Printf("Updated OS types cache with %d entries", len(osTypes))
}

func (s *SyncService) processAppsResponse(resp *http.Response) {
	var apps []api.AppType
	if err := json.NewDecoder(resp.Body).Decode(&apps); err != nil {
		log.Printf("Error parsing apps response: %v", err)
		return
	}

	ctx := context.Background()
	if _, err := s.db.ExecContext(ctx, "TRUNCATE TABLE apps_cache"); err != nil {
		log.Printf("Error clearing apps_cache: %v", err)
		return
	}

	for _, app := range apps {
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO apps_cache (id, app, name, last_updated) 
			VALUES (?, ?, ?, now())
		`, app.ID, app.App, app.Name)

		if err != nil {
			log.Printf("Error inserting app %d: %v", app.ID, err)
		}
	}

	log.Printf("Updated apps cache with %d entries", len(apps))
}

func (s *SyncService) processServerDetailResponse(resp *http.Response, serverID int) {
	var serverDetail api.ServerDetail
	if err := json.NewDecoder(resp.Body).Decode(&serverDetail); err != nil {
		log.Printf("Error parsing server detail response for server %d: %v", serverID, err)
		return
	}

	ctx := context.Background()

	// Delete existing entry for this server
	_, err := s.db.ExecContext(ctx, "DELETE FROM server_details_cache WHERE server_id = ?", serverID)
	if err != nil {
		log.Printf("Error deleting existing server detail for %d: %v", serverID, err)
		return
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO server_details_cache (server_id, name, state, ip_address, operating_system, memory, disk, cpu, vnc_status, daily_snapshots, last_updated) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now())
	`, serverID, serverDetail.Name, serverDetail.State, serverDetail.IpAddress, serverDetail.OperatingSystem,
		serverDetail.Memory, serverDetail.Disk, serverDetail.Cpu, serverDetail.VncStatus, serverDetail.DailySnapshots)

	if err != nil {
		log.Printf("Error inserting server detail for %d: %v", serverID, err)
	} else {
		log.Printf("Updated server detail for server %d", serverID)
	}
}

func (s *SyncService) processSnapshotsResponse(resp *http.Response, serverID int) {
	var snapshots []api.ServerSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snapshots); err != nil {
		log.Printf("Error parsing snapshots response for server %d: %v", serverID, err)
		return
	}

	ctx := context.Background()

	// Delete existing snapshots for this server
	_, err := s.db.ExecContext(ctx, "DELETE FROM server_snapshots_cache WHERE server_id = ?", serverID)
	if err != nil {
		log.Printf("Error deleting existing snapshots for server %d: %v", serverID, err)
		return
	}

	for _, snapshot := range snapshots {
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO server_snapshots_cache (server_id, snapshot_id, name, created_at, size_gb, status, last_updated) 
			VALUES (?, ?, ?, ?, ?, ?, now())
		`, serverID, snapshot.ID, snapshot.Name, snapshot.CreatedAt, snapshot.SizeGB, snapshot.Status)

		if err != nil {
			log.Printf("Error inserting snapshot %s for server %d: %v", snapshot.ID, serverID, err)
		}
	}

	log.Printf("Updated %d snapshots for server %d", len(snapshots), serverID)
}

func (s *SyncService) processRootPasswordResponse(resp *http.Response, serverID int) {
	var rootPass api.RootPasswordResponse
	if err := json.NewDecoder(resp.Body).Decode(&rootPass); err != nil {
		log.Printf("Error parsing root password response for server %d: %v", serverID, err)
		return
	}

	ctx := context.Background()

	// Delete existing entry and insert new one (ClickHouse approach)
	_, err := s.db.ExecContext(ctx, "DELETE FROM server_credentials_cache WHERE server_id = ?", serverID)
	if err != nil {
		log.Printf("Error deleting existing credentials for server %d: %v", serverID, err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO server_credentials_cache (server_id, root_password, vnc_host, vnc_port, vnc_password, last_updated) 
		VALUES (?, ?, '', 0, '', now())
	`, serverID, rootPass.Password)

	if err != nil {
		log.Printf("Error inserting root password for server %d: %v", serverID, err)
	} else {
		log.Printf("Updated root password for server %d", serverID)
	}
}

// Grab the credentials from the StatusMessage:  "msg": "Your VM console credentials are: vnc_port: 9063, vnc_pass: Nc#mq6Rv."
func (s *SyncService) parseVncConsoleCredentials(msg string, serverID int) api.VncConsoleCredentials {
	var creds api.VncConsoleCredentials
	creds.Host = "23.227.199.130"
	parts := strings.Split(msg, ": vnc_port: ")
	for _, part := range parts {
		kv := strings.SplitN(part, ", vnc_pass: ", 2)
		if len(kv) != 2 {
			continue
		}
		var err error
		portInt64, err := strconv.ParseInt(kv[0], 10, 64)
		if err != nil {
			creds.Port = 0
			log.Printf("Error parsing vnc_port for server %d: %v", serverID, err)
			continue
		}
		creds.Port = int(portInt64)
		creds.Password = kv[1] // Remove the last period if present
		creds.Password = strings.TrimSuffix(creds.Password, ".")
		break
	}

	if creds.Port == 0 || creds.Password == "" {
		// Set the VNC to be enabled
		log.Printf("Incomplete VNC credentials parsed for server %d: port=%d, password=%s, reenabling console server", serverID, creds.Port, creds.Password)
		req, err := http.NewRequest("POST", fmt.Sprintf("/servers/%d/console/enable", serverID), nil)
		if err != nil {
			log.Printf("Error creating console enable request for server %d: %v", serverID, err)
			return creds
		}

		apiKey := s.lruKeyManager.GetLeastRecentlyUsedKey()
		if apiKey == "" {
			log.Printf("No API key available to re-enable console for server %d", serverID)
			return creds
		}

		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.client.Do(req)
		if err != nil {
			log.Printf("Error making console enable request for server %d: %v", serverID, err)
			return creds
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("API returned status %s when re-enabling console for server %d", resp.Status, serverID)
			return creds
		}

		// If we reach here, the console has been successfully re-enabled
		log.Printf("Successfully re-enabled console for server %d", serverID)
	}

	return creds
}

func (s *SyncService) processConsoleCredentialsResponse(resp *http.Response, serverID int) {
	var creds api.StatusMessage
	if err := json.NewDecoder(resp.Body).Decode(&creds); err != nil {
		log.Printf("Error parsing console credentials response for server %d: %v", serverID, err)
		return
	}

	// Grab the credentials from the StatusMessage:  "msg": "Your VM console credentials are: vnc_port: 9063, vnc_pass: Nc#mq6Rv."
	// The host will always be 23.227.199.130
	credsData := s.parseVncConsoleCredentials(creds.Message, serverID)

	ctx := context.Background()

	// Check if entry exists and update appropriately
	var existingPassword string
	err := s.db.QueryRowContext(ctx, "SELECT root_password FROM server_credentials_cache WHERE server_id = ?", serverID).Scan(&existingPassword)

	if err == sql.ErrNoRows {
		// Insert new entry
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO server_credentials_cache (server_id, root_password, vnc_host, vnc_port, vnc_password, last_updated) 
			VALUES (?, '', ?, ?, ?, now())
		`, serverID, credsData.Host, credsData.Port, credsData.Password)
	} else if err == nil {
		// Update existing entry
		_, err = s.db.ExecContext(ctx, `
			DELETE FROM server_credentials_cache WHERE server_id = ?
		`, serverID)
		if err == nil {
			_, err = s.db.ExecContext(ctx, `
				INSERT INTO server_credentials_cache (server_id, root_password, vnc_host, vnc_port, vnc_password, last_updated) 
				VALUES (?, ?, ?, ?, ?, now())
			`, serverID, existingPassword, credsData.Host, credsData.Port, credsData.Password)
		}
	}

	if err != nil {
		log.Printf("Error updating console credentials for server %d: %v", serverID, err)
	} else {
		log.Printf("Updated console credentials for server %d", serverID)
	}
}
