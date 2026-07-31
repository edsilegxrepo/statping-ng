# Encryption Architecture

Statping-ng uses a **single external master key** for all secret encryption. This ensures that sensitive data (passwords, tokens, API keys) is never stored in plaintext, and the encryption key is never stored alongside the encrypted data.

## Design Principles

1. **External Key Storage**: Master key stored outside the application (environment variable or protected file)
2. **Single Key**: One master key for all encryption (config.yml and database fields)
3. **Zero Plaintext**: Secrets never written to disk or logs in plaintext
4. **Memory Safety**: Key buffers zeroed immediately after use
5. **Mandatory Security**: No master key = application refuses to start (no fallback)

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                   EXTERNAL MASTER KEY                            │
│                                                                  │
│  Stored OUTSIDE the system:                                      │
│  - ENV: STATPING_MASTER_KEY (Docker/K8s secrets, CI/CD)         │
│  - File: $STATPING_DIR/master.key or custom path (chmod 0400)   │
│  - Future: HashiCorp Vault, AWS KMS, Azure Key Vault            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ Same key for ALL encryption
                              ▼
        ┌─────────────────────┴─────────────────────┐
        │                                           │
        ▼                                           ▼
┌───────────────────┐                    ┌───────────────────┐
│   config.yml      │                    │   Database        │
│                   │                    │                   │
│ • db_password     │                    │ • OAuth secrets   │
│ • api_secret      │                    │ • Notifier tokens │
│ • log_ship_token  │                    │ • LDAP password   │
│ • admin_password  │                    │ • SMTP password   │
└───────────────────┘                    └───────────────────┘
     Decrypted at                         Encrypted/decrypted
     startup BEFORE                       AFTER DB connection
     DB connection
```

## Implementation Design

### Dependency

Use `secretprotector/pkg/libsecsecrets` as a Go module dependency (no code duplication):

```go
import "github.com/your-org/secretprotector/pkg/libsecsecrets"
```

### Abstraction Layer

Keep `utils/encryption.go` as the statping interface, internally calling libsecsecrets:

```
┌─────────────────────────────────────────────────────────────────┐
│                     Statping Codebase                            │
│                                                                  │
│   handlers/         types/           notifiers/                  │
│      │                 │                  │                      │
│      └────────────────┼──────────────────┘                      │
│                       │                                          │
│                       ▼                                          │
│            ┌─────────────────────┐                              │
│            │  utils/encryption.go │  ◄── Statping's interface   │
│            │                      │                              │
│            │  • Encrypt(string)   │      Same API as before     │
│            │  • Decrypt(string)   │      Existing code unchanged │
│            │  • InitMasterKey()   │                              │
│            │  • ZeroMasterKey()   │                              │
│            └──────────┬──────────┘                              │
│                       │                                          │
└───────────────────────┼──────────────────────────────────────────┘
                        │
                        ▼
            ┌─────────────────────┐
            │    libsecsecrets    │  ◄── Actual crypto implementation
            │                     │
            │  • ResolveKey()     │      AES-256-GCM
            │  • Encrypt()        │      Key resolution
            │  • Decrypt()        │      Memory safety
            │  • ZeroBuffer()     │      Platform security checks
            └─────────────────────┘
```

**Benefits of this abstraction:**
- Single point of change if we swap crypto libraries
- Consistent API for the rest of statping
- Existing code keeps calling `utils.Encrypt()` / `utils.Decrypt()`
- Under the hood: libsecsecrets does the actual AES-256-GCM work

### Master Key Lifecycle

```go
// Package-level master key (initialized once at startup)
var (
    masterKey     []byte
    masterKeyOnce sync.Once
)

// InitMasterKey resolves and stores the master key at startup
func InitMasterKey() error {
    var initErr error
    masterKeyOnce.Do(func() {
        ctx := context.Background()
        key, err := libsecsecrets.ResolveKey(ctx, 
            "",                              // no direct key flag
            "STATPING_MASTER_KEY",           // env var name
            getMasterKeyFilePath(),          // file path
        )
        if err != nil {
            initErr = fmt.Errorf("master key required: %w", err)
            return
        }
        masterKey = key
    })
    return initErr
}

// ZeroMasterKey clears the master key from memory (call on shutdown)
func ZeroMasterKey() {
    if masterKey != nil {
        libsecsecrets.ZeroBuffer(masterKey)
        masterKey = nil
    }
}

// Encrypt encrypts plaintext using the master key
func Encrypt(plaintext string) (string, error) {
    if masterKey == nil {
        return "", errors.New("master key not initialized")
    }
    return libsecsecrets.Encrypt(context.Background(), plaintext, masterKey)
}

// Decrypt decrypts ciphertext using the master key
func Decrypt(ciphertext string) (string, error) {
    if masterKey == nil {
        return "", errors.New("master key not initialized")
    }
    // Auto-detect encrypted values by prefix
    if !strings.HasPrefix(ciphertext, "v1:gcm:") {
        return ciphertext, nil  // plaintext passthrough for backward compat
    }
    return libsecsecrets.Decrypt(context.Background(), ciphertext, masterKey)
}
```

## Master Key Resolution

Priority order (first found wins):

1. **Direct**: `STATPING_MASTER_KEY` environment variable (64-char hex string)
2. **File**: `STATPING_MASTER_KEY_FILE` environment variable pointing to key file
3. **Default File**: `$STATPING_DIR/master.key` if it exists

If no master key is found, the application will **refuse to start**. This is mandatory with no fallback - security is not optional.

## Encrypted Value Format

Encrypted values use a versioned prefix format:

```
v1:gcm:<base64-encoded-nonce+ciphertext>
```

- `v1`: Version identifier (allows future algorithm changes)
- `gcm`: Algorithm identifier (AES-256-GCM)
- `<base64>`: Base64-encoded 12-byte nonce + ciphertext + 16-byte auth tag

Example in config.yml:
```yaml
db_password: "v1:gcm:A1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0..."
```

Plaintext values (no prefix) are used as-is for backward compatibility.

## Cryptographic Standards

- **Algorithm**: AES-256-GCM (Galois/Counter Mode)
- **Key Size**: 256 bits (32 bytes, represented as 64-char hex string)
- **Nonce Size**: 96 bits (12 bytes, CSPRNG-generated per encryption)
- **Authentication**: GCM provides authenticated encryption (detects tampering)

## Startup Sequence

```
1. Run `statping init` (pre-flight checks)
   - Verify master key is configured (ENV or file)
   - Validate key format (64-char hex)
   - Check file permissions (if using file)
   - Fail with clear instructions if anything missing

2. Application startup
   - Call InitMasterKey() - fails if no key
   - Read config.yml
   - Decrypt "v1:gcm:..." values in memory
   - Connect to database
   - For DB secrets: decrypt on read, encrypt on write

3. Shutdown
   - Call ZeroMasterKey() to clear from memory
```

## CLI Commands

### `statping init`

Pre-flight check that verifies all dependencies. If no master key is found and `--key-file` is specified, it generates one and writes it to the file with secure permissions.

**Key found in ENV:**
```bash
$ statping init

Checking dependencies...
✓ Master key: found in STATPING_MASTER_KEY
✓ Key format: valid 64-character hex string

Ready to start.
```

**Key found via file reference:**
```bash
$ export STATPING_MASTER_KEY_FILE="/etc/statping/master.key"
$ statping init

Checking dependencies...
✓ Master key: found in /etc/statping/master.key (via STATPING_MASTER_KEY_FILE)
✓ Key format: valid 64-character hex string
✓ File permissions: 0400

Ready to start.
```

**No key found, no --key-file specified:**
```bash
$ statping init

Checking dependencies...
✗ Master key: NOT FOUND

No master key configured. To generate one:
  statping init --key-file /path/to/master.key

Or set STATPING_MASTER_KEY environment variable directly.
```

**Generate and store key (first-time setup):**
```bash
# Linux/macOS
$ statping init --key-file /etc/statping/master.key

Checking dependencies...
✗ Master key: NOT FOUND in STATPING_MASTER_KEY or STATPING_MASTER_KEY_FILE

Generating new master key...
✓ Generated 256-bit key
✓ Written to /etc/statping/master.key
✓ Permissions set to 0400

Add to your environment:
  export STATPING_MASTER_KEY_FILE="/etc/statping/master.key"

# Windows
$ statping init --key-file C:\Users\admin\.statping\master.key

Checking dependencies...
✗ Master key: NOT FOUND in STATPING_MASTER_KEY or STATPING_MASTER_KEY_FILE

Generating new master key...
✓ Generated 256-bit key
✓ Written to C:\Users\admin\.statping\master.key

Add to your environment:
  $env:STATPING_MASTER_KEY_FILE = "C:\Users\admin\.statping\master.key"
```

**Key file already exists (safety check):**
```bash
$ statping init --key-file /etc/statping/master.key

Error: /etc/statping/master.key already exists.
Use --force to overwrite (DANGER: will invalidate all encrypted data).
```

**Container/CI setup (env var):**
```bash
# For containers, generate key externally and inject via env var:
$ secretprotector -generate
4f7e2d9a3b1c4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f

# Docker
docker run -e STATPING_MASTER_KEY="4f7e2d9a..." statping/statping-ng

# Kubernetes
kubectl create secret generic statping-secrets \
  --from-literal=master-key="4f7e2d9a..."
```

### Encrypting Secrets with secretprotector

```bash
# Generate a master key
secretprotector -generate
# Output: 4f7e2d9a3b1c4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f

# Store the master key
export STATPING_MASTER_KEY="4f7e2d9a3b1c4e5f..."
# OR
echo "4f7e2d9a3b1c4e5f..." > /etc/statping/master.key
chmod 0400 /etc/statping/master.key

# Encrypt a secret
secretprotector -encrypt "MyDatabasePassword123"
# Output: v1:gcm:A1B2C3D4...

# Add to config.yml
db_password: "v1:gcm:A1B2C3D4..."
```

## Testing Strategy

### Unit Tests

```go
func TestEncryptDecryptRoundTrip(t *testing.T) {
    // Mock key resolution
    testKey := make([]byte, 32)
    rand.Read(testKey)
    
    plaintext := "secret-password-123"
    
    ciphertext, err := EncryptWithKey(plaintext, testKey)
    require.NoError(t, err)
    assert.True(t, strings.HasPrefix(ciphertext, "v1:gcm:"))
    
    decrypted, err := DecryptWithKey(ciphertext, testKey)
    require.NoError(t, err)
    assert.Equal(t, plaintext, decrypted)
}

func TestDecryptPlaintextPassthrough(t *testing.T) {
    // Plaintext (no prefix) should pass through unchanged
    plaintext := "not-encrypted"
    result, err := Decrypt(plaintext)
    require.NoError(t, err)
    assert.Equal(t, plaintext, result)
}

func TestNoMasterKeyFails(t *testing.T) {
    // Without InitMasterKey, operations should fail
    _, err := Encrypt("test")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "master key not initialized")
}
```

### Integration Tests

```go
func TestIntegrationWithRealKey(t *testing.T) {
    // Set up real master key
    os.Setenv("STATPING_MASTER_KEY", "4f7e2d9a3b1c4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f")
    defer os.Unsetenv("STATPING_MASTER_KEY")
    
    err := InitMasterKey()
    require.NoError(t, err)
    defer ZeroMasterKey()
    
    // Full round-trip
    secret := "database-password-prod-123!"
    encrypted, err := Encrypt(secret)
    require.NoError(t, err)
    
    decrypted, err := Decrypt(encrypted)
    require.NoError(t, err)
    assert.Equal(t, secret, decrypted)
}

func TestKeyFileResolution(t *testing.T) {
    // Create temp key file
    keyFile := filepath.Join(t.TempDir(), "master.key")
    os.WriteFile(keyFile, []byte("4f7e2d9a3b1c4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f"), 0400)
    
    os.Setenv("STATPING_MASTER_KEY_FILE", keyFile)
    defer os.Unsetenv("STATPING_MASTER_KEY_FILE")
    
    err := InitMasterKey()
    require.NoError(t, err)
    defer ZeroMasterKey()
    
    // Verify encryption works
    _, err = Encrypt("test")
    require.NoError(t, err)
}
```

## Security Considerations

### Key File Protection

**Linux/Unix:**
```bash
chmod 0400 /etc/statping/master.key
chown statping:statping /etc/statping/master.key
```

**Windows:**
- Store in user profile directory (not Public or Temp)
- Use NTFS permissions to restrict access

### Environment Variable Security

- Use Docker/Kubernetes secrets to inject `STATPING_MASTER_KEY`
- Never log or print the master key
- Never commit the master key to version control

### Memory Safety

- Master key buffer zeroed on shutdown via `ZeroMasterKey()`
- Decrypted secrets kept in memory only (never written to disk)
- libsecsecrets uses `ZeroBuffer()` for sensitive byte slices

## Startup Behavior

| Scenario | Behavior |
|----------|----------|
| Master key set, encrypted values | Decrypt and use |
| Master key set, plaintext values | Use as-is (backward compatible) |
| No master key | **Fail to start** (mandatory, no fallback) |

The master key is **mandatory**. There is no insecure fallback mode.

## Future Enhancements

- **Key Rotation**: Support for multiple key versions, gradual re-encryption
- **Key Vault Integration**: HashiCorp Vault, AWS KMS, Azure Key Vault, GCP KMS
- **Hardware Security**: TPM/HSM support for key storage
- **Envelope Encryption**: Data keys encrypted by master key (for large data)

## Dependencies

This implementation uses `libsecsecrets` from the `secretprotector` project:
- Pure Go standard library (no external dependencies)
- AES-256-GCM authenticated encryption
- Platform-aware security checks (file permissions)
- Memory-safe buffer handling

Reference as Go module dependency (no code duplication):
```go
import "github.com/your-org/secretprotector/pkg/libsecsecrets"
```
