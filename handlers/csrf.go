package handlers

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

const (
	csrfTokenHeader = "X-CSRF-Token"  // #nosec G101 - not a credential, just a header name
	csrfTokenCookie = "csrf_token"
	csrfTokenLength = 32
	csrfTokenExpiry = 24 * time.Hour
)

// csrfTokenStore stores valid CSRF tokens with their expiration
var (
	csrfTokenStore     = make(map[string]time.Time)
	csrfTokenStoreLock sync.RWMutex
)

// generateCSRFToken creates a new cryptographically secure CSRF token
func generateCSRFToken() (string, error) {
	b := make([]byte, csrfTokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(b)

	csrfTokenStoreLock.Lock()
	csrfTokenStore[token] = time.Now().Add(csrfTokenExpiry)
	csrfTokenStoreLock.Unlock()

	return token, nil
}

// validateCSRFToken checks if the provided token is valid
func validateCSRFToken(token string) bool {
	if token == "" {
		return false
	}

	csrfTokenStoreLock.RLock()
	expiry, exists := csrfTokenStore[token]
	csrfTokenStoreLock.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(expiry) {
		// Clean up expired token
		csrfTokenStoreLock.Lock()
		delete(csrfTokenStore, token)
		csrfTokenStoreLock.Unlock()
		return false
	}

	return true
}

// cleanupExpiredCSRFTokens removes expired tokens (called periodically)
func cleanupExpiredCSRFTokens() {
	csrfTokenStoreLock.Lock()
	defer csrfTokenStoreLock.Unlock()

	now := time.Now()
	for token, expiry := range csrfTokenStore {
		if now.After(expiry) {
			delete(csrfTokenStore, token)
		}
	}
}

// csrfMiddleware validates CSRF tokens on state-changing requests
func csrfMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF check for safe methods
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			// For GET requests, ensure a CSRF token cookie exists
			if _, err := r.Cookie(csrfTokenCookie); err != nil {
				token, err := generateCSRFToken()
				if err == nil {
					setCSRFCookie(w, r, token)
				}
			}
			next.ServeHTTP(w, r)
			return
		}

		// For state-changing methods, validate the CSRF token
		headerToken := r.Header.Get(csrfTokenHeader)
		cookieToken := ""
		if cookie, err := r.Cookie(csrfTokenCookie); err == nil {
			cookieToken = cookie.Value
		}

		// Both header and cookie must be present and match
		if headerToken == "" || cookieToken == "" {
			log.Warnln("CSRF token missing from request")
			http.Error(w, "CSRF token required", http.StatusForbidden)
			return
		}

		// Constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(headerToken), []byte(cookieToken)) != 1 {
			log.Warnln("CSRF token mismatch")
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		// Validate the token exists in our store
		if !validateCSRFToken(headerToken) {
			log.Warnln("CSRF token not found in store or expired")
			http.Error(w, "Invalid or expired CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// setCSRFCookie sets a CSRF token cookie with the given token value
func setCSRFCookie(w http.ResponseWriter, r *http.Request, token string) {
	secure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"

	// #nosec G124 - HttpOnly=false is intentional for CSRF double-submit cookie pattern
	http.SetCookie(w, &http.Cookie{
		Name:     csrfTokenCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: false, // JavaScript must read this for double-submit pattern
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(csrfTokenExpiry.Seconds()),
	})
}

// csrfTokenHandler returns the current CSRF token (for the frontend)
func csrfTokenHandler(w http.ResponseWriter, r *http.Request) {
	token := ""
	if cookie, err := r.Cookie(csrfTokenCookie); err == nil {
		token = cookie.Value
	} else {
		var err error
		token, err = generateCSRFToken()
		if err != nil {
			sendErrorJson(err, w, r)
			return
		}
		setCSRFCookie(w, r, token)
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"token":"` + token + `"}`))
}

// Start cleanup goroutine
func init() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for range ticker.C {
			cleanupExpiredCSRFTokens()
			cleanupExpiredOAuthStates()
		}
	}()
}
