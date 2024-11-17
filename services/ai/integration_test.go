package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"textadventureservices/services/worldgen/logging"
)

// mockLoggingServer simulates the logging service
type mockLoggingServer struct {
	logs []struct {
		Level   string
		Message string
	}
}

func newMockLoggingServer() (*mockLoggingServer, *httptest.Server) {
	mock := &mockLoggingServer{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var logReq struct {
			Level   string `json:"level"`
			Message string `json:"message"`
		}
		json.NewDecoder(r.Body).Decode(&logReq)
		mock.logs = append(mock.logs, struct {
			Level   string
			Message string
		}{logReq.Level, logReq.Message})
		w.WriteHeader(http.StatusOK)
	}))
	return mock, server
}

// mockOpenAIServer simulates the OpenAI API
type mockOpenAIServer struct {
	responses []OpenAIResponse
	requests  []OpenAIRequest
}

func newMockOpenAIServer() (*mockOpenAIServer, *httptest.Server) {
	mock := &mockOpenAIServer{
		responses: []OpenAIResponse{
			{
				ID:      "test-1",
				Model:   "gpt-4o",
				Created: time.Now().Unix(),
				Usage: Usage{
					PromptTokens:     10,
					CompletionTokens: 20,
					TotalTokens:      30,
				},
				Choices: []Choice{
					{
						Message: Message{
							Role: "assistant",
							Content: `{
								"updatedState": {
									"currentRoom": "room_1",
									"rooms": {
										"room_1": {
											"description": "A dark forest clearing with ancient trees",
											"exits": {
												"north": "room_2"
											},
											"items": ["sword", "shield"]
										}
									}
								},
								"actionSummary": "You examine the clearing and notice a sword and shield lying against a tree"
							}`,
						},
						FinishReason: "stop",
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req OpenAIRequest
		json.NewDecoder(r.Body).Decode(&req)
		mock.requests = append(mock.requests, req)

		resp := mock.responses[len(mock.requests)-1]
		json.NewEncoder(w).Encode(resp)
	}))

	return mock, server
}

func TestIntegrationWithLogging(t *testing.T) {
	// Set up mock logging server
	mockLogger, logServer := newMockLoggingServer()
	defer logServer.Close()

	// Create logger that points to our mock server
	logger := logging.NewHTTPLogger(logServer.URL)

	// Create AI service
	cfg := DefaultConfig()
	cfg.APIKey = "test-key"
	cfg.Endpoint = "http://localhost:8080" // Will be overridden in the actual test

	service, err := NewService(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Set up mock OpenAI server
	mockOpenAI, openaiServer := newMockOpenAIServer()
	defer openaiServer.Close()
	service.config.Endpoint = openaiServer.URL

	// Test processing input
	req := ProcessInputRequest{
		Input: "look around",
		GameState: map[string]interface{}{
			"location": "start",
			"inventory": []string{},
		},
	}

	_, err = service.ProcessInput(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to process input: %v", err)
	}

	// Verify logs were sent
	if len(mockLogger.logs) == 0 {
		t.Error("No logs were sent")
	}

	// Verify OpenAI requests
	if len(mockOpenAI.requests) == 0 {
		t.Error("No requests were sent to OpenAI")
	}
}

func TestIntegrationWithWorldGen(t *testing.T) {
	// Set up mock OpenAI server
	mockOpenAI, openaiServer := newMockOpenAIServer()
	defer openaiServer.Close()

	// Create AI service
	cfg := DefaultConfig()
	cfg.APIKey = "test-key"
	cfg.Endpoint = openaiServer.URL

	service, err := NewService(cfg, logging.NewNoopLogger())
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Test world generation integration
	worldState := map[string]interface{}{
		"currentRoom": "room_1",
		"rooms": map[string]interface{}{
			"room_1": map[string]interface{}{
				"description": "A dark forest clearing",
				"exits": map[string]string{
					"north": "room_2",
				},
			},
		},
	}

	req := ProcessInputRequest{
		Input:     "examine the clearing",
		GameState: worldState,
	}

	resp, err := service.ProcessInput(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to process world state: %v", err)
	}

	// Verify the response updates the world state correctly
	newState, ok := resp.UpdatedState.(map[string]interface{})
	if !ok {
		t.Fatal("Failed to get updated state")
	}

	if _, ok := newState["currentRoom"]; !ok {
		t.Error("Updated state missing currentRoom")
	}

	// Verify OpenAI was called with correct prompt
	lastReq := mockOpenAI.requests[len(mockOpenAI.requests)-1]
	if len(lastReq.Messages) == 0 {
		t.Error("No messages in OpenAI request")
	}
}

// loadTestConfig loads a test configuration
func loadTestConfig() *Config {
	return &Config{
		DefaultModel:  "gpt-4o",
		FallbackModel: "gpt-3.5-turbo",
		Parameters: ModelParameters{
			Temperature: 0.8,
			MaxTokens:   200,
		},
		RateLimits: RateLimits{
			RequestsPerMinute: 60,
			TokensPerMinute:   90000,
		},
		APIKey:   "test-env-key",
		Endpoint: "https://api.openai.com/v1",
	}
}

func TestIntegrationWithConfig(t *testing.T) {
	// Load test config
	cfg := loadTestConfig()

	// Create service
	service, err := NewService(cfg, logging.NewNoopLogger())
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Verify config was loaded correctly
	if service.config.DefaultModel != "gpt-4o" {
		t.Errorf("Expected model gpt-4o, got %s", service.config.DefaultModel)
	}
	if service.config.Parameters.Temperature != 0.8 {
		t.Errorf("Expected temperature 0.8, got %f", service.config.Parameters.Temperature)
	}
	if service.config.Parameters.MaxTokens != 200 {
		t.Errorf("Expected max tokens 200, got %d", service.config.Parameters.MaxTokens)
	}
}
