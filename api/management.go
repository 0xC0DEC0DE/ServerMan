package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
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

	return lruNode.key
}

// Management API client
type ManagementAPIClient struct {
	baseURL       string
	client        *http.Client
	lruKeyManager *LRUAPIKeyManager
}

var (
	managementClient *ManagementAPIClient
	once             sync.Once
)

// GetManagementClient returns a singleton instance of the management API client
func GetManagementClient() *ManagementAPIClient {
	once.Do(func() {
		// Get API keys from environment
		var apiKeys []string
		for _, env := range os.Environ() {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 && strings.HasPrefix(parts[0], "SERVER_MANAGEMENT_KEY") {
				if parts[1] != "" {
					apiKeys = append(apiKeys, parts[1])
				}
			}
		}

		managementClient = &ManagementAPIClient{
			baseURL:       os.Getenv("SERVER_MANAGEMENT_URL"),
			client:        &http.Client{Timeout: 30 * time.Second},
			lruKeyManager: NewLRUAPIKeyManager(apiKeys),
		}
	})
	return managementClient
}

// makeRequest makes an HTTP request to the management API
func (m *ManagementAPIClient) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, m.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, err
	}

	// Get the least recently used API key
	apiKey := m.lruKeyManager.GetLeastRecentlyUsedKey()
	if apiKey == "" {
		return nil, fmt.Errorf("no API key available")
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	return m.client.Do(req)
}

// Server action request structure
type ServerActionRequest struct {
	Action string `json:"action"`
}

// Reinstall request structure
type ReinstallRequest struct {
	ReinstallType  string `json:"reinstall_type"`  // "os" or "app"
	OsAppId        int    `json:"os_app_id"`
	Authentication string `json:"authentication"`  // "ssh" or "password"
	SshKey         string `json:"ssh_key,omitempty"`
}

// Snapshot restore request structure
type RestoreSnapshotRequest struct {
	SnapshotName string `json:"snapshot_name"`
}

// Console action request structure
type ConsoleActionRequest struct {
	Action string `json:"action"` // "enable" or "disable"
}

// ServerAction handles server power actions (start, stop, restart)
func ServerAction(c *gin.Context) {
	serverID := c.Param("id")
	action := c.Param("action")

	// Validate action
	validActions := map[string]bool{
		"start":   true,
		"stop":    true,
		"restart": true,
	}

	if !validActions[action] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action. Must be start, stop, or restart"})
		return
	}

	client := GetManagementClient()
	endpoint := fmt.Sprintf("/servers/%s/action/%s", serverID, action)

	resp, err := client.makeRequest("POST", endpoint, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "Management API request failed"})
		return
	}

	var result []StatusMessage
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode response"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ReinstallServer handles server reinstallation
func ReinstallServer(c *gin.Context) {
	serverID := c.Param("id")

	var req ReinstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate reinstall type
	if req.ReinstallType != "os" && req.ReinstallType != "app" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reinstall_type must be 'os' or 'app'"})
		return
	}

	// Validate authentication type
	if req.Authentication != "ssh" && req.Authentication != "password" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "authentication must be 'ssh' or 'password'"})
		return
	}

	// If SSH authentication, validate SSH key is provided
	if req.Authentication == "ssh" && req.SshKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ssh_key is required when authentication is 'ssh'"})
		return
	}

	client := GetManagementClient()
	endpoint := fmt.Sprintf("/servers/%s/reinstall", serverID)

	resp, err := client.makeRequest("POST", endpoint, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "Management API request failed"})
		return
	}

	var result []StatusMessage
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode response"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RestoreSnapshot handles server snapshot restoration
func RestoreSnapshot(c *gin.Context) {
	serverID := c.Param("id")

	var req RestoreSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.SnapshotName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "snapshot_name is required"})
		return
	}

	client := GetManagementClient()
	endpoint := fmt.Sprintf("/servers/%s/restore-snapshot", serverID)

	resp, err := client.makeRequest("POST", endpoint, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "Management API request failed"})
		return
	}

	var result []StatusMessage
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode response"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ResetRootPassword handles resetting the server root password
func ResetRootPassword(c *gin.Context) {
	serverID := c.Param("id")

	client := GetManagementClient()
	endpoint := fmt.Sprintf("/servers/%s/reset-root-password", serverID)

	resp, err := client.makeRequest("POST", endpoint, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "Management API request failed"})
		return
	}

	var result []StatusMessage
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode response"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ChangeConsoleStatus handles enabling/disabling console
func ChangeConsoleStatus(c *gin.Context) {
	serverID := c.Param("id")
	action := c.Param("action")

	// Validate action
	if action != "enable" && action != "disable" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action. Must be enable or disable"})
		return
	}

	client := GetManagementClient()
	endpoint := fmt.Sprintf("/servers/%s/console/%s", serverID, action)

	resp, err := client.makeRequest("POST", endpoint, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "Management API request failed"})
		return
	}

	var result []StatusMessage
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode response"})
		return
	}

	c.JSON(http.StatusOK, result)
}
