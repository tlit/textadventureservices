package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"textadventureservices/services/ai"
	"textadventureservices/services/worldgen"
	"textadventureservices/services/worldgen/config"
	"textadventureservices/services/worldgen/logging"
)

func TestEndToEnd(t *testing.T) {
	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: OPENAI_API_KEY environment variable not set")
	}

	// Initialize logger
	logger := logging.NewQuantumLogger(
		logging.WithBufferSize(1000),
		logging.WithRetentionPeriod(24 * time.Hour),
	)
	defer logger.Shutdown()

	// Initialize world generation service
	worldgenConfig := config.DefaultConfig()
	worldgenConfig.DefaultRooms = 3
	worldgenConfig.AIProvider = ai.Config{
		Model:          "gpt-3.5-turbo",
		MaxTokens:      150,
		Temperature:    0.7,
		RateLimit:      60,
		TimeoutSeconds: 30,
		APIKey:         apiKey,
	}

	worldService, err := worldgen.NewService(worldgenConfig)
	if err != nil {
		t.Fatalf("Failed to create world service: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Test scenario: Generate and interact with a medieval fantasy world
	t.Run("Complete Game Flow", func(t *testing.T) {
		// 1. Generate world
		world, err := worldService.GenerateWorld(ctx, "A medieval fantasy town with a mysterious tower")
		if err != nil {
			t.Fatalf("Failed to generate world: %v", err)
		}
		if world == nil {
			t.Fatal("Generated world is nil")
		}
		if len(world.Scenes) == 0 {
			t.Fatal("No scenes generated in world")
		}

		// Log world creation
		logger.Info(ctx, "World generated successfully")

		// 2. Test navigation
		var startScene *worldgen.Scene
		for _, scene := range world.Scenes {
			startScene = scene
			break
		}
		if len(startScene.Exits) == 0 {
			t.Fatal("Start scene has no exits")
		}

		// Try moving to connected room
		nextRoomID := ""
		for _, exitID := range startScene.Exits {
			nextRoomID = exitID
			break
		}

		nextScene, err := world.GetScene(nextRoomID)
		if err != nil {
			t.Fatalf("Failed to get next scene: %v", err)
		}
		if nextScene == nil {
			t.Fatal("Next scene is nil")
		}

		// Log navigation
		logger.Info(ctx, "Successfully navigated between scenes")

		// 3. Test object interaction
		if len(startScene.Objects) > 0 {
			obj := startScene.Objects[0]
			// Simulate examining object
			response, err := worldService.ProcessInput(ctx, "examine "+obj.Name)
			if err != nil {
				t.Fatalf("Failed to process examine command: %v", err)
			}
			if response == "" {
				t.Fatal("Empty response from AI service")
			}
			logger.Info(ctx, "Successfully examined object: "+obj.Name)
		}

		// 4. Test world persistence
		err = world.Save("test_world_e2e.json")
		if err != nil {
			t.Fatalf("Failed to save world: %v", err)
		}

		loadedWorld, err := worldgen.LoadWorld("test_world_e2e.json")
		if err != nil {
			t.Fatalf("Failed to load world: %v", err)
		}
		if len(loadedWorld.Scenes) != len(world.Scenes) {
			t.Fatalf("Loaded world has different number of scenes. Expected %d, got %d",
				len(world.Scenes), len(loadedWorld.Scenes))
		}

		logger.Info(ctx, "Successfully tested world persistence")

		// 5. Verify logging
		time.Sleep(100 * time.Millisecond) // Wait for logs to be processed
		logs := logger.Fetch(ctx, time.Now().Add(-1*time.Hour), time.Now())
		if len(logs) == 0 {
			t.Fatal("No logs recorded during test")
		}

		// Test completion
		logger.Info(ctx, "End-to-end test completed successfully")

		// Cleanup
		os.Remove("test_world_e2e.json")
	})
}
