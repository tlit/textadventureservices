package worldgen

// Room represents a location in the text adventure world
type Room struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Objects     []Object               `json:"objects"`
	Exits       map[string]string      `json:"exits"`
	Properties  map[string]interface{} `json:"properties"`
}

// Object represents an interactive item in a room
type Object struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Properties  map[string]interface{} `json:"properties"`
}
