package ai

import (
	"context"
	"net/http"
	"sync"
	"time"

	"textadventureservices/services/worldgen/logging"
)

// Service represents the AI interaction service
type Service struct {
	config      *Config
	client      *http.Client
	logger      logging.Logger
	rateLimiter *RateLimiter
	mu          sync.Mutex
}

// Provider defines the interface for AI services
type Provider interface {
	ProcessInput(ctx context.Context, req ProcessInputRequest) (*ProcessInputResponse, error)
	GenerateDescription(ctx context.Context, prompt string) (string, error)
}

// Config represents the configuration for the AI service
type Config struct {
	Model          string  `json:"model"`
	MaxTokens      int     `json:"max_tokens"`
	Temperature    float64 `json:"temperature"`
	RateLimit      int     `json:"rate_limit"`
	TimeoutSeconds int     `json:"timeout_seconds"`
	APIKey         string  `json:"api_key"`
}

// ProcessInputRequest represents the request for processing user input
type ProcessInputRequest struct {
	Input     string      `json:"input"`
	GameState interface{} `json:"gameState"`
}

// ProcessInputResponse represents the response from processing user input
type ProcessInputResponse struct {
	UpdatedState  interface{} `json:"updatedState"`
	ActionSummary string      `json:"actionSummary"`
	ModelInfo     ModelInfo   `json:"modelInfo"`
}

// ModelInfo contains information about the AI model used
type ModelInfo struct {
	Model       string  `json:"model"`
	TokensUsed  int     `json:"tokensUsed"`
	Temperature float64 `json:"temperature"`
}

// RateLimiter represents a rate limiter
type RateLimiter struct {
	limit     int
	lastReset time.Time
	tokens    int
	mu        sync.Mutex
}

// ChatMessage represents a message in the chat
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIRequest represents a request to the OpenAI API
type OpenAIRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int          `json:"max_tokens"`
}

// OpenAIResponse represents a response from the OpenAI API
type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}
