# World Generation Service

A sophisticated world generation service for creating dynamic, interconnected game environments with AI-enhanced descriptions.

## Status: 
Current Version: 1.0.0

## Features

- **AI-Powered World Generation**
  - Dynamic scene descriptions using GPT-4o
  - Advanced language model capabilities
  - Contextual object generation
  - Prompt enhancement for better results
  - OpenAI integration with configurable models

- **Scene Graph Management**
  - Robust bidirectional scene connections
  - Unique room identification system
  - Multiple direction support (N, S, E, W, Up, Down)
  - Path finding between scenes
  - Scene validation and error checking

- **Object Interaction**
  - Dynamic object properties
  - Interactive object generation
  - Property management
  - Error validation

## Room Connection System

The service implements a robust room connection system with the following features:

- **Unique Room Identification**
  - Timestamp-based IDs with counter to prevent duplicates
  - Thread-safe room creation
  - Consistent ID format: `room_[timestamp]_[counter]`

- **Bidirectional Connections**
  - Automatic two-way connections between rooms
  - Direction/opposite direction mapping
  - Proper exit map initialization
  - Connection validation

- **Direction Support**
  - Cardinal directions (North, South, East, West)
  - Vertical movement (Up, Down)
  - Extensible for custom directions

## Configuration

The service is configured via environment files:

### OpenAI Configuration (.env.openai)
```env
OPENAI_API_KEY=your-api-key-here
OPENAI_MODEL=gpt-4o  # Supported models: gpt-4o, gpt-4, gpt-3.5-turbo

# Optional parameters
OPENAI_TEMPERATURE=0.7
OPENAI_MAX_TOKENS=150
OPENAI_TOP_P=1.0
OPENAI_FREQUENCY_PENALTY=0.0
OPENAI_PRESENCE_PENALTY=0.0
```

## API Endpoints

### Generate World
- **Path**: `/api/v1/generate-world`
- **Method**: POST
- **Request Body**:
  ```json
  {
    "prompt": "A haunted castle on a stormy night",
    "num_rooms": 5,
    "theme": "gothic_horror"
  }
  ```
- **Response**:
  ```json
  {
    "title": "The Haunted Castle",
    "world": {
      "id": "world_xyz",
      "scenes": [
        {
          "id": "room_1704999999_1",
          "description": "You stand in a dimly lit hallway...",
          "objects": [
            {
              "id": "obj_1",
              "name": "rusty candlestick",
              "description": "A weathered brass candlestick..."
            }
          ],
          "exits": {
            "north": "room_1704999999_2",
            "east": "room_1704999999_3"
          }
        }
      ]
    }
  }
  ```

## Architecture

### Components

1. **AI Provider Interface**
   - Generic interface for AI providers
   - Supports multiple provider implementations
   - Configurable endpoints and models
   - Methods for generating content

2. **World Generation Engine**
   - Scene graph management
   - Room connection handling
   - Object generation and management
   - State persistence

3. **API Layer**
   - RESTful endpoints
   - JSON request/response handling
   - Input validation
   - Error handling

## Usage

### Command Line
```bash
# Generate a world with 5 rooms
go run ./services/worldgen/cmd/worldgen \
  -prompt "surrealist inflatable spacestation insanity" \
  -rooms 5

# Save world to file
go run ./services/worldgen/cmd/worldgen \
  -prompt "cyberpunk neon city" \
  -rooms 10 \
  -output world.json
```

### API
```bash
# Generate a world using the API
curl -X POST http://localhost:8080/api/v1/generate-world \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "underwater crystal palace",
    "num_rooms": 5,
    "theme": "aquatic_fantasy"
  }'
```

## Testing

Run all tests:
```bash
go test ./services/worldgen/... -v
```

Run specific test:
```bash
go test ./services/worldgen -run TestRoomConnections
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request
