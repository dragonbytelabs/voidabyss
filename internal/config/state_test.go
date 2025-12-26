package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStateAtomicWrite(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "state.json")

	state := &State{
		data: make(map[string]interface{}),
		path: statePath,
	}

	// Set some values
	state.Set("key1", "value1")
	state.Set("key2", 42)

	// Verify state file exists
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("State file was not created")
	}

	// Verify temp file doesn't exist (atomic write cleaned up)
	tempFile := statePath + ".tmp"
	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Error("Temp file was not cleaned up")
	}

	// Verify values
	if got := state.Get("key1", ""); got != "value1" {
		t.Errorf("Expected value1, got %v", got)
	}

	// JSON unmarshals numbers as float64, but we store directly as interface{}
	// so it could be int or float64 depending on whether it was loaded from JSON
	key2Val := state.Get("key2", 0)
	switch v := key2Val.(type) {
	case int:
		if v != 42 {
			t.Errorf("Expected 42, got %v", v)
		}
	case float64:
		if v != 42.0 {
			t.Errorf("Expected 42.0, got %v", v)
		}
	default:
		t.Errorf("Expected int or float64, got %T: %v", key2Val, key2Val)
	}
}

func TestStateCorruptionRecovery(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "state.json")

	// Write corrupted JSON
	corruptJSON := []byte(`{"key": "value", broken`)
	if err := os.WriteFile(statePath, corruptJSON, 0644); err != nil {
		t.Fatalf("Failed to write corrupt file: %v", err)
	}

	// Try to load
	state := &State{
		data: make(map[string]interface{}),
		path: statePath,
	}

	if err := state.Load(); err != nil {
		t.Errorf("Load should recover from corruption: %v", err)
	}

	// Verify backup was created
	backupPath := statePath + ".corrupt"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup file was not created")
	}

	// Verify state is empty after recovery
	if len(state.data) != 0 {
		t.Errorf("Expected empty state after corruption recovery, got %d keys", len(state.data))
	}

	// Verify new state file is valid JSON
	newData, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("Failed to read recovered state: %v", err)
	}

	if string(newData) == "" {
		t.Error("Recovered state file is empty")
	}
}

func TestStateLoadNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "nonexistent.json")

	state := &State{
		data: make(map[string]interface{}),
		path: statePath,
	}

	// Should not error on non-existent file
	if err := state.Load(); err != nil {
		t.Errorf("Load should not error on non-existent file: %v", err)
	}

	// Should have empty data
	if len(state.data) != 0 {
		t.Errorf("Expected empty state, got %d keys", len(state.data))
	}
}

func TestStateConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "state.json")

	state := &State{
		data: make(map[string]interface{}),
		path: statePath,
	}

	// Concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			state.Set("key", n)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have exactly one value (last write wins)
	val := state.Get("key", nil)
	if val == nil {
		t.Error("Expected a value, got nil")
	}
}
