package worldgen

import (
	"testing"
)

func TestSceneCreation(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		sceneName   string
		description string
		wantErr     bool
	}{
		{
			name:        "Valid scene creation",
			id:          "scene_1",
			sceneName:   "Forest Clearing",
			description: "A peaceful forest clearing with sunlight filtering through the trees.",
			wantErr:     false,
		},
		{
			name:        "Another valid scene",
			id:          "scene_2",
			sceneName:   "Dark Cave",
			description: "A dark cave with mysterious echoes.",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scene := NewScene(tt.id, tt.sceneName, tt.description)
			if scene == nil && !tt.wantErr {
				t.Errorf("NewScene() returned nil, wanted non-nil")
			}

			if scene != nil {
				if scene.ID != tt.id {
					t.Errorf("Scene.ID = %v, want %v", scene.ID, tt.id)
				}
				if scene.Name != tt.sceneName {
					t.Errorf("Scene.Name = %v, want %v", scene.Name, tt.sceneName)
				}
				if scene.Description != tt.description {
					t.Errorf("Scene.Description = %v, want %v", scene.Description, tt.description)
				}
			}
		})
	}
}

func TestRoomConnections(t *testing.T) {
	tests := []struct {
		name      string
		fromDesc  string
		toDesc    string
		direction Direction
		wantErr   bool
	}{
		{
			name:      "Connect north-south",
			fromDesc:  "Starting room",
			toDesc:    "Northern room",
			direction: North,
			wantErr:   false,
		},
		{
			name:      "Connect east-west",
			fromDesc:  "Eastern room",
			toDesc:    "Western room",
			direction: East,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scene := NewScene("test_scene", "Test Scene", "A test scene")

			// Create the rooms
			from, err := GenerateRoom(tt.fromDesc, nil)
			if err != nil {
				t.Fatalf("Failed to create from room: %v", err)
			}

			to, err := GenerateRoom(tt.toDesc, nil)
			if err != nil {
				t.Fatalf("Failed to create to room: %v", err)
			}

			// Add rooms to scene
			scene.AddRoom(from)
			scene.AddRoom(to)

			// Debug info
			t.Logf("From room ID: %s", from.ID)
			t.Logf("To room ID: %s", to.ID)
			t.Logf("Direction: %s", tt.direction)

			// Connect the rooms
			err = scene.ConnectRooms(from.ID, to.ID, tt.direction)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConnectRooms() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Debug info after connection
			t.Logf("From room exits: %v", from.Exits)
			t.Logf("To room exits: %v", to.Exits)

			// Verify connections
			if !tt.wantErr {
				// Check forward connection
				targetID, ok := from.Exits[string(tt.direction)]
				if !ok || targetID != to.ID {
					t.Errorf("Forward connection not found or incorrect. Got %v, want %v", targetID, to.ID)
					t.Logf("From room exits map: %v", from.Exits)
				}

				// Check reverse connection
				if tt.direction == North {
					sourceID, ok := to.Exits[string(South)]
					if !ok || sourceID != from.ID {
						t.Errorf("Reverse connection not found or incorrect. Got %v, want %v", sourceID, from.ID)
						t.Logf("To room exits map: %v", to.Exits)
					}
				} else if tt.direction == East {
					sourceID, ok := to.Exits[string(West)]
					if !ok || sourceID != from.ID {
						t.Errorf("Reverse connection not found or incorrect. Got %v, want %v", sourceID, from.ID)
						t.Logf("To room exits map: %v", to.Exits)
					}
				}
			}
		})
	}
}
