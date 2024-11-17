package state

import (
	"encoding/json"
	"fmt"
	"sync"
)

// StateManager handles game state management
type StateManager struct {
	state    map[string]interface{}
	stateMux sync.RWMutex
}

// NewStateManager creates a new state manager
func NewStateManager() *StateManager {
	return &StateManager{
		state: make(map[string]interface{}),
	}
}

// GetState returns a copy of the current state
func (sm *StateManager) GetState() map[string]interface{} {
	sm.stateMux.RLock()
	defer sm.stateMux.RUnlock()

	// Create a deep copy of the state
	stateCopy := make(map[string]interface{})
	data, _ := json.Marshal(sm.state)
	json.Unmarshal(data, &stateCopy)

	return stateCopy
}

// UpdateState updates the game state
func (sm *StateManager) UpdateState(updates map[string]interface{}) error {
	sm.stateMux.Lock()
	defer sm.stateMux.Unlock()

	// Merge updates into current state
	for key, value := range updates {
		sm.state[key] = value
	}

	return nil
}

// GetValue retrieves a specific value from the state
func (sm *StateManager) GetValue(key string) (interface{}, error) {
	sm.stateMux.RLock()
	defer sm.stateMux.RUnlock()

	value, exists := sm.state[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	return value, nil
}

// SetValue sets a specific value in the state
func (sm *StateManager) SetValue(key string, value interface{}) {
	sm.stateMux.Lock()
	defer sm.stateMux.Unlock()

	sm.state[key] = value
}

// Reset clears the current state
func (sm *StateManager) Reset() {
	sm.stateMux.Lock()
	defer sm.stateMux.Unlock()

	sm.state = make(map[string]interface{})
}

// SaveState returns the current state as JSON
func (sm *StateManager) SaveState() ([]byte, error) {
	sm.stateMux.RLock()
	defer sm.stateMux.RUnlock()

	return json.Marshal(sm.state)
}

// LoadState loads a state from JSON
func (sm *StateManager) LoadState(data []byte) error {
	var newState map[string]interface{}
	if err := json.Unmarshal(data, &newState); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	sm.stateMux.Lock()
	defer sm.stateMux.Unlock()

	sm.state = newState
	return nil
}
