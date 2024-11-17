package ollama

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestOllamaClient(t *testing.T) {
	t.Run("NewOllamaClient uses default endpoint", func(t *testing.T) {
		client := NewOllamaClient()
		if client.endpoint != "http://localhost:11434" {
			t.Errorf("expected default endpoint, got %s", client.endpoint)
		}
	})

	t.Run("Generate sends correct request", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request method and path
			if r.Method != "POST" {
				t.Errorf("expected POST request, got %s", r.Method)
			}
			if r.URL.Path != "/api/generate" {
				t.Errorf("expected /api/generate path, got %s", r.URL.Path)
			}

			// Verify request body
			var req GenerateRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}

			// Send response
			response := GenerateResponse{
				Response: "Test response",
				Context: map[string]interface{}{
					"test": "context",
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		// Create client with test server endpoint
		client := &OllamaClient{
			endpoint: server.URL,
			client:   &http.Client{},
		}

		// Test request
		req := &GenerateRequest{
			Model:  "llama2",
			Prompt: "test prompt",
			Context: map[string]interface{}{
				"test": "data",
			},
		}

		resp, err := client.Generate(req)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		if resp.Response != "Test response" {
			t.Errorf("expected 'Test response', got %s", resp.Response)
		}
		if resp.Context["test"] != "context" {
			t.Errorf("expected context['test'] = 'context', got %v", resp.Context["test"])
		}
	})

	t.Run("ProcessGameCommand formats prompt correctly", func(t *testing.T) {
		var capturedPrompt string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req GenerateRequest
			json.NewDecoder(r.Body).Decode(&req)
			capturedPrompt = req.Prompt

			response := GenerateResponse{
				Response: "Command processed",
				Context: map[string]interface{}{
					"updated": true,
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := &OllamaClient{
			endpoint: server.URL,
			client:   &http.Client{},
		}

		gameState := map[string]interface{}{
			"location": "start_room",
			"health":   100,
		}

		response, updatedState, err := client.ProcessGameCommand("go north", gameState)
		if err != nil {
			t.Fatalf("ProcessGameCommand failed: %v", err)
		}

		// Verify prompt format
		expectedPromptPrefix := "Process this game command in the context of a text adventure:"
		if capturedPrompt[:len(expectedPromptPrefix)] != expectedPromptPrefix {
			t.Errorf("prompt doesn't start with expected prefix")
		}

		if response != "Command processed" {
			t.Errorf("expected 'Command processed', got %s", response)
		}

		if !reflect.DeepEqual(updatedState, map[string]interface{}{"updated": true}) {
			t.Errorf("unexpected state update: %v", updatedState)
		}
	})

	t.Run("Generate handles server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := &OllamaClient{
			endpoint: server.URL,
			client:   &http.Client{},
		}

		_, err := client.Generate(&GenerateRequest{
			Model:  "llama2",
			Prompt: "test",
		})

		if err == nil {
			t.Error("expected error for server error response")
		}
	})

	t.Run("Generate handles invalid response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		client := &OllamaClient{
			endpoint: server.URL,
			client:   &http.Client{},
		}

		_, err := client.Generate(&GenerateRequest{
			Model:  "llama2",
			Prompt: "test",
		})

		if err == nil {
			t.Error("expected error for invalid response")
		}
	})
}

func TestOllamaClientIntegration(t *testing.T) {
	// Skip by default unless OLLAMA_TEST_INTEGRATION=true
	if os.Getenv("OLLAMA_TEST_INTEGRATION") != "true" {
		t.Skip("skipping integration test; set OLLAMA_TEST_INTEGRATION=true to run")
	}

	client := NewOllamaClient()
	gameState := map[string]interface{}{
		"location": "start_room",
		"inventory": []string{"torch", "key"},
		"health": 100,
	}

	response, updatedState, err := client.ProcessGameCommand("go north", gameState)
	if err != nil {
		t.Fatalf("ProcessGameCommand failed: %v", err)
	}

	if response == "" {
		t.Error("expected non-empty response")
	}

	if updatedState == nil {
		t.Error("expected non-nil updated state")
	}
}
