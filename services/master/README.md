# Master Service

The Master Service is a central orchestrator for the Text Adventure Services system. It provides service registration, health monitoring, game state management, and AI-powered text processing through Ollama integration.

## Features

- Service Registration and Health Monitoring
  - Register and deregister services
  - Track service health status
  - Manage service endpoints

- Game State Management
  - Thread-safe state operations
  - Persistent state storage via JSON
  - Deep state copying to prevent mutations
  - Concurrent access support

- Ollama AI Integration
  - Natural language processing
  - Game command interpretation
  - Configurable AI model endpoint

## API Endpoints

### Service Management

- `POST /api/v1/services/register`
  - Register a new service
  - Request: `{ "name": "string", "endpoint": "string" }`

- `POST /api/v1/services/deregister`
  - Deregister an existing service
  - Request: `{ "name": "string" }`

- `GET /api/v1/services/status`
  - Get status of all registered services
  - Response: `{ "services": [{ "name": "string", "endpoint": "string", "status": "string" }] }`

### Game State

- `POST /api/v1/process-input`
  - Process game commands using Ollama
  - Request: `{ "input": "string" }`
  - Response: `{ "response": "string", "state_updates": {} }`

- `GET /api/v1/game-state`
  - Get current game state
  - Response: `{ "current_room": "string", "inventory": ["string"], "state": {} }`

## Configuration

Environment variables:
- `MASTER_PORT`: Service listening port (default: 8080)
- `OLLAMA_ENDPOINT`: Ollama service URL (default: http://localhost:11434)

## Development

### Prerequisites

- Go 1.20 or later
- Ollama service running locally or accessible via network

### Running Tests

```bash
# Run all tests
go test ./... -v

# Run specific package tests
go test ./state -v
go test ./ollama -v

# Run integration tests (requires Ollama)
OLLAMA_TEST_INTEGRATION=true go test ./ollama -v
```

### Project Structure

```
services/master/
├── main.go           # Service entry point and HTTP handlers
├── state/
│   ├── manager.go    # Game state management
│   └── manager_test.go
├── ollama/
│   ├── client.go     # Ollama API client
│   └── client_test.go
├── go.mod           # Go module file
└── README.md        # This file
```

## Security Considerations

- Thread-safe state management
- Configurable endpoints
- Basic error handling
- Isolated service design

## Future Improvements

- Enhanced error handling
- More granular service health checks
- Advanced logging
- Expanded Ollama integration
- Performance optimization
- UI service integration

## Contributing

1. Fork the repository
2. Create your feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

[Add your license information here]
