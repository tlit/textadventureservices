package worldgen

// Direction represents a possible direction of movement
type Direction string

// Direction constants
const (
	North Direction = "north"
	South Direction = "south"
	East  Direction = "east"
	West  Direction = "west"
)

// GetOppositeDirection returns the opposite direction
func (d Direction) GetOppositeDirection() Direction {
	switch d {
	case North:
		return South
	case South:
		return North
	case East:
		return West
	case West:
		return East
	default:
		return ""
	}
}

// IsValidDirection checks if a direction is valid
func IsValidDirection(dir Direction) bool {
	switch dir {
	case North, South, East, West:
		return true
	default:
		return false
	}
}
