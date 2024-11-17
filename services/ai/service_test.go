package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type mockLogger struct{}

func (m *mockLogger) Debug(ctx context.Context, msg string) {}
func (m *mockLogger) Info(ctx context.Context, msg string)  {}
func (m *mockLogger) Warn(ctx context.Context, msg string)  {}
func (m *mockLogger) Error(ctx context.Context, msg string) {}

func TestNewService(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = "test-key"
	
	service, err := NewService(cfg, &mockLogger{})
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	
	if service.config.DefaultModel != "gpt-4o" {
		t.Errorf("Expected default model gpt-4o, got %s", service.config.DefaultModel)
	}
}

func TestProcessInput(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("Missing or invalid Authorization header")
		}

		response := OpenAIResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4o",
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
							"updatedState": {"location": "forest"},
							"actionSummary": "Moved to forest"
						}`,
					},
					FinishReason: "stop",
				},
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create service with mock server
	cfg := DefaultConfig()
	cfg.APIKey = "test-key"
	cfg.Endpoint = server.URL

	service, err := NewService(cfg, &mockLogger{})
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Test processing input
	req := ProcessInputRequest{
		Input:     "go to forest",
		GameState: map[string]interface{}{"location": "town"},
	}

	resp, err := service.ProcessInput(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to process input: %v", err)
	}

	// Verify response
	if resp.ModelInfo.Model != "gpt-4o" {
		t.Errorf("Expected model gpt-4o, got %s", resp.ModelInfo.Model)
	}

	if resp.ModelInfo.TokensUsed != 30 {
		t.Errorf("Expected 30 tokens used, got %d", resp.ModelInfo.TokensUsed)
	}

	state, ok := resp.UpdatedState.(map[string]interface{})
	if !ok {
		t.Fatal("Failed to parse updated state")
	}

	if state["location"] != "forest" {
		t.Errorf("Expected location forest, got %v", state["location"])
	}

	if resp.ActionSummary != "Moved to forest" {
		t.Errorf("Expected 'Moved to forest', got %s", resp.ActionSummary)
	}
}

func TestRateLimiting(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = "test-key"
	cfg.RateLimits.RequestsPerMinute = 2

	service, err := NewService(cfg, &mockLogger{})
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Make requests up to the limit
	for i := 0; i < cfg.RateLimits.RequestsPerMinute; i++ {
		err := service.rateLimiter.CheckLimit()
		if err != nil {
			t.Errorf("Request %d should not exceed limit: %v", i+1, err)
		}
	}

	// Next request should fail
	err = service.rateLimiter.CheckLimit()
	if err == nil {
		t.Error("Expected rate limit error, got nil")
	}
}

// TestFallbackModel tests the fallback model behavior when primary model fails
func TestFallbackModel(t *testing.T) {
	// Create a mock server that fails for primary model but succeeds for fallback
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req OpenAIRequest
		json.NewDecoder(r.Body).Decode(&req)
		requestCount++

		if req.Model == "gpt-4o" {
			// Primary model fails
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			return
		}

		// Fallback model succeeds
		response := OpenAIResponse{
			ID:      "test-id",
			Model:   "gpt-3.5-turbo",
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
							"updatedState": {"location": "forest"},
							"actionSummary": "Moved to forest"
						}`,
					},
					FinishReason: "stop",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create service with mock server
	cfg := DefaultConfig()
	cfg.APIKey = "test-key"
	cfg.Endpoint = server.URL

	service, err := NewService(cfg, &mockLogger{})
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Test processing input
	req := ProcessInputRequest{
		Input:     "go to forest",
		GameState: map[string]interface{}{"location": "town"},
	}

	resp, err := service.ProcessInput(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to process input: %v", err)
	}

	// Verify fallback model was used
	if resp.ModelInfo.Model != "gpt-3.5-turbo" {
		t.Errorf("Expected fallback model gpt-3.5-turbo, got %s", resp.ModelInfo.Model)
	}

	// Verify two requests were made (primary + fallback)
	if requestCount != 2 {
		t.Errorf("Expected 2 requests, got %d", requestCount)
	}
}

// TestErrorRecovery tests various error scenarios and recovery mechanisms
func TestErrorRecovery(t *testing.T) {
	tests := []struct {
		name          string
		serverBehavior func(w http.ResponseWriter, r *http.Request)
		wantErr       bool
		errorContains string
	}{
		{
			name: "timeout_recovery",
			serverBehavior: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(2 * time.Second)
			},
			wantErr:       true,
			errorContains: "deadline exceeded",
		},
		{
			name: "malformed_response",
			serverBehavior: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"invalid": json`))
			},
			wantErr:       true,
			errorContains: "invalid character",
		},
		{
			name: "rate_limit_error",
			serverBehavior: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			},
			wantErr:       true,
			errorContains: "rate limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverBehavior))
			defer server.Close()

			cfg := DefaultConfig()
			cfg.APIKey = "test-key"
			cfg.Endpoint = server.URL

			service, err := NewService(cfg, &mockLogger{})
			if err != nil {
				t.Fatalf("Failed to create service: %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			req := ProcessInputRequest{
				Input:     "test command",
				GameState: map[string]interface{}{},
			}

			_, err = service.ProcessInput(ctx, req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && !strings.Contains(err.Error(), tt.errorContains) {
				t.Errorf("Expected error containing %q, got %v", tt.errorContains, err)
			}
		})
	}
}

// TestConcurrentRequests tests handling of multiple concurrent requests
func TestConcurrentRequests(t *testing.T) {
	// Create a mock server that tracks concurrent requests
	var (
		activeRequests int32
		maxRequests    int32
		totalRequests  int32
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&activeRequests, 1)
		atomic.AddInt32(&totalRequests, 1)
		current := atomic.LoadInt32(&activeRequests)
		for {
			max := atomic.LoadInt32(&maxRequests)
			if current <= max {
				break
			}
			if atomic.CompareAndSwapInt32(&maxRequests, max, current) {
				break
			}
		}

		// Simulate processing time
		time.Sleep(100 * time.Millisecond)

		response := OpenAIResponse{
			ID:      "test-id",
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
							"updatedState": {"location": "forest"},
							"actionSummary": "Moved to forest"
						}`,
					},
					FinishReason: "stop",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
		atomic.AddInt32(&activeRequests, -1)
	}))
	defer server.Close()

	// Create service with mock server
	cfg := DefaultConfig()
	cfg.APIKey = "test-key"
	cfg.Endpoint = server.URL
	cfg.RateLimits.RequestsPerMinute = 100 // Allow high concurrency

	service, err := NewService(cfg, &mockLogger{})
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Launch concurrent requests
	const numRequests = 10
	var wg sync.WaitGroup
	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := ProcessInputRequest{
				Input:     "test command",
				GameState: map[string]interface{}{},
			}
			_, err := service.ProcessInput(context.Background(), req)
			if err != nil {
				errors <- err
			}
		}()
	}

	// Wait for all requests to complete
	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent request error: %v", err)
	}

	// Verify concurrent execution
	if maxRequests <= 1 {
		t.Error("Requests were not executed concurrently")
	}
	if totalRequests != numRequests {
		t.Errorf("Expected %d total requests, got %d", numRequests, totalRequests)
	}
}

// TestTokenUsageTracking tests the token usage tracking functionality
func TestTokenUsageTracking(t *testing.T) {
	// Create a mock server that returns varying token usage
	var totalTokens int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokens := 100 // Each request uses 100 tokens
		totalTokens += tokens

		response := OpenAIResponse{
			ID:      "test-id",
			Model:   "gpt-4o",
			Created: time.Now().Unix(),
			Usage: Usage{
				PromptTokens:     50,
				CompletionTokens: 50,
				TotalTokens:      tokens,
			},
			Choices: []Choice{
				{
					Message: Message{
						Role: "assistant",
						Content: `{
							"updatedState": {"location": "forest"},
							"actionSummary": "Moved to forest"
						}`,
					},
					FinishReason: "stop",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create service with mock server and token limit
	cfg := DefaultConfig()
	cfg.APIKey = "test-key"
	cfg.Endpoint = server.URL
	cfg.RateLimits.TokensPerMinute = 250 // Allow only 2 requests worth of tokens

	service, err := NewService(cfg, &mockLogger{})
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Make requests until we hit the token limit
	for i := 0; i < 3; i++ {
		req := ProcessInputRequest{
			Input:     "test command",
			GameState: map[string]interface{}{},
		}

		resp, err := service.ProcessInput(context.Background(), req)
		if i < 2 {
			// First two requests should succeed
			if err != nil {
				t.Errorf("Request %d failed unexpectedly: %v", i+1, err)
			}
			if resp.ModelInfo.TokensUsed != 100 {
				t.Errorf("Expected 100 tokens used, got %d", resp.ModelInfo.TokensUsed)
			}
		} else {
			// Third request should fail due to token limit
			if err == nil {
				t.Error("Expected token limit error, got nil")
			}
			if !strings.Contains(err.Error(), "token limit exceeded") {
				t.Errorf("Expected token limit error, got: %v", err)
			}
		}
	}

	// Verify total token usage
	if totalTokens != 200 { // Only 2 successful requests
		t.Errorf("Expected 200 total tokens used, got %d", totalTokens)
	}
}
