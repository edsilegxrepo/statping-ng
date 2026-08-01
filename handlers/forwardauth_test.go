package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/stretchr/testify/assert"
)

func TestForwardAuthExtract(t *testing.T) {
	// Setup core with forward auth enabled
	core.App = &core.Core{
		ForwardAuth: core.ForwardAuth{
			ForwardAuthEnabled:        null.NewNullBool(true),
			ForwardAuthHeaderUser:     "Remote-User",
			ForwardAuthHeaderEmail:    "Remote-Email",
			ForwardAuthHeaderGroups:   "Remote-Groups",
			ForwardAuthHeaderName:     "Remote-Name",
			ForwardAuthAdminGroups:    "admins;statping-admins",
			ForwardAuthTrustedProxies: "127.0.0.1/32;10.0.0.0/8",
		},
	}

	tests := []struct {
		name          string
		headers       map[string]string
		remoteAddr    string
		wantUser      bool
		wantUsername  string
		wantEmail     string
		wantGroups    []string
		wantAdmin     bool
	}{
		{
			name: "valid headers from trusted proxy",
			headers: map[string]string{
				"Remote-User":   "john",
				"Remote-Email":  "john@example.com",
				"Remote-Groups": "users,admins",
				"Remote-Name":   "John Doe",
			},
			remoteAddr:   "127.0.0.1:12345",
			wantUser:     true,
			wantUsername: "john",
			wantEmail:    "john@example.com",
			wantGroups:   []string{"users", "admins"},
			wantAdmin:    true,
		},
		{
			name: "valid headers from 10.x network",
			headers: map[string]string{
				"Remote-User":   "jane",
				"Remote-Email":  "jane@example.com",
				"Remote-Groups": "users",
			},
			remoteAddr:   "10.0.0.5:12345",
			wantUser:     true,
			wantUsername: "jane",
			wantEmail:    "jane@example.com",
			wantGroups:   []string{"users"},
			wantAdmin:    false,
		},
		{
			name: "headers from untrusted IP rejected",
			headers: map[string]string{
				"Remote-User": "hacker",
			},
			remoteAddr: "1.2.3.4:12345",
			wantUser:   false,
		},
		{
			name:       "no headers",
			headers:    map[string]string{},
			remoteAddr: "127.0.0.1:12345",
			wantUser:   false,
		},
		{
			name: "admin group match - statping-admins",
			headers: map[string]string{
				"Remote-User":   "admin2",
				"Remote-Groups": "developers,statping-admins",
			},
			remoteAddr:   "127.0.0.1:12345",
			wantUser:     true,
			wantUsername: "admin2",
			wantGroups:   []string{"developers", "statping-admins"},
			wantAdmin:    true,
		},
		{
			name: "no admin group",
			headers: map[string]string{
				"Remote-User":   "regular",
				"Remote-Groups": "users,developers",
			},
			remoteAddr:   "127.0.0.1:12345",
			wantUser:     true,
			wantUsername: "regular",
			wantGroups:   []string{"users", "developers"},
			wantAdmin:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			info := forwardAuthExtract(req)

			if !tt.wantUser {
				assert.Nil(t, info)
				return
			}

			assert.NotNil(t, info)
			assert.Equal(t, tt.wantUsername, info.Username)
			assert.Equal(t, tt.wantEmail, info.Email)
			assert.Equal(t, tt.wantGroups, info.Groups)
			assert.Equal(t, tt.wantAdmin, info.IsAdmin)
		})
	}
}

func TestForwardAuthDisabled(t *testing.T) {
	core.App = &core.Core{
		ForwardAuth: core.ForwardAuth{
			ForwardAuthEnabled: null.NewNullBool(false),
		},
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("Remote-User", "john")

	info := forwardAuthExtract(req)
	assert.Nil(t, info, "should return nil when forward auth is disabled")
}

func TestIsForwardAuthAdmin(t *testing.T) {
	core.App = &core.Core{
		ForwardAuth: core.ForwardAuth{
			ForwardAuthAdminGroups: "admins;super-admins;wheel",
		},
	}

	tests := []struct {
		name      string
		groups    []string
		wantAdmin bool
	}{
		{"single admin group", []string{"admins"}, true},
		{"admin in list", []string{"users", "admins", "developers"}, true},
		{"super-admins match", []string{"super-admins"}, true},
		{"wheel match", []string{"wheel"}, true},
		{"no match", []string{"users", "developers"}, false},
		{"empty groups", []string{}, false},
		{"similar but not exact", []string{"admin", "admins2"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isForwardAuthAdmin(tt.groups)
			assert.Equal(t, tt.wantAdmin, result)
		})
	}
}

func TestIsFromForwardAuthTrustedProxy(t *testing.T) {
	core.App = &core.Core{
		ForwardAuth: core.ForwardAuth{
			ForwardAuthTrustedProxies: "10.0.0.0/8;192.168.1.0/24;172.16.0.1",
		},
	}

	tests := []struct {
		name       string
		remoteAddr string
		want       bool
	}{
		{"10.x range", "10.0.0.5:12345", true},
		{"10.x edge", "10.255.255.255:12345", true},
		{"192.168.1.x range", "192.168.1.100:12345", true},
		{"192.168.2.x out of range", "192.168.2.1:12345", false},
		{"single IP match", "172.16.0.1:12345", true},
		{"single IP no match", "172.16.0.2:12345", false},
		{"external IP", "8.8.8.8:12345", false},
		{"localhost", "127.0.0.1:12345", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			result := isFromForwardAuthTrustedProxy(req)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestForwardAuthNoTrustedProxies(t *testing.T) {
	core.App = &core.Core{
		ForwardAuth: core.ForwardAuth{
			ForwardAuthEnabled:        null.NewNullBool(true),
			ForwardAuthTrustedProxies: "", // Empty = reject all
		},
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("Remote-User", "john")

	info := forwardAuthExtract(req)
	assert.Nil(t, info, "should reject when no trusted proxies configured")
}

func TestForwardAuthSettingsHandler(t *testing.T) {
	ensureHandlerSetup(t)

	core.App.ForwardAuth = core.ForwardAuth{
		ForwardAuthEnabled:        null.NewNullBool(true),
		ForwardAuthHeaderUser:     "X-Auth-User",
		ForwardAuthHeaderEmail:    "X-Auth-Email",
		ForwardAuthHeaderGroups:   "X-Auth-Groups",
		ForwardAuthHeaderName:     "X-Auth-Name",
		ForwardAuthAdminGroups:    "admins",
		ForwardAuthTrustedProxies: "10.0.0.0/8",
		ForwardAuthLogoutURL:      "https://auth.example.com/logout",
	}

	req := httptest.NewRequest("GET", "/api/forwardauth", nil)
	rec := httptest.NewRecorder()

	forwardAuthSettingsHandler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"forward_auth_enabled":true`)
	assert.Contains(t, rec.Body.String(), `"forward_auth_header_user":"X-Auth-User"`)
	assert.Contains(t, rec.Body.String(), `"forward_auth_admin_groups":"admins"`)
}

func TestForwardAuthSaveHandler(t *testing.T) {
	ensureHandlerSetup(t)

	body := `{
		"forward_auth_enabled": true,
		"forward_auth_header_user": "Remote-User",
		"forward_auth_header_email": "Remote-Email",
		"forward_auth_header_groups": "Remote-Groups",
		"forward_auth_header_name": "Remote-Name",
		"forward_auth_admin_groups": "admins;ops",
		"forward_auth_trusted_proxies": "10.0.0.0/8;192.168.0.0/16",
		"forward_auth_logout_url": "https://auth.example.com/logout"
	}`

	req := httptest.NewRequest("POST", "/api/forwardauth", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	forwardAuthSaveHandler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, core.App.ForwardAuthEnabled.Bool)
	assert.Equal(t, "admins;ops", core.App.ForwardAuthAdminGroups)
	assert.Equal(t, "10.0.0.0/8;192.168.0.0/16", core.App.ForwardAuthTrustedProxies)
}

func TestForwardAuthSaveHandlerValidation(t *testing.T) {
	ensureHandlerSetup(t)

	// Test: enabled without trusted proxies should fail
	body := `{
		"forward_auth_enabled": true,
		"forward_auth_trusted_proxies": ""
	}`

	req := httptest.NewRequest("POST", "/api/forwardauth", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	forwardAuthSaveHandler(rec, req)

	assert.Contains(t, rec.Body.String(), "Trusted proxies required")
}

func TestForwardAuthSaveHandlerInvalidCIDR(t *testing.T) {
	ensureHandlerSetup(t)

	body := `{
		"forward_auth_enabled": true,
		"forward_auth_trusted_proxies": "not-a-cidr"
	}`

	req := httptest.NewRequest("POST", "/api/forwardauth", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	forwardAuthSaveHandler(rec, req)

	assert.Contains(t, rec.Body.String(), "Invalid CIDR")
}

func TestCustomHeaderNames(t *testing.T) {
	core.App = &core.Core{
		ForwardAuth: core.ForwardAuth{
			ForwardAuthEnabled:        null.NewNullBool(true),
			ForwardAuthHeaderUser:     "X-Authentik-Username",
			ForwardAuthHeaderEmail:    "X-Authentik-Email",
			ForwardAuthHeaderGroups:   "X-Authentik-Groups",
			ForwardAuthHeaderName:     "X-Authentik-Name",
			ForwardAuthAdminGroups:    "authentik Admins",
			ForwardAuthTrustedProxies: "127.0.0.1/32",
		},
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("X-Authentik-Username", "authuser")
	req.Header.Set("X-Authentik-Email", "auth@example.com")
	req.Header.Set("X-Authentik-Groups", "users,authentik Admins")
	req.Header.Set("X-Authentik-Name", "Auth User")

	info := forwardAuthExtract(req)

	assert.NotNil(t, info)
	assert.Equal(t, "authuser", info.Username)
	assert.Equal(t, "auth@example.com", info.Email)
	assert.Equal(t, "Auth User", info.Name)
	assert.True(t, info.IsAdmin)
}
