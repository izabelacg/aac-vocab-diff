// Package version holds build metadata injected at link time (-ldflags -X).
package version

// Default strings match an unstamped local `go build` (no -ldflags).
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)
