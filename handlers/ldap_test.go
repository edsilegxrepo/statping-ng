package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/statping-ng/statping-ng/types/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLDAPTemplates(t *testing.T) {
	// Verify built-in templates exist
	assert.Contains(t, LDAPTemplates, "openldap")
	assert.Contains(t, LDAPTemplates, "activedirectory")
	assert.Contains(t, LDAPTemplates, "freeipa")

	// Verify template structure
	ad := LDAPTemplates["activedirectory"]
	assert.Equal(t, "Microsoft Active Directory", ad.Name)
	assert.Contains(t, ad.UserFilter, "sAMAccountName")
	assert.Equal(t, "sAMAccountName", ad.UsernameAttr)
	assert.Equal(t, "mail", ad.EmailAttr)

	openldap := LDAPTemplates["openldap"]
	assert.Equal(t, "OpenLDAP", openldap.Name)
	assert.Contains(t, openldap.UserFilter, "uid")
}

func TestApiLdapTemplatesHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/ldap/templates", nil)
	w := httptest.NewRecorder()

	apiLdapTemplatesHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]LDAPTemplate
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Len(t, resp, 3)
	assert.Contains(t, resp, "openldap")
	assert.Contains(t, resp, "activedirectory")
	assert.Contains(t, resp, "freeipa")
}

func TestApiLdapSettingsHandler(t *testing.T) {
	if core.App == nil {
		t.Skip("requires database setup")
	}

	req := httptest.NewRequest("GET", "/api/ldap", nil)
	w := httptest.NewRecorder()

	apiLdapSettingsHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Verify expected fields
	expectedFields := []string{
		"ldap_enabled", "ldap_host", "ldap_port", "ldap_start_tls",
		"ldap_skip_verify", "ldap_bind_dn", "ldap_base_dn",
		"ldap_user_filter", "ldap_username_attr", "ldap_email_attr",
		"ldap_authorized_group_enabled", "ldap_authorized_group", "ldap_template",
	}

	for _, field := range expectedFields {
		_, exists := resp[field]
		assert.True(t, exists, "response should have field: %s", field)
	}
}

func TestApiLdapSaveHandler(t *testing.T) {
	if core.App == nil {
		t.Skip("requires database setup")
	}

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid LDAP settings",
			payload: map[string]interface{}{
				"ldap_enabled":     true,
				"ldap_host":        "ldap.example.com",
				"ldap_port":        636,
				"ldap_start_tls":   false,
				"ldap_skip_verify": false,
				"ldap_bind_dn":     "cn=admin,dc=example,dc=com",
				"ldap_base_dn":     "dc=example,dc=com",
				"ldap_user_filter": "(&(objectClass=user)(sAMAccountName=%s))",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "disable LDAP",
			payload: map[string]interface{}{
				"ldap_enabled": false,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "OpenLDAP template settings",
			payload: map[string]interface{}{
				"ldap_enabled":       true,
				"ldap_host":          "openldap.local",
				"ldap_port":          389,
				"ldap_start_tls":     true,
				"ldap_user_filter":   "(&(objectClass=inetOrgPerson)(uid=%s))",
				"ldap_username_attr": "uid",
				"ldap_email_attr":    "mail",
				"ldap_template":      "openldap",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/ldap", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			apiLdapSaveHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestApiLdapSaveHandlerInvalidJSON(t *testing.T) {
	if core.App == nil {
		t.Skip("requires database setup")
	}

	req := httptest.NewRequest("POST", "/api/ldap", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	apiLdapSaveHandler(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestApiLdapTestHandlerInvalidJSON(t *testing.T) {
	if core.App == nil {
		t.Skip("requires database setup")
	}

	req := httptest.NewRequest("POST", "/api/ldap/test", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	apiLdapTestHandler(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestApiLdapTestHandlerConnectionFailure(t *testing.T) {
	if core.App == nil {
		t.Skip("requires database setup")
	}

	payload := map[string]interface{}{
		"ldap_host":        "nonexistent.invalid",
		"ldap_port":        636,
		"ldap_skip_verify": true,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/ldap/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	apiLdapTestHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp LDAPTestResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.False(t, resp.Success)
	assert.Contains(t, resp.Message, "Connection failed")
}

func TestIsUserInGroup(t *testing.T) {
	tests := []struct {
		name      string
		memberOf  []string
		groupDN   string
		expected  bool
	}{
		{
			name:     "empty group DN allows all",
			memberOf: []string{"cn=users,dc=example,dc=com"},
			groupDN:  "",
			expected: true,
		},
		{
			name:     "exact match",
			memberOf: []string{"cn=admins,dc=example,dc=com", "cn=users,dc=example,dc=com"},
			groupDN:  "cn=admins,dc=example,dc=com",
			expected: true,
		},
		{
			name:     "case insensitive match",
			memberOf: []string{"CN=Admins,DC=Example,DC=Com"},
			groupDN:  "cn=admins,dc=example,dc=com",
			expected: true,
		},
		{
			name:     "no match",
			memberOf: []string{"cn=users,dc=example,dc=com"},
			groupDN:  "cn=admins,dc=example,dc=com",
			expected: false,
		},
		{
			name:     "semicolon-separated groups - first match",
			memberOf: []string{"cn=admins,dc=example,dc=com"},
			groupDN:  "cn=admins,dc=example,dc=com; cn=superadmins,dc=example,dc=com",
			expected: true,
		},
		{
			name:     "semicolon-separated groups - second match",
			memberOf: []string{"cn=superadmins,dc=example,dc=com"},
			groupDN:  "cn=admins,dc=example,dc=com; cn=superadmins,dc=example,dc=com",
			expected: true,
		},
		{
			name:     "semicolon-separated groups - no match",
			memberOf: []string{"cn=users,dc=example,dc=com"},
			groupDN:  "cn=admins,dc=example,dc=com; cn=superadmins,dc=example,dc=com",
			expected: false,
		},
		{
			name:     "empty memberOf",
			memberOf: []string{},
			groupDN:  "cn=admins,dc=example,dc=com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isUserInGroup(tt.memberOf, tt.groupDN)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateRandomPassword(t *testing.T) {
	// Test various lengths
	lengths := []int{8, 16, 32, 64}

	for _, length := range lengths {
		t.Run("length_"+string(rune(length)), func(t *testing.T) {
			password := generateRandomPassword(length)
			assert.Len(t, password, length)
		})
	}

	// Test uniqueness
	passwords := make(map[string]bool)
	for i := 0; i < 100; i++ {
		p := generateRandomPassword(32)
		assert.False(t, passwords[p], "generated duplicate password")
		passwords[p] = true
	}
}

func TestLDAPTestRequest(t *testing.T) {
	// Test JSON marshaling/unmarshaling
	req := LDAPTestRequest{
		Host:         "ldap.example.com",
		Port:         636,
		StartTLS:     false,
		SkipVerify:   true,
		BindDN:       "cn=admin,dc=example,dc=com",
		BindPassword: "secret",
		BaseDN:       "dc=example,dc=com",
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var decoded LDAPTestRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, req.Host, decoded.Host)
	assert.Equal(t, req.Port, decoded.Port)
	assert.Equal(t, req.StartTLS, decoded.StartTLS)
	assert.Equal(t, req.SkipVerify, decoded.SkipVerify)
	assert.Equal(t, req.BindDN, decoded.BindDN)
	assert.Equal(t, req.BindPassword, decoded.BindPassword)
	assert.Equal(t, req.BaseDN, decoded.BaseDN)
}

func TestLDAPTestResponse(t *testing.T) {
	// Test success response
	success := LDAPTestResponse{Success: true, Message: "OK"}
	data, _ := json.Marshal(success)
	assert.Contains(t, string(data), `"success":true`)

	// Test failure response
	failure := LDAPTestResponse{Success: false, Message: "Connection failed"}
	data, _ = json.Marshal(failure)
	assert.Contains(t, string(data), `"success":false`)
}
