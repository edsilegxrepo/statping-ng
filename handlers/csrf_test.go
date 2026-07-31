package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/statping-ng/statping-ng/utils"
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

func TestCSRFMiddlewareBlocksPOSTWithoutOriginOrReferer(t *testing.T) {
	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/test", nil)
	req.Host = "localhost:8080"
	// No Origin or Referer - blocked by default for security
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code, "POST without Origin/Referer should be blocked by default")
}

func TestCSRFMiddlewareAllowsPOSTWithoutOriginOrRefererWhenConfigured(t *testing.T) {
	// Set ALLOW_NO_ORIGIN for this test
	utils.Params.Set("ALLOW_NO_ORIGIN", true)
	defer utils.Params.Set("ALLOW_NO_ORIGIN", false)

	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/test", nil)
	req.Host = "localhost:8080"
	// No Origin or Referer - allowed when ALLOW_NO_ORIGIN is set
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "POST without Origin/Referer should be allowed when ALLOW_NO_ORIGIN is set")
}

func TestCSRFMiddlewareUsesXForwardedHost(t *testing.T) {
	// Configure trusted proxy so X-Forwarded-Host is trusted
	// Reset the trustedProxiesLoaded flag to force reload
	trustedProxiesLoadMu.Lock()
	trustedProxiesLoaded = false
	trustedProxyCIDRs = nil
	trustedProxiesLoadMu.Unlock()

	utils.Params.Set("TRUSTED_PROXIES", "192.0.2.1/32")
	defer func() {
		utils.Params.Set("TRUSTED_PROXIES", "")
		trustedProxiesLoadMu.Lock()
		trustedProxiesLoaded = false
		trustedProxyCIDRs = nil
		trustedProxiesLoadMu.Unlock()
	}()

	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/test", nil)
	req.RemoteAddr = "192.0.2.1:12345" // From trusted proxy
	req.Host = "127.0.0.1:8080" // Internal host
	req.Header.Set("X-Forwarded-Host", "statping.example.com")
	req.Header.Set("Origin", "https://statping.example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Should use X-Forwarded-Host for validation when from trusted proxy")
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
