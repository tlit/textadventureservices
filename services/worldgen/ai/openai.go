package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type OpenAIProvider struct {
	endpoint    string
	model       string
	apiKey      string
	client      *http.Client
	temperature float64
	maxTokens   int
}

type openAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{
		client:      &http.Client{},
		temperature: 0.8,  // Increased for more creativity
		maxTokens:   500,  // Increased for longer descriptions
	}
}

func (p *OpenAIProvider) Initialize(config ProviderConfig) error {
	if config.APIKey == "" {
		return fmt.Errorf("OpenAI API key is required")
	}
	if config.Endpoint == "" {
		return fmt.Errorf("OpenAI endpoint is required")
	}
	if config.Model == "" {
		return fmt.Errorf("OpenAI model is required")
	}

	p.endpoint = config.Endpoint
	p.model = config.Model
	p.apiKey = config.APIKey

	// Apply custom parameters if provided
	if config.Parameters != nil {
		if temp, ok := config.Parameters["temperature"].(float64); ok {
			p.temperature = temp
		}
		if tokens, ok := config.Parameters["max_tokens"].(float64); ok {
			p.maxTokens = int(tokens)
		}
	}

	return nil
}

func (p *OpenAIProvider) makeRequest(ctx context.Context, messages []Message) (string, error) {
	reqBody := openAIRequest{
		Model:       p.model,
		Messages:    messages,
		Temperature: p.temperature,
		MaxTokens:   p.maxTokens,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var result openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("OpenAI error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return result.Choices[0].Message.Content, nil
}

func (p *OpenAIProvider) GenerateDescription(ctx context.Context, prompt string) (string, error) {
	messages := []Message{
		{
			Role: "system",
			Content: `You are a creative writing AI specializing in text adventure game descriptions.
Your task is to create rich, immersive descriptions that engage all senses and highlight interactive elements.

Follow these guidelines:
1. Write detailed descriptions of 50-100 words
2. Include vivid sensory details (sights, sounds, smells, textures)
3. Highlight 2-3 interactive objects or features
4. Set a clear mood and atmosphere
5. Use evocative language that draws players in
6. Maintain consistency with the scene's theme

Format your response as a single, cohesive paragraph that flows naturally.`,
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	return p.makeRequest(ctx, messages)
}

func (p *OpenAIProvider) GenerateObjects(ctx context.Context, sceneDescription string) ([]string, error) {
	messages := []Message{
		{
			Role: "system",
			Content: "You are an AI that generates interactive objects for text adventure games. " +
				"List objects that would naturally be found in the scene and could be interesting for player interaction. " +
				"Provide 3-5 objects as a comma-separated list.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("List the interactive objects that would be found in this scene: %s", sceneDescription),
		},
	}

	response, err := p.makeRequest(ctx, messages)
	if err != nil {
		return nil, err
	}

	// Clean and split the response
	objects := strings.Split(response, ",")
	for i := range objects {
		objects[i] = strings.TrimSpace(objects[i])
	}

	return objects, nil
}

func (p *OpenAIProvider) EnhancePrompt(ctx context.Context, basePrompt string) (string, error) {
	messages := []Message{
		{
			Role: "system",
			Content: "You are an AI that enhances basic scene prompts for text adventure games. " +
				"Add atmospheric details, time period context, and mood elements while keeping the prompt concise.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Enhance this basic scene prompt with more detail: %s", basePrompt),
		},
	}

	return p.makeRequest(ctx, messages)
}
