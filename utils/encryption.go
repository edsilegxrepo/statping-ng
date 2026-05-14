package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"

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
