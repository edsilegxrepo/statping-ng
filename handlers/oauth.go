package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/errors"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/types/users"
	"golang.org/x/oauth2"
)

type oAuth struct {
	Email    string
	Username string
	*oauth2.Token
}

// oauthStateStore stores OAuth state tokens with expiration for CSRF protection
var (
	oauthStateStore     = make(map[string]time.Time)
	oauthStateStoreLock sync.RWMutex
)

// generateOAuthState creates a cryptographically secure state token
func generateOAuthState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	oauthStateStoreLock.Lock()
	oauthStateStore[state] = time.Now().Add(10 * time.Minute)
	oauthStateStoreLock.Unlock()

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
	case "custom":
		oauth, err = customOAuth(r)
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
	// First, check if this OAuth user already exists in the database
	existingUser, err := users.FindByEmail(oauth.Email)
	if err == nil && existingUser != nil {
		// Existing user - use their existing permissions
		log.Infoln(fmt.Sprintf("OAuth %s User %s (existing) logged in from IP %s", oauth.Type(), oauth.Email, r.RemoteAddr))
		setJwtToken(existingUser, w)
		http.Redirect(w, r, core.App.Domain+"/dashboard", http.StatusPermanentRedirect)
		return
	}

	// New OAuth user - create as non-admin by default
	// Admins must be promoted manually through the admin interface
	user := &users.User{
		Id:       0,
		Username: oauth.Username,
		Email:    oauth.Email,
		Admin:    null.NewNullBool(false), // SECURITY: OAuth users are NOT admin by default
	}

	// Create the user in the database
	if err := user.Create(); err != nil {
		log.Errorln(fmt.Sprintf("Failed to create OAuth user %s: %v", oauth.Email, err))
		sendErrorJson(errors.New("failed to create user account"), w, r)
		return
	}

	log.Infoln(fmt.Sprintf("OAuth %s User %s (new, non-admin) logged in from IP %s", oauth.Type(), oauth.Email, r.RemoteAddr))
	setJwtToken(user, w)

	http.Redirect(w, r, core.App.Domain+"/dashboard", http.StatusPermanentRedirect)
}
