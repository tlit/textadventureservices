package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"textadventureservices/services/worldgen/ai"
)

// LoadProviderFromEnv loads AI provider configuration from environment files
func LoadProviderFromEnv(providerType ai.ProviderType) (ai.ProviderConfig, error) {
	// Only support OpenAI configuration
	if providerType != ai.ProviderOpenAI {
		return ai.ProviderConfig{}, fmt.Errorf("unsupported provider type: %s", providerType)
	}

	envFile := ".env.openai"

	// Find the environment file by walking up directories
	currentDir, err := os.Getwd()
	if err != nil {
		return ai.ProviderConfig{}, fmt.Errorf("failed to get current directory: %w", err)
	}

	envPath := findEnvFile(currentDir, envFile)
	if envPath == "" {
		return ai.ProviderConfig{}, fmt.Errorf("environment file %s not found", envFile)
	}

	// Load environment variables from file
	if err := loadEnvFile(envPath); err != nil {
		return ai.ProviderConfig{}, fmt.Errorf("failed to load environment file: %w", err)
	}

	return loadOpenAIConfig(), nil
}

// findEnvFile walks up the directory tree looking for the environment file
func findEnvFile(startDir, filename string) string {
	dir := startDir
	for {
		path := filepath.Join(dir, filename)
		if _, err := os.Stat(path); err == nil {
			return path
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			return "" // Reached root directory
		}
		dir = parent
	}
}

// loadEnvFile loads environment variables from a file
func loadEnvFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove comments from value
		if idx := strings.Index(value, "#"); idx != -1 {
			value = strings.TrimSpace(value[:idx])
		}

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		os.Setenv(key, value)
	}

	return nil
}

func loadOpenAIConfig() ai.ProviderConfig {
	// Load optional parameters
	params := make(map[string]interface{})
	
	if temp, err := strconv.ParseFloat(os.Getenv("OPENAI_TEMPERATURE"), 64); err == nil {
		params["temperature"] = temp
	}
	if tokens, err := strconv.ParseFloat(os.Getenv("OPENAI_MAX_TOKENS"), 64); err == nil {
		params["max_tokens"] = tokens
	}
	if topP, err := strconv.ParseFloat(os.Getenv("OPENAI_TOP_P"), 64); err == nil {
		params["top_p"] = topP
	}
	if freqPen, err := strconv.ParseFloat(os.Getenv("OPENAI_FREQUENCY_PENALTY"), 64); err == nil {
		params["frequency_penalty"] = freqPen
	}
	if presPen, err := strconv.ParseFloat(os.Getenv("OPENAI_PRESENCE_PENALTY"), 64); err == nil {
		params["presence_penalty"] = presPen
	}
	
	return ai.ProviderConfig{
		Type:       ai.ProviderOpenAI,
		Endpoint:   os.Getenv("OPENAI_ENDPOINT"),
		Model:      os.Getenv("OPENAI_MODEL"),
		APIKey:     os.Getenv("OPENAI_API_KEY"),
		Parameters: params,
	}
}
