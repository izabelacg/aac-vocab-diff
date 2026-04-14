package server

import (
	"net"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

// ipLimiter holds a per-IP token-bucket rate limiter.
type ipLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	r        rate.Limit
	b        int
}

func newIPLimiter(r rate.Limit, b int) *ipLimiter {
	return &ipLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        r,
		b:        b,
	}
}

// get returns the limiter for the given IP, creating one on first use.
func (l *ipLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	if lim, ok := l.limiters[ip]; ok {
		return lim
	}
	lim := rate.NewLimiter(l.r, l.b)
	l.limiters[ip] = lim
	return lim
}

// realIP returns the client's IP address. When the TCP connection comes from a
// localhost proxy (cloudflared), it trusts the CF-Connecting-IP header first,
// then X-Forwarded-For. This is safe because we only trust those headers when
// the connection is from loopback — a remote client cannot inject them.
func realIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	if ip == "127.0.0.1" || ip == "::1" {
		if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
			return cfIP
		}
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			return strings.TrimSpace(strings.SplitN(fwd, ",", 2)[0])
		}
	}
	return ip
}

// rateLimitMiddleware wraps next and returns 429 if the requesting IP exceeds
// the configured rate.
func rateLimitMiddleware(next http.Handler, lim *ipLimiter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !lim.get(realIP(r)).Allow() {
			http.Error(w, "Too many requests. Please wait a moment before trying again.", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
