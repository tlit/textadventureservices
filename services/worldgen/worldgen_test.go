package worldgen

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"textadventureservices/services/worldgen/ai"
	"textadventureservices/services/worldgen/config"
)

func TestWorldGeneration_GPT4o(t *testing.T) {
	// Set required environment variables for test
	os.Setenv("OPENAI_ENDPOINT", "https://api.openai.com/v1")

	// Ensure we can load the OpenAI configuration
	cfg, err := config.LoadProviderFromEnv(config.ProviderOpenAI)
	if err != nil {
		t.Fatalf("Failed to load OpenAI config: %v", err)
	}

	// Validate OpenAI configuration
	if !strings.HasPrefix(cfg.Model, "gpt-4o") {
		t.Fatalf("Expected model to start with gpt-4o, got %s", cfg.Model)
	}

	// Create a new world generation service
	service, err := NewService(&config.Config{
		AIProvider: cfg,
		Server: config.ServerConfig{Port: 8080},
	})
	if err != nil {
		t.Fatalf("Failed to create world generation service: %v", err)
	}

	// Test cases for different world themes
	testCases := []struct {
		name        string
		prompt      string
		expectWords int // Minimum number of words expected in description
	}{
		{
			name:        "Fantasy Castle",
			prompt:      "A mysterious castle with magical elements",
			expectWords: 20,
		},
		{
			name:        "Sci-Fi Station",
			prompt:      "An abandoned space station with alien technology",
			expectWords: 20,
		},
		{
			name:        "Horror Mansion",
			prompt:      "A Victorian mansion with a dark secret",
			expectWords: 20,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate a world based on the prompt
			world, err := service.GenerateWorld(ctx, tc.prompt)
			if err != nil {
				t.Fatalf("Failed to generate world: %v", err)
			}

			// Basic validation
			if world == nil {
				t.Fatal("Generated world is nil")
			}

			// Check if we have at least one room
			if len(world.Rooms) == 0 {
				t.Fatal("No rooms generated")
			}

			// Get the starting room (first room)
			var startRoom *Room
			for _, room := range world.Rooms {
				startRoom = room
				break
			}

			// Validate room description
			if startRoom.Description == "" {
				t.Error("Room description is empty")
			}

			// Count words in description (simple split-based count)
			words := len(splitWords(startRoom.Description))
			if words < tc.expectWords {
				t.Errorf("Description too short. Got %d words, expected at least %d. Description: %s",
					words, tc.expectWords, startRoom.Description)
			}

			// Optional: Save the generated world to a test file
			testOutputDir := "test_output"
			os.MkdirAll(testOutputDir, 0755)
			outputFile := filepath.Join(testOutputDir, 
				"world_"+sanitizeFilename(tc.name)+".json")
			
			if err := world.Save(outputFile); err != nil {
				t.Errorf("Failed to save world to file: %v", err)
			} else {
				t.Logf("Saved generated world to: %s", outputFile)
			}

			// Test loading the saved world
			loadedWorld, err := LoadWorld(outputFile)
			if err != nil {
				t.Errorf("Failed to load saved world: %v", err)
			}

			// Verify loaded world matches original
			if len(loadedWorld.Rooms) != len(world.Rooms) {
				t.Errorf("Loaded world has different number of rooms. Got %d, expected %d",
					len(loadedWorld.Rooms), len(world.Rooms))
			}
		})
	}
}

func TestGenerateDescription(t *testing.T) {
	// Create a test configuration
	cfg := &ai.ProviderConfig{
		Type:   ai.ProviderOpenAI,
		Model:  "gpt-4o",
		APIKey: "test-key",
		Parameters: map[string]interface{}{
			"temperature": 0.7,
			"max_tokens": 100,
			"rate_limit": 10,
			"timeout":    30,
		},
	}

	// Create a new world generation service
	service, err := NewService(cfg)
	if err != nil {
		t.Fatalf("Failed to create world generation service: %v", err)
	}

	// Test generating a world
	world, err := service.GenerateWorld(context.Background(), "A mysterious castle")
	if err != nil {
		t.Fatalf("Failed to generate world: %v", err)
	}

	// Verify world properties
	if world == nil {
		t.Fatal("Generated world is nil")
	}

	if len(world.Rooms) == 0 {
		t.Error("No rooms generated")
	}

	// Check each room
	for _, room := range world.Rooms {
		// Description should not be empty
		if room.Description == "" {
			t.Error("Room description is empty")
		}

		// Objects should be initialized
		if room.Objects == nil {
			t.Error("Room objects not initialized")
		}

		// Exits should be initialized
		if room.Exits == nil {
			t.Error("Room exits not initialized")
		}

		// Properties should be initialized
		if room.Properties == nil {
			t.Error("Room properties not initialized")
		}

		// Check that each exit leads to a valid room
		for _, targetID := range room.Exits {
			if targetID != "" {
				if _, ok := world.Rooms[targetID]; !ok {
					t.Errorf("Exit leads to non-existent room: %s", targetID)
				}
			}
		}
	}
}

// Helper function to count words in a string
func splitWords(s string) []string {
	words := strings.Fields(s)
	return words
}

// Helper function to create safe filenames
func sanitizeFilename(name string) string {
	// Replace spaces and special characters with underscores
	safe := strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			return r
		}
		return '_'
	}, name)
	return strings.ToLower(safe)
}
