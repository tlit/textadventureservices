package ai

import (
	"context"
	"fmt"
)

type ProviderType string

const (
	ProviderOpenAI ProviderType = "openai"
)

// Provider defines the interface for AI providers
type Provider interface {
	Initialize(config ProviderConfig) error
	EnhancePrompt(ctx context.Context, basePrompt string) (string, error)
	GenerateObjects(ctx context.Context, sceneDescription string) ([]string, error)
	GenerateDescription(ctx context.Context, prompt string) (string, error)
}

// ProviderConfig holds the configuration for an AI provider
type ProviderConfig struct {
	Type       ProviderType
	Endpoint   string
	Model      string
	APIKey     string
	Parameters map[string]interface{}
}

// NewProvider creates a new AI provider based on the configuration
func NewProvider(config ProviderConfig) (Provider, error) {
	switch config.Type {
	case ProviderOpenAI:
		return NewOpenAIProvider(), nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", config.Type)
	}
}
