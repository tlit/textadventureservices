package worldgen

import (
	"os"
	"testing"
)

func TestNewWorld(t *testing.T) {
	seed := int64(42)
	world, err := NewWorld(seed)
	if err != nil {
		t.Fatalf("Failed to create new world: %v", err)
	}

	if world.Seed != seed {
		t.Errorf("Expected seed %d, got %d", seed, world.Seed)
	}

	if world.Rooms == nil {
		t.Error("Expected rooms map to be initialized")
	}

	if world.rng == nil {
		t.Error("Expected RNG to be initialized")
	}
}

func TestAddRoom(t *testing.T) {
	world, err := NewWorld(42)
	if err != nil {
		t.Fatalf("Failed to create new world: %v", err)
	}

	room := &Room{
		ID:          "room1",
		Description: "Test Room",
		Objects:     make([]Object, 0),
		Exits:       make(map[string]string),
		Properties:  make(map[string]interface{}),
	}

	world.AddRoom(room)

	if len(world.Rooms) != 1 {
		t.Errorf("Expected 1 room, got %d", len(world.Rooms))
	}

	if world.Rooms["room1"] != room {
		t.Error("Room not added correctly")
	}
}

func TestConnectRooms(t *testing.T) {
	world, err := NewWorld(42)
	if err != nil {
		t.Fatalf("Failed to create new world: %v", err)
	}

	room1 := &Room{
		ID:          "room1",
		Description: "Room 1",
		Objects:     make([]Object, 0),
		Exits:       make(map[string]string),
		Properties:  make(map[string]interface{}),
	}
	room2 := &Room{
		ID:          "room2",
		Description: "Room 2",
		Objects:     make([]Object, 0),
		Exits:       make(map[string]string),
		Properties:  make(map[string]interface{}),
	}

	world.AddRoom(room1)
	world.AddRoom(room2)

	err = world.ConnectRooms(room1.ID, room2.ID, North)
	if err != nil {
		t.Errorf("Failed to connect rooms: %v", err)
	}

	// Check if room1 has an exit to room2
	if targetID, ok := room1.Exits[string(North)]; !ok || targetID != room2.ID {
		t.Errorf("Room1 should have a north exit to room2")
	}

	// Check if room2 has an exit back to room1
	if targetID, ok := room2.Exits[string(South)]; !ok || targetID != room1.ID {
		t.Errorf("Room2 should have a south exit to room1")
	}

	// Test connecting non-existent rooms
	err = world.ConnectRooms("nonexistent1", "nonexistent2", East)
	if err == nil {
		t.Error("Expected error when connecting non-existent rooms")
	}
}

func TestGenerateWorld(t *testing.T) {
	world, err := NewWorld(42)
	if err != nil {
		t.Fatalf("Failed to create new world: %v", err)
	}

	basePrompt := "A mysterious castle"
	numRooms := 3

	err = world.GenerateWorld(basePrompt, numRooms)
	if err != nil {
		t.Fatalf("Failed to generate world: %v", err)
	}

	if len(world.Rooms) != numRooms {
		t.Errorf("Expected %d rooms, got %d", numRooms, len(world.Rooms))
	}

	// Check that each room has proper connections
	for _, room := range world.Rooms {
		if room.Description == "" {
			t.Error("Room description should not be empty")
		}

		if room.Objects == nil {
			t.Error("Room objects should be initialized")
		}

		if room.Exits == nil {
			t.Error("Room exits should be initialized")
		}

		if room.Properties == nil {
			t.Error("Room properties should be initialized")
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

func TestSaveAndLoadWorld(t *testing.T) {
	// Create a test world
	world, err := NewWorld(42)
	if err != nil {
		t.Fatalf("Failed to create new world: %v", err)
	}

	room := &Room{
		ID:          "room1",
		Description: "Test Room",
		Objects: []Object{
			{
				ID:          "obj1",
				Name:        "Test Object",
				Description: "A test object",
				Properties: map[string]interface{}{
					"portable": true,
					"weight":   1.5,
				},
			},
		},
		Exits: map[string]string{
			string(North): "room2",
		},
		Properties: map[string]interface{}{
			"lit":         true,
			"temperature": 20.5,
		},
	}
	world.AddRoom(room)

	// Save the world
	filename := "test_world.json"
	err = world.Save(filename)
	if err != nil {
		t.Fatalf("Failed to save world: %v", err)
	}
	defer os.Remove(filename) // Clean up after test

	// Load the world
	loadedWorld, err := LoadWorld(filename)
	if err != nil {
		t.Fatalf("Failed to load world: %v", err)
	}

	// Verify loaded world matches original
	if loadedWorld.Seed != world.Seed {
		t.Errorf("Expected world seed %d, got %d", world.Seed, loadedWorld.Seed)
	}

	if len(loadedWorld.Rooms) != len(world.Rooms) {
		t.Errorf("Expected %d rooms, got %d", len(world.Rooms), len(loadedWorld.Rooms))
	}

	// Check that the loaded room matches the original
	loadedRoom, ok := loadedWorld.Rooms["room1"]
	if !ok {
		t.Fatal("Room not found in loaded world")
	}

	if loadedRoom.Description != room.Description {
		t.Errorf("Expected room description %s, got %s", room.Description, loadedRoom.Description)
	}

	if len(loadedRoom.Objects) != len(room.Objects) {
		t.Errorf("Expected %d objects, got %d", len(room.Objects), len(loadedRoom.Objects))
	}

	// Check that the loaded object matches the original
	loadedObj := loadedRoom.Objects[0]
	originalObj := room.Objects[0]

	if loadedObj.Name != originalObj.Name {
		t.Errorf("Expected object name %s, got %s", originalObj.Name, loadedObj.Name)
	}

	if loadedObj.Description != originalObj.Description {
		t.Errorf("Expected object description %s, got %s", originalObj.Description, loadedObj.Description)
	}

	// Check properties
	if v, ok := loadedObj.Properties["portable"]; !ok || v != true {
		t.Errorf("Expected object property 'portable' to be true")
	}

	if v, ok := loadedObj.Properties["weight"]; !ok || v != 1.5 {
		t.Errorf("Expected object property 'weight' to be 1.5")
	}

	// Check exits
	if v, ok := loadedRoom.Exits[string(North)]; !ok || v != "room2" {
		t.Errorf("Expected room exit 'north' to lead to 'room2'")
	}

	// Check room properties
	if v, ok := loadedRoom.Properties["lit"]; !ok || v != true {
		t.Errorf("Expected room property 'lit' to be true")
	}

	if v, ok := loadedRoom.Properties["temperature"]; !ok || v != 20.5 {
		t.Errorf("Expected room property 'temperature' to be 20.5")
	}
}
