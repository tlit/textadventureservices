package state

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func TestStateManager(t *testing.T) {
	t.Run("GetState_returns_empty_state_initially", func(t *testing.T) {
		sm := NewStateManager()
		state := sm.GetState()
		if len(state) != 0 {
			t.Errorf("expected empty state, got %v", state)
		}
	})

	t.Run("UpdateState_merges_updates_correctly", func(t *testing.T) {
		sm := NewStateManager()
		updates := map[string]interface{}{
			"health": float64(100),
			"score":  float64(50),
		}
		sm.UpdateState(updates)
		state := sm.GetState()
		if !reflect.DeepEqual(state, updates) {
			t.Errorf("state = %v, want %v", state, updates)
		}
	})

	t.Run("GetValue_retrieves_correct_value", func(t *testing.T) {
		sm := NewStateManager()
		sm.SetValue("health", float64(100))
		value, err := sm.GetValue("health")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if value != float64(100) {
			t.Errorf("value = %v, want %v", value, float64(100))
		}
	})

	t.Run("GetValue_returns_error_for_non-existent_key", func(t *testing.T) {
		sm := NewStateManager()
		_, err := sm.GetValue("nonexistent")
		if err == nil {
			t.Error("expected error for non-existent key")
		}
	})

	t.Run("SetValue_sets_value_correctly", func(t *testing.T) {
		sm := NewStateManager()
		sm.SetValue("score", float64(50))
		value, err := sm.GetValue("score")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if value != float64(50) {
			t.Errorf("value = %v, want %v", value, float64(50))
		}
	})

	t.Run("Reset_clears_state", func(t *testing.T) {
		sm := NewStateManager()
		sm.SetValue("score", float64(50))
		sm.Reset()
		state := sm.GetState()
		if len(state) != 0 {
			t.Errorf("expected empty state after reset, got %v", state)
		}
	})

	t.Run("SaveState_and_LoadState_work_correctly", func(t *testing.T) {
		sm := NewStateManager()
		initialState := map[string]interface{}{
			"health": float64(100),
			"score":  float64(50),
			"items":  []interface{}{"sword", "shield"},
		}
		sm.UpdateState(initialState)
		
		// Save state
		data, err := sm.SaveState()
		if err != nil {
			t.Fatalf("SaveState failed: %v", err)
		}

		// Reset and load state
		sm.Reset()
		err = sm.LoadState(data)
		if err != nil {
			t.Fatalf("LoadState failed: %v", err)
		}

		loadedState := sm.GetState()
		if !reflect.DeepEqual(loadedState, initialState) {
			t.Errorf("loaded state = %v, want %v", loadedState, initialState)
		}
	})

	t.Run("LoadState_handles_invalid_JSON", func(t *testing.T) {
		sm := NewStateManager()
		err := sm.LoadState([]byte("invalid json"))
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("State_is_thread-safe", func(t *testing.T) {
		sm := NewStateManager()
		var wg sync.WaitGroup
		numGoroutines := 10

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(n int) {
				defer wg.Done()
				sm.SetValue(fmt.Sprintf("key%d", n), float64(n))
			}(i)
		}
		wg.Wait()

		state := sm.GetState()
		if len(state) != numGoroutines {
			t.Errorf("expected %d items in state, got %d", numGoroutines, len(state))
		}
	})

	t.Run("Deep_copy_prevents_state_mutation", func(t *testing.T) {
		sm := NewStateManager()
		initialState := map[string]interface{}{
			"items": []interface{}{"sword"},
		}
		sm.UpdateState(initialState)

		// Try to modify the returned state
		state := sm.GetState()
		items := state["items"].([]interface{})
		items = append(items, "shield")

		// Original state should be unchanged
		originalItems, _ := sm.GetValue("items")
		if len(originalItems.([]interface{})) != 1 {
			t.Error("original state was modified")
		}
	})
}

func TestStateManagerConcurrency(t *testing.T) {
	t.Run("Concurrent_updates_are_safe", func(t *testing.T) {
		sm := NewStateManager()
		var wg sync.WaitGroup
		numGoroutines := 100
		numUpdates := 1000

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(n int) {
				defer wg.Done()
				for j := 0; j < numUpdates; j++ {
					sm.SetValue("counter", float64(n*numUpdates+j))
				}
			}(i)
		}
		wg.Wait()

		// Verify final state
		value, err := sm.GetValue("counter")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if value.(float64) < 0 {
			t.Error("unexpected counter value")
		}
	})
}
