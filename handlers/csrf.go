package handlers

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/statping-ng/statping-ng/utils"
)

// csrfMiddleware protects against CSRF attacks using Origin/Referer validation.
// Primary defense is SameSite=Strict on auth cookies; this is defense-in-depth.
func csrfMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Safe methods don't need CSRF protection
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Get the expected host
		// Only trust X-Forwarded-Host if request comes from a trusted proxy
		expectedHost := r.Host
		if isTrustedProxy(r.RemoteAddr) {
			if fwdHost := r.Header.Get("X-Forwarded-Host"); fwdHost != "" {
				expectedHost = fwdHost
			}
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
		// Allow if using API authentication (Bearer token or API key) - these are not vulnerable to CSRF
		if hasAuthorizationHeader(r) || hasAPIQuery(r) {
			next.ServeHTTP(w, r)
			return
		}

		// For cookie-only auth without Origin/Referer, block by default for security
		// Set ALLOW_NO_ORIGIN=true to allow (for legacy clients/privacy tools)
		if utils.Params.GetBool("ALLOW_NO_ORIGIN") {
			log.Debugln("CSRF: Request without Origin or Referer header (allowed via ALLOW_NO_ORIGIN)")
			next.ServeHTTP(w, r)
			return
		}

		log.Warnln("CSRF: Blocking request without Origin or Referer header (set ALLOW_NO_ORIGIN=true to allow)")
		http.Error(w, "Missing Origin or Referer header", http.StatusForbidden)
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
