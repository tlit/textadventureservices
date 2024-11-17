package logging

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestHTTPLogger(t *testing.T) {
	// Create a test server to mock the logging service
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/logs" {
			t.Errorf("Expected /api/v1/logs path, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create logger with test server endpoint
	logger := NewHTTPLogger(server.URL)

	// Test all log levels
	tests := []struct {
		name     string
		logFunc  func(context.Context, string)
		message  string
		logLevel LogLevel2
	}{
		{"info_log", logger.Info, "test info message", LogLevelInfo2},
		{"warn_log", logger.Warn, "test warn message", LogLevelWarn2},
		{"error_log", logger.Error, "test error message", LogLevelError2},
		{"debug_log", logger.Debug, "test debug message", LogLevelDebug2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			// Log should not block or panic
			tt.logFunc(ctx, tt.message)

			// Give some time for async log processing
			time.Sleep(100 * time.Millisecond)
		})
	}
}

func TestNoopLogger(t *testing.T) {
	logger := NewNoopLogger()
	ctx := context.Background()

	// Test all log levels - should not panic
	logger.Info(ctx, "test info")
	logger.Warn(ctx, "test warn")
	logger.Error(ctx, "test error")
	logger.Debug(ctx, "test debug")
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		level    LogLevel2
		expected string
	}{
		{LogLevelInfo2, "INFO"},
		{LogLevelWarn2, "WARN"},
		{LogLevelError2, "ERROR"},
		{LogLevelDebug2, "DEBUG"},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			if tt.level.String() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.level.String())
			}
		})
	}
}

func TestHTTPLoggerEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		wantErr  bool
	}{
		{"valid_endpoint", "http://localhost:8081", false},
		{"empty_endpoint", "", false}, // Empty endpoint should not cause error, just fail silently
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewHTTPLogger(tt.endpoint)
			if logger == nil {
				t.Error("NewHTTPLogger returned nil")
			}
			if logger.endpoint != tt.endpoint {
				t.Errorf("Logger endpoint not set correctly. Got %s, want %s",
					logger.endpoint, tt.endpoint)
			}
			if logger.service != "worldgen" {
				t.Errorf("Logger service not set correctly. Got %s, want worldgen",
					logger.service)
			}
		})
	}
}

func TestHTTPLoggerTimeout(t *testing.T) {
	// Create a slow server that will trigger timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Longer than client timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := NewHTTPLogger(server.URL)
	ctx := context.Background()

	// Log should not block despite server being slow
	done := make(chan bool)
	go func() {
		logger.Info(ctx, "test timeout message")
		done <- true
	}()

	select {
	case <-done:
		// Success - log call returned quickly
	case <-time.After(time.Second):
		t.Error("Logger did not return within timeout period")
	}
}

func TestLogRetention(t *testing.T) {
	logger := NewQuantumLogger(
		WithBufferSize(100),
		WithRetentionPeriod(time.Second),
	)
	defer logger.Shutdown()

	ctx := context.Background()
	now := time.Now()

	// Create logs with different timestamps
	logs := []struct {
		message string
		age     time.Duration
		expect  bool
	}{
		{"recent log", 0, true},
		{"old log", -2 * time.Second, false},
		{"very old log", -5 * time.Second, false},
	}

	for _, log := range logs {
		logger.logWithTime(ctx, LogLevelInfo2, log.message, now.Add(log.age))
	}

	// Wait for logs to be processed
	time.Sleep(100 * time.Millisecond)

	// Fetch logs and verify retention
	fetchedLogs := logger.Fetch(ctx, now.Add(-time.Second), now.Add(time.Second))
	for _, log := range logs {
		found := false
		for _, fetched := range fetchedLogs {
			if fetched.Message == log.message {
				found = true
				break
			}
		}
		if found != log.expect {
			t.Errorf("Log %q: expected found=%v, got found=%v", log.message, log.expect, found)
		}
	}
}

func TestLogRotation(t *testing.T) {
	logger := NewQuantumLogger(
		WithBufferSize(5), // Small buffer to force rotation
	)
	defer logger.Shutdown()

	ctx := context.Background()

	// Generate more logs than buffer size
	for i := 0; i < 10; i++ {
		logger.Info(ctx, fmt.Sprintf("Log message %d", i))
	}

	// Wait for logs to be processed
	time.Sleep(100 * time.Millisecond)

	// Fetch recent logs
	logs := logger.Fetch(ctx, time.Now().Add(-time.Hour), time.Now())

	// Verify log count matches buffer size
	if len(logs) > 5 {
		t.Errorf("Expected at most 5 logs after rotation, got %d", len(logs))
	}

	// Verify logs are the most recent ones
	for i, log := range logs {
		expected := fmt.Sprintf("Log message %d", i+5)
		if log.Message != expected {
			t.Errorf("Expected log message %q, got %q", expected, log.Message)
		}
	}
}

func TestHighLoadLogging(t *testing.T) {
	logger := NewQuantumLogger(
		WithBufferSize(1000),
	)
	defer logger.Shutdown()

	ctx := context.Background()
	const numRoutines = 10
	const logsPerRoutine = 100

	var wg sync.WaitGroup
	startTime := time.Now()

	// Launch multiple goroutines to write logs concurrently
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()
			for j := 0; j < logsPerRoutine; j++ {
				logger.Info(ctx, fmt.Sprintf("Routine %d: Log %d", routineID, j))
				// Add small delay to prevent overwhelming the buffer
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	// Wait for all routines to finish
	wg.Wait()

	// Wait for logs to be processed
	time.Sleep(200 * time.Millisecond)

	// Fetch all logs
	logs := logger.Fetch(ctx, startTime, time.Now().Add(time.Second))

	// Verify log count
	expectedLogs := numRoutines * logsPerRoutine
	if len(logs) != expectedLogs {
		t.Errorf("Expected %d logs, got %d", expectedLogs, len(logs))
	}

	// Create a map to track logs from each routine
	logCounts := make(map[int]int)
	for _, log := range logs {
		var routineID, logNum int
		if n, err := fmt.Sscanf(log.Message, "Routine %d: Log %d", &routineID, &logNum); n == 2 && err == nil {
			logCounts[routineID]++
		}
	}

	// Verify each routine's logs were captured
	for i := 0; i < numRoutines; i++ {
		if count := logCounts[i]; count != logsPerRoutine {
			t.Errorf("Routine %d: expected %d logs, got %d", i, logsPerRoutine, count)
		}
	}
}

func TestLogFiltering(t *testing.T) {
	logger := NewQuantumLogger(WithBufferSize(100))
	defer logger.Shutdown()

	ctx := context.Background()
	now := time.Now()

	// Create logs with different levels and patterns
	testLogs := []struct {
		level   LogLevel2
		message string
		tags    []string
	}{
		{LogLevelInfo2, "User login successful", []string{"auth", "success"}},
		{LogLevelError2, "Database connection failed", []string{"db", "error"}},
		{LogLevelWarn2, "High memory usage", []string{"system", "performance"}},
		{LogLevelDebug2, "Processing request", []string{"api", "debug"}},
	}

	// Write test logs
	for _, log := range testLogs {
		logger.Log(ctx, log.level, log.message)
	}

	// Wait for logs to be processed
	time.Sleep(100 * time.Millisecond)

	// Test filtering by level
	tests := []struct {
		name          string
		filterFunc    func(LogEntry) bool
		expectedLogs  int
		expectedLevel LogLevel2
	}{
		{
			name: "filter_errors",
			filterFunc: func(entry LogEntry) bool {
				return entry.Level == LogLevelError2
			},
			expectedLogs:  1,
			expectedLevel: LogLevelError2,
		},
		{
			name: "filter_info_and_warn",
			filterFunc: func(entry LogEntry) bool {
				return entry.Level == LogLevelInfo2 || entry.Level == LogLevelWarn2
			},
			expectedLogs: 2,
		},
		{
			name: "filter_by_message",
			filterFunc: func(entry LogEntry) bool {
				return strings.Contains(entry.Message, "login")
			},
			expectedLogs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs := logger.Fetch(ctx, now, time.Now())
			filtered := filterLogs(logs, tt.filterFunc)

			if len(filtered) != tt.expectedLogs {
				t.Errorf("Expected %d logs, got %d", tt.expectedLogs, len(filtered))
			}

			if tt.expectedLevel != 0 {
				for _, log := range filtered {
					if log.Level != tt.expectedLevel {
						t.Errorf("Expected log level %v, got %v", tt.expectedLevel, log.Level)
					}
				}
			}
		})
	}
}

func filterLogs(logs []LogEntry, filterFunc func(LogEntry) bool) []LogEntry {
	var filtered []LogEntry
	for _, log := range logs {
		if filterFunc(log) {
			filtered = append(filtered, log)
		}
	}
	return filtered
}
