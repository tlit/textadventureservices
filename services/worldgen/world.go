package worldgen

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
)

// World represents the entire game world
type World struct {
	Seed   int64             `json:"seed"`
	Rooms  map[string]*Room  `json:"rooms"`
	Scenes map[string]*Scene `json:"scenes"` // Added for compatibility with e2e tests
	rng    *rand.Rand        `json:"-"`
}

// NewWorld creates a new world instance with the given seed
func NewWorld(seed int64) (*World, error) {
	world := &World{
		Seed:   seed,
		Rooms:  make(map[string]*Room),
		Scenes: make(map[string]*Scene), // Initialize Scenes map
		rng:    rand.New(rand.NewSource(seed)),
	}

	// No longer create initial room here as it's handled by GenerateWorld
	return world, nil
}

// AddRoom adds a room to the world and updates the scenes map
func (w *World) AddRoom(room *Room) {
	if w.Rooms == nil {
		w.Rooms = make(map[string]*Room)
	}
	w.Rooms[room.ID] = room
	// Convert room to scene and add to scenes map
	scene := &Scene{
		ID:          room.ID,
		Description: room.Description,
		Objects:     room.Objects,
		Exits:       room.Exits,
		Properties:  room.Properties,
	}
	w.Scenes[room.ID] = scene
}

// GetRoom returns a room by its ID
func (w *World) GetRoom(id string) (*Room, bool) {
	if w.Rooms == nil {
		return nil, false
	}
	room, ok := w.Rooms[id]
	return room, ok
}

// GetScene returns a scene by its ID
func (w *World) GetScene(id string) (*Scene, error) {
	if w.Scenes == nil {
		return nil, fmt.Errorf("scenes map is not initialized")
	}
	
	scene, exists := w.Scenes[id]
	if !exists {
		return nil, fmt.Errorf("scene with ID %s not found", id)
	}
	return scene, nil
}

// GenerateWorld creates a multi-room world based on a base prompt
func (w *World) GenerateWorld(basePrompt string, numRooms int) error {
	fmt.Printf("Generating world with prompt '%s' and %d rooms...\n", basePrompt, numRooms)
	
	if numRooms <= 0 {
		return fmt.Errorf("number of rooms must be positive")
	}

	if basePrompt == "" {
		return fmt.Errorf("base prompt cannot be empty")
	}

	// Clear any existing rooms
	w.Rooms = make(map[string]*Room)
	w.Scenes = make(map[string]*Scene)

	// Create the first room (central hub)
	fmt.Println("Generating central hub room...")
	centralRoom, err := w.generateRoom(fmt.Sprintf("%s - Central Hub", basePrompt), nil) // Start with no exits
	if err != nil {
		return fmt.Errorf("failed to generate central room: %w", err)
	}
	w.AddRoom(centralRoom)
	fmt.Printf("Created central room with ID: %s\n", centralRoom.ID)

	// Generate additional rooms and connect them
	directions := []Direction{North, South, East, West}
	remainingRooms := numRooms - 1 // Subtract 1 for the central room
	currentRooms := []*Room{centralRoom}

	fmt.Printf("Generating %d additional rooms...\n", remainingRooms)
	attemptLimit := numRooms * 10 // Prevent infinite loops
	attempts := 0

	for remainingRooms > 0 && attempts < attemptLimit {
		attempts++
		// Pick a random existing room to connect from
		sourceRoom := currentRooms[w.rng.Intn(len(currentRooms))]
		fmt.Printf("Selected source room: %s\n", sourceRoom.ID)
		
		// Pick a random available direction
		var availableDirections []Direction
		for _, dir := range directions {
			if targetID, exists := sourceRoom.Exits[string(dir)]; !exists || targetID == "" {
				availableDirections = append(availableDirections, dir)
			}
		}

		if len(availableDirections) == 0 {
			fmt.Printf("No available directions from room %s, trying another room...\n", sourceRoom.ID)
			continue // Try another room if this one is fully connected
		}

		// Choose a random direction and create a new room
		direction := availableDirections[w.rng.Intn(len(availableDirections))]
		oppositeDir := direction.GetOppositeDirection()
		fmt.Printf("Selected direction: %s (opposite: %s)\n", direction, oppositeDir)

		// Generate a themed room based on its direction from the center
		roomPrompt := generateDirectionalPrompt(basePrompt, string(direction))
		fmt.Printf("Generating room with prompt: %s\n", roomPrompt)
		
		newRoom, err := w.generateRoom(roomPrompt, []string{string(oppositeDir)})
		if err != nil {
			return fmt.Errorf("failed to generate room: %w", err)
		}
		fmt.Printf("Created new room with ID: %s\n", newRoom.ID)

		// Connect the rooms bidirectionally
		if err := w.ConnectRooms(sourceRoom.ID, newRoom.ID, direction); err != nil {
			return fmt.Errorf("failed to connect rooms: %w", err)
		}
		fmt.Printf("Connected rooms: %s -%s-> %s\n", sourceRoom.ID, direction, newRoom.ID)

		w.AddRoom(newRoom)
		currentRooms = append(currentRooms, newRoom)
		remainingRooms--
		fmt.Printf("Remaining rooms to generate: %d\n", remainingRooms)
	}

	if remainingRooms > 0 {
		return fmt.Errorf("failed to generate all rooms after %d attempts", attemptLimit)
	}

	fmt.Println("World generation complete!")
	return nil
}

// ConnectRooms connects two rooms bidirectionally
func (w *World) ConnectRooms(sourceID string, targetID string, direction Direction) error {
	sourceRoom, ok := w.GetRoom(sourceID)
	if !ok {
		return fmt.Errorf("source room %s not found", sourceID)
	}

	targetRoom, ok := w.GetRoom(targetID)
	if !ok {
		return fmt.Errorf("target room %s not found", targetID)
	}

	// Get opposite direction
	oppositeDir := direction.GetOppositeDirection()

	// Connect rooms bidirectionally
	sourceRoom.Exits[string(direction)] = targetID
	targetRoom.Exits[string(oppositeDir)] = sourceID

	return nil
}

// getOppositeDirection returns the opposite direction
func getOppositeDirection(dir string) string {
	direction := Direction(dir)
	return string(direction.GetOppositeDirection())
}

// generateDirectionalPrompt creates a themed prompt based on direction
func generateDirectionalPrompt(basePrompt string, direction string) string {
	switch direction {
	case "north":
		return fmt.Sprintf("%s - Upper Level", basePrompt)
	case "south":
		return fmt.Sprintf("%s - Lower Level", basePrompt)
	case "east":
		return fmt.Sprintf("%s - Eastern Wing", basePrompt)
	case "west":
		return fmt.Sprintf("%s - Western Wing", basePrompt)
	default:
		return basePrompt
	}
}

// generateRoom creates a new room using the world's RNG
func (w *World) generateRoom(description string, exits []string) (*Room, error) {
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

	room := &Room{
		ID:          fmt.Sprintf("room_%d", w.rng.Int63()),
		Description: enhancedDesc,
		Objects:     make([]Object, 0),
		Exits:       exitMap,
		Properties:  make(map[string]interface{}),
	}

	return room, nil
}

// Save writes the world to a JSON file
func (w *World) Save(filename string) error {
	data, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal world: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write world file: %w", err)
	}

	return nil
}

// LoadWorld reads a world from a JSON file
func LoadWorld(filename string) (*World, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read world file: %w", err)
	}

	var world World
	if err := json.Unmarshal(data, &world); err != nil {
		return nil, fmt.Errorf("failed to unmarshal world: %w", err)
	}

	// Recreate the RNG with the saved seed
	world.rng = rand.New(rand.NewSource(world.Seed))

	return &world, nil
}

// generateID creates a unique ID for a room
// In a real implementation, this would use a more sophisticated ID generation method
func generateID() string {
	return fmt.Sprintf("room_%d", rand.Int63())
}

// generateRandomString creates a random string of the specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
