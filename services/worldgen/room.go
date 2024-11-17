package worldgen

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"textadventureservices/services/worldgen/ai"
)

var aiProvider ai.Provider
var roomCounter int64

func init() {
	// Read OpenAI API key from .env.openai file
	file, err := os.Open(".env.openai")
	if err != nil {
		fmt.Printf("Warning: Failed to open .env.openai: %v\n", err)
		return
	}
	defer file.Close()

	var apiKey string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "OPENAI_API_KEY=") {
			apiKey = strings.TrimPrefix(line, "OPENAI_API_KEY=")
			break
		}
	}

	if apiKey == "" {
		fmt.Println("Warning: OPENAI_API_KEY not found in .env.openai")
		return
	}

	// Initialize OpenAI provider with configuration
	providerCfg := ai.ProviderConfig{
		Type:     ai.ProviderOpenAI,
		Model:    "gpt-4o",
		APIKey:   apiKey,
		Endpoint: "https://api.openai.com/v1",
	}

	provider, err := ai.NewProvider(providerCfg)
	if err != nil {
		fmt.Printf("Warning: Failed to create AI provider: %v\n", err)
		return
	}

	if err := provider.Initialize(providerCfg); err != nil {
		fmt.Printf("Warning: Failed to initialize AI provider: %v\n", err)
		return
	}

	aiProvider = provider
}

// GenerateRoom creates a new room with the given description and exits
func GenerateRoom(description string, exits []string) (*Room, error) {
	if description == "" {
		return nil, fmt.Errorf("room description cannot be empty")
	}

	// Enhance the description using AI if available
	enhancedDesc := description
	if aiProvider != nil {
		desc, err := aiProvider.GenerateDescription(context.Background(), description)
		if err != nil {
			fmt.Printf("Warning: Failed to enhance description: %v\n", err)
		} else {
			enhancedDesc = desc
		}
	}

	// Create a map of exits, but don't fill in the target IDs yet
	exitMap := make(map[string]string)
	for _, exit := range exits {
		exitMap[exit] = "" // Leave target ID empty, to be filled in later
	}

	roomCounter++
	room := &Room{
		ID:          fmt.Sprintf("room_%d_%d", time.Now().UnixNano(), roomCounter),
		Description: enhancedDesc,
		Objects:     make([]Object, 0),
		Exits:       exitMap,
		Properties:  make(map[string]interface{}),
	}

	return room, nil
}

// AddObject adds an object to the room
func (r *Room) AddObject(obj Object) {
	r.Objects = append(r.Objects, obj)
}

// AddExit adds an exit to the room
func (r *Room) AddExit(direction, targetID string) {
	if r.Exits == nil {
		r.Exits = make(map[string]string)
	}
	r.Exits[direction] = targetID
}

// GetExit returns the target room ID for a given direction
func (r *Room) GetExit(direction string) (string, bool) {
	if r.Exits == nil {
		return "", false
	}
	id, ok := r.Exits[direction]
	if !ok || id == "" {
		return "", false
	}
	return id, true
}

// SetProperty sets a property value for the room
func (r *Room) SetProperty(key string, value interface{}) {
	if r.Properties == nil {
		r.Properties = make(map[string]interface{})
	}
	r.Properties[key] = value
}

// GetProperty gets a property value from the room
func (r *Room) GetProperty(key string) (interface{}, bool) {
	if r.Properties == nil {
		return nil, false
	}
	value, ok := r.Properties[key]
	return value, ok
}
