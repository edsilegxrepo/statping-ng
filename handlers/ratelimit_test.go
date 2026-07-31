package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiterAllow(t *testing.T) {
	rl := &rateLimiter{
		requests: make(map[string]*bucket),
		rate:     3,
		window:   1 * time.Minute,
	}

	ip := "192.168.1.100"

	// First 3 requests should be allowed
	assert.True(t, rl.allow(ip), "First request should be allowed")
	assert.True(t, rl.allow(ip), "Second request should be allowed")
	assert.True(t, rl.allow(ip), "Third request should be allowed")

	// Fourth request should be blocked
	assert.False(t, rl.allow(ip), "Fourth request should be blocked")
}

func TestRateLimiterDifferentIPs(t *testing.T) {
	rl := &rateLimiter{
		requests: make(map[string]*bucket),
		rate:     2,
		window:   1 * time.Minute,
	}

	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	// Both IPs should have their own limits
	assert.True(t, rl.allow(ip1), "IP1 first request should be allowed")
	assert.True(t, rl.allow(ip2), "IP2 first request should be allowed")
	assert.True(t, rl.allow(ip1), "IP1 second request should be allowed")
	assert.True(t, rl.allow(ip2), "IP2 second request should be allowed")
	assert.False(t, rl.allow(ip1), "IP1 third request should be blocked")
	assert.False(t, rl.allow(ip2), "IP2 third request should be blocked")
}

func TestRateLimiterWindowReset(t *testing.T) {
	rl := &rateLimiter{
		requests: make(map[string]*bucket),
		rate:     2,
		window:   100 * time.Millisecond,
	}

	ip := "192.168.1.50"

	// Use up the rate limit
	assert.True(t, rl.allow(ip))
	assert.True(t, rl.allow(ip))
	assert.False(t, rl.allow(ip))

	// Wait for window to reset
	time.Sleep(150 * time.Millisecond)

	// Should be allowed again
	assert.True(t, rl.allow(ip), "Request should be allowed after window reset")
}

func TestRateLimiterRemaining(t *testing.T) {
	rl := &rateLimiter{
		requests: make(map[string]*bucket),
		rate:     5,
		window:   1 * time.Minute,
	}

	ip := "192.168.1.200"

	assert.Equal(t, 5, rl.remaining(ip), "Should start with full rate")

	rl.allow(ip)
	assert.Equal(t, 4, rl.remaining(ip), "Should have 4 remaining after 1 request")

	rl.allow(ip)
	rl.allow(ip)
	assert.Equal(t, 2, rl.remaining(ip), "Should have 2 remaining after 3 requests")
}

func TestRateLimiterCleanup(t *testing.T) {
	rl := &rateLimiter{
		requests: make(map[string]*bucket),
		rate:     5,
		window:   50 * time.Millisecond,
	}

	ip1 := "192.168.1.10"
	ip2 := "192.168.1.20"

	rl.allow(ip1)
	rl.allow(ip2)

	// Wait for entries to expire
	time.Sleep(150 * time.Millisecond)

	rl.cleanup()

	rl.RLock()
	_, exists1 := rl.requests[ip1]
	_, exists2 := rl.requests[ip2]
	rl.RUnlock()

	assert.False(t, exists1, "Expired entry should be cleaned up")
	assert.False(t, exists2, "Expired entry should be cleaned up")
}

func TestGetClientIP(t *testing.T) {
	// Configure trusted proxy so X-Forwarded-For/X-Real-IP are trusted
	trustedProxiesLoadMu.Lock()
	trustedProxiesLoaded = false
	trustedProxyCIDRs = nil
	trustedProxiesLoadMu.Unlock()

	utils.Params.Set("TRUSTED_PROXIES", "127.0.0.1/32")
	defer func() {
		utils.Params.Set("TRUSTED_PROXIES", "")
		trustedProxiesLoadMu.Lock()
		trustedProxiesLoaded = false
		trustedProxyCIDRs = nil
		trustedProxiesLoadMu.Unlock()
	}()

	tests := []struct {
		name     string
		headers  map[string]string
		remoteAddr string
		expected string
	}{
		{
			name:       "X-Forwarded-For single IP from trusted proxy",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.195"},
			remoteAddr: "127.0.0.1:8080",
			expected:   "203.0.113.195",
		},
		{
			name:       "X-Forwarded-For multiple IPs from trusted proxy",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.195, 70.41.3.18, 150.172.238.178"},
			remoteAddr: "127.0.0.1:8080",
			expected:   "203.0.113.195",
		},
		{
			name:       "X-Real-IP from trusted proxy",
			headers:    map[string]string{"X-Real-IP": "198.51.100.178"},
			remoteAddr: "127.0.0.1:8080",
			expected:   "198.51.100.178",
		},
		{
			name:       "No proxy headers",
			headers:    map[string]string{},
			remoteAddr: "192.0.2.1:12345",
			expected:   "192.0.2.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			result := getClientIP(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetClientIPIgnoresUntrustedProxy(t *testing.T) {
	// Reset trusted proxies (no trusted proxies configured)
	trustedProxiesLoadMu.Lock()
	trustedProxiesLoaded = false
	trustedProxyCIDRs = nil
	trustedProxiesLoadMu.Unlock()

	utils.Params.Set("TRUSTED_PROXIES", "")
	defer func() {
		trustedProxiesLoadMu.Lock()
		trustedProxiesLoaded = false
		trustedProxyCIDRs = nil
		trustedProxiesLoadMu.Unlock()
	}()

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.0.2.1:8080"
	req.Header.Set("X-Forwarded-For", "203.0.113.195") // Should be ignored

	result := getClientIP(req)
	assert.Equal(t, "192.0.2.1", result, "Should use RemoteAddr when no trusted proxies configured")
}

func TestRateLimitMiddleware(t *testing.T) {
	// Reset the login rate limiter for testing
	loginRateLimiter = &rateLimiter{
		requests: make(map[string]*bucket),
		rate:     3,
		window:   1 * time.Minute,
	}

	handler := rateLimitLoginMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// First 3 requests should succeed
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("POST", "/api/login", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		rr := httptest.NewRecorder()
		handler(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code, "Request %d should succeed", i+1)
	}

	// Fourth request should be rate limited
	req := httptest.NewRequest("POST", "/api/login", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rr := httptest.NewRecorder()
	handler(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code, "Fourth request should be rate limited")
}
