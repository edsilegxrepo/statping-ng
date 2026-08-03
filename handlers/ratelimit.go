package handlers

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/statping-ng/statping-ng/utils"
)

// rateLimiter implements a simple token bucket rate limiter per IP
type rateLimiter struct {
	sync.RWMutex
	requests   map[string]*bucket
	rate       int           // requests per window
	window     time.Duration // time window
	maxEntries int           // maximum number of tracked IPs (DoS protection)
	stopCh     chan struct{} // channel to stop cleanup goroutine
}

type bucket struct {
	tokens    int
	lastReset time.Time
}

var loginRateLimiter = &rateLimiter{
	requests:   make(map[string]*bucket),
	rate:       5,               // 5 login attempts
	window:     5 * time.Minute, // per 5 minutes
	maxEntries: 100000,          // max 100k tracked IPs to prevent memory exhaustion
	stopCh:     make(chan struct{}),
}

var oauthStateRateLimiter = &rateLimiter{
	requests:   make(map[string]*bucket),
	rate:       20,              // 20 OAuth state requests
	window:     1 * time.Minute, // per minute
	maxEntries: 100000,
	stopCh:     make(chan struct{}),
}

// cleanup removes expired entries periodically
func init() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				loginRateLimiter.cleanup()
				oauthStateRateLimiter.cleanup()
				cleanupExpiredOAuthStates()
			case <-loginRateLimiter.stopCh:
				return
			}
		}
	}()
}

// StopRateLimiterCleanup stops the cleanup goroutine for graceful shutdown
func StopRateLimiterCleanup() {
	close(loginRateLimiter.stopCh)
}

func (rl *rateLimiter) cleanup() {
	rl.Lock()
	defer rl.Unlock()
	now := time.Now()
	for ip, b := range rl.requests {
		if now.Sub(b.lastReset) > rl.window*2 {
			delete(rl.requests, ip)
		}
	}
}

// allow checks if a request from the given IP should be allowed
func (rl *rateLimiter) allow(ip string) bool {
	rl.Lock()
	defer rl.Unlock()

	now := time.Now()
	b, exists := rl.requests[ip]

	if !exists {
		// Check if we've hit the max entries limit (DoS protection)
		if len(rl.requests) >= rl.maxEntries {
			// Evict oldest entries (simple LRU-like eviction)
			rl.evictOldestLocked(rl.maxEntries / 10) // Evict 10%
		}
		rl.requests[ip] = &bucket{
			tokens:    rl.rate - 1,
			lastReset: now,
		}
		return true
	}

	// Reset tokens if window has passed
	if now.Sub(b.lastReset) > rl.window {
		b.tokens = rl.rate - 1
		b.lastReset = now
		return true
	}

	// Check if we have tokens left
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// evictOldestLocked removes the oldest entries (must be called with lock held)
func (rl *rateLimiter) evictOldestLocked(count int) {
	// Find and remove oldest entries
	type entry struct {
		ip   string
		time time.Time
	}
	entries := make([]entry, 0, len(rl.requests))
	for ip, b := range rl.requests {
		entries = append(entries, entry{ip, b.lastReset})
	}

	// Sort by time (oldest first) - simple bubble sort for small eviction batches
	for i := 0; i < count && i < len(entries); i++ {
		minIdx := i
		for j := i + 1; j < len(entries); j++ {
			if entries[j].time.Before(entries[minIdx].time) {
				minIdx = j
			}
		}
		entries[i], entries[minIdx] = entries[minIdx], entries[i]
		delete(rl.requests, entries[i].ip)
	}
}

// remaining returns the number of requests remaining for the IP
func (rl *rateLimiter) remaining(ip string) int {
	rl.RLock()
	defer rl.RUnlock()

	b, exists := rl.requests[ip]
	if !exists {
		return rl.rate
	}

	now := time.Now()
	if now.Sub(b.lastReset) > rl.window {
		return rl.rate
	}

	return b.tokens
}

// resetTime returns when the rate limit will reset for the IP
func (rl *rateLimiter) resetTime(ip string) time.Time {
	rl.RLock()
	defer rl.RUnlock()

	b, exists := rl.requests[ip]
	if !exists {
		return time.Now()
	}

	return b.lastReset.Add(rl.window)
}

// trustedProxies contains IP ranges that are trusted to set X-Forwarded-For
// Set TRUSTED_PROXIES env var to comma-separated CIDR ranges (e.g., "10.0.0.0/8,192.168.0.0/16")
var (
	trustedProxyCIDRs    []*net.IPNet
	trustedProxiesLoaded bool
	trustedProxiesLoadMu sync.Mutex
)

// loadTrustedProxies parses TRUSTED_PROXIES from environment (called lazily)
func loadTrustedProxies() {
	trustedProxiesLoadMu.Lock()
	defer trustedProxiesLoadMu.Unlock()

	if trustedProxiesLoaded {
		return
	}
	trustedProxiesLoaded = true

	if utils.Params == nil {
		return
	}

	proxies := utils.Params.GetString("TRUSTED_PROXIES")
	if proxies == "" {
		return
	}

	for _, cidr := range strings.Split(proxies, ",") {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			// Try as single IP
			ip := net.ParseIP(cidr)
			if ip != nil {
				mask := net.CIDRMask(128, 128)
				if ip.To4() != nil {
					mask = net.CIDRMask(32, 32)
				}
				ipNet = &net.IPNet{IP: ip, Mask: mask}
			} else {
				log.Warnf("Invalid TRUSTED_PROXIES entry: %s", cidr)
				continue
			}
		}
		trustedProxyCIDRs = append(trustedProxyCIDRs, ipNet)
	}
}

// isTrustedProxy checks if an IP is in the trusted proxy list
func isTrustedProxy(ip string) bool {
	loadTrustedProxies() // Lazy load on first use

	if len(trustedProxyCIDRs) == 0 {
		return false // No trusted proxies configured, don't trust headers
	}
	// Strip port if present
	host, _, err := net.SplitHostPort(ip)
	if err != nil {
		host = ip
	}
	parsedIP := net.ParseIP(host)
	if parsedIP == nil {
		return false
	}
	for _, cidr := range trustedProxyCIDRs {
		if cidr.Contains(parsedIP) {
			return true
		}
	}
	return false
}

// getClientIP extracts the client IP from the request
// Only trusts X-Forwarded-For/X-Real-IP if request comes from a trusted proxy
func getClientIP(r *http.Request) string {
	remoteIP := r.RemoteAddr

	// Only trust proxy headers if request comes from a trusted proxy
	if isTrustedProxy(remoteIP) {
		// Check X-Forwarded-For header (for proxied requests)
		xff := r.Header.Get("X-Forwarded-For")
		if xff != "" {
			// Take the first IP in the chain (original client)
			if idx := strings.Index(xff, ","); idx != -1 {
				return strings.TrimSpace(xff[:idx])
			}
			return strings.TrimSpace(xff)
		}

		// Check X-Real-IP header
		xri := r.Header.Get("X-Real-IP")
		if xri != "" {
			return strings.TrimSpace(xri)
		}
	}

	// Fall back to RemoteAddr (strip port)
	host, _, err := net.SplitHostPort(remoteIP)
	if err != nil {
		return remoteIP
	}
	return host
}

// rateLimitLoginMiddleware applies rate limiting to login attempts
func rateLimitLoginMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		if !loginRateLimiter.allow(ip) {
			w.Header().Set("Retry-After", loginRateLimiter.resetTime(ip).Format(time.RFC1123))
			w.Header().Set("X-RateLimit-Remaining", "0")
			log.Warnln("Rate limit exceeded for IP:", ip)
			http.Error(w, "Too many login attempts. Please try again later.", http.StatusTooManyRequests)
			return
		}

		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", loginRateLimiter.remaining(ip)))
		next(w, r)
	}
}

// rateLimitOAuthStateMiddleware applies rate limiting to OAuth state generation
func rateLimitOAuthStateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		if !oauthStateRateLimiter.allow(ip) {
			w.Header().Set("Retry-After", oauthStateRateLimiter.resetTime(ip).Format(time.RFC1123))
			log.Warnln("OAuth state rate limit exceeded for IP:", ip)
			http.Error(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next(w, r)
	}
}
