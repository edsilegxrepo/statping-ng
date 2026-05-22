package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword returns the bcrypt hash of a password string
func HashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes)
}

// CheckHash returns true if the password matches with a hashed bcrypt password
func CheckHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// NewSHA256Hash returns a random SHA256 hash
func NewSHA256Hash() string {
	d := make([]byte, 32)
	if _, err := rand.Read(d); err != nil {
		return ""
	}
	return fmt.Sprintf("%x", sha256.Sum256(d))
}

// Sha256Hash returns a SHA256 hash of a string
func Sha256Hash(val string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(val)))
}

var characterRunes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandomString generates a random string of n length using crypto/rand
func RandomString(n int) string {
	res := make([]byte, n)
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	for i := range b {
		res[i] = characterRunes[int(b[i])%len(characterRunes)]
	}
	return string(res)
}

// IsHash returns true if the string is already a bcrypt hash
func IsHash(password string) bool {
	return strings.HasPrefix(password, "$2a$") || strings.HasPrefix(password, "$2b$") || strings.HasPrefix(password, "$2y$")
}

// ComplexityCheck returns true if the password meets the complexity requirements:
// 30 characters minimum, with upper, lower, and digits.
func ComplexityCheck(password string) bool {
	if len(password) < 30 {
		return false
	}
	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		if unicode.IsUpper(r) {
			hasUpper = true
		} else if unicode.IsLower(r) {
			hasLower = true
		} else if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	return hasUpper && hasLower && hasDigit
}
