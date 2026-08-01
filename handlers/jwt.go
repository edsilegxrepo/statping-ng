package handlers

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/users"
)

type JwtClaim struct {
	Username string `json:"username"`
	Admin    bool   `json:"admin"`
	Scopes   string `json:"scopes"`
	jwt.RegisteredClaims
}

// isSecureRequest returns true if the request is over HTTPS,
// either directly or via a reverse proxy (X-Forwarded-Proto header)
func isSecureRequest(r *http.Request) bool {
	if usingSSL {
		return true
	}
	return r.Header.Get("X-Forwarded-Proto") == "https"
}

func removeJwtToken(w http.ResponseWriter, r *http.Request) {
	c := http.Cookie{ // #nosec G124 - HttpOnly, Secure, SameSite are set below
		Name:     cookieName,
		Value:    "deleted",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &c)
}

const (
	defaultSessionTimeout = 720   // 12 hours in minutes
	maxSessionTimeout     = 43200 // 30 days in minutes
)

func setJwtToken(user *users.User, w http.ResponseWriter, r *http.Request) (JwtClaim, string) {
	// Get session timeout from settings (in minutes), default to 720 (12 hours)
	timeoutMinutes := core.App.SessionTimeout
	if timeoutMinutes <= 0 {
		timeoutMinutes = defaultSessionTimeout
	}
	if timeoutMinutes > maxSessionTimeout {
		timeoutMinutes = maxSessionTimeout
	}
	expirationTime := time.Now().Add(time.Duration(timeoutMinutes) * time.Minute)
	jwtClaim := JwtClaim{
		Username: user.Username,
		Admin:    user.Admin.Bool,
		Scopes:   user.Scopes,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaim)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Errorln("error setting token: ", err)
	}
	user.Token = tokenString
	http.SetCookie(w, &http.Cookie{ // #nosec G124 - HttpOnly, Secure, SameSite are set below
		Name:     cookieName,
		Value:    tokenString,
		Expires:  expirationTime,
		MaxAge:   timeoutMinutes * 60,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteStrictMode,
	})
	return jwtClaim, tokenString
}

func parseToken(token string) (JwtClaim, error) {
	var claims JwtClaim
	tkn, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return JwtClaim{}, err
		}
		return JwtClaim{}, err
	}
	if !tkn.Valid {
		return claims, errors.New("token is not valid")
	}
	return claims, nil
}

func getJwtToken(r *http.Request) (JwtClaim, error) {
	c, err := r.Cookie(cookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			return JwtClaim{}, err
		}
		return JwtClaim{}, err
	}
	return parseToken(c.Value)
}
