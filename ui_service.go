package main

import (
	"encoding/json"
	"net/http"
)

// UIState represents the current state of the UI
type UIState struct {
	State string `json:"state"`
}

// Command represents a command sent from the UI
type Command struct {
	Command string `json:"command"`
}

// registerUIHandlers registers the UI service handlers
func registerUIHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/ui/state", getUIState)
	mux.HandleFunc("/api/v1/ui/command", postUICommand)
}

// getUIState handles the retrieval of the current UI state
func getUIState(w http.ResponseWriter, r *http.Request) {
	uiState := UIState{State: "active"} // Example state
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(uiState)
}

// postUICommand handles commands sent from the UI
func postUICommand(w http.ResponseWriter, r *http.Request) {
	var cmd Command
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Process the command (this is a placeholder)
	response := map[string]string{"message": "Command processed successfully"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
