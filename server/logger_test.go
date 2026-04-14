package server

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

func TestEventLogger_WritesTimestampedLine(t *testing.T) {
	var buf bytes.Buffer
	l := &eventLogger{w: &buf}
	l.logEvent("diff_ok")

	got := buf.String()
	if !strings.Contains(got, "diff_ok") {
		t.Errorf("log line missing event name, got: %q", got)
	}
	// RFC3339 timestamps start with the year and contain a 'T' separator.
	if !strings.Contains(got, "T") || !strings.HasSuffix(strings.TrimSpace(got), "diff_ok") {
		t.Errorf("log line does not look like '<timestamp> diff_ok', got: %q", got)
	}
}

func TestEventLogger_NilIsNoop(t *testing.T) {
	var l *eventLogger
	// Must not panic.
	l.logEvent("page_view")
}

func TestEventLogger_ConcurrentSafety(t *testing.T) {
	var buf syncBuffer
	l := &eventLogger{w: &buf}
	const goroutines = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			l.logEvent("diff_ok")
		}()
	}
	wg.Wait()

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != goroutines {
		t.Errorf("expected %d log lines, got %d", goroutines, len(lines))
	}
}

// syncBuffer is a bytes.Buffer safe for concurrent use (the test races many
// goroutines against the same writer; the eventLogger's mu protects it but
// the race detector checks the buffer itself too).
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}
