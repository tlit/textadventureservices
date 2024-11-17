package worldgen

import (
	"testing"
	"strings"
)

func TestWorldGeneration(t *testing.T) {
	tests := []struct {
		name      string
		numRooms  int
		basePrompt string
		wantErr   bool
	}{
		{
			name:      "valid_small_world",
			numRooms:  3,
			basePrompt: "test world",
			wantErr:   false,
		},
		{
			name:      "zero_rooms",
			numRooms:  0,
			basePrompt: "test world",
			wantErr:   true,
		},
		{
			name:      "negative_rooms",
			numRooms:  -1,
			basePrompt: "test world",
			wantErr:   true,
		},
		{
			name:      "empty_prompt",
			numRooms:  3,
			basePrompt: "",
			wantErr:   true,
		},
		{
			name:      "large_world",
			numRooms:  10,
			basePrompt: "test world",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world, err := NewWorld(12345) // Use consistent seed for reproducibility
			if err != nil {
				t.Fatalf("Failed to create world: %v", err)
			}

			err = world.GenerateWorld(tt.basePrompt, tt.numRooms)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateWorld() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify number of rooms
				if len(world.Rooms) != tt.numRooms {
					t.Errorf("Expected %d rooms, got %d", tt.numRooms, len(world.Rooms))
				}

				// Verify room connections
				for _, room := range world.Rooms {
					// Check each exit leads to a valid room
					for dir, targetID := range room.Exits {
						if targetID == "" {
							t.Errorf("Room %s has empty exit %s", room.ID, dir)
							continue
						}

						targetRoom, exists := world.Rooms[targetID]
						if !exists {
							t.Errorf("Room %s has exit %s leading to non-existent room %s", room.ID, dir, targetID)
							continue
						}

						// Verify bidirectional connection
						oppositeDir := getOppositeDirection(dir)
						if targetRoom.Exits[oppositeDir] != room.ID {
							t.Errorf("Room %s -> %s connection not bidirectional", room.ID, targetID)
						}
					}
				}

				// Verify room IDs are unique
				roomIDs := make(map[string]bool)
				for id := range world.Rooms {
					if roomIDs[id] {
						t.Errorf("Duplicate room ID found: %s", id)
					}
					roomIDs[id] = true
				}

				// Verify directional theming
				for _, room := range world.Rooms {
					if strings.Contains(room.Description, "north") && !strings.Contains(room.Description, "Upper Level") {
						t.Errorf("Northern room %s doesn't mention Upper Level: %s", room.ID, room.Description)
					}
					if strings.Contains(room.Description, "south") && !strings.Contains(room.Description, "Lower Level") {
						t.Errorf("Southern room %s doesn't mention Lower Level: %s", room.ID, room.Description)
					}
				}
			}
		})
	}
}

func TestWorldGenRoomConnections(t *testing.T) {
	world, err := NewWorld(12345)
	if err != nil {
		t.Fatalf("Failed to create world: %v", err)
	}

	// Test connecting rooms in all directions
	directions := []string{"north", "south", "east", "west"}
	rooms := make([]*Room, len(directions)+1)

	// Create central room
	rooms[0], err = world.generateRoom("Central Room", nil)
	if err != nil {
		t.Fatalf("Failed to generate central room: %v", err)
	}
	world.AddRoom(rooms[0])

	// Create and connect rooms in each direction
	for i, dir := range directions {
		rooms[i+1], err = world.generateRoom(
			"Room "+dir,
			[]string{getOppositeDirection(dir)},
		)
		if err != nil {
			t.Fatalf("Failed to generate room for direction %s: %v", dir, err)
		}
		world.AddRoom(rooms[i+1])

		// Connect rooms
		rooms[0].Exits[dir] = rooms[i+1].ID
		rooms[i+1].Exits[getOppositeDirection(dir)] = rooms[0].ID

		// Verify connection
		if rooms[0].Exits[dir] != rooms[i+1].ID {
			t.Errorf("Central room not connected to room %d in direction %s", i+1, dir)
		}
		if rooms[i+1].Exits[getOppositeDirection(dir)] != rooms[0].ID {
			t.Errorf("Room %d not connected back to central room", i+1)
		}
	}

	// Verify no duplicate connections
	for _, room := range rooms {
		seenExits := make(map[string]bool)
		for dir := range room.Exits {
			if seenExits[dir] {
				t.Errorf("Room %s has duplicate exit in direction %s", room.ID, dir)
			}
			seenExits[dir] = true
		}
	}
}
