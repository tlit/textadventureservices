# Quantum Logger Service

A high-performance, concurrent logging service for the Text Adventure Engine, featuring quantum-inspired buffering and sophisticated synchronization mechanisms.

## Status: ✅ Completed

Current Version: 1.0.0

## Features

- ✅ High-performance concurrent log processing
- ✅ Non-blocking log submission
- ✅ Time-based log querying
- ✅ Configurable buffer sizes
- ✅ Thread-safe operations
- ✅ HTTP API endpoints
- ✅ Comprehensive test coverage

## Architecture

The service uses a quantum-inspired architecture with the following components:

- **QuantumLogger**: Core logging engine with buffered channels
- **LoggingHandler**: HTTP interface for external services
- **Concurrent Processing**: Background goroutine for log persistence
- **Thread-safe Storage**: Mutex-protected log storage

## API Endpoints

### Submit Log
```http
POST /api/v1/logs
Content-Type: application/json

{
    "level": "INFO",
    "message": "Player entered the dungeon",
    "service": "game-engine",
    "metadata": {
        "playerId": "12345",
        "location": "dungeon-entrance"
    }
}
```

### Fetch Logs
```http
GET /api/v1/logs/fetch?start=2024-01-01T00:00:00Z&end=2024-01-02T00:00:00Z
```

## Usage

### Basic Usage

```go
// Create a new logger
logger := logging.NewQuantumLogger(
    logging.WithBufferSize(4096),
)
defer logger.Shutdown()

// Log a message
ctx := context.Background()
logger.Log(ctx, logging.LogLevelInfo, "Game started")

// Fetch recent logs
logs := logger.Fetch(
    ctx,
    time.Now().Add(-1*time.Hour),
    time.Now(),
)
```

### HTTP Server Setup

```go
func main() {
    logger := logging.NewQuantumLogger()
    handler := logging.NewLoggingHandler(logger)
    
    mux := http.NewServeMux()
    handler.RegisterRoutes(mux)
    
    http.ListenAndServe(":8080", mux)
}
```

## Configuration

The logger can be configured using functional options:

```go
logger := logging.NewQuantumLogger(
    logging.WithBufferSize(8192),  // Increase buffer size
)
```

## Log Levels

- `LogLevelInfo`: General information
- `LogLevelWarn`: Warning conditions
- `LogLevelError`: Error conditions
- `LogLevelDebug`: Debug information

## Performance Considerations

- Uses buffered channels for non-blocking log submission
- Concurrent processing of logs in background
- Read-Write mutex for thread-safe storage access
- Configurable buffer sizes for memory management

## Testing

Run the test suite:

```bash
go test ./services/logging/... -v
```

The test suite includes:
- Basic logging functionality
- Concurrent logging stress tests
- Time-based log retrieval
- Error handling

## Dependencies

- Go 1.21 or higher
- Standard library only (no external dependencies)

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

MIT License - see LICENSE file for details
