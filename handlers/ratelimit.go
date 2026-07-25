package handlers

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// rateLimiter implements a simple token bucket rate limiter per IP
type rateLimiter struct {
	sync.RWMutex
	requests map[string]*bucket
	rate     int           // requests per window
	window   time.Duration // time window
}

type bucket struct {
	tokens    int
	lastReset time.Time
}

var loginRateLimiter = &rateLimiter{
	requests: make(map[string]*bucket),
	rate:     5,              // 5 login attempts
	window:   5 * time.Minute, // per 5 minutes
}

// cleanup removes expired entries periodically
func init() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		for range ticker.C {
			loginRateLimiter.cleanup()
		}
	}()
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

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxied requests)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in the chain
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
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
