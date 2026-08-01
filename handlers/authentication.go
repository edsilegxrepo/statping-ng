package handlers

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/users"
)

// hasAuthorizationHeader check to see if the Authorization header is the correct API Secret Key
func hasAuthorizationHeader(r *http.Request) bool {
	var token string
	tokens, ok := r.Header["Authorization"]
	if ok && len(tokens) >= 1 {
		token = tokens[0]
		token = strings.TrimPrefix(token, "Bearer ")
		if subtle.ConstantTimeCompare([]byte(token), []byte(core.App.ApiSecret)) == 1 {
			return true
		}
		user, err := users.FindByAPIKey(token)
		if err == nil && user != nil {
			return user.Admin.Bool
		}
	}
	return false
}

// hasForwardAuth checks for valid forward auth headers from a trusted proxy
func hasForwardAuth(r *http.Request) (*users.User, bool) {
	user := forwardAuthUser(r)
	return user, user != nil
}
