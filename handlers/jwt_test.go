package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/statping-ng/statping-ng/types/users"
	"github.com/stretchr/testify/assert"
)

func TestJwtTokenOperations(t *testing.T) {
	ensureHandlerSetup(t)

	t.Run("setJwtToken creates valid token", func(t *testing.T) {
		user := &users.User{
			Username: "testuser",
			Scopes:   "admin",
		}
		user.Admin.Bool = true
		user.Admin.Valid = true

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/login", nil)
		claim, tokenString := setJwtToken(user, w, r)

		assert.NotEmpty(t, tokenString, "Token should not be empty")
		assert.Equal(t, "testuser", claim.Username)
		assert.True(t, claim.Admin)
		assert.Equal(t, "admin", claim.Scopes)

		cookies := w.Result().Cookies()
		assert.NotEmpty(t, cookies, "Should set cookie")
	})

	t.Run("parseToken with valid token", func(t *testing.T) {
		user := &users.User{
			Username: "parseuser",
			Scopes:   "user",
		}
		user.Admin.Bool = false
		user.Admin.Valid = true

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/login", nil)
		_, tokenString := setJwtToken(user, w, r)

		claim, err := parseToken(tokenString)
		assert.Nil(t, err)
		assert.Equal(t, "parseuser", claim.Username)
		assert.False(t, claim.Admin)
	})

	t.Run("parseToken with invalid token", func(t *testing.T) {
		_, err := parseToken("invalid.token.string")
		assert.NotNil(t, err)
	})

	t.Run("getJwtToken with no cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		_, err := getJwtToken(req)
		assert.NotNil(t, err)
		assert.Equal(t, http.ErrNoCookie, err)
	})

	t.Run("getJwtToken with valid cookie", func(t *testing.T) {
		user := &users.User{
			Username: "cookieuser",
			Scopes:   "admin",
		}
		user.Admin.Bool = true
		user.Admin.Valid = true

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/login", nil)
		_, tokenString := setJwtToken(user, w, r)

		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  cookieName,
			Value: tokenString,
		})

		claim, err := getJwtToken(req)
		assert.Nil(t, err)
		assert.Equal(t, "cookieuser", claim.Username)
	})

	t.Run("removeJwtToken clears cookie", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/logout", nil)
		removeJwtToken(w, r)

		cookies := w.Result().Cookies()
		assert.NotEmpty(t, cookies)
		for _, c := range cookies {
			if c.Name == cookieName {
				assert.Equal(t, "deleted", c.Value)
				assert.True(t, c.MaxAge < 0)
			}
		}
	})

	t.Run("isSecureRequest detects HTTPS via header", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		assert.False(t, isSecureRequest(r))

		r.Header.Set("X-Forwarded-Proto", "https")
		assert.True(t, isSecureRequest(r))
	})

	t.Run("cookie has SameSite=Strict", func(t *testing.T) {
		user := &users.User{
			Username: "strictuser",
			Scopes:   "admin",
		}
		user.Admin.Bool = true
		user.Admin.Valid = true

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/login", nil)
		setJwtToken(user, w, r)

		cookies := w.Result().Cookies()
		for _, c := range cookies {
			if c.Name == cookieName {
				assert.Equal(t, http.SameSiteStrictMode, c.SameSite)
			}
		}
	})
}
