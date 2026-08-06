package utils

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"

	"github.com/edsilegxrepo/secretprotector/pkg/libsecsecrets"
	"golang.org/x/crypto/bcrypt"
)

// Encrypted value prefix for libsecsecrets format
const encryptedPrefixV1 = "v1:gcm:"

var (
	// ErrMasterKeyNotInitialized is returned when encryption is attempted without a master key
	ErrMasterKeyNotInitialized = errors.New("master key not initialized - run 'statping init' first")

	// ErrMasterKeyNotFound is returned when no master key source is available
	ErrMasterKeyNotFound = errors.New("master key not found in STATPING_MASTER_KEY or STATPING_MASTER_KEY_FILE")
)

var (
	masterKey     []byte
	masterKeyMu   sync.RWMutex
	masterKeyOnce sync.Once
	masterKeyErr  error
)

// InitMasterKey resolves and stores the master key at startup.
// Priority: STATPING_MASTER_KEY env > STATPING_MASTER_KEY_FILE env > $STATPING_DIR/master.key
// Returns an error if no master key is found (mandatory, no fallback).
func InitMasterKey() error {
	masterKeyOnce.Do(func() {
		ctx := context.Background()

		// Determine file path: from env or default location
		keyFile := os.Getenv("STATPING_MASTER_KEY_FILE")
		if keyFile == "" {
			defaultPath := filepath.Join(Directory, "master.key")
			if FileExists(defaultPath) {
				keyFile = defaultPath
			}
		}

		key, err := libsecsecrets.ResolveKey(ctx,
			"",                    // no direct key flag
			"STATPING_MASTER_KEY", // env var name
			keyFile,               // file path
		)
		if err != nil {
			if errors.Is(err, libsecsecrets.ErrNoKeySource) {
				masterKeyErr = ErrMasterKeyNotFound
			} else {
				masterKeyErr = fmt.Errorf("failed to resolve master key: %w", err)
			}
			return
		}

		masterKeyMu.Lock()
		masterKey = key
		masterKeyMu.Unlock()
	})
	return masterKeyErr
}

// MasterKeyInitialized returns true if the master key has been successfully initialized
func MasterKeyInitialized() bool {
	masterKeyMu.RLock()
	defer masterKeyMu.RUnlock()
	return masterKey != nil
}

// GetMasterKey returns a copy of the master key for use by healthchecker probes.
// Returns nil if master key is not initialized.
func GetMasterKey() []byte {
	masterKeyMu.RLock()
	defer masterKeyMu.RUnlock()
	if masterKey == nil {
		return nil
	}
	keyCopy := make([]byte, len(masterKey))
	copy(keyCopy, masterKey)
	return keyCopy
}

// ZeroMasterKey clears the master key from memory (call on shutdown)
func ZeroMasterKey() {
	masterKeyMu.Lock()
	defer masterKeyMu.Unlock()
	if masterKey != nil {
		libsecsecrets.ZeroBuffer(masterKey)
		masterKey = nil
	}
}

// ResetMasterKey resets the master key state (for testing only)
func ResetMasterKey() {
	masterKeyMu.Lock()
	defer masterKeyMu.Unlock()
	if masterKey != nil {
		libsecsecrets.ZeroBuffer(masterKey)
		masterKey = nil
	}
	// Reset sync.Once inside the lock to prevent race conditions
	masterKeyOnce = sync.Once{}
	masterKeyErr = nil
}

// GenerateMasterKey generates a new 256-bit master key and returns it as a hex string
func GenerateMasterKey() (string, error) {
	return libsecsecrets.GenerateKey()
}

// Encrypt encrypts plaintext using the master key with AES-256-GCM.
// Returns a v1:gcm: prefixed ciphertext.
func Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	masterKeyMu.RLock()
	key := masterKey
	masterKeyMu.RUnlock()

	if key == nil {
		return "", ErrMasterKeyNotInitialized
	}

	ctx := context.Background()
	ciphertext, err := libsecsecrets.Encrypt(ctx, plaintext, key)
	if err != nil {
		return "", fmt.Errorf("encryption failed: %w", err)
	}

	return encryptedPrefixV1 + ciphertext, nil
}

// Decrypt decrypts ciphertext using the master key.
// Handles both v1:gcm: (new) and enc: (legacy) prefixes.
// Plaintext values without a prefix are returned as-is for backward compatibility.
func Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Plaintext passthrough (no prefix)
	if !strings.HasPrefix(ciphertext, encryptedPrefixV1) {
		return ciphertext, nil
	}

	masterKeyMu.RLock()
	key := masterKey
	masterKeyMu.RUnlock()

	if key == nil {
		return "", ErrMasterKeyNotInitialized
	}

	ctx := context.Background()

	// Handle v1:gcm: prefix (new format)
	if strings.HasPrefix(ciphertext, encryptedPrefixV1) {
		encoded := strings.TrimPrefix(ciphertext, encryptedPrefixV1)
		plaintext, err := libsecsecrets.Decrypt(ctx, encoded, key)
		if err != nil {
			return "", fmt.Errorf("decryption failed: %w", err)
		}
		return plaintext, nil
	}

	return ciphertext, nil
}

// EncryptWithKey encrypts plaintext using a provided key (for testing or migration)
func EncryptWithKey(plaintext string, keyHex string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	key, err := parseKeyHex(keyHex)
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	ciphertext, err := libsecsecrets.Encrypt(ctx, plaintext, key)
	if err != nil {
		return "", fmt.Errorf("encryption failed: %w", err)
	}

	return encryptedPrefixV1 + ciphertext, nil
}

// DecryptWithKey decrypts ciphertext using a provided key (for testing or migration)
func DecryptWithKey(ciphertext string, keyHex string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	key, err := parseKeyHex(keyHex)
	if err != nil {
		return "", err
	}

	ctx := context.Background()

	// Strip prefix and decrypt
	if strings.HasPrefix(ciphertext, encryptedPrefixV1) {
		encoded := strings.TrimPrefix(ciphertext, encryptedPrefixV1)
		plaintext, err := libsecsecrets.Decrypt(ctx, encoded, key)
		if err != nil {
			return "", fmt.Errorf("decryption failed: %w", err)
		}
		return plaintext, nil
	}

	// No prefix - return as-is
	return ciphertext, nil
}

// parseKeyHex parses a 64-char hex key or 32-byte raw key into bytes
func parseKeyHex(keyHex string) ([]byte, error) {
	ctx := context.Background()
	return libsecsecrets.ResolveKey(ctx, keyHex, "", "")
}

// IsEncrypted checks if a string is encrypted (has v1:gcm: prefix)
func IsEncrypted(s string) bool {
	return strings.HasPrefix(s, encryptedPrefixV1)
}

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
// Uses rejection sampling to avoid modulo bias
func RandomString(n int) string {
	result, err := RandomStringSecure(n)
	if err != nil {
		// Panic on crypto/rand failure - this indicates a severe system issue
		panic("crypto/rand failed: " + err.Error())
	}
	return result
}

// RandomStringSecure generates a random string of n length using crypto/rand
// Returns an error if crypto/rand fails. Uses rejection sampling to avoid modulo bias.
func RandomStringSecure(n int) (string, error) {
	res := make([]byte, n)
	charLen := len(characterRunes)

	// Calculate the maximum value that gives uniform distribution
	// For 62 characters: 256 % 62 = 8, so we reject values >= 256 - 8 = 248
	maxValid := 256 - (256 % charLen)

	for i := 0; i < n; {
		b := make([]byte, 1)
		if _, err := rand.Read(b); err != nil {
			return "", err
		}
		// Rejection sampling: discard values that would cause bias
		if int(b[0]) < maxValid {
			res[i] = characterRunes[int(b[0])%charLen]
			i++
		}
	}
	return string(res), nil
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
