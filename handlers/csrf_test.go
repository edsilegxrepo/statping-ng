package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCSRFToken(t *testing.T) {
	token1, err := generateCSRFToken()
	assert.Nil(t, err, "Should generate CSRF token without error")
	assert.NotEmpty(t, token1, "Generated token should not be empty")

	token2, err := generateCSRFToken()
	assert.Nil(t, err, "Should generate second CSRF token without error")
	assert.NotEqual(t, token1, token2, "Generated tokens should be unique")
}

func TestValidateCSRFToken(t *testing.T) {
	token, err := generateCSRFToken()
	assert.Nil(t, err)

	// Token should be valid
	assert.True(t, validateCSRFToken(token), "Valid token should validate")

	// Token should still be valid (not consumed like OAuth state)
	assert.True(t, validateCSRFToken(token), "Token should remain valid")
}

func TestValidateCSRFTokenEmpty(t *testing.T) {
	assert.False(t, validateCSRFToken(""), "Empty token should not validate")
}

func TestValidateCSRFTokenInvalid(t *testing.T) {
	assert.False(t, validateCSRFToken("invalid-csrf-token"), "Invalid token should not validate")
}

func TestValidateCSRFTokenExpired(t *testing.T) {
	// Manually add an expired token
	expiredToken := "expired-csrf-token-" + time.Now().String()
	csrfTokenStoreLock.Lock()
	csrfTokenStore[expiredToken] = time.Now().Add(-1 * time.Hour)
	csrfTokenStoreLock.Unlock()

	assert.False(t, validateCSRFToken(expiredToken), "Expired token should not validate")

	// Token should be cleaned up after failed validation
	csrfTokenStoreLock.RLock()
	_, exists := csrfTokenStore[expiredToken]
	csrfTokenStoreLock.RUnlock()
	assert.False(t, exists, "Expired token should be removed from store")
}

func TestCleanupExpiredCSRFTokens(t *testing.T) {
	prefix := time.Now().UnixNano()
	expired1 := "csrf-expired1-" + string(rune(prefix))
	expired2 := "csrf-expired2-" + string(rune(prefix))
	valid := "csrf-valid-" + string(rune(prefix))

	csrfTokenStoreLock.Lock()
	csrfTokenStore[expired1] = time.Now().Add(-1 * time.Hour)
	csrfTokenStore[expired2] = time.Now().Add(-2 * time.Hour)
	csrfTokenStore[valid] = time.Now().Add(1 * time.Hour)
	csrfTokenStoreLock.Unlock()

	cleanupExpiredCSRFTokens()

	csrfTokenStoreLock.RLock()
	_, exists1 := csrfTokenStore[expired1]
	_, exists2 := csrfTokenStore[expired2]
	_, existsValid := csrfTokenStore[valid]
	csrfTokenStoreLock.RUnlock()

	assert.False(t, exists1, "expired1 should have been cleaned up")
	assert.False(t, exists2, "expired2 should have been cleaned up")
	assert.True(t, existsValid, "valid token should still exist")

	// Cleanup test token
	csrfTokenStoreLock.Lock()
	delete(csrfTokenStore, valid)
	csrfTokenStoreLock.Unlock()
}

func TestCSRFMiddlewareAllowsGET(t *testing.T) {
	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "GET request should be allowed without CSRF token")
}

func TestCSRFMiddlewareBlocksPOSTWithoutToken(t *testing.T) {
	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code, "POST request without CSRF token should be blocked")
}

func TestCSRFMiddlewareAllowsPOSTWithValidToken(t *testing.T) {
	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Generate a valid token
	token, err := generateCSRFToken()
	assert.Nil(t, err)

	req := httptest.NewRequest("POST", "/api/test", nil)
	req.Header.Set(csrfTokenHeader, token)
	req.AddCookie(&http.Cookie{Name: csrfTokenCookie, Value: token})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "POST request with valid CSRF token should be allowed")
}

func TestCSRFMiddlewareBlocksMismatchedToken(t *testing.T) {
	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Generate two different tokens
	token1, _ := generateCSRFToken()
	token2, _ := generateCSRFToken()

	req := httptest.NewRequest("POST", "/api/test", nil)
	req.Header.Set(csrfTokenHeader, token1)
	req.AddCookie(&http.Cookie{Name: csrfTokenCookie, Value: token2})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code, "POST request with mismatched CSRF tokens should be blocked")
}

func TestCSRFTokenHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/csrf", nil)
	rr := httptest.NewRecorder()
	csrfTokenHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "token", "Response should contain token")

	// Check that cookie was set
	cookies := rr.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == csrfTokenCookie {
			found = true
			assert.NotEmpty(t, c.Value, "CSRF cookie should have a value")
		}
	}
	assert.True(t, found, "CSRF cookie should be set")
}
