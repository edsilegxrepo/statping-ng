package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// encryptedPrefix marks a value as encrypted to avoid heuristic detection
const encryptedPrefix = "enc:"

// HashPassword returns the bcrypt hash of a password string
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		Log.Errorf("bcrypt hash failed: %v", err)
		return ""
	}
	return string(bytes)
}

// CheckHash returns true if the password matches with a hashed bcrypt password
func CheckHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// NewSHA256Hash returns a random SHA256 hash.
// Panics if crypto/rand fails (indicates severe system issue).
func NewSHA256Hash() string {
	hash, err := GenerateSHA256Hash()
	if err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return hash
}

// GenerateSHA256Hash returns a random SHA256 hash or an error.
// Use this when you need to handle the error gracefully instead of panicking.
func GenerateSHA256Hash() (string, error) {
	d := make([]byte, 32)
	if _, err := rand.Read(d); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(d)), nil
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

// Encrypt encrypts plaintext using AES-256-GCM with the given key.
// The key is hashed with SHA-256 to ensure it's exactly 32 bytes.
// Returns base64-encoded ciphertext with nonce prepended.
func Encrypt(plaintext, key string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	keyHash := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(keyHash[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return encryptedPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64-encoded ciphertext using AES-256-GCM with the given key.
// The key is hashed with SHA-256 to ensure it's exactly 32 bytes.
func Decrypt(ciphertext, key string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Strip the encrypted prefix if present
	ciphertext = strings.TrimPrefix(ciphertext, encryptedPrefix)

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	keyHash := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(keyHash[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(data) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce := data[:gcm.NonceSize()]
	ciphertextBytes := data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// IsEncrypted checks if a string is encrypted by looking for the encrypted prefix
func IsEncrypted(s string) bool {
	return strings.HasPrefix(s, encryptedPrefix)
}
