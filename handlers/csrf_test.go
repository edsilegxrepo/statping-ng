package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCSRFMiddlewareAllowsGET(t *testing.T) {
	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "GET request should be allowed")
}

func TestCSRFMiddlewareAllowsHEAD(t *testing.T) {
	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("HEAD", "/api/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "HEAD request should be allowed")
}

func TestCSRFMiddlewareAllowsOPTIONS(t *testing.T) {
	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/api/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "OPTIONS request should be allowed")
}

func TestCSRFMiddlewareAllowsPOSTWithValidOrigin(t *testing.T) {
	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/test", nil)
	req.Host = "localhost:8080"
	req.Header.Set("Origin", "http://localhost:8080")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "POST with matching Origin should be allowed")
}

func TestCSRFMiddlewareBlocksPOSTWithInvalidOrigin(t *testing.T) {
	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/test", nil)
	req.Host = "localhost:8080"
	req.Header.Set("Origin", "http://evil.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code, "POST with mismatched Origin should be blocked")
}

func TestCSRFMiddlewareAllowsPOSTWithValidReferer(t *testing.T) {
	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/test", nil)
	req.Host = "localhost:8080"
	req.Header.Set("Referer", "http://localhost:8080/dashboard")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "POST with matching Referer should be allowed")
}

func TestCSRFMiddlewareBlocksPOSTWithInvalidReferer(t *testing.T) {
	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/test", nil)
	req.Host = "localhost:8080"
	req.Header.Set("Referer", "http://evil.com/attack")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code, "POST with mismatched Referer should be blocked")
}

func TestCSRFMiddlewareAllowsPOSTWithoutOriginOrReferer(t *testing.T) {
	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/test", nil)
	req.Host = "localhost:8080"
	// No Origin or Referer - allowed because SameSite=Strict is primary defense
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "POST without Origin/Referer should be allowed (SameSite is primary defense)")
}

func TestCSRFMiddlewareUsesXForwardedHost(t *testing.T) {
	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/test", nil)
	req.Host = "127.0.0.1:8080" // Internal host
	req.Header.Set("X-Forwarded-Host", "statping.example.com")
	req.Header.Set("Origin", "https://statping.example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Should use X-Forwarded-Host for validation")
}

func TestCSRFMiddlewareBlocksWithXForwardedHostMismatch(t *testing.T) {
	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/test", nil)
	req.Host = "127.0.0.1:8080"
	req.Header.Set("X-Forwarded-Host", "statping.example.com")
	req.Header.Set("Origin", "https://evil.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code, "Should block when Origin doesn't match X-Forwarded-Host")
}

func TestIsValidOrigin(t *testing.T) {
	assert.True(t, isValidOrigin("http://localhost:8080", "localhost:8080"))
	assert.True(t, isValidOrigin("https://example.com", "example.com"))
	assert.True(t, isValidOrigin("https://Example.Com", "example.com")) // case insensitive
	assert.False(t, isValidOrigin("http://evil.com", "localhost:8080"))
	assert.False(t, isValidOrigin("not-a-url", "localhost:8080"))
}

func TestIsValidReferer(t *testing.T) {
	assert.True(t, isValidReferer("http://localhost:8080/dashboard", "localhost:8080"))
	assert.True(t, isValidReferer("https://example.com/some/path", "example.com"))
	assert.False(t, isValidReferer("http://evil.com/attack", "localhost:8080"))
	assert.False(t, isValidReferer("not-a-url", "localhost:8080"))
}
