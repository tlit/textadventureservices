package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// LogLevel represents the severity of a log entry
type LogLevel string

const (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
	LogLevelDebug LogLevel = "DEBUG"
)

// Logger defines the interface for logging in the worldgen service
type Logger interface {
	Info(ctx context.Context, msg string)
	Warn(ctx context.Context, msg string)
	Error(ctx context.Context, msg string)
	Debug(ctx context.Context, msg string)
}

// HTTPLogger implements Logger using the logging service HTTP API
type HTTPLogger struct {
	endpoint string
	client   *http.Client
	service  string
}

// NewHTTPLogger creates a new logger that sends logs to the logging service
func NewHTTPLogger(endpoint string) *HTTPLogger {
	return &HTTPLogger{
		endpoint: endpoint,
		client:   &http.Client{Timeout: 5 * time.Second},
		service:  "worldgen",
	}
}

type logRequest struct {
	Level    string                 `json:"level"`
	Message  string                 `json:"message"`
	Service  string                 `json:"service"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (l *HTTPLogger) log(ctx context.Context, level LogLevel, msg string) {
	logReq := logRequest{
		Level:    string(level),
		Message:  msg,
		Service:  l.service,
		Metadata: map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	// Send log asynchronously to avoid blocking
	go func() {
		jsonData, err := json.Marshal(logReq)
		if err != nil {
			fmt.Printf("Failed to marshal log request: %v\n", err)
			return
		}

		resp, err := l.client.Post(fmt.Sprintf("%s/api/v1/logs", l.endpoint), "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Failed to send log: %v\n", err)
			return
		}
		defer resp.Body.Close()
	}()
}

func (l *HTTPLogger) Info(ctx context.Context, msg string) {
	l.log(ctx, LogLevelInfo, msg)
}

func (l *HTTPLogger) Warn(ctx context.Context, msg string) {
	l.log(ctx, LogLevelWarn, msg)
}

func (l *HTTPLogger) Error(ctx context.Context, msg string) {
	l.log(ctx, LogLevelError, msg)
}

func (l *HTTPLogger) Debug(ctx context.Context, msg string) {
	l.log(ctx, LogLevelDebug, msg)
}

// NoopLogger implements Logger but does nothing
type NoopLogger struct{}

func NewNoopLogger() *NoopLogger {
	return &NoopLogger{}
}

func (l *NoopLogger) Info(ctx context.Context, msg string)  {}
func (l *NoopLogger) Warn(ctx context.Context, msg string)  {}
func (l *NoopLogger) Error(ctx context.Context, msg string) {}
func (l *NoopLogger) Debug(ctx context.Context, msg string) {}

// LogLevel2 represents the severity of a log entry
type LogLevel2 int

const (
	LogLevelDebug2 LogLevel2 = iota
	LogLevelInfo2
	LogLevelWarn2
	LogLevelError2
)

// String converts LogLevel2 to a string representation
func (l LogLevel2) String() string {
	switch l {
	case LogLevelDebug2:
		return "DEBUG"
	case LogLevelInfo2:
		return "INFO"
	case LogLevelWarn2:
		return "WARN"
	case LogLevelError2:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a single log message
type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel2
	Message   string
}

// QuantumLogger is a high-performance, thread-safe logger with retention and rotation capabilities
type QuantumLogger struct {
	mu              sync.RWMutex
	buffer          []LogEntry
	bufferSize      int
	retentionPeriod time.Duration
	bufferChan      chan LogEntry
	done            chan struct{}
}

// LoggerOption is a function type for configuring the logger
type LoggerOption func(*QuantumLogger)

// WithBufferSize sets the buffer size for the logger
func WithBufferSize(size int) LoggerOption {
	return func(l *QuantumLogger) {
		l.bufferSize = size
	}
}

// WithRetentionPeriod sets how long logs should be retained
func WithRetentionPeriod(duration time.Duration) LoggerOption {
	return func(l *QuantumLogger) {
		l.retentionPeriod = duration
	}
}

// NewQuantumLogger creates a new logger with the given options
func NewQuantumLogger(opts ...LoggerOption) *QuantumLogger {
	l := &QuantumLogger{
		bufferSize:      1000,                // default buffer size
		retentionPeriod: 24 * time.Hour,      // default retention period
		bufferChan:      make(chan LogEntry, 1000),
		done:            make(chan struct{}),
	}

	for _, opt := range opts {
		opt(l)
	}

	l.buffer = make([]LogEntry, 0, l.bufferSize)
	
	// Start background worker to process logs
	go l.processLogs()
	
	return l
}

// processLogs handles incoming log entries in the background
func (l *QuantumLogger) processLogs() {
	for {
		select {
		case entry := <-l.bufferChan:
			l.mu.Lock()
			// If buffer is full, remove oldest entry
			if len(l.buffer) >= l.bufferSize {
				l.buffer = l.buffer[1:]
			}
			l.buffer = append(l.buffer, entry)
			l.mu.Unlock()
		case <-l.done:
			return
		}
	}
}

// Log adds a new log entry
func (l *QuantumLogger) Log(ctx context.Context, level LogLevel2, message string) {
	l.logWithTime(ctx, level, message, time.Now())
}

// logWithTime adds a new log entry with a specific timestamp (used for testing)
func (l *QuantumLogger) logWithTime(ctx context.Context, level LogLevel2, message string, timestamp time.Time) {
	select {
	case l.bufferChan <- LogEntry{
		Timestamp: timestamp,
		Level:     level,
		Message:   message,
	}:
	default:
		// Buffer is full, log will be dropped
	}
}

// Debug logs a debug message
func (l *QuantumLogger) Debug(ctx context.Context, message string) {
	l.Log(ctx, LogLevelDebug2, message)
}

// Info logs an info message
func (l *QuantumLogger) Info(ctx context.Context, message string) {
	l.Log(ctx, LogLevelInfo2, message)
}

// Warn logs a warning message
func (l *QuantumLogger) Warn(ctx context.Context, message string) {
	l.Log(ctx, LogLevelWarn2, message)
}

// Error logs an error message
func (l *QuantumLogger) Error(ctx context.Context, message string) {
	l.Log(ctx, LogLevelError2, message)
}

// Fetch retrieves logs within the specified time range
func (l *QuantumLogger) Fetch(ctx context.Context, start, end time.Time) []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var result []LogEntry
	for _, entry := range l.buffer {
		if (entry.Timestamp.Equal(start) || entry.Timestamp.After(start)) &&
		   (entry.Timestamp.Equal(end) || entry.Timestamp.Before(end)) {
			result = append(result, entry)
		}
	}
	return result
}

// Shutdown performs any cleanup needed
func (l *QuantumLogger) Shutdown() {
	close(l.done)
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buffer = nil
}
