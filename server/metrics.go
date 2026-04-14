package server

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

// counters holds in-process event counts. The zero value is ready to use.
// All fields are accessed via sync/atomic (atomic.Int64 embeds the lock-free ops).
type counters struct {
	pageViews atomic.Int64 // GET / → 200
	diffsOK   atomic.Int64 // POST /diff → 200
	diffsErr  atomic.Int64 // POST /diff → 4xx/5xx
}

// serveMetrics writes current counter values in Prometheus text exposition
// format. No auth is required — counts are not sensitive.
func (s *Server) serveMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "page_views %d\n", s.counters.pageViews.Load())
	fmt.Fprintf(w, "diffs_ok   %d\n", s.counters.diffsOK.Load())
	fmt.Fprintf(w, "diffs_err  %d\n", s.counters.diffsErr.Load())
}
