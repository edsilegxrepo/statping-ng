package handlers

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/users"
)

// hasSetupEnv checks to see if the Statping instance has not been setup yet
func hasSetupEnv() bool {
	if core.App == nil {
		return false
	}
	if !core.App.Setup {
		return false
	}
	return false
}

// hasAPIQuery checks the `api` query parameter against the API Secret Key
func hasAPIQuery(r *http.Request) bool {
	query := r.URL.Query()
	key := query.Get("api")
	if key == "" {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(key), []byte(core.App.ApiSecret)) == 1 {
		return true
	}
	// find user with API key and verify Admin status
	user, err := users.FindByAPIKey(key)
	if err != nil || user == nil {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(key), []byte(user.ApiKey)) == 1 {
		return user.Admin.Bool
	}
	return false
}

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
