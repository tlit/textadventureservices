package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"textadventureservices/services/worldgen/logging"
)

const openAIEndpoint = "https://api.openai.com/v1"

// NewService creates a new AI interaction service
func NewService(cfg *Config, logger logging.Logger) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	client := &http.Client{
		Timeout: time.Duration(cfg.TimeoutSeconds) * time.Second,
	}

	rateLimiter := &RateLimiter{
		limit:     cfg.RateLimit,
		lastReset: time.Now(),
	}

	return &Service{
		config:      cfg,
		client:      client,
		logger:      logger,
		rateLimiter: rateLimiter,
	}, nil
}

// ProcessInput processes user input and returns a response
func (s *Service) ProcessInput(ctx context.Context, req ProcessInputRequest) (*ProcessInputResponse, error) {
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	response, err := s.callOpenAI(ctx, req.Input)
	if err != nil {
		return nil, fmt.Errorf("OpenAI request failed: %w", err)
	}

	return &ProcessInputResponse{
		UpdatedState:  req.GameState,
		ActionSummary: response,
		ModelInfo: ModelInfo{
			Model:       s.config.Model,
			Temperature: s.config.Temperature,
		},
	}, nil
}

// GenerateDescription generates a description using the AI model
func (s *Service) GenerateDescription(ctx context.Context, prompt string) (string, error) {
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limit exceeded: %w", err)
	}

	response, err := s.callOpenAI(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("OpenAI request failed: %w", err)
	}

	return response, nil
}

// callOpenAI makes a request to the OpenAI API
func (s *Service) callOpenAI(ctx context.Context, prompt string) (string, error) {
	reqBody := OpenAIRequest{
		Model: s.config.Model,
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: s.config.Temperature,
		MaxTokens:   s.config.MaxTokens,
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", openAIEndpoint+"/chat/completions", bytes.NewReader(reqData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices from OpenAI")
	}

	return openAIResp.Choices[0].Message.Content, nil
}
