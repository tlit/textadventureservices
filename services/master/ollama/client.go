package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type OllamaClient struct {
	endpoint string
	client   *http.Client
}

type GenerateRequest struct {
	Model    string                 `json:"model"`
	Prompt   string                 `json:"prompt"`
	Context  map[string]interface{} `json:"context,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

type GenerateResponse struct {
	Response string                 `json:"response"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

func NewOllamaClient() *OllamaClient {
	endpoint := os.Getenv("OLLAMA_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	return &OllamaClient{
		endpoint: endpoint,
		client:   &http.Client{},
	}
}

func (c *OllamaClient) Generate(req *GenerateRequest) (*GenerateResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(c.endpoint+"/api/generate", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

func (c *OllamaClient) ProcessGameCommand(input string, gameState map[string]interface{}) (string, map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Process this game command in the context of a text adventure:
Input: %s
Current Game State: %v
Respond with actions to take and any state updates.`, input, gameState)

	req := &GenerateRequest{
		Model:   "llama2",
		Prompt:  prompt,
		Context: gameState,
	}

	resp, err := c.Generate(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to process command: %w", err)
	}

	// Parse the response to extract actions and state updates
	// This is a simplified version - in practice, you'd want more structured parsing
	return resp.Response, resp.Context, nil
}
