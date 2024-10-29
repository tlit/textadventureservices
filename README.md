# AI-Powered Text Adventure Game Engine

A sophisticated text adventure game engine powered by AI, featuring dynamic world generation, intelligent game state management, and interactive storytelling.

## Project Overview

This project is a modern take on classic text adventure games, leveraging AI to create dynamic, responsive game worlds. The system consists of multiple microservices working together to provide an immersive gaming experience.

### Core Features

- Dynamic world generation based on user prompts
- AI-powered game state management
- Real-time text processing and response generation
- Web-based user interface with visual elements
- Service-oriented architecture for scalability

## Architecture

The project follows a microservices architecture with the following key components:

```
textadventureservices/
├── Specs/                    # Service specifications and API definitions
├── src/
│   ├── config/              # Configuration settings
│   ├── models/              # Data models and types
│   ├── services/            # Core services implementation
│   │   ├── auth/           # Authentication service
│   │   ├── master/         # Master coordination service
│   │   ├── ui/             # User interface service
│   │   └── worldgen/       # World generation service
│   └── utils/              # Utility functions and helpers
├── go-services/             # Go implementations of core services
└── tests/                   # Test suites
```

### Services

1. **Master Service**
   - Central orchestrator for all other services
   - Handles service registration and coordination
   - Manages game state and user sessions

2. **Auth Service**
   - User authentication and session management
   - Secure token generation and validation
   - User permissions and access control

3. **World Generation Service**
   - Creates dynamic game worlds based on prompts
   - Generates scene descriptions and connections
   - Manages world state and persistence

4. **UI Service**
   - Web-based user interface
   - Command input and processing
   - Game state visualization
   - Real-time updates

## API Endpoints

### Master Service Endpoints
- `POST /api/v1/services/register` - Register new services
- `POST /api/v1/services/deregister` - Deregister services
- `GET /api/v1/services/status` - Get service status
- `POST /api/v1/process-input` - Process user commands
- `GET /api/v1/game-state` - Get current game state

### Authentication Endpoints
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/validate` - Validate session tokens

### World Generation Endpoints
- `POST /api/v1/generate-world` - Generate new game world
- `POST /api/v1/render-image` - Generate scene visualizations

## Installation

1. System Requirements:
   - Python 3.11+
   - Go 1.x+
   - Redis (for state management)

2. Install Python dependencies:
```bash
pip install -r requirements.txt
```

3. Install Go dependencies:
```bash
cd go-services
go mod download
```

## Running the Services

1. Start the Master Service:
```bash
cd services/master
go run .
```

2. Start the Auth Service:
```bash
cd services/auth
go run .
```

3. Start the World Generation Service:
```bash
cd services/worldgen
python -m src.services.worldgen
```

4. Start the UI Service:
```bash
cd services/ui
python -m src.services.ui
```

## Development

### Code Style
- Python: Follow PEP 8
- Go: Use standard Go formatting
- Run `go fmt ./...` before committing Go code

### Testing
```bash
# Run Go tests
cd go-services
go test ./...

# Run Python tests
python -m pytest
```

### Data Models

The system uses several key data models:

1. **Scene**
   - Unique identifier
   - Description
   - Objects
   - Connectors to other scenes

2. **Object**
   - Name
   - Description
   - Properties
   - Interaction rules

3. **GameState**
   - Current scene
   - Inventory
   - Game progress
   - Player status

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit changes
4. Push to the branch
5. Create a Pull Request

## License

This project is proprietary and confidential.

## Version History

- v1.3.0 - Added world generation integration
- v1.2.0 - Complete service specification
- v1.1.0 - Added AI integration
- v1.0.0 - Initial release