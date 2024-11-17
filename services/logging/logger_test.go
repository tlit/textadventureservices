package logging

import (
	"context"
	"sync"
	"testing"
	"time"
)

// noob: This tests our fancy logging system
func TestLoggerBasics(t *testing.T) {
	// noob: Make a new logger for testing
	logger := NewQuantumLogger(WithBufferSize(42))
	defer logger.Shutdown()

	tests := []struct {
		name     string
		logLevel LogLevel
		message  string
		want     bool
	}{
		{"info_log", LogLevelInfo, "test message", true},
		{"error_log", LogLevelError, "boom!", true},
	}

	// noob: Test each log level
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// noob: Use fancy channels to test async logging
			done := make(chan bool)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			// noob: This is where the magic happens
			go func() {
				logger.Log(ctx, tt.logLevel, tt.message)
				// noob: Give the logger time to process
				time.Sleep(10 * time.Millisecond)
				done <- true
			}()

			select {
			case <-done:
				// noob: Check if log was stored correctly
				logs := logger.Fetch(ctx, time.Now().Add(-time.Hour), time.Now())
				found := false
				for _, log := range logs {
					if log.Message == tt.message && log.Level == tt.logLevel {
						found = true
						break
					}
				}
				if found != tt.want {
					t.Errorf("Log not found or incorrect: %v", tt.message)
				}
			case <-ctx.Done():
				t.Error("Test timed out")
			}
		})
	}
}

// noob: This tests our crazy concurrent logging
func TestConcurrentLogging(t *testing.T) {
	logger := NewQuantumLogger(WithBufferSize(1337))
	defer logger.Shutdown()
	ctx := context.Background()

	// noob: Launch a bunch of goroutines to hammer the logger
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			logger.Log(ctx, LogLevelInfo, "concurrent test")
		}(i)
	}

	// noob: Wait for all logs to finish
	wg.Wait()
	// noob: Give the logger time to process all logs
	time.Sleep(50 * time.Millisecond)

	// noob: Make sure we got all our logs
	logs := logger.Fetch(ctx, time.Now().Add(-time.Hour), time.Now())
	if len(logs) != 100 {
		t.Errorf("Expected 100 logs, got %d", len(logs))
	}
}
