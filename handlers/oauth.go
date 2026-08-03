package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/errors"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/types/users"
	"github.com/statping-ng/statping-ng/utils"
	"golang.org/x/oauth2"
)

// oauthUserProvisionMu prevents race conditions during OAuth user creation
var oauthUserProvisionMu sync.Mutex

type oAuth struct {
	Email        string
	Username     string
	ProviderType string // e.g., "oauth_google", "oauth_github", etc.
	IsAdmin      bool   // Set by OIDC group claim check
	*oauth2.Token
}

// Type returns the OAuth provider type for logging purposes
func (o *oAuth) Type() string {
	if o.ProviderType != "" {
		return o.ProviderType
	}
	return "oauth"
}

// oauthStateStore stores OAuth state tokens with expiration for CSRF protection
var (
	oauthStateStore     = make(map[string]time.Time)
	oauthStateStoreLock sync.RWMutex
)

// maxOAuthStates limits the number of pending OAuth states to prevent memory exhaustion
const maxOAuthStates = 10000

// generateOAuthState creates a cryptographically secure state token
func generateOAuthState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	oauthStateStoreLock.Lock()
	defer oauthStateStoreLock.Unlock()

	// Prevent memory exhaustion from DoS attacks
	if len(oauthStateStore) >= maxOAuthStates {
		// Cleanup expired states first
		now := time.Now()
		for s, expiry := range oauthStateStore {
			if now.After(expiry) {
				delete(oauthStateStore, s)
			}
		}
		// If still at limit after cleanup, reject
		if len(oauthStateStore) >= maxOAuthStates {
			return "", errors.New("too many pending OAuth requests, please try again later")
		}
	}

	oauthStateStore[state] = time.Now().Add(10 * time.Minute)

	return state, nil
}

// validateOAuthState validates and consumes the state token (one-time use)
func validateOAuthState(state string) bool {
	if state == "" {
		return false
	}

	oauthStateStoreLock.Lock()
	defer oauthStateStoreLock.Unlock()

	expiry, exists := oauthStateStore[state]
	if !exists {
		return false
	}

	// Delete the state (one-time use)
	delete(oauthStateStore, state)

	// Check if expired
	if time.Now().After(expiry) {
		return false
	}

	return true
}

// cleanupExpiredOAuthStates removes expired state tokens (called periodically)
func cleanupExpiredOAuthStates() {
	oauthStateStoreLock.Lock()
	defer oauthStateStoreLock.Unlock()

	now := time.Now()
	for state, expiry := range oauthStateStore {
		if now.After(expiry) {
			delete(oauthStateStore, state)
		}
	}
}

// oauthStateHandler generates a new OAuth state token for CSRF protection
func oauthStateHandler(w http.ResponseWriter, r *http.Request) {
	state, err := generateOAuthState()
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"state":"` + state + `"}`))
}

func oauthHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider"]

	// Validate OAuth state parameter for CSRF protection
	state := r.URL.Query().Get("state")
	if !validateOAuthState(state) {
		log.Warnln(fmt.Sprintf("Invalid OAuth state from IP %s", r.RemoteAddr))
		sendErrorJson(errors.New("invalid or expired OAuth state"), w, r)
		return
	}

	var err error
	var oauth *oAuth
	switch provider {
	case "google":
		oauth, err = googleOAuth(r)
	case "github":
		oauth, err = githubOAuth(r)
	case "slack":
		oauth, err = slackOAuth(r)
	case "oidc":
		oauth, err = oidcOAuth(r)
	default:
		err = errors.New("unknown oauth provider")
	}

	if err != nil {
		log.Error(err)
		sendErrorJson(err, w, r)
		return
	}

	oauthLogin(oauth, w, r)
}

func oauthLogin(oauth *oAuth, w http.ResponseWriter, r *http.Request) {
	// Use mutex to prevent race condition during user lookup/creation
	oauthUserProvisionMu.Lock()
	defer oauthUserProvisionMu.Unlock()

	// Normalize email for lookup (already lowercased in provider handlers)
	email := strings.ToLower(oauth.Email)

	// First, check if this OAuth user already exists in the database
	existingUser, err := users.FindByEmail(email)
	if err == nil && existingUser != nil {
		// Check if user is enabled
		if !existingUser.Enabled.Bool {
			AuditLog(AuditOAuthFailed, r, map[string]interface{}{
				"email":    email,
				"provider": oauth.ProviderType,
				"reason":   "account_disabled",
			})
			http.Redirect(w, r, core.App.Domain+"/login?error=pending_approval", http.StatusTemporaryRedirect)
			return
		}
		// Existing user - use their existing permissions
		AuditLog(AuditOAuthLogin, r, map[string]interface{}{
			"username": existingUser.Username,
			"email":    email,
			"provider": oauth.ProviderType,
		})
		setJwtToken(existingUser, w, r)
		http.Redirect(w, r, core.App.Domain+"/dashboard", http.StatusFound)
		return
	}

	// Normalize username
	username := strings.ToLower(oauth.Username)

	// Check if username already exists (could be a local user or from another OAuth provider)
	existingByUsername, _ := users.FindByUsername(username)
	if existingByUsername != nil {
		// Username collision - append a unique suffix
		username = fmt.Sprintf("%s_%s", username, utils.RandomString(6))
		log.Warnf("OAuth: username collision, using %s for %s", username, email)
	}

	// New OAuth user - create as non-admin and disabled by default
	// Requires admin approval before they can access the system
	// Exception: OIDC can set admin based on group claims
	user := &users.User{
		Id:           0,
		Username:     username,
		Email:        email,
		Password:     utils.HashPassword(utils.RandomString(32)), // Random password - user authenticates via OAuth
		AuthProvider: oauth.ProviderType,
		Admin:        null.NewNullBool(oauth.IsAdmin), // OIDC can set admin via group claims
		Enabled:      null.NewNullBool(false),         // Requires admin approval
	}

	// Create the user in the database
	if err := user.Create(); err != nil {
		AuditLog(AuditOAuthFailed, r, map[string]interface{}{
			"email":    email,
			"provider": oauth.ProviderType,
			"reason":   "create_failed",
			"error":    err.Error(),
		})
		sendErrorJson(errors.New("failed to create user account"), w, r)
		return
	}

	AuditLog(AuditOAuthUserCreated, r, map[string]interface{}{
		"username": username,
		"email":    email,
		"provider": oauth.ProviderType,
		"is_admin": oauth.IsAdmin,
	})
	http.Redirect(w, r, core.App.Domain+"/login?error=pending_approval", http.StatusTemporaryRedirect)
}
