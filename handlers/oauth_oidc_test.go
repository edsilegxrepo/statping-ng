package handlers

import (
	"testing"

	"github.com/statping-ng/statping-ng/types/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePKCE(t *testing.T) {
	verifier1, challenge1, err := generatePKCE()
	require.NoError(t, err)
	assert.NotEmpty(t, verifier1, "verifier should not be empty")
	assert.NotEmpty(t, challenge1, "challenge should not be empty")
	assert.NotEqual(t, verifier1, challenge1, "verifier and challenge should be different")

	// Generate another and verify they're unique
	verifier2, challenge2, err := generatePKCE()
	require.NoError(t, err)
	assert.NotEqual(t, verifier1, verifier2, "verifiers should be unique")
	assert.NotEqual(t, challenge1, challenge2, "challenges should be unique")
}

func TestPKCEVerifierStorage(t *testing.T) {
	state := "test-state-12345"
	verifier := "test-verifier-abcdef"

	// Store verifier
	storePKCEVerifier(state, verifier)

	// First retrieval should succeed
	retrieved, ok := consumePKCEVerifier(state)
	assert.True(t, ok, "first retrieval should succeed")
	assert.Equal(t, verifier, retrieved)

	// Second retrieval should fail (one-time use)
	_, ok = consumePKCEVerifier(state)
	assert.False(t, ok, "second retrieval should fail (consumed)")
}

func TestPKCEVerifierNotFound(t *testing.T) {
	_, ok := consumePKCEVerifier("nonexistent-state")
	assert.False(t, ok, "nonexistent state should not be found")
}

func TestParseScopes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string returns defaults",
			input:    "",
			expected: []string{"openid", "profile", "email"},
		},
		{
			name:     "custom scopes with openid",
			input:    "openid,profile,email,groups",
			expected: []string{"openid", "profile", "email", "groups"},
		},
		{
			name:     "custom scopes without openid adds it",
			input:    "profile,email",
			expected: []string{"openid", "profile", "email"},
		},
		{
			name:     "handles whitespace",
			input:    " profile , email , groups ",
			expected: []string{"openid", "profile", "email", "groups"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseScopes(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractClaim(t *testing.T) {
	claims := map[string]interface{}{
		"sub":                "user123",
		"email":              "user@example.com",
		"preferred_username": "johndoe",
		"empty_string":       "",
		"number":             42,
	}

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"extract email", "email", "user@example.com"},
		{"extract username", "preferred_username", "johndoe"},
		{"extract sub", "sub", "user123"},
		{"missing key returns empty", "missing", ""},
		{"empty string returns empty", "empty_string", ""},
		{"non-string returns empty", "number", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := extractClaim(claims, tc.key)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidateOIDCUser(t *testing.T) {
	tests := []struct {
		name         string
		email        string
		allowedUsers string
		expected     bool
	}{
		{
			name:         "empty allowed list allows all",
			email:        "anyone@example.com",
			allowedUsers: "",
			expected:     true,
		},
		{
			name:         "exact email match",
			email:        "admin@example.com",
			allowedUsers: "admin@example.com,other@example.com",
			expected:     true,
		},
		{
			name:         "case insensitive email match",
			email:        "Admin@Example.COM",
			allowedUsers: "admin@example.com",
			expected:     true,
		},
		{
			name:         "domain match",
			email:        "user@company.com",
			allowedUsers: "@company.com",
			expected:     true,
		},
		{
			name:         "domain match case insensitive",
			email:        "USER@Company.COM",
			allowedUsers: "@company.com",
			expected:     true,
		},
		{
			name:         "not in allowed list",
			email:        "hacker@evil.com",
			allowedUsers: "admin@example.com,@company.com",
			expected:     false,
		},
		{
			name:         "partial domain doesn't match",
			email:        "user@notcompany.com",
			allowedUsers: "@company.com",
			expected:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			auth := core.OAuth{OidcAllowedUsers: tc.allowedUsers}
			result := validateOIDCUser(tc.email, auth)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCheckOIDCAdminGroups(t *testing.T) {
	tests := []struct {
		name        string
		claims      map[string]interface{}
		adminGroups string
		groupClaim  string
		expected    bool
	}{
		{
			name:        "empty admin groups returns false",
			claims:      map[string]interface{}{"groups": []interface{}{"users", "devs"}},
			adminGroups: "",
			groupClaim:  "",
			expected:    false,
		},
		{
			name:        "user in admin group",
			claims:      map[string]interface{}{"groups": []interface{}{"users", "admins"}},
			adminGroups: "admins",
			groupClaim:  "",
			expected:    true,
		},
		{
			name:        "user not in admin group",
			claims:      map[string]interface{}{"groups": []interface{}{"users", "devs"}},
			adminGroups: "admins",
			groupClaim:  "",
			expected:    false,
		},
		{
			name:        "multiple admin groups with match",
			claims:      map[string]interface{}{"groups": []interface{}{"users", "superadmins"}},
			adminGroups: "admins;superadmins;wheel",
			groupClaim:  "",
			expected:    true,
		},
		{
			name:        "case insensitive group match",
			claims:      map[string]interface{}{"groups": []interface{}{"Admins"}},
			adminGroups: "admins",
			groupClaim:  "",
			expected:    true,
		},
		{
			name:        "custom group claim",
			claims:      map[string]interface{}{"roles": []interface{}{"admin"}},
			adminGroups: "admin",
			groupClaim:  "roles",
			expected:    true,
		},
		{
			name:        "no groups claim in token",
			claims:      map[string]interface{}{"email": "user@example.com"},
			adminGroups: "admins",
			groupClaim:  "",
			expected:    false,
		},
		{
			name:        "groups as string array",
			claims:      map[string]interface{}{"groups": []string{"users", "admins"}},
			adminGroups: "admins",
			groupClaim:  "",
			expected:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			auth := core.OAuth{
				OidcAdminGroups:  tc.adminGroups,
				OidcClaimGroups:  tc.groupClaim,
			}
			result := checkOIDCAdminGroups(tc.claims, auth)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestOIDCAuthURLHandlerNotEnabled(t *testing.T) {
	ensureHandlerSetup(t)

	// Ensure OIDC is disabled
	core.App.OidcEnabled.Bool = false

	tests := []HTTPTest{
		{
			Name:             "OIDC auth URL when disabled",
			URL:              "/api/oauth/oidc/auth-url",
			Method:           "GET",
			ExpectedStatus:   200,
			ExpectedContains: []string{"OIDC is not enabled"},
			NoAuth:           true,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestOIDCAuthURLHandlerNotConfigured(t *testing.T) {
	ensureHandlerSetup(t)

	// Enable OIDC but don't configure it
	core.App.OidcEnabled.Bool = true
	core.App.OidcIssuerURL = ""
	core.App.OidcClientID = ""

	tests := []HTTPTest{
		{
			Name:             "OIDC auth URL when not configured",
			URL:              "/api/oauth/oidc/auth-url",
			Method:           "GET",
			ExpectedStatus:   200,
			ExpectedContains: []string{"OIDC is not configured"},
			NoAuth:           true,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}

	// Reset
	core.App.OidcEnabled.Bool = false
}
