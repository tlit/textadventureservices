# AI Service

The AI Service is a core component of the text adventure game system, responsible for processing user input and generating contextually appropriate responses using large language models.

## Status: 
Current Version: 1.0.0

## Features

- OpenAI Integration
  - Model configuration
  - Response parsing
  - Error handling

- Rate Limiting
  - Token usage tracking
  - Request rate control
  - Concurrent request management

- Error Recovery
  - Network error handling
  - API rate limits
  - Token quota management
  - Invalid response handling
  - Timeout handling

- Testing
  - Unit tests
  - Integration tests
  - Mock responses
  - Error scenario validation

## Overview

The AI Service provides:
- Natural language processing of user commands
- Context-aware response generation
- Fallback mechanisms for reliability
- Rate limiting for API usage control
- Configurable model selection

## Components

### 1. Service (`service.go`)
- Main service implementation
- Request handling and response processing
- Error management and recovery
- Context management

### 2. OpenAI Integration (`openai.go`)
- OpenAI API integration
- Model configuration
- Response parsing
- Error handling

### 3. Rate Limiter (`ratelimiter.go`)
- Token usage tracking
- Request rate control
- Concurrent request management

### 4. Configuration (`config.go`)
- Service configuration
- Model settings
- API credentials management
- Rate limit settings

## Testing

The service includes comprehensive testing:
- Unit tests (`service_test.go`)
- Integration tests (`integration_test.go`)
- Mock responses for testing
- Error scenario validation

### Test Coverage
- Fallback model behavior
- Error recovery mechanisms
- Concurrent request handling
- Token usage tracking
- Response validation

## Usage

### Basic Usage
```go
aiService := ai.NewService(config)
response, err := aiService.ProcessInput(ctx, "look around")
if err != nil {
    log.Printf("Error processing input: %v", err)
    return
}
fmt.Println(response)
```

### With Configuration
```go
config := ai.Config{
    Model:           "gpt-3.5-turbo",
    MaxTokens:       150,
    Temperature:     0.7,
    RateLimit:       60, // requests per minute
    TimeoutSeconds:  30,
}
```

## Error Handling

The service implements robust error handling:
1. Network errors
2. API rate limits
3. Token quota exceeded
4. Invalid responses
5. Timeout handling

## Rate Limiting

Rate limiting is implemented to:
- Prevent API quota exhaustion
- Manage concurrent requests
- Track token usage
- Handle backpressure

## Integration

The AI Service integrates with:
1. World Generation Service
2. Logging Service
3. Game State Management

## Configuration

### Environment Variables
```env
AI_MODEL=gpt-3.5-turbo
AI_MAX_TOKENS=150
AI_TEMPERATURE=0.7
AI_RATE_LIMIT=60
AI_TIMEOUT_SECONDS=30
```

### Model Configuration
- Primary model selection
- Fallback model settings
- Response parameters
- Token limits

## Best Practices

1. Always use context for cancellation
2. Handle rate limit errors gracefully
3. Implement fallback mechanisms
4. Monitor token usage
5. Log errors appropriately

## Dependencies

- OpenAI Go Client
- Context package
- Logging service
- Configuration management

## Future Improvements

1. Additional model support
2. Enhanced error recovery
3. Advanced rate limiting
4. Performance optimizations
5. Extended testing coverage
