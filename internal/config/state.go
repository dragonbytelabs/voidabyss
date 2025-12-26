package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// State manages persistent key-value storage
type State struct {
	mu   sync.RWMutex
	data map[string]interface{}
	path string
}

// NewState creates a new state manager
func NewState() *State {
	return &State{
		data: make(map[string]interface{}),
		path: GetStatePath(),
	}
}

// GetStatePath returns the path to state.json
func GetStatePath() string {
	configDir := GetConfigDir()
	stateDir := filepath.Join(filepath.Dir(configDir), "state", "voidabyss")
	return filepath.Join(stateDir, "state.json")
}

// Load loads state from disk with corruption recovery
func (s *State) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stateDir := filepath.Dir(s.path)
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return err
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.data = make(map[string]interface{})
			return nil
		}
		return err
	}

	// Try to unmarshal the data
	if err := json.Unmarshal(data, &s.data); err != nil {
		// Corruption detected - backup and reset
		backupPath := s.path + ".corrupt"
		if backupErr := os.WriteFile(backupPath, data, 0644); backupErr == nil {
			// Successfully backed up, now reset
			s.data = make(map[string]interface{})
			// Save empty state
			s.mu.Unlock()
			saveErr := s.Save()
			s.mu.Lock()
			return saveErr
		}
		// Couldn't backup, return original error
		return err
	}

	return nil
}

// Save saves state to disk atomically
func (s *State) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stateDir := filepath.Dir(s.path)
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}

	// Atomic write: write to temp file then rename
	tempFile := s.path + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return err
	}

	// Rename is atomic on POSIX systems
	if err := os.Rename(tempFile, s.path); err != nil {
		os.Remove(tempFile) // Clean up temp file on failure
		return err
	}

	return nil
}

// Get retrieves a value from state
func (s *State) Get(key string, defaultValue interface{}) interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if val, ok := s.data[key]; ok {
		return val
	}
	return defaultValue
}

// Set stores a value in state
func (s *State) Set(key string, value interface{}) {
	s.mu.Lock()
	s.data[key] = value
	s.mu.Unlock()

	// Save synchronously to ensure persistence
	s.Save()
}

// Delete removes a key from state
func (s *State) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
	go s.Save()
}

// Keys returns all keys in state
func (s *State) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

// GetPath returns the state file path
func (s *State) GetPath() string {
	return s.path
}

// TestWrite tests if the state file is writable
func (s *State) TestWrite() error {
	stateDir := filepath.Dir(s.path)
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return err
	}

	// Try to create/write a test file
	testFile := filepath.Join(stateDir, ".test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return err
	}
	os.Remove(testFile)

	return nil
}
