package config

import (
	"encoding/json"
	"fmt"
	"os"
	
	"textadventureservices/services/ai"
)

// Config represents the configuration for the world generation service
type Config struct {
	DefaultRooms    int       `json:"default_rooms"`
	LoggingEndpoint string    `json:"logging_endpoint"`
	AIProvider      ai.Config `json:"ai_provider"`
	Server          ServerConfig      `json:"server"`
}

// ServerConfig holds the HTTP server configuration
type ServerConfig struct {
	Port int `json:"port"`
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return &config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.DefaultRooms <= 0 {
		return fmt.Errorf("default_rooms must be positive")
	}
	if err := c.AIProvider.Validate(); err != nil {
		return fmt.Errorf("invalid AI provider config: %w", err)
	}
	if c.Server.Port <= 0 {
		return fmt.Errorf("invalid server port")
	}
	return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultRooms: 5,
		AIProvider: ai.Config{
			Model:          "gpt-3.5-turbo",
			MaxTokens:      150,
			Temperature:    0.7,
			RateLimit:      60,
			TimeoutSeconds: 30,
		},
		Server: ServerConfig{
			Port: 8080,
		},
		LoggingEndpoint: "http://localhost:8081", // Default logging service endpoint
	}
}
