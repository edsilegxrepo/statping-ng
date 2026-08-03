package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/errors"
	"github.com/statping-ng/statping-ng/types/users"
	"golang.org/x/oauth2"
)

// PKCE verifier storage (code_verifier must survive the redirect)
type pkceEntry struct {
	verifier string
	expiry   time.Time
}

var (
	pkceStore     = make(map[string]pkceEntry) // state -> pkceEntry
	pkceStoreLock sync.RWMutex
)

const pkceExpiration = 10 * time.Minute

// oidcProvider caches the OIDC provider to avoid repeated discovery
var (
	oidcProvider       *oidc.Provider
	oidcProviderLock   sync.RWMutex
	oidcProviderExpiry time.Time
	oidcProviderIssuer string // Track which issuer the cached provider is for
)

// getOIDCProvider returns cached or creates new OIDC provider
func getOIDCProvider(ctx context.Context) (*oidc.Provider, error) {
	auth := core.App.OAuth

	oidcProviderLock.RLock()
	if oidcProvider != nil && time.Now().Before(oidcProviderExpiry) && oidcProviderIssuer == auth.OidcIssuerURL {
		provider := oidcProvider
		oidcProviderLock.RUnlock()
		return provider, nil
	}
	oidcProviderLock.RUnlock()

	oidcProviderLock.Lock()
	defer oidcProviderLock.Unlock()

	// Double-check after acquiring write lock
	if oidcProvider != nil && time.Now().Before(oidcProviderExpiry) && oidcProviderIssuer == auth.OidcIssuerURL {
		return oidcProvider, nil
	}

	// Standard discovery from issuer URL
	provider, err := oidc.NewProvider(ctx, auth.OidcIssuerURL)
	if err != nil {
		return nil, errors.Wrap(err, "OIDC discovery failed")
	}

	oidcProvider = provider
	oidcProviderExpiry = time.Now().Add(1 * time.Hour)
	oidcProviderIssuer = auth.OidcIssuerURL

	return provider, nil
}

// generatePKCE creates code_verifier and code_challenge for PKCE flow
func generatePKCE() (verifier, challenge string, err error) {
	// Generate 32 random bytes for verifier
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	verifier = base64.RawURLEncoding.EncodeToString(b)

	// SHA256 hash for challenge (S256 method)
	h := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h[:])

	return verifier, challenge, nil
}

// storePKCEVerifier stores verifier keyed by state with expiration
func storePKCEVerifier(state, verifier string) {
	pkceStoreLock.Lock()
	defer pkceStoreLock.Unlock()

	// Cleanup expired entries first
	now := time.Now()
	for k, v := range pkceStore {
		if now.After(v.expiry) {
			delete(pkceStore, k)
		}
	}

	// Check limit after cleanup
	if len(pkceStore) >= maxOAuthStates {
		for k := range pkceStore {
			delete(pkceStore, k)
			break
		}
	}

	pkceStore[state] = pkceEntry{
		verifier: verifier,
		expiry:   now.Add(pkceExpiration),
	}
}

// consumePKCEVerifier retrieves and deletes verifier (one-time use)
func consumePKCEVerifier(state string) (string, bool) {
	pkceStoreLock.Lock()
	defer pkceStoreLock.Unlock()

	entry, ok := pkceStore[state]
	if !ok {
		return "", false
	}

	delete(pkceStore, state)

	// Check if expired
	if time.Now().After(entry.expiry) {
		return "", false
	}

	return entry.verifier, true
}

// oidcOAuth handles the OIDC callback after user authenticates with IdP
func oidcOAuth(r *http.Request) (*oAuth, error) {
	ctx := r.Context()
	auth := core.App.OAuth
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if !auth.OidcEnabled.Bool {
		return nil, errors.New("OIDC is not enabled")
	}

	// Get OIDC provider (cached)
	provider, err := getOIDCProvider(ctx)
	if err != nil {
		return nil, err
	}

	// Build scopes list
	scopes := parseScopes(auth.OidcScopes)

	// Build OAuth2 config
	oauth2Config := oauth2.Config{
		ClientID:     auth.OidcClientID,
		ClientSecret: auth.OidcClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  core.App.Domain + basePath + "oauth/oidc",
		Scopes:       scopes,
	}

	// Exchange code for token
	var opts []oauth2.AuthCodeOption

	// Add PKCE verifier if enabled
	if auth.OidcUsePKCE.Bool {
		verifier, ok := consumePKCEVerifier(state)
		if !ok {
			log.Warnf("OIDC: PKCE verifier not found for state (may have expired)")
			// Continue without PKCE - some IdPs don't require it
		} else {
			opts = append(opts, oauth2.SetAuthURLParam("code_verifier", verifier))
		}
	}

	token, err := oauth2Config.Exchange(ctx, code, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "OIDC token exchange failed")
	}

	// Extract ID token from response
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token in OIDC response")
	}

	// Verify ID token signature and claims
	verifier := provider.Verifier(&oidc.Config{
		ClientID: auth.OidcClientID,
	})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, errors.Wrap(err, "OIDC id_token verification failed")
	}

	// Extract claims from ID token
	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		return nil, errors.Wrap(err, "failed to parse OIDC claims")
	}

	// Get username and email from configured claim names
	usernameClaim := auth.OidcClaimUsername
	if usernameClaim == "" {
		usernameClaim = "preferred_username"
	}
	emailClaim := auth.OidcClaimEmail
	if emailClaim == "" {
		emailClaim = "email"
	}

	username := extractClaim(claims, usernameClaim)
	email := extractClaim(claims, emailClaim)

	// Fallback to userinfo endpoint if claims missing
	if username == "" || email == "" {
		userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
		if err == nil {
			var infoClaims map[string]interface{}
			if err := userInfo.Claims(&infoClaims); err == nil {
				if username == "" {
					username = extractClaim(infoClaims, usernameClaim)
				}
				if email == "" {
					email = extractClaim(infoClaims, emailClaim)
				}
			}
		} else {
			log.Warnf("OIDC: userinfo endpoint failed: %v", err)
		}
	}

	// Final fallback: use subject as username
	if username == "" {
		username = idToken.Subject
	}
	if email == "" {
		return nil, errors.New("OIDC: email claim not found in id_token or userinfo")
	}

	// Validate against allowed users list
	if !validateOIDCUser(email, auth) {
		return nil, errors.New("OIDC: user not authorized")
	}

	// Check admin groups
	isAdmin := checkOIDCAdminGroups(claims, auth)

	log.Infof("OIDC: authenticated user %s (%s) admin=%v", username, email, isAdmin)

	return &oAuth{
		Token:        token,
		Username:     strings.ToLower(username),
		Email:        strings.ToLower(email),
		ProviderType: users.AuthProviderOIDC,
		IsAdmin:      isAdmin,
	}, nil
}

// parseScopes splits comma-separated scopes and ensures openid is present
func parseScopes(scopeStr string) []string {
	if scopeStr == "" {
		return []string{"openid", "profile", "email"}
	}

	var scopes []string
	hasOpenID := false
	for _, s := range strings.Split(scopeStr, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			scopes = append(scopes, s)
			if s == "openid" {
				hasOpenID = true
			}
		}
	}

	// Ensure openid scope is present (required for OIDC)
	if !hasOpenID {
		scopes = append([]string{"openid"}, scopes...)
	}

	return scopes
}

// extractClaim gets a string claim value from claims map
func extractClaim(claims map[string]interface{}, key string) string {
	if v, ok := claims[key].(string); ok {
		return v
	}
	return ""
}

// validateOIDCUser checks if user is in allowed list
func validateOIDCUser(email string, auth core.OAuth) bool {
	if auth.OidcAllowedUsers == "" {
		return true // No restrictions
	}

	email = strings.ToLower(email)
	allowed := strings.Split(auth.OidcAllowedUsers, ",")
	for _, u := range allowed {
		u = strings.TrimSpace(strings.ToLower(u))
		if u == "" {
			continue
		}
		// Exact email match
		if email == u {
			return true
		}
		// Domain match (e.g., "@company.com")
		if strings.HasPrefix(u, "@") && strings.HasSuffix(email, u) {
			return true
		}
	}
	return false
}

// checkOIDCAdminGroups checks if user is in admin groups
func checkOIDCAdminGroups(claims map[string]interface{}, auth core.OAuth) bool {
	if auth.OidcAdminGroups == "" {
		return false
	}

	groupClaim := auth.OidcClaimGroups
	if groupClaim == "" {
		groupClaim = "groups"
	}

	// Groups can be []interface{} or []string depending on IdP
	var userGroups []string
	switch g := claims[groupClaim].(type) {
	case []interface{}:
		for _, v := range g {
			if s, ok := v.(string); ok {
				userGroups = append(userGroups, s)
			}
		}
	case []string:
		userGroups = g
	}

	if len(userGroups) == 0 {
		return false
	}

	adminGroups := strings.Split(auth.OidcAdminGroups, ";")
	for _, adminGroup := range adminGroups {
		adminGroup = strings.TrimSpace(adminGroup)
		if adminGroup == "" {
			continue
		}
		for _, userGroup := range userGroups {
			if strings.EqualFold(adminGroup, userGroup) {
				return true
			}
		}
	}
	return false
}

// oidcAuthURLHandler returns the authorization URL with PKCE for frontend
func oidcAuthURLHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := core.App.OAuth

	if !auth.OidcEnabled.Bool {
		sendErrorJson(errors.New("OIDC is not enabled"), w, r)
		return
	}

	if auth.OidcIssuerURL == "" || auth.OidcClientID == "" {
		sendErrorJson(errors.New("OIDC is not configured"), w, r)
		return
	}

	provider, err := getOIDCProvider(ctx)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	// Generate state token
	state, err := generateOAuthState()
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	scopes := parseScopes(auth.OidcScopes)

	oauth2Config := oauth2.Config{
		ClientID:     auth.OidcClientID,
		ClientSecret: auth.OidcClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  core.App.Domain + basePath + "oauth/oidc",
		Scopes:       scopes,
	}

	var opts []oauth2.AuthCodeOption

	// Add PKCE if enabled
	if auth.OidcUsePKCE.Bool {
		verifier, challenge, err := generatePKCE()
		if err != nil {
			sendErrorJson(err, w, r)
			return
		}
		storePKCEVerifier(state, verifier)
		opts = append(opts,
			oauth2.SetAuthURLParam("code_challenge", challenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		)
	}

	authURL := oauth2Config.AuthCodeURL(state, opts...)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"url":   authURL,
		"state": state,
	})
}
