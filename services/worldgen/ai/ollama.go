package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type OllamaProvider struct {
	endpoint string
	model    string
	client   *http.Client
}

type ollamaRequest struct {
	Model    string                 `json:"model"`
	Prompt   string                 `json:"prompt"`
	Stream   bool                   `json:"stream"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

func NewOllamaProvider() *OllamaProvider {
	return &OllamaProvider{
		client: &http.Client{},
	}
}

func (p *OllamaProvider) Initialize(config ProviderConfig) error {
	if config.Endpoint == "" {
		return fmt.Errorf("ollama endpoint is required")
	}
	if config.Model == "" {
		return fmt.Errorf("ollama model is required")
	}
	
	p.endpoint = config.Endpoint
	p.model = config.Model
	return nil
}

func (p *OllamaProvider) makeRequest(ctx context.Context, prompt string) (string, error) {
	reqBody := ollamaRequest{
		Model:  p.model,
		Prompt: prompt,
		Stream: false,
	}
	
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/api/generate", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	var result ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	
	if result.Error != "" {
		return "", fmt.Errorf("ollama error: %s", result.Error)
	}
	
	return result.Response, nil
}

func (p *OllamaProvider) GenerateDescription(ctx context.Context, prompt string) (string, error) {
	enhancedPrompt := fmt.Sprintf(
		"Generate a vivid and detailed description of a scene in a text adventure game. The scene should be based on this prompt: %s. "+
			"Focus on atmospheric details, sensory information, and notable features that a player might interact with. "+
			"Keep the description between 2-4 sentences.",
		prompt,
	)
	return p.makeRequest(ctx, enhancedPrompt)
}

func (p *OllamaProvider) GenerateObjects(ctx context.Context, sceneDescription string) ([]string, error) {
	prompt := fmt.Sprintf(
		"Based on this scene description: '%s'\n"+
			"List 3-5 interactive objects that would logically be found in this scene. "+
			"Format the response as a comma-separated list of simple object names.",
		sceneDescription,
	)
	
	response, err := p.makeRequest(ctx, prompt)
	if err != nil {
		return nil, err
	}
	
	// Split the comma-separated response into a slice
	objects := splitAndTrim(response)
	return objects, nil
}

func (p *OllamaProvider) EnhancePrompt(ctx context.Context, basePrompt string) (string, error) {
	enhancedPrompt := fmt.Sprintf(
		"Take this basic scene prompt: '%s'\n"+
			"Enhance it by adding more specific details about the atmosphere, time period, and mood. "+
			"Keep the enhanced prompt concise but vivid.",
		basePrompt,
	)
	return p.makeRequest(ctx, enhancedPrompt)
}

// Helper function to split and trim a comma-separated string
func splitAndTrim(input string) []string {
	// Implementation omitted for brevity - splits on commas and trims whitespace
	return []string{"TODO: Implement splitting logic"}
}
