package server

import (
	"sync"
	"testing"
)

func TestCounters_IncrementAndLoad(t *testing.T) {
	var c counters
	c.pageViews.Add(1)
	c.pageViews.Add(1)
	c.diffsOK.Add(1)
	c.diffsErr.Add(1)
	c.diffsErr.Add(1)

	if got := c.pageViews.Load(); got != 2 {
		t.Errorf("pageViews: got %d, want 2", got)
	}
	if got := c.diffsOK.Load(); got != 1 {
		t.Errorf("diffsOK: got %d, want 1", got)
	}
	if got := c.diffsErr.Load(); got != 2 {
		t.Errorf("diffsErr: got %d, want 2", got)
	}
}

// TestCounters_ConcurrentSafety races many goroutines against all three
// counters simultaneously. Run with -race to detect data races.
func TestCounters_ConcurrentSafety(t *testing.T) {
	var c counters
	const goroutines = 100

	var wg sync.WaitGroup
	wg.Add(goroutines * 3)
	for range goroutines {
		go func() { defer wg.Done(); c.pageViews.Add(1) }()
		go func() { defer wg.Done(); c.diffsOK.Add(1) }()
		go func() { defer wg.Done(); c.diffsErr.Add(1) }()
	}
	wg.Wait()

	if got := c.pageViews.Load(); got != goroutines {
		t.Errorf("pageViews: got %d, want %d", got, goroutines)
	}
	if got := c.diffsOK.Load(); got != goroutines {
		t.Errorf("diffsOK: got %d, want %d", got, goroutines)
	}
	if got := c.diffsErr.Load(); got != goroutines {
		t.Errorf("diffsErr: got %d, want %d", got, goroutines)
	}
}
