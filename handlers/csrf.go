package handlers

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

// csrfMiddleware protects against CSRF attacks using Origin/Referer validation.
// Primary defense is SameSite=Strict on auth cookies; this is defense-in-depth.
// Since the app runs behind a reverse proxy on 127.0.0.1, we trust X-Forwarded-* headers.
func csrfMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Safe methods don't need CSRF protection
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Get the expected host (trust X-Forwarded-Host from reverse proxy)
		expectedHost := r.Header.Get("X-Forwarded-Host")
		if expectedHost == "" {
			expectedHost = r.Host
		}

		// Check Origin header first (most reliable)
		origin := r.Header.Get("Origin")
		if origin != "" {
			if !isValidOrigin(origin, expectedHost) {
				log.Warnf("CSRF: Origin mismatch - got %s, expected host %s", origin, expectedHost)
				http.Error(w, "Invalid origin", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		// Fallback to Referer header
		referer := r.Header.Get("Referer")
		if referer != "" {
			if !isValidReferer(referer, expectedHost) {
				log.Warnf("CSRF: Referer mismatch - got %s, expected host %s", referer, expectedHost)
				http.Error(w, "Invalid referer", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		// Neither Origin nor Referer present
		// This can happen with some privacy tools, old browsers, or direct API calls
		// SameSite=Strict on auth cookie is our primary defense, so we allow this
		// but log it for monitoring
		log.Debugln("CSRF: Request without Origin or Referer header (allowed, relying on SameSite cookie)")
		next.ServeHTTP(w, r)
	})
}

// isValidOrigin checks if the Origin header matches the expected host
func isValidOrigin(origin, expectedHost string) bool {
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return strings.EqualFold(parsed.Host, expectedHost)
}

// isValidReferer checks if the Referer header matches the expected host
func isValidReferer(referer, expectedHost string) bool {
	parsed, err := url.Parse(referer)
	if err != nil {
		return false
	}
	return strings.EqualFold(parsed.Host, expectedHost)
}

// Start cleanup goroutine for OAuth states (CSRF tokens no longer used)
func init() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for range ticker.C {
			cleanupExpiredOAuthStates()
		}
	}()
}
