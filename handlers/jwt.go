package handlers

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/statping-ng/statping-ng/types/users"
)

type JwtClaim struct {
	Username string `json:"username"`
	Admin    bool   `json:"admin"`
	Scopes   string `json:"scopes"`
	jwt.RegisteredClaims
}

func removeJwtToken(w http.ResponseWriter) {
	c := http.Cookie{
		Name:     cookieName,
		Value:    "deleted",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		Secure:   usingSSL,
		SameSite: http.SameSiteLaxMode,
	} // #nosec G124
	http.SetCookie(w, &c)
}

func setJwtToken(user *users.User, w http.ResponseWriter) (JwtClaim, string) {
	expirationTime := time.Now().Add(72 * time.Hour)
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
	// set cookies
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    tokenString,
		Expires:  expirationTime,
		MaxAge:   int(time.Duration(72 * time.Hour).Seconds()),
		Path:     "/",
		HttpOnly: true,
		Secure:   usingSSL,
		SameSite: http.SameSiteLaxMode,
	}) // #nosec G124
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
