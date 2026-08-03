package handlers

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/users"
)

// hasAuthorizationHeader check to see if the Authorization header or ?api= query param is a valid API key
func hasAuthorizationHeader(r *http.Request) bool {
	var token string
	// Check Authorization header first
	tokens, ok := r.Header["Authorization"]
	if ok && len(tokens) >= 1 {
		token = tokens[0]
		token = strings.TrimPrefix(token, "Bearer ")
	} else {
		// Fall back to ?api= query parameter
		token = r.URL.Query().Get("api")
	}

	if token == "" {
		return false
	}

	// Check if it's the app API secret
	if subtle.ConstantTimeCompare([]byte(token), []byte(core.App.ApiSecret)) == 1 {
		return true
	}
	// Check if it's a user's API key
	user, err := users.FindByAPIKey(token)
	if err == nil && user != nil {
		return user.Admin.Bool
	}
	return false
}

// AuthResult indicates the result of an authentication check
type AuthResult int

const (
	AuthResultOK AuthResult = iota
	AuthResultNoUser
	AuthResultPendingApproval
)

// hasForwardAuth checks for valid forward auth headers from a trusted proxy
// Returns the user, success bool, and auth result for specific error handling
func hasForwardAuth(r *http.Request) (*users.User, bool, AuthResult) {
	user := forwardAuthUser(r)
	if user == nil {
		return nil, false, AuthResultNoUser
	}
	// Check if user is enabled (approved by admin)
	if !user.Enabled.Bool {
		log.Infof("Forward auth user %s rejected - account pending approval", user.Username)
		return user, false, AuthResultPendingApproval
	}
	return user, true, AuthResultOK
}
