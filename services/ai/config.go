package ai

import (
	"fmt"
)

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		TimeoutSeconds: 30,
		RateLimit:      60,
		Model:          "gpt-3.5-turbo",
		Temperature:    0.7,
		MaxTokens:      150,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.TimeoutSeconds <= 0 {
		return fmt.Errorf("timeoutSeconds must be greater than 0")
	}
	if c.RateLimit <= 0 {
		return fmt.Errorf("rateLimit must be greater than 0")
	}
	if c.Model == "" {
		return fmt.Errorf("model must not be empty")
	}
	if c.APIKey == "" {
		return fmt.Errorf("apiKey must not be empty")
	}
	if c.MaxTokens <= 0 {
		return fmt.Errorf("maxTokens must be greater than 0")
	}
	if c.Temperature < 0 || c.Temperature > 1 {
		return fmt.Errorf("temperature must be between 0 and 1")
	}
	return nil
}
