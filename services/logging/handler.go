package logging

import (
	"encoding/json"
	"net/http"
	"time"
)

// noob: This handles all our HTTP stuff
type LoggingHandler struct {
	logger *QuantumLogger
}

// noob: This is what the client sends us
type LogRequest struct {
	Level    LogLevel                `json:"level"`
	Message  string                  `json:"message"`
	Service  string                  `json:"service"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// noob: This is what we send back
type LogResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// noob: This makes a new handler
func NewLoggingHandler(logger *QuantumLogger) *LoggingHandler {
	return &LoggingHandler{logger: logger}
}

// noob: This handles log submissions
func (h *LoggingHandler) HandleLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// noob: Decode the fancy JSON
	var req LogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// noob: Log it in our quantum system
	err := h.logger.Log(r.Context(), req.Level, req.Message)
	resp := LogResponse{Success: err == nil}
	if err != nil {
		resp.Error = err.Error()
	}

	// noob: Send back the result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// noob: This gets logs for a time range
func (h *LoggingHandler) HandleFetch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// noob: Parse the quantum time parameters
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		http.Error(w, "Invalid start time", http.StatusBadRequest)
		return
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		http.Error(w, "Invalid end time", http.StatusBadRequest)
		return
	}

	// noob: Get logs from our time-space continuum
	logs := h.logger.Fetch(r.Context(), start, end)

	// noob: Send them back through the quantum tunnel
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// noob: This sets up all our routes
func (h *LoggingHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/logs", h.HandleLog)
	mux.HandleFunc("/api/v1/logs/fetch", h.HandleFetch)
}
