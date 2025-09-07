package sync

import (
	"testing"
	"time"

	"github.com/callowaysutton/servercon/api"
)

func TestLRUAPIKeyManager(t *testing.T) {
	// Test with sample API keys
	apiKeys := []string{
		"key1_12345678",
		"key2_87654321",
		"key3_abcdefgh",
	}

	// Create LRU manager
	manager := NewLRUAPIKeyManager(apiKeys)
	if manager == nil {
		t.Fatal("Failed to create LRU manager")
	}

	// Test that all keys are loaded
	status := manager.GetAllKeysStatus()
	if len(status) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(status))
	}

	// Test LRU selection - should cycle through keys
	usedKeys := make([]string, 6)
	for i := 0; i < 6; i++ {
		key := manager.GetLeastRecentlyUsedKey()
		usedKeys[i] = key
		if key == "" {
			t.Errorf("Got empty key at iteration %d", i)
		}
		// Small delay to ensure different timestamps
		time.Sleep(1 * time.Millisecond)
	}

	// Verify we got all 3 keys twice (since we made 6 requests)
	keyCount := make(map[string]int)
	for _, key := range usedKeys {
		keyCount[key]++
	}

	if len(keyCount) != 3 {
		t.Errorf("Expected 3 unique keys, got %d", len(keyCount))
	}

	for key, count := range keyCount {
		if count != 2 {
			t.Errorf("Key %s was used %d times, expected 2", key, count)
		}
	}
}

func TestLRUAPIKeyManagerEmpty(t *testing.T) {
	// Test with empty keys
	manager := NewLRUAPIKeyManager([]string{})
	if manager != nil {
		t.Error("Expected nil manager for empty keys")
	}

	// Test with nil keys
	manager = NewLRUAPIKeyManager(nil)
	if manager != nil {
		t.Error("Expected nil manager for nil keys")
	}
}

func TestLRUAPIKeyManagerFiltersEmpty(t *testing.T) {
	// Test that empty strings are filtered out
	apiKeys := []string{
		"key1_12345678",
		"", // Empty key should be filtered
		"key3_abcdefgh",
		"", // Another empty key
	}

	manager := NewLRUAPIKeyManager(apiKeys)
	if manager == nil {
		t.Fatal("Failed to create LRU manager")
	}

	status := manager.GetAllKeysStatus()
	if len(status) != 2 {
		t.Errorf("Expected 2 keys after filtering, got %d", len(status))
	}
}

// Grab the credentials from the StatusMessage:  "msg": "Your VM console credentials are: vnc_port: 9063, vnc_pass: Nc#mq6Rv."
func TestParseVncConsoleCredentials(t *testing.T) {
	tests := []struct {
		input    string
		expected api.VncConsoleCredentials
	}{
		{"Your VM console credentials are: vnc_port: 9063, vnc_pass: Nc#mq6Rv.", api.VncConsoleCredentials{Host: "23.227.199.130", Port: 9063, Password: "Nc#mq6Rv"}},
		{"Your VM console credentials are: vnc_port: 8365, vnc_pass: e5Raiu#3.", api.VncConsoleCredentials{Host: "23.227.199.130", Port: 8365, Password: "e5Raiu#3"}},
	}

	service := &SyncService{}

	for _, test := range tests {
		result := service.parseVncConsoleCredentials(test.input, 0)
		if result != test.expected {
			t.Errorf("For input '%s', expected '%v' but got '%v'", test.input, test.expected, result)
		}
	}
}
