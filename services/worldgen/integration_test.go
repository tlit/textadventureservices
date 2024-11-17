package worldgen

import (
	"context"
	"testing"
	"time"

	"textadventureservices/services/logging"
)

// WorldGenLogger wraps the world generation service with logging capabilities
type WorldGenLogger struct {
	*World
	logger *logging.QuantumLogger
}

func NewWorldGenLogger(seed int64, logger *logging.QuantumLogger) (*WorldGenLogger, error) {
	world, err := NewWorld(seed)
	if err != nil {
		return nil, err
	}

	return &WorldGenLogger{
		World:  world,
		logger: logger,
	}, nil
}

func TestWorldGenWithLogging(t *testing.T) {
	// Setup logger
	logger := logging.NewQuantumLogger(logging.WithBufferSize(100))
	defer logger.Shutdown()

	ctx := context.Background()
	
	// Create world with logging
	worldGen, err := NewWorldGenLogger(12345, logger)
	if err != nil {
		t.Fatalf("Failed to create world with logger: %v", err)
	}

	// Generate a test world
	err = worldGen.GenerateWorld("test world", 1)
	if err != nil {
		t.Fatalf("Failed to generate world: %v", err)
	}

	// Verify world state
	if len(worldGen.Rooms) != 1 {
		t.Errorf("Expected 1 initial room, got %d", len(worldGen.Rooms))
	}

	// Test room generation with logging
	t.Run("generate_room_with_logging", func(t *testing.T) {
		description := "A test room with logging"
		exits := []string{"north", "south"}

		// Log the attempt to create a room
		err := logger.Log(ctx, logging.LogLevelInfo, "Attempting to generate new room")
		if err != nil {
			t.Errorf("Failed to log room generation attempt: %v", err)
		}

		room, err := GenerateRoom(description, exits)
		if err != nil {
			t.Fatalf("Failed to generate room: %v", err)
		}

		// Log successful room creation
		err = logger.Log(ctx, logging.LogLevelInfo, "Successfully generated room")
		if err != nil {
			t.Errorf("Failed to log room generation success: %v", err)
		}

		// Verify room properties
		if room.Description != description {
			t.Errorf("Room description mismatch. Got %s, want %s", room.Description, description)
		}
		if len(room.Exits) != len(exits) {
			t.Errorf("Room exits count mismatch. Got %d, want %d", len(room.Exits), len(exits))
		}
	})

	// Test log retrieval
	t.Run("verify_logs", func(t *testing.T) {
		// Wait a bit for logs to be processed
		time.Sleep(100 * time.Millisecond)

		start := time.Now().Add(-1 * time.Hour)
		end := time.Now()
		logs := logger.Fetch(ctx, start, end)

		// We expect at least 2 logs: attempt and success
		if len(logs) < 2 {
			t.Errorf("Expected at least 2 logs, got %d", len(logs))
		}

		// Verify log contents
		foundAttempt := false
		foundSuccess := false
		for _, log := range logs {
			switch log.Message {
			case "Attempting to generate new room":
				foundAttempt = true
			case "Successfully generated room":
				foundSuccess = true
			}
		}

		if !foundAttempt {
			t.Error("Missing log entry for room generation attempt")
		}
		if !foundSuccess {
			t.Error("Missing log entry for room generation success")
		}
	})
}

func TestWorldPersistenceWithLogging(t *testing.T) {
	// Setup logger
	logger := logging.NewQuantumLogger(logging.WithBufferSize(100))
	defer logger.Shutdown()

	ctx := context.Background()

	// Create world with logging
	worldGen, err := NewWorldGenLogger(12345, logger)
	if err != nil {
		t.Fatalf("Failed to create world with logger: %v", err)
	}

	// Generate a test world
	err = worldGen.GenerateWorld("test world", 1)
	if err != nil {
		t.Fatalf("Failed to generate world: %v", err)
	}

	// Verify initial world state
	if len(worldGen.Rooms) != 1 {
		t.Errorf("Expected 1 initial room, got %d", len(worldGen.Rooms))
	}

	// Test saving world state
	t.Run("save_world_with_logging", func(t *testing.T) {
		err := logger.Log(ctx, logging.LogLevelInfo, "Attempting to save world state")
		if err != nil {
			t.Errorf("Failed to log save attempt: %v", err)
		}

		err = worldGen.Save("test_world_with_logging.json")
		if err != nil {
			t.Fatalf("Failed to save world: %v", err)
		}

		err = logger.Log(ctx, logging.LogLevelInfo, "Successfully saved world state")
		if err != nil {
			t.Errorf("Failed to log save success: %v", err)
		}
	})

	// Test loading world state
	t.Run("load_world_with_logging", func(t *testing.T) {
		err := logger.Log(ctx, logging.LogLevelInfo, "Attempting to load world state")
		if err != nil {
			t.Errorf("Failed to log load attempt: %v", err)
		}

		loadedWorld, err := LoadWorld("test_world_with_logging.json")
		if err != nil {
			t.Fatalf("Failed to load world: %v", err)
		}

		err = logger.Log(ctx, logging.LogLevelInfo, "Successfully loaded world state")
		if err != nil {
			t.Errorf("Failed to log load success: %v", err)
		}

		if len(loadedWorld.Rooms) != len(worldGen.Rooms) {
			t.Errorf("Loaded world has different number of rooms. Expected %d, got %d",
				len(worldGen.Rooms), len(loadedWorld.Rooms))
		}
	})

	// Verify all operation logs
	t.Run("verify_persistence_logs", func(t *testing.T) {
		// Wait a bit for logs to be processed
		time.Sleep(100 * time.Millisecond)

		start := time.Now().Add(-1 * time.Hour)
		end := time.Now()
		logs := logger.Fetch(ctx, start, end)

		expectedMessages := map[string]bool{
			"Attempting to save world state":    false,
			"Successfully saved world state":    false,
			"Attempting to load world state":    false,
			"Successfully loaded world state":   false,
		}

		for _, log := range logs {
			if _, exists := expectedMessages[log.Message]; exists {
				expectedMessages[log.Message] = true
			}
		}

		for msg, found := range expectedMessages {
			if !found {
				t.Errorf("Missing expected log message: %s", msg)
			}
		}
	})
}
