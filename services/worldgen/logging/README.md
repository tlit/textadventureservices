# World Generation Service Logging

This package provides logging capabilities for the World Generation Service, integrating with the Quantum Logger Service.

## Features

- 🔄 Non-blocking log submission
- 🌐 HTTP-based logging integration
- 📝 Multiple log levels (INFO, WARN, ERROR, DEBUG)
- 🔌 Pluggable logger interface
- 🚫 No-op logger for testing/development

## Usage

### Basic Usage

```go
// Create a logger that sends logs to the logging service
logger := logging.NewHTTPLogger("http://localhost:8081")

// Log messages at different levels
ctx := context.Background()
logger.Info(ctx, "World generation started")
logger.Debug(ctx, "Using seed: 12345")
logger.Warn(ctx, "Room connection failed, retrying")
logger.Error(ctx, "Failed to generate description")
```

### Using with World Generation Service

The logger is automatically configured when creating a new service:

```go
cfg := config.DefaultConfig()
cfg.LoggingEndpoint = "http://localhost:8081"

service, err := worldgen.NewService(cfg)
if err != nil {
    log.Fatalf("Failed to create service: %v", err)
}
```

### Testing with NoopLogger

For testing or development without the logging service:

```go
logger := logging.NewNoopLogger()
```

## Log Levels

- **INFO**: General information about service operation
- **WARN**: Warning conditions that might need attention
- **ERROR**: Error conditions that need immediate attention
- **DEBUG**: Detailed information for debugging

## Integration with Quantum Logger

The HTTPLogger sends logs to the Quantum Logger Service with the following format:

```json
{
    "level": "INFO",
    "message": "World generation started",
    "service": "worldgen",
    "metadata": {
        "timestamp": "2024-01-21T15:04:05Z"
    }
}
```

## Configuration

Configure the logging endpoint in your service configuration:

```json
{
    "ai_provider": {
        "type": "ollama",
        "endpoint": "http://localhost:11434",
        "model": "llama2"
    },
    "server": {
        "port": 8080
    },
    "logging_endpoint": "http://localhost:8081"
}
```

## Error Handling

- Log submission failures are handled gracefully and won't block service operation
- Failed log submissions are reported to stdout for monitoring
- HTTP client has a 5-second timeout to prevent hanging

## Best Practices

1. Always provide context in log messages
2. Use appropriate log levels
3. Include relevant metadata when possible
4. Keep log messages concise but informative
5. Use DEBUG level for detailed operation information

## Development

### Adding a New Logger Implementation

1. Implement the `Logger` interface:
```go
type Logger interface {
    Info(ctx context.Context, msg string)
    Warn(ctx context.Context, msg string)
    Error(ctx context.Context, msg string)
    Debug(ctx context.Context, msg string)
}
```

2. Add any necessary configuration
3. Implement error handling
4. Add tests for the new implementation

### Testing

Run the tests:
```bash
go test ./services/worldgen/logging/... -v
```
