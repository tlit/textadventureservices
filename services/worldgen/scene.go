package worldgen

import (
	"encoding/json"
	"fmt"
	"os"
)

// Scene represents a location in the game world
type Scene struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Objects     []Object               `json:"objects"`
	Exits       map[string]string      `json:"exits"`
	Properties  map[string]interface{} `json:"properties"`
}

// NewScene creates a new scene with the given properties
func NewScene(id, description string) *Scene {
	return &Scene{
		ID:          id,
		Description: description,
		Objects:     make([]Object, 0),
		Exits:       make(map[string]string),
		Properties:  make(map[string]interface{}),
	}
}

// AddObject adds an object to the scene
func (s *Scene) AddObject(object Object) {
	s.Objects = append(s.Objects, object)
}

// GetObject returns an object by its index
func (s *Scene) GetObject(index int) (Object, bool) {
	if index < 0 || index >= len(s.Objects) {
		return Object{}, false
	}
	return s.Objects[index], true
}

// AddExit adds an exit to the scene
func (s *Scene) AddExit(direction string, destination string) {
	s.Exits[direction] = destination
}

// GetExit returns the destination of an exit
func (s *Scene) GetExit(direction string) (string, bool) {
	destination, ok := s.Exits[direction]
	return destination, ok
}

// AddProperty adds a property to the scene
func (s *Scene) AddProperty(key string, value interface{}) {
	s.Properties[key] = value
}

// GetProperty returns the value of a property
func (s *Scene) GetProperty(key string) (interface{}, bool) {
	value, ok := s.Properties[key]
	return value, ok
}

// Save writes the scene to a JSON file
func (s *Scene) Save(filename string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal scene: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write scene file: %w", err)
	}

	return nil
}

// LoadScene reads a scene from a JSON file
func LoadScene(filename string) (*Scene, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read scene file: %w", err)
	}

	var scene Scene
	if err := json.Unmarshal(data, &scene); err != nil {
		return nil, fmt.Errorf("failed to unmarshal scene: %w", err)
	}

	return &scene, nil
}
