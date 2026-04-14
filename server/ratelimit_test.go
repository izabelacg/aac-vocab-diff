package server

// Internal tests for ipLimiter and realIP — lives in package server (not server_test)
// so it can access the unexported types directly.

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestIPRateLimiter_AllowsUpToBurst(t *testing.T) {
	lim := newIPLimiter(rate.Every(time.Hour), 3) // refills once per hour — effectively never
	ip := "1.2.3.4"
	for i := range 3 {
		if !lim.get(ip).Allow() {
			t.Fatalf("request %d should be allowed (burst=3)", i+1)
		}
	}
}

func TestIPRateLimiter_BlocksAfterBurst(t *testing.T) {
	lim := newIPLimiter(rate.Every(time.Hour), 3)
	ip := "1.2.3.4"
	for range 3 {
		lim.get(ip).Allow() // exhaust burst
	}
	if lim.get(ip).Allow() {
		t.Error("4th request should be blocked after burst exhausted")
	}
}

func TestIPRateLimiter_IsolatesPerIP(t *testing.T) {
	lim := newIPLimiter(rate.Every(time.Hour), 1)
	lim.get("1.2.3.4").Allow() // exhaust first IP
	if !lim.get("5.6.7.8").Allow() {
		t.Error("different IP should have its own independent limiter")
	}
}

func TestRealIP(t *testing.T) {
	cases := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       string
	}{
		{
			name:       "direct connection uses RemoteAddr",
			remoteAddr: "1.2.3.4:5000",
			want:       "1.2.3.4",
		},
		{
			name:       "cloudflared with CF-Connecting-IP",
			remoteAddr: "127.0.0.1:12345",
			headers:    map[string]string{"CF-Connecting-IP": "9.8.7.6"},
			want:       "9.8.7.6",
		},
		{
			name:       "cloudflared with X-Forwarded-For uses first entry",
			remoteAddr: "127.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "9.8.7.6, 10.0.0.1"},
			want:       "9.8.7.6",
		},
		{
			name:       "cloudflared with no forwarded header falls back to loopback",
			remoteAddr: "127.0.0.1:12345",
			want:       "127.0.0.1",
		},
		{
			name:       "IPv6 loopback with CF-Connecting-IP",
			remoteAddr: "[::1]:12345",
			headers:    map[string]string{"CF-Connecting-IP": "2001:db8::1"},
			want:       "2001:db8::1",
		},
		{
			name:       "non-loopback remote ignores CF-Connecting-IP",
			remoteAddr: "5.6.7.8:9000",
			headers:    map[string]string{"CF-Connecting-IP": "9.8.7.6"},
			want:       "5.6.7.8",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tc.remoteAddr
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			got := realIP(req)
			if got != tc.want {
				t.Errorf("realIP() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestRateLimitMiddleware_CloudflaredIsolatesPerIP verifies that two clients
// arriving through cloudflared (both with RemoteAddr 127.0.0.1) get separate
// rate-limit buckets based on their real IPs.
func TestRateLimitMiddleware_CloudflaredIsolatesPerIP(t *testing.T) {
	lim := newIPLimiter(rate.Every(time.Hour), 1) // burst of 1 — exhausted on first request

	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := rateLimitMiddleware(ok, lim)

	makeReq := func(cfIP string) *http.Request {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("CF-Connecting-IP", cfIP)
		return req
	}

	// First client exhausts their own bucket.
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, makeReq("1.1.1.1"))
	if rr.Code != http.StatusOK {
		t.Fatalf("first request for 1.1.1.1: got %d, want 200", rr.Code)
	}

	// First client is now rate-limited.
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, makeReq("1.1.1.1"))
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("second request for 1.1.1.1: got %d, want 429", rr.Code)
	}

	// Second client still has a full bucket — must not be affected.
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, makeReq("2.2.2.2"))
	if rr.Code != http.StatusOK {
		t.Errorf("first request for 2.2.2.2: got %d, want 200 (should have its own bucket)", rr.Code)
	}
}
