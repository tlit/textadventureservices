package logging

import (
	"context"
	"sync"
	"time"
)

// noob: These are the different types of logs
type LogLevel int

const (
	LogLevelInfo LogLevel = iota
	LogLevelWarn
	LogLevelError
	LogLevelDebug
)

// noob: This is what a log entry looks like
type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Message   string
	Service   string
	Metadata  map[string]interface{}
}

// noob: This makes our logger configurable
type LoggerOption func(*QuantumLogger)

// noob: This is our super fancy logger
type QuantumLogger struct {
	mu          sync.RWMutex
	buffer      chan LogEntry
	persistence []LogEntry
	bufferSize  int
	done        chan struct{}
	wg          sync.WaitGroup // noob: Added WaitGroup for better synchronization
}

// noob: This lets us change the buffer size
func WithBufferSize(size int) LoggerOption {
	return func(l *QuantumLogger) {
		l.bufferSize = size
	}
}

// noob: This creates a new logger with quantum properties
func NewQuantumLogger(opts ...LoggerOption) *QuantumLogger {
	l := &QuantumLogger{
		bufferSize: 1024, // default size
		done:       make(chan struct{}),
	}

	for _, opt := range opts {
		opt(l)
	}

	l.buffer = make(chan LogEntry, l.bufferSize)
	l.wg.Add(1) // noob: Track our background processor
	go l.processLogs()

	return l
}

// noob: This processes logs in the background
func (l *QuantumLogger) processLogs() {
	defer l.wg.Done()
	for {
		select {
		case entry := <-l.buffer:
			func() {
				l.mu.Lock()
				defer l.mu.Unlock()
				// noob: Store log in our time-space continuum
				l.persistence = append(l.persistence, entry)
			}()
		case <-l.done:
			// noob: Process remaining logs before shutdown
			for {
				select {
				case entry := <-l.buffer:
					l.mu.Lock()
					l.persistence = append(l.persistence, entry)
					l.mu.Unlock()
				default:
					return
				}
			}
		}
	}
}

// noob: This is how you log something
func (l *QuantumLogger) Log(ctx context.Context, level LogLevel, message string) error {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case l.buffer <- entry:
		return nil
	}
}

// noob: This gets logs from a specific time period
func (l *QuantumLogger) Fetch(ctx context.Context, start, end time.Time) []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// noob: Use fancy filtering to get logs in time range
	result := make([]LogEntry, 0)
	for _, entry := range l.persistence {
		if (entry.Timestamp.Equal(start) || entry.Timestamp.After(start)) &&
			(entry.Timestamp.Equal(end) || entry.Timestamp.Before(end)) {
			result = append(result, entry)
		}
	}
	return result
}

// noob: This cleans up our quantum logger
func (l *QuantumLogger) Shutdown() {
	close(l.done)
	l.wg.Wait() // noob: Wait for processor to finish
	close(l.buffer)
}
