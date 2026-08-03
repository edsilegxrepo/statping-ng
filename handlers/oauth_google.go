package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/errors"
	"github.com/statping-ng/statping-ng/types/users"
	"github.com/statping-ng/statping-ng/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func googleOAuth(r *http.Request) (*oAuth, error) {
	auth := core.App.OAuth
	code := r.URL.Query().Get("code")

	config := &oauth2.Config{
		ClientID:     auth.GoogleClientID,
		ClientSecret: auth.GoogleClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  core.App.Domain + basePath + "oauth/google",
	}

	gg, err := config.Exchange(r.Context(), code)
	if err != nil {
		return nil, err
	}

	if !gg.Valid() {
		return nil, errors.New("oauth token is not valid")
	}

	info, err := returnGoogleInfo(gg.AccessToken)
	if err != nil {
		return nil, err
	}

	if !validateGoogle(info) {
		return nil, errors.New("google user is not allowed to login")
	}

	return &oAuth{
		Token:        gg,
		Username:     strings.ToLower(info.Name),
		Email:        strings.ToLower(info.Email),
		ProviderType: users.AuthProviderOAuthGoogle,
	}, nil
}

func validateGoogle(info googleUserInfo) bool {
	auth := core.App.OAuth
	if auth.GoogleUsers == "" {
		return true
	}

	if auth.GoogleUsers != "" {
		users := strings.Split(auth.GoogleUsers, ",")
		for _, u := range users {
			if strings.EqualFold(info.Email, u) {
				return true
			}
			if strings.EqualFold(info.Hd, u) {
				return true
			}
		}
	}

	return false
}

func returnGoogleInfo(token string) (googleUserInfo, error) {
	// Use Authorization header instead of query parameter to prevent token leakage in logs
	headers := []string{"Authorization=Bearer " + token}
	resp, _, err := utils.HttpRequest("https://www.googleapis.com/oauth2/v1/userinfo?alt=json", "GET", nil, headers, nil, 10*time.Second, true, nil)
	if err != nil {
		return googleUserInfo{}, err
	}
	var user googleUserInfo
	if err := json.Unmarshal(resp, &user); err != nil {
		return googleUserInfo{}, err
	}
	return user, nil
}

type googleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Link          string `json:"link"`
	Picture       string `json:"picture"`
	Gender        string `json:"gender"`
	Locale        string `json:"locale"`
	Hd            string `json:"hd"`
}
