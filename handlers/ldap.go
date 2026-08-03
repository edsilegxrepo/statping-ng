package handlers

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/errors"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/types/users"
	"github.com/statping-ng/statping-ng/utils"
)

// LDAPTemplates provides preset configurations for common LDAP servers
var LDAPTemplates = map[string]LDAPTemplate{
	"openldap": {
		Name:         "OpenLDAP",
		UserFilter:   "(&(objectClass=inetOrgPerson)(uid=%s))",
		UsernameAttr: "uid",
		EmailAttr:    "mail",
	},
	"activedirectory": {
		Name:         "Microsoft Active Directory",
		UserFilter:   "(&(objectClass=user)(sAMAccountName=%s))",
		UsernameAttr: "sAMAccountName",
		EmailAttr:    "mail",
	},
	"freeipa": {
		Name:         "FreeIPA",
		UserFilter:   "(&(objectClass=person)(uid=%s))",
		UsernameAttr: "uid",
		EmailAttr:    "mail",
	},
}

type LDAPTemplate struct {
	Name         string `json:"name"`
	UserFilter   string `json:"user_filter"`
	UsernameAttr string `json:"username_attr"`
	EmailAttr    string `json:"email_attr"`
}

type LDAPTestRequest struct {
	Host         string `json:"ldap_host"`
	Port         int    `json:"ldap_port"`
	StartTLS     bool   `json:"ldap_start_tls"`
	SkipVerify   bool   `json:"ldap_skip_verify"`
	BindDN       string `json:"ldap_bind_dn"`
	BindPassword string `json:"ldap_bind_password"`
	BaseDN       string `json:"ldap_base_dn"`
}

type LDAPTestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type LDAPAuthResult struct {
	Authenticated bool
	UserDN        string
	Email         string
	MemberOf      []string
}

func connectLDAP(host string, port int, skipVerify bool, startTLS bool) (*ldap.Conn, error) {
	address := fmt.Sprintf("%s:%d", host, port)

	// Ports 636 and 3269 require implicit SSL/TLS (LDAPS)
	requiresImplicitTLS := port == 636 || port == 3269

	if requiresImplicitTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: skipVerify, // #nosec G402 - user-configurable for self-signed certs
			ServerName:         host,
			MinVersion:         tls.VersionTLS12,
		}
		return ldap.DialURL("ldaps://"+address, ldap.DialWithTLSConfig(tlsConfig))
	}

	// Plain connection REQUIRES StartTLS upgrade for security
	if !startTLS {
		return nil, fmt.Errorf("LDAP encryption is mandatory: enable StartTLS or use LDAPS (port 636/3269)")
	}

	conn, err := ldap.DialURL("ldap://" + address)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: skipVerify, // #nosec G402 - user-configurable for self-signed certs
		ServerName:         host,
		MinVersion:         tls.VersionTLS12,
	}
	if err := conn.StartTLS(tlsConfig); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("StartTLS failed: %w", err)
	}

	return conn, nil
}

func generateRandomPassword(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}

func isUserInGroup(memberOf []string, groupDN string) bool {
	if groupDN == "" {
		return true
	}
	// Support semicolon-separated group DNs (commas are part of DN syntax)
	groups := strings.Split(groupDN, ";")
	for _, group := range groups {
		group = strings.TrimSpace(group)
		if group == "" {
			continue
		}
		for _, userGroup := range memberOf {
			if strings.EqualFold(userGroup, group) {
				return true
			}
		}
	}
	return false
}

// authenticateLDAP authenticates a user via LDAP and returns auth result and user info
func authenticateLDAP(username, password string) (*LDAPAuthResult, error) {
	c := core.App
	if !c.LdapEnabled.Bool {
		return nil, errors.New("LDAP is not enabled")
	}

	conn, err := connectLDAP(c.LdapHost, c.LdapPort, c.LdapSkipVerify.Bool, c.LdapStartTLS.Bool)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP server: %w", err)
	}
	defer func() { _ = conn.Close() }()

	// Bind with service account (password already decrypted by AfterFind hook)
	if c.LdapBindDN != "" {
		if err := conn.Bind(c.LdapBindDN, c.LdapBindPassword); err != nil {
			return nil, fmt.Errorf("service account bind failed: %w", err)
		}
	}

	// Search for user
	filter := strings.ReplaceAll(c.LdapUserFilter, "%s", ldap.EscapeFilter(username))
	searchReq := ldap.NewSearchRequest(
		c.LdapBaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		1,
		0,
		false,
		filter,
		[]string{"dn", c.LdapUsernameAttr, c.LdapEmailAttr, "memberOf"},
		nil,
	)

	result, err := conn.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("user search failed: %w", err)
	}

	if len(result.Entries) != 1 {
		return &LDAPAuthResult{Authenticated: false}, nil
	}

	entry := result.Entries[0]
	userDN := entry.DN
	email := entry.GetAttributeValue(c.LdapEmailAttr)
	memberOf := entry.GetAttributeValues("memberOf")

	// Authenticate user with their password
	if err := conn.Bind(userDN, password); err != nil {
		return &LDAPAuthResult{Authenticated: false}, nil
	}

	return &LDAPAuthResult{
		Authenticated: true,
		UserDN:        userDN,
		Email:         email,
		MemberOf:      memberOf,
	}, nil
}

// processLDAPLogin handles the full LDAP login flow including group check and user provisioning
func processLDAPLogin(username, password string) (*users.User, error) {
	c := core.App

	authResult, err := authenticateLDAP(username, password)
	if err != nil {
		return nil, err
	}

	if !authResult.Authenticated {
		return nil, nil
	}

	// Check authorized group membership if enabled
	if c.LdapAuthorizedGroupEnabled.Bool && c.LdapAuthorizedGroup != "" {
		if !isUserInGroup(authResult.MemberOf, c.LdapAuthorizedGroup) {
			return nil, errors.New("user is not a member of the authorized group")
		}
	}

	// Find or create user
	user, _ := users.FindByUsername(username)
	if user == nil {
		// Create new user with disabled state
		randomPass := generateRandomPassword(32)
		user = &users.User{
			Username:     username,
			Email:        authResult.Email,
			Password:     utils.HashPassword(randomPass),
			Admin:        null.NewNullBool(false),
			AuthProvider: users.AuthProviderLDAP,
			Enabled:      null.NewNullBool(false), // Requires admin approval
		}
		if err := user.Create(); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		log.Infof("Created new LDAP user: %s (pending approval)", username)
	}

	return user, nil
}

func apiLdapTemplatesHandler(w http.ResponseWriter, r *http.Request) {
	returnJson(LDAPTemplates, w, r)
}

func apiLdapSettingsHandler(w http.ResponseWriter, r *http.Request) {
	c := core.App
	settings := map[string]interface{}{
		"ldap_enabled":                  c.LdapEnabled.Bool,
		"ldap_host":                     c.LdapHost,
		"ldap_port":                     c.LdapPort,
		"ldap_start_tls":                c.LdapStartTLS.Bool,
		"ldap_skip_verify":              c.LdapSkipVerify.Bool,
		"ldap_bind_dn":                  c.LdapBindDN,
		"ldap_base_dn":                  c.LdapBaseDN,
		"ldap_user_filter":              c.LdapUserFilter,
		"ldap_username_attr":            c.LdapUsernameAttr,
		"ldap_email_attr":               c.LdapEmailAttr,
		"ldap_authorized_group_enabled": c.LdapAuthorizedGroupEnabled.Bool,
		"ldap_authorized_group":         c.LdapAuthorizedGroup,
		"ldap_template":                 c.LdapTemplate,
	}
	returnJson(settings, w, r)
}

func apiLdapSaveHandler(w http.ResponseWriter, r *http.Request) {
	var ldapSettings core.LDAP
	if err := DecodeJSON(r, &ldapSettings); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	c := core.App
	c.LdapEnabled = ldapSettings.LdapEnabled
	c.LdapHost = ldapSettings.LdapHost
	c.LdapPort = ldapSettings.LdapPort
	c.LdapStartTLS = ldapSettings.LdapStartTLS
	c.LdapSkipVerify = ldapSettings.LdapSkipVerify
	c.LdapBindDN = ldapSettings.LdapBindDN
	if ldapSettings.LdapBindPassword != "" {
		// Password will be encrypted by BeforeSave hook
		c.LdapBindPassword = ldapSettings.LdapBindPassword
	}
	c.LdapBaseDN = ldapSettings.LdapBaseDN
	c.LdapUserFilter = ldapSettings.LdapUserFilter
	c.LdapUsernameAttr = ldapSettings.LdapUsernameAttr
	c.LdapEmailAttr = ldapSettings.LdapEmailAttr
	c.LdapAuthorizedGroupEnabled = ldapSettings.LdapAuthorizedGroupEnabled
	c.LdapAuthorizedGroup = ldapSettings.LdapAuthorizedGroup
	c.LdapTemplate = ldapSettings.LdapTemplate

	if err := c.Update(); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	returnJson(map[string]string{"status": "success"}, w, r)
}

func apiLdapTestHandler(w http.ResponseWriter, r *http.Request) {
	var req LDAPTestRequest
	if err := DecodeJSON(r, &req); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	// Test connection
	conn, err := connectLDAP(req.Host, req.Port, req.SkipVerify, req.StartTLS)
	if err != nil {
		returnJson(LDAPTestResponse{
			Success: false,
			Message: fmt.Sprintf("Connection failed: %v", err),
		}, w, r)
		return
	}
	defer func() { _ = conn.Close() }()

	// Test bind with service account
	if req.BindDN != "" {
		bindPassword := req.BindPassword
		// If password is empty, use stored password (already decrypted by AfterFind)
		if bindPassword == "" && core.App.LdapBindPassword != "" {
			bindPassword = core.App.LdapBindPassword
		}
		if err := conn.Bind(req.BindDN, bindPassword); err != nil {
			returnJson(LDAPTestResponse{
				Success: false,
				Message: fmt.Sprintf("Service account bind failed: %v", err),
			}, w, r)
			return
		}
	}

	// Test base DN search
	if req.BaseDN != "" {
		searchReq := ldap.NewSearchRequest(
			req.BaseDN,
			ldap.ScopeBaseObject,
			ldap.NeverDerefAliases,
			1,
			5,
			false,
			"(objectClass=*)",
			[]string{"dn"},
			nil,
		)
		if _, err := conn.Search(searchReq); err != nil {
			returnJson(LDAPTestResponse{
				Success: false,
				Message: fmt.Sprintf("Base DN search failed: %v", err),
			}, w, r)
			return
		}
	}

	returnJson(LDAPTestResponse{
		Success: true,
		Message: "LDAP connection and authentication successful",
	}, w, r)
}
