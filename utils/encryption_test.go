package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test key for testing (64-char hex = 32 bytes)
const testKeyHex = "4f7e2d9a3b1c4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f"

func TestGenerateMasterKey(t *testing.T) {
	key, err := GenerateMasterKey()
	require.NoError(t, err)
	assert.Len(t, key, 64, "master key should be 64-char hex string")

	// Verify it's valid hex
	for _, c := range key {
		assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'),
			"key should only contain hex characters")
	}

	// Generate another and verify they're different
	key2, err := GenerateMasterKey()
	require.NoError(t, err)
	assert.NotEqual(t, key, key2, "generated keys should be unique")
}

func TestEncryptDecryptWithKey(t *testing.T) {
	testCases := []struct {
		name      string
		plaintext string
	}{
		{"simple", "hello world"},
		{"with spaces", "hello world with spaces"},
		{"with special chars", "p@ssw0rd!#$%^&*()"},
		{"unicode", "日本語テスト"},
		{"long text", strings.Repeat("a", 1000)},
		{"empty", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encrypted, err := EncryptWithKey(tc.plaintext, testKeyHex)
			if tc.plaintext == "" {
				assert.NoError(t, err)
				assert.Empty(t, encrypted)
				return
			}

			require.NoError(t, err)
			assert.True(t, strings.HasPrefix(encrypted, "v1:gcm:"),
				"encrypted value should have v1:gcm: prefix")
			assert.NotEqual(t, tc.plaintext, encrypted)

			decrypted, err := DecryptWithKey(encrypted, testKeyHex)
			require.NoError(t, err)
			assert.Equal(t, tc.plaintext, decrypted)
		})
	}
}

func TestDecryptPlaintextPassthrough(t *testing.T) {
	// Set up master key for this test
	os.Setenv("STATPING_MASTER_KEY", testKeyHex)
	defer os.Unsetenv("STATPING_MASTER_KEY")
	ResetMasterKey()
	require.NoError(t, InitMasterKey())
	defer ZeroMasterKey()

	// Plaintext without prefix should pass through unchanged
	plaintext := "not-encrypted-value"
	result, err := Decrypt(plaintext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, result)
}

func TestIsEncrypted(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"v1:gcm:ABC123", true},
		{"enc:ABC123", true},
		{"plaintext", false},
		{"v1:ABC123", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsEncrypted(tc.input))
		})
	}
}

func TestEncryptWithoutMasterKey(t *testing.T) {
	// Ensure no master key is set
	ResetMasterKey()

	_, err := Encrypt("test")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrMasterKeyNotInitialized)
}

func TestDecryptWithoutMasterKey(t *testing.T) {
	// Ensure no master key is set
	ResetMasterKey()

	// Encrypted value should fail
	_, err := Decrypt("v1:gcm:someencryptedvalue")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrMasterKeyNotInitialized)

	// Plaintext should still work (passthrough)
	result, err := Decrypt("plaintext")
	assert.NoError(t, err)
	assert.Equal(t, "plaintext", result)
}

func TestInitMasterKeyFromEnv(t *testing.T) {
	// Clean state
	ResetMasterKey()

	os.Setenv("STATPING_MASTER_KEY", testKeyHex)
	defer os.Unsetenv("STATPING_MASTER_KEY")

	err := InitMasterKey()
	require.NoError(t, err)
	assert.True(t, MasterKeyInitialized())

	// Test encrypt/decrypt works
	encrypted, err := Encrypt("secret")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(encrypted, "v1:gcm:"))

	decrypted, err := Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, "secret", decrypted)

	ZeroMasterKey()
	assert.False(t, MasterKeyInitialized())
}

func TestInitMasterKeyFromFile(t *testing.T) {
	// Skip on Windows because libsecsecrets blocks temp directories
	if os.Getenv("OS") == "Windows_NT" || filepath.Separator == '\\' {
		t.Skip("Skipping file-based key test on Windows (temp dir restrictions)")
	}

	// Clean state
	ResetMasterKey()

	// Create temp key file
	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "master.key")
	err := os.WriteFile(keyFile, []byte(testKeyHex), 0400)
	require.NoError(t, err)

	os.Setenv("STATPING_MASTER_KEY_FILE", keyFile)
	defer os.Unsetenv("STATPING_MASTER_KEY_FILE")

	err = InitMasterKey()
	require.NoError(t, err)
	assert.True(t, MasterKeyInitialized())

	// Test encrypt/decrypt works
	encrypted, err := Encrypt("secret-from-file")
	require.NoError(t, err)

	decrypted, err := Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, "secret-from-file", decrypted)

	ZeroMasterKey()
}

func TestInitMasterKeyNotFound(t *testing.T) {
	// Clean state
	ResetMasterKey()

	// Ensure no key sources
	os.Unsetenv("STATPING_MASTER_KEY")
	os.Unsetenv("STATPING_MASTER_KEY_FILE")

	// Save and restore Directory
	oldDir := Directory
	Directory = t.TempDir()
	defer func() { Directory = oldDir }()

	err := InitMasterKey()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrMasterKeyNotFound)
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	// Set up master key
	os.Setenv("STATPING_MASTER_KEY", testKeyHex)
	defer os.Unsetenv("STATPING_MASTER_KEY")
	ResetMasterKey()
	require.NoError(t, InitMasterKey())
	defer ZeroMasterKey()

	secrets := []string{
		"password123",
		"oauth-client-secret-xyz",
		"api-key-with-special-chars!@#$%",
		"ldap-bind-password",
	}

	for _, secret := range secrets {
		encrypted, err := Encrypt(secret)
		require.NoError(t, err)
		assert.True(t, IsEncrypted(encrypted))

		decrypted, err := Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, secret, decrypted)
	}
}

func TestDifferentKeysCannotDecrypt(t *testing.T) {
	key1 := "4f7e2d9a3b1c4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f"
	key2 := "1111111111111111111111111111111111111111111111111111111111111111"

	encrypted, err := EncryptWithKey("secret", key1)
	require.NoError(t, err)

	_, err = DecryptWithKey(encrypted, key2)
	assert.Error(t, err, "decryption with wrong key should fail")
}

