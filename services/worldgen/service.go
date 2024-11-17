package worldgen

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
	
	"textadventureservices/services/ai"
	"textadventureservices/services/worldgen/config"
	"textadventureservices/services/worldgen/logging"
)

// Service represents the world generation service
type Service struct {
	aiProvider ai.Provider
	config     *config.Config
	logger     logging.Logger
}

// NewService creates a new world generation service
func NewService(cfg *config.Config) (*Service, error) {
	aiService, err := ai.NewService(&ai.Config{
		Model:          cfg.AIProvider.Model,
		MaxTokens:      cfg.AIProvider.MaxTokens,
		Temperature:    cfg.AIProvider.Temperature,
		RateLimit:      cfg.AIProvider.RateLimit,
		TimeoutSeconds: cfg.AIProvider.TimeoutSeconds,
	}, logging.NewNoopLogger())
	if err != nil {
		return nil, fmt.Errorf("failed to create AI service: %w", err)
	}

	// Create logger based on configuration
	var logger logging.Logger
	if cfg.LoggingEndpoint != "" {
		logger = logging.NewHTTPLogger(cfg.LoggingEndpoint)
	} else {
		logger = logging.NewNoopLogger()
	}
	
	return &Service{
		aiProvider: aiService,
		config:     cfg,
		logger:     logger,
	}, nil
}

// GenerateWorld generates a new world based on a prompt
func (s *Service) GenerateWorld(ctx context.Context, prompt string) (*World, error) {
	s.logger.Info(ctx, fmt.Sprintf("Starting world generation with prompt: %s", prompt))

	// Generate initial description
	description, err := s.aiProvider.GenerateDescription(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate world description: %w", err)
	}

	// Create world with initial room
	seed := time.Now().UnixNano()
	world, err := NewWorld(seed)
	if err != nil {
		return nil, fmt.Errorf("failed to create world: %w", err)
	}

	startRoom, err := GenerateRoom(description, []string{})
	if err != nil {
		return nil, fmt.Errorf("failed to create start room: %w", err)
	}
	world.AddRoom(startRoom)

	// Generate additional rooms based on config
	for i := 1; i < s.config.DefaultRooms; i++ {
		roomPrompt := fmt.Sprintf("Generate a connected room for: %s", description)
		roomDesc, err := s.aiProvider.GenerateDescription(ctx, roomPrompt)
		if err != nil {
			s.logger.Error(ctx, fmt.Sprintf("Failed to generate room %d: %v", i, err))
			continue
		}
		
		room, err := GenerateRoom(roomDesc, []string{})
		if err != nil {
			s.logger.Error(ctx, fmt.Sprintf("Failed to create room %d: %v", i, err))
			continue
		}
		world.AddRoom(room)
		world.ConnectRooms(startRoom.ID, room.ID, getRandomDirection())
	}

	s.logger.Info(ctx, "World generation completed successfully")
	return world, nil
}

// ProcessInput handles user input and returns a response
func (s *Service) ProcessInput(ctx context.Context, input string) (string, error) {
	s.logger.Info(ctx, fmt.Sprintf("Processing input: %s", input))

	// Extract command and object from input
	parts := strings.Fields(input)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid input format: expected 'command object'")
	}

	command := strings.ToLower(parts[0])
	object := strings.Join(parts[1:], " ")

	// Handle different commands
	switch command {
	case "examine":
		// Generate description for the object using AI
		prompt := fmt.Sprintf("Describe the '%s' in detail", object)
		description, err := s.aiProvider.GenerateDescription(ctx, prompt)
		if err != nil {
			s.logger.Error(ctx, fmt.Sprintf("Failed to generate object description: %v", err))
			return "", fmt.Errorf("failed to examine object: %w", err)
		}
		return description, nil

	// Add more commands here as needed
	default:
		return "", fmt.Errorf("unknown command: %s", command)
	}
}

func getRandomDirection() Direction {
	directions := []Direction{North, South, East, West}
	return directions[rand.Intn(len(directions))]
}
