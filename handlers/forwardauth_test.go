package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/types/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// ============================================================================
// IPv6 Support Tests
// ============================================================================

func TestForwardAuthIPv6TrustedProxies(t *testing.T) {
	core.App = &core.Core{
		ForwardAuth: core.ForwardAuth{
			ForwardAuthEnabled:        null.NewNullBool(true),
			ForwardAuthHeaderUser:     "Remote-User",
			ForwardAuthTrustedProxies: "::1/128;fd00::/8;2001:db8::/32",
		},
	}

	tests := []struct {
		name       string
		remoteAddr string
		want       bool
	}{
		{"IPv6 localhost", "[::1]:12345", true},
		{"IPv6 private fd00::", "[fd00::1]:12345", true},
		{"IPv6 private fd00::abc", "[fd00::abc:def]:12345", true},
		{"IPv6 documentation prefix", "[2001:db8::1]:12345", true},
		{"IPv6 outside range", "[2001:db9::1]:12345", false},
		{"IPv6 public", "[2607:f8b0:4004:800::200e]:12345", false},
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

func TestForwardAuthMixedIPv4IPv6(t *testing.T) {
	core.App = &core.Core{
		ForwardAuth: core.ForwardAuth{
			ForwardAuthEnabled:        null.NewNullBool(true),
			ForwardAuthHeaderUser:     "Remote-User",
			ForwardAuthTrustedProxies: "10.0.0.0/8;::1/128;192.168.0.0/16",
		},
	}

	tests := []struct {
		name       string
		remoteAddr string
		want       bool
	}{
		{"IPv4 in range", "10.1.2.3:12345", true},
		{"IPv4 192.168", "192.168.1.1:12345", true},
		{"IPv6 localhost", "[::1]:12345", true},
		{"IPv4 out of range", "172.16.0.1:12345", false},
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

// ============================================================================
// User Provisioning Tests
// ============================================================================

func TestForwardAuthUserProvisioning(t *testing.T) {
	// Use ensureHandlerSetup to get proper database state
	ensureHandlerSetup(t)

	// Save original ForwardAuth settings
	origForwardAuth := core.App.ForwardAuth

	// Configure forward auth for this test
	core.App.ForwardAuth = core.ForwardAuth{
		ForwardAuthEnabled:        null.NewNullBool(true),
		ForwardAuthHeaderUser:     "Remote-User",
		ForwardAuthHeaderEmail:    "Remote-Email",
		ForwardAuthHeaderGroups:   "Remote-Groups",
		ForwardAuthAdminGroups:    "admins",
		ForwardAuthTrustedProxies: "127.0.0.1/32",
	}

	// Restore original settings after test
	t.Cleanup(func() {
		core.App.ForwardAuth = origForwardAuth
	})

	t.Run("creates new user on first login", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("Remote-User", "newuser1")
		req.Header.Set("Remote-Email", "newuser1@example.com")
		req.Header.Set("Remote-Groups", "users")

		user := forwardAuthUser(req)

		require.NotNil(t, user)
		assert.Equal(t, "newuser1", user.Username)
		assert.Equal(t, "newuser1@example.com", user.Email)
		assert.False(t, user.Admin.Bool)

		// Verify user was persisted
		found, err := users.FindByUsername("newuser1")
		require.NoError(t, err)
		assert.Equal(t, "newuser1@example.com", found.Email)
	})

	t.Run("creates admin user when in admin group", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("Remote-User", "adminuser1")
		req.Header.Set("Remote-Email", "admin@example.com")
		req.Header.Set("Remote-Groups", "users,admins")

		user := forwardAuthUser(req)

		require.NotNil(t, user)
		assert.Equal(t, "adminuser1", user.Username)
		assert.True(t, user.Admin.Bool)
	})

	t.Run("returns existing user on subsequent login", func(t *testing.T) {
		// First login
		req1 := httptest.NewRequest("GET", "/", nil)
		req1.RemoteAddr = "127.0.0.1:12345"
		req1.Header.Set("Remote-User", "existinguser")
		req1.Header.Set("Remote-Email", "existing@example.com")
		req1.Header.Set("Remote-Groups", "users")

		user1 := forwardAuthUser(req1)
		require.NotNil(t, user1)
		firstId := user1.Id

		// Second login
		req2 := httptest.NewRequest("GET", "/", nil)
		req2.RemoteAddr = "127.0.0.1:12345"
		req2.Header.Set("Remote-User", "existinguser")
		req2.Header.Set("Remote-Email", "existing@example.com")
		req2.Header.Set("Remote-Groups", "users")

		user2 := forwardAuthUser(req2)
		require.NotNil(t, user2)
		assert.Equal(t, firstId, user2.Id, "should return same user")
	})

	t.Run("updates admin status when groups change", func(t *testing.T) {
		// Create non-admin user
		req1 := httptest.NewRequest("GET", "/", nil)
		req1.RemoteAddr = "127.0.0.1:12345"
		req1.Header.Set("Remote-User", "promoteduser")
		req1.Header.Set("Remote-Groups", "users")

		user1 := forwardAuthUser(req1)
		require.NotNil(t, user1)
		assert.False(t, user1.Admin.Bool)

		// Login again with admin group
		req2 := httptest.NewRequest("GET", "/", nil)
		req2.RemoteAddr = "127.0.0.1:12345"
		req2.Header.Set("Remote-User", "promoteduser")
		req2.Header.Set("Remote-Groups", "users,admins")

		user2 := forwardAuthUser(req2)
		require.NotNil(t, user2)
		assert.True(t, user2.Admin.Bool, "admin status should be updated")
	})

	t.Run("updates email when changed", func(t *testing.T) {
		// Create user
		req1 := httptest.NewRequest("GET", "/", nil)
		req1.RemoteAddr = "127.0.0.1:12345"
		req1.Header.Set("Remote-User", "emailchangeuser")
		req1.Header.Set("Remote-Email", "old@example.com")

		user1 := forwardAuthUser(req1)
		require.NotNil(t, user1)
		assert.Equal(t, "old@example.com", user1.Email)

		// Login with new email
		req2 := httptest.NewRequest("GET", "/", nil)
		req2.RemoteAddr = "127.0.0.1:12345"
		req2.Header.Set("Remote-User", "emailchangeuser")
		req2.Header.Set("Remote-Email", "new@example.com")

		user2 := forwardAuthUser(req2)
		require.NotNil(t, user2)
		assert.Equal(t, "new@example.com", user2.Email)
	})
}

// ============================================================================
// Edge Cases Tests
// ============================================================================

func TestForwardAuthEdgeCases(t *testing.T) {
	core.App = &core.Core{
		ForwardAuth: core.ForwardAuth{
			ForwardAuthEnabled:        null.NewNullBool(true),
			ForwardAuthHeaderUser:     "Remote-User",
			ForwardAuthHeaderEmail:    "Remote-Email",
			ForwardAuthHeaderGroups:   "Remote-Groups",
			ForwardAuthAdminGroups:    "admins",
			ForwardAuthTrustedProxies: "127.0.0.1/32",
		},
	}

	t.Run("empty username header rejected", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("Remote-User", "")

		info := forwardAuthExtract(req)
		assert.Nil(t, info)
	})

	t.Run("whitespace in groups handled", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("Remote-User", "user1")
		req.Header.Set("Remote-Groups", "  users  ,  admins  ,  developers  ")

		info := forwardAuthExtract(req)
		require.NotNil(t, info)
		assert.Equal(t, []string{"users", "admins", "developers"}, info.Groups)
		assert.True(t, info.IsAdmin)
	})

	t.Run("empty groups string", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("Remote-User", "user2")
		req.Header.Set("Remote-Groups", "")

		info := forwardAuthExtract(req)
		require.NotNil(t, info)
		assert.Nil(t, info.Groups)
		assert.False(t, info.IsAdmin)
	})

	t.Run("groups with only commas", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("Remote-User", "user3")
		req.Header.Set("Remote-Groups", ",,,")

		info := forwardAuthExtract(req)
		require.NotNil(t, info)
		assert.Nil(t, info.Groups)
	})

	t.Run("single group no comma", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("Remote-User", "user4")
		req.Header.Set("Remote-Groups", "admins")

		info := forwardAuthExtract(req)
		require.NotNil(t, info)
		assert.Equal(t, []string{"admins"}, info.Groups)
		assert.True(t, info.IsAdmin)
	})

	t.Run("case sensitive group matching", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("Remote-User", "user5")
		req.Header.Set("Remote-Groups", "Admins,ADMINS")

		info := forwardAuthExtract(req)
		require.NotNil(t, info)
		assert.False(t, info.IsAdmin, "group matching should be case sensitive")
	})
}

func TestForwardAuthMalformedCIDRs(t *testing.T) {
	tests := []struct {
		name   string
		cidrs  string
		valid  bool
	}{
		{"valid single CIDR", "10.0.0.0/8", true},
		{"valid multiple CIDRs", "10.0.0.0/8;192.168.0.0/16", true},
		{"valid with single IP", "10.0.0.1", true},
		{"empty string", "", false},
		{"only semicolons", ";;;", false},
		{"invalid CIDR notation", "10.0.0.0/33", false},
		{"garbage", "not-an-ip", false},
		{"partial IP", "10.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core.App = &core.Core{
				ForwardAuth: core.ForwardAuth{
					ForwardAuthEnabled:        null.NewNullBool(true),
					ForwardAuthTrustedProxies: tt.cidrs,
				},
			}

			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = "10.0.0.1:12345"
			result := isFromForwardAuthTrustedProxy(req)

			if tt.valid && tt.cidrs != "" {
				assert.True(t, result, "should accept valid CIDR: %s", tt.cidrs)
			} else {
				assert.False(t, result, "should reject invalid CIDR config: %s", tt.cidrs)
			}
		})
	}
}

// ============================================================================
// Integration with Auth Middleware Tests
// ============================================================================

func TestForwardAuthIntegrationWithIsFullAuthenticated(t *testing.T) {
	ensureHandlerSetup(t)

	core.App.ForwardAuth = core.ForwardAuth{
		ForwardAuthEnabled:        null.NewNullBool(true),
		ForwardAuthHeaderUser:     "Remote-User",
		ForwardAuthHeaderGroups:   "Remote-Groups",
		ForwardAuthAdminGroups:    "admins",
		ForwardAuthTrustedProxies: "127.0.0.1/32",
	}

	t.Run("admin user via forward auth is fully authenticated", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/users", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("Remote-User", "faadmin")
		req.Header.Set("Remote-Groups", "admins")

		result := IsFullAuthenticated(req)
		assert.True(t, result)
	})

	t.Run("non-admin user via forward auth is NOT fully authenticated", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/users", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("Remote-User", "fauser")
		req.Header.Set("Remote-Groups", "users")

		result := IsFullAuthenticated(req)
		assert.False(t, result)
	})

	t.Run("forward auth from untrusted IP is not authenticated", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/users", nil)
		req.RemoteAddr = "8.8.8.8:12345"
		req.Header.Set("Remote-User", "hacker")
		req.Header.Set("Remote-Groups", "admins")

		result := IsFullAuthenticated(req)
		assert.False(t, result)
	})
}

func TestForwardAuthIntegrationWithIsUser(t *testing.T) {
	ensureHandlerSetup(t)

	core.App.ForwardAuth = core.ForwardAuth{
		ForwardAuthEnabled:        null.NewNullBool(true),
		ForwardAuthHeaderUser:     "Remote-User",
		ForwardAuthTrustedProxies: "127.0.0.1/32",
	}

	t.Run("any forward auth user is a user", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/services", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("Remote-User", "anyuser")

		result := IsUser(req)
		assert.True(t, result)
	})

	t.Run("forward auth from untrusted IP is not a user", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/services", nil)
		req.RemoteAddr = "8.8.8.8:12345"
		req.Header.Set("Remote-User", "anyuser")

		result := IsUser(req)
		assert.False(t, result)
	})
}

func TestForwardAuthIntegrationWithScopeName(t *testing.T) {
	ensureHandlerSetup(t)

	core.App.ForwardAuth = core.ForwardAuth{
		ForwardAuthEnabled:        null.NewNullBool(true),
		ForwardAuthHeaderUser:     "Remote-User",
		ForwardAuthHeaderGroups:   "Remote-Groups",
		ForwardAuthAdminGroups:    "admins",
		ForwardAuthTrustedProxies: "127.0.0.1/32",
	}

	t.Run("admin forward auth user gets admin scope", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("Remote-User", "scopeadmin")
		req.Header.Set("Remote-Groups", "admins")

		scope := ScopeName(req)
		assert.Equal(t, "admin", scope)
	})

	t.Run("non-admin forward auth user gets user scope", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("Remote-User", "scopeuser")
		req.Header.Set("Remote-Groups", "users")

		scope := ScopeName(req)
		assert.Equal(t, "user", scope)
	})

	t.Run("untrusted IP gets no scope", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api", nil)
		req.RemoteAddr = "8.8.8.8:12345"
		req.Header.Set("Remote-User", "hacker")
		req.Header.Set("Remote-Groups", "admins")

		scope := ScopeName(req)
		assert.Equal(t, "", scope)
	})
}

// ============================================================================
// Default Header Names Tests
// ============================================================================

func TestForwardAuthDefaultHeaderNames(t *testing.T) {
	core.App = &core.Core{
		ForwardAuth: core.ForwardAuth{
			ForwardAuthEnabled:        null.NewNullBool(true),
			ForwardAuthHeaderUser:     "", // Empty = use default "Remote-User"
			ForwardAuthHeaderEmail:    "", // Empty = use default "Remote-Email"
			ForwardAuthHeaderGroups:   "", // Empty = use default "Remote-Groups"
			ForwardAuthHeaderName:     "", // Empty = use default "Remote-Name"
			ForwardAuthAdminGroups:    "admins",
			ForwardAuthTrustedProxies: "127.0.0.1/32",
		},
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("Remote-User", "defaultuser")
	req.Header.Set("Remote-Email", "default@example.com")
	req.Header.Set("Remote-Groups", "users,admins")
	req.Header.Set("Remote-Name", "Default User")

	info := forwardAuthExtract(req)

	require.NotNil(t, info, "should use default header names when config is empty")
	assert.Equal(t, "defaultuser", info.Username)
	assert.Equal(t, "default@example.com", info.Email)
	assert.Equal(t, "Default User", info.Name)
	assert.Equal(t, []string{"users", "admins"}, info.Groups)
	assert.True(t, info.IsAdmin)
}

func TestForwardAuthPartialDefaultHeaders(t *testing.T) {
	core.App = &core.Core{
		ForwardAuth: core.ForwardAuth{
			ForwardAuthEnabled:        null.NewNullBool(true),
			ForwardAuthHeaderUser:     "X-Custom-User", // Custom
			ForwardAuthHeaderEmail:    "",              // Default
			ForwardAuthHeaderGroups:   "X-Custom-Groups", // Custom
			ForwardAuthHeaderName:     "",              // Default
			ForwardAuthTrustedProxies: "127.0.0.1/32",
		},
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("X-Custom-User", "mixeduser")
	req.Header.Set("Remote-Email", "mixed@example.com")
	req.Header.Set("X-Custom-Groups", "developers")
	req.Header.Set("Remote-Name", "Mixed User")

	info := forwardAuthExtract(req)

	require.NotNil(t, info)
	assert.Equal(t, "mixeduser", info.Username)
	assert.Equal(t, "mixed@example.com", info.Email)
	assert.Equal(t, "Mixed User", info.Name)
	assert.Equal(t, []string{"developers"}, info.Groups)
}

// ============================================================================
// hasForwardAuth Function Tests
// ============================================================================

func TestHasForwardAuth(t *testing.T) {
	// Use shared test setup
	ensureHandlerSetup(t)

	// Save and restore original settings
	origForwardAuth := core.App.ForwardAuth
	t.Cleanup(func() {
		core.App.ForwardAuth = origForwardAuth
	})

	core.App.ForwardAuth = core.ForwardAuth{
		ForwardAuthEnabled:        null.NewNullBool(true),
		ForwardAuthHeaderUser:     "Remote-User",
		ForwardAuthHeaderGroups:   "Remote-Groups",
		ForwardAuthAdminGroups:    "admins",
		ForwardAuthTrustedProxies: "127.0.0.1/32",
	}

	t.Run("returns user and true when valid", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("Remote-User", "hasfahelper")

		user, ok := hasForwardAuth(req)
		assert.True(t, ok)
		assert.NotNil(t, user)
		assert.Equal(t, "hasfahelper", user.Username)
	})

	t.Run("returns nil and false when disabled", func(t *testing.T) {
		core.App.ForwardAuthEnabled = null.NewNullBool(false)
		defer func() { core.App.ForwardAuthEnabled = null.NewNullBool(true) }()

		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("Remote-User", "disabled")

		user, ok := hasForwardAuth(req)
		assert.False(t, ok)
		assert.Nil(t, user)
	})

	t.Run("returns nil and false when untrusted", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "8.8.8.8:12345"
		req.Header.Set("Remote-User", "untrusted")

		user, ok := hasForwardAuth(req)
		assert.False(t, ok)
		assert.Nil(t, user)
	})
}
