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

// Load loads state from disk
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

	if err := json.Unmarshal(data, &s.data); err != nil {
		return err
	}

	return nil
}

// Save saves state to disk
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

	return os.WriteFile(s.path, data, 0644)
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
