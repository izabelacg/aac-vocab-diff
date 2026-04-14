package server

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// eventLogger appends one timestamped line per event to an io.Writer.
// All methods are nil-safe: a nil *eventLogger is a no-op logger.
// Concurrent writes are serialised by mu.
type eventLogger struct {
	mu sync.Mutex
	w  io.Writer
}

// newEventLogger opens path for append-only writing (creating it if needed)
// and returns a ready-to-use eventLogger. The caller is responsible for
// closing the underlying file when the server shuts down; for a long-running
// process this is effectively the lifetime of the program.
func newEventLogger(path string) (*eventLogger, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &eventLogger{w: f}, nil
}

// logEvent writes "<RFC3339> <event>\n" to the underlying writer.
// It is a no-op when l is nil.
func (l *eventLogger) logEvent(event string) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, err := fmt.Fprintf(l.w, "%s %s\n", time.Now().UTC().Format(time.RFC3339), event); err != nil {
		log.Printf("analytics log: %v", err)
	}
}
