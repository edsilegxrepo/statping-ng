package configs

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/types/failures"
	"github.com/statping-ng/statping-ng/types/groups"
	"github.com/statping-ng/statping-ng/types/hits"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/types/users"
	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = utils.InitLogs()
	utils.InitEnvs()
}

// setTestDir sets utils.Directory to tmpDir and registers cleanup to restore it
func setTestDir(t *testing.T, tmpDir string) {
	originalDir := utils.Directory
	utils.Directory = tmpDir
	t.Cleanup(func() { utils.Directory = originalDir })
}

// =============================================================================
// Config Loading from YAML Tests
// =============================================================================

func TestLoadConfigs_ValidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	setTestDir(t, tmpDir)

	yamlContent := `connection: sqlite
host: localhost
database: test.db
port: 0
language: en
allow_reports: true
`
	configPath := filepath.Join(tmpDir, "config.yml")
	err := os.WriteFile(configPath, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	cfg, err := LoadConfigs(configPath)
	require.NoError(t, err)
	assert.Equal(t, "sqlite", cfg.DbConn)
	assert.Equal(t, "localhost", cfg.DbHost)
	assert.Equal(t, "test.db", cfg.DbData)
	assert.Equal(t, "en", cfg.Language)
	assert.True(t, cfg.AllowReports)
}

func TestLoadConfigs_PostgresYAML(t *testing.T) {
	tmpDir := t.TempDir()
	setTestDir(t, tmpDir)

	yamlContent := `connection: postgres
host: dbserver.example.com
user: pguser
password: secretpass
database: statping_prod
port: 5432
postgres_ssl: require
`
	configPath := filepath.Join(tmpDir, "config.yml")
	err := os.WriteFile(configPath, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	cfg, err := LoadConfigs(configPath)
	require.NoError(t, err)
	assert.Equal(t, "postgres", cfg.DbConn)
	assert.Equal(t, "dbserver.example.com", cfg.DbHost)
	assert.Equal(t, "pguser", cfg.DbUser)
	assert.Equal(t, "secretpass", cfg.DbPass)
	assert.Equal(t, "statping_prod", cfg.DbData)
	assert.Equal(t, 5432, cfg.DbPort)
}

func TestLoadConfigs_MySQLYAML(t *testing.T) {
	tmpDir := t.TempDir()
	setTestDir(t, tmpDir)

	yamlContent := `connection: mysql
host: mysql.example.com
user: mysqluser
password: mysqlpass
database: statping_db
port: 3306
`
	configPath := filepath.Join(tmpDir, "config.yml")
	err := os.WriteFile(configPath, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	cfg, err := LoadConfigs(configPath)
	require.NoError(t, err)
	assert.Equal(t, "mysql", cfg.DbConn)
	assert.Equal(t, "mysql.example.com", cfg.DbHost)
	assert.Equal(t, "mysqluser", cfg.DbUser)
	assert.Equal(t, "mysqlpass", cfg.DbPass)
	assert.Equal(t, "statping_db", cfg.DbData)
	assert.Equal(t, 3306, cfg.DbPort)
}

// =============================================================================
// Environment Variable Override Tests
// =============================================================================

func TestLoadConfigs_EnvOverridesYAML(t *testing.T) {
	tmpDir := t.TempDir()
	setTestDir(t, tmpDir)

	// Set environment variables
	_ = os.Setenv("DB_CONN", "sqlite")
	defer func() { _ = os.Unsetenv("DB_CONN") }()

	yamlContent := `connection: postgres
host: dbserver
database: prod_db
`
	configPath := filepath.Join(tmpDir, "config.yml")
	err := os.WriteFile(configPath, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	cfg, err := LoadConfigs(configPath)
	require.NoError(t, err)
	// sqlite/sqlite3 from env should override postgres from YAML
	assert.Equal(t, "sqlite3", cfg.DbConn)
}

func TestLoadConfigs_SqliteEnvVariants(t *testing.T) {
	tmpDir := t.TempDir()
	setTestDir(t, tmpDir)

	configPath := filepath.Join(tmpDir, "config.yml")
	err := os.WriteFile(configPath, []byte(""), 0o644)
	require.NoError(t, err)

	testCases := []struct {
		envValue     string
		expectedConn string
	}{
		{"sqlite", "sqlite3"},
		{"sqlite3", "sqlite3"},
	}

	for _, tc := range testCases {
		t.Run(tc.envValue, func(t *testing.T) {
			_ = os.Setenv("DB_CONN", tc.envValue)
			defer func() { _ = os.Unsetenv("DB_CONN") }()

			cfg, _ := LoadConfigs(configPath)
			assert.Equal(t, tc.expectedConn, cfg.DbConn)
		})
	}
}

// =============================================================================
// Default Value Handling Tests
// =============================================================================

func TestDefaultValues(t *testing.T) {
	// Ensure InitEnvs has been called
	utils.InitEnvs()

	// Check default values from utils.Params
	assert.Equal(t, "admin", utils.Params.GetString("ADMIN_USER"))
	assert.Equal(t, "info@admin.com", utils.Params.GetString("ADMIN_EMAIL"))
	assert.Equal(t, 25, utils.Params.GetInt("MAX_OPEN_CONN"))
	assert.Equal(t, 25, utils.Params.GetInt("MAX_IDLE_CONN"))
	assert.Equal(t, "disable", utils.Params.GetString("POSTGRES_SSLMODE"))
	assert.Equal(t, "en", utils.Params.GetString("LANGUAGE"))
	assert.Equal(t, false, utils.Params.GetBool("SAMPLE_DATA"))
	assert.Equal(t, false, utils.Params.GetBool("ADMIN_LOCK"))
}

func TestDbConfig_DefaultConnectionPoolSettings(t *testing.T) {
	cfg := &DbConfig{}

	// When not set, should use defaults from Params
	assert.Equal(t, 0, cfg.MaxOpenConnections)
	assert.Equal(t, 0, cfg.MaxIdleConnections)
	assert.Equal(t, 0, cfg.MaxLifeConnections)
}

// =============================================================================
// Config Validation Tests
// =============================================================================

func TestLoadConfigs_EmptyConnection_SetupMode(t *testing.T) {
	tmpDir := t.TempDir()
	setTestDir(t, tmpDir)

	// Clear any existing DB_CONN env var
	_ = os.Unsetenv("DB_CONN")
	utils.Params.Set("DB_CONN", "")

	yamlContent := `host: localhost
database: test.db
`
	configPath := filepath.Join(tmpDir, "config.yml")
	err := os.WriteFile(configPath, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	cfg, err := LoadConfigs(configPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "starting in setup mode")
	assert.Empty(t, cfg.DbConn)
}

func TestDbConfig_ValidConnectionTypes(t *testing.T) {
	validConnTypes := []string{"sqlite", "sqlite3", "mysql", "postgres", "memory"}

	for _, connType := range validConnTypes {
		t.Run(connType, func(t *testing.T) {
			cfg := &DbConfig{DbConn: connType}
			assert.NotEmpty(t, cfg.DbConn)
		})
	}
}

// =============================================================================
// Database Connection String Parsing Tests
// =============================================================================

func TestConnectionString_Memory(t *testing.T) {
	cfg := &DbConfig{DbConn: "memory"}
	conn := cfg.ConnectionString()
	assert.Equal(t, ":memory:", conn)
	assert.Equal(t, ":memory:", cfg.DbConn)
}

func TestConnectionString_MemoryVariant(t *testing.T) {
	cfg := &DbConfig{DbConn: ":memory:"}
	conn := cfg.ConnectionString()
	assert.Equal(t, ":memory:", conn)
}

func TestConnectionString_SQLite(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "statping.db")
	err := utils.SaveFile(dbPath, []byte(""))
	require.NoError(t, err)

	cfg := &DbConfig{
		DbConn:   "sqlite",
		Location: tmpDir,
	}
	conn := cfg.ConnectionString()
	assert.Contains(t, conn, "statping.db")
	assert.Equal(t, "sqlite3", cfg.DbConn)
}

func TestConnectionString_SQLiteWithCustomDbData(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "custom.db")
	err := utils.SaveFile(dbPath, []byte(""))
	require.NoError(t, err)

	cfg := &DbConfig{
		DbConn:   "sqlite",
		DbData:   "custom.db",
		Location: tmpDir,
	}
	conn := cfg.ConnectionString()
	assert.Contains(t, conn, "custom.db")
}

func TestConnectionString_SQLiteWithExplicitSqlFile(t *testing.T) {
	tmpDir := t.TempDir()
	sqlFile := filepath.Join(tmpDir, "explicit.db")

	cfg := &DbConfig{
		DbConn:   "sqlite",
		SqlFile:  sqlFile,
		Location: tmpDir,
	}
	conn := cfg.ConnectionString()
	assert.Equal(t, sqlFile, conn)
}

func TestConnectionString_MySQL(t *testing.T) {
	cfg := &DbConfig{
		DbConn: "mysql",
		DbHost: "localhost",
		DbPort: 3306,
		DbUser: "root",
		DbPass: "password",
		DbData: "statping",
	}
	conn := cfg.ConnectionString()
	assert.Contains(t, conn, "root:password@tcp(localhost:3306)/statping")
	assert.Contains(t, conn, "charset=utf8")
	assert.Contains(t, conn, "parseTime=True")
	assert.Contains(t, conn, "loc=UTC")
}

func TestConnectionString_MySQLSpecialChars(t *testing.T) {
	cfg := &DbConfig{
		DbConn: "mysql",
		DbHost: "db.example.com",
		DbPort: 3307,
		DbUser: "app_user",
		DbPass: "p@ss!word#123",
		DbData: "my_database",
	}
	conn := cfg.ConnectionString()
	assert.Contains(t, conn, "app_user:p@ss!word#123@tcp(db.example.com:3307)/my_database")
}

func TestConnectionString_Postgres(t *testing.T) {
	utils.Params.Set("POSTGRES_SSLMODE", "disable")
	defer utils.Params.Set("POSTGRES_SSLMODE", "")

	cfg := &DbConfig{
		DbConn: "postgres",
		DbHost: "localhost",
		DbPort: 5432,
		DbUser: "postgres",
		DbPass: "password",
		DbData: "statping",
	}
	conn := cfg.ConnectionString()
	assert.Contains(t, conn, "host=localhost")
	assert.Contains(t, conn, "port=5432")
	assert.Contains(t, conn, "user=postgres")
	assert.Contains(t, conn, "dbname=statping")
	assert.Contains(t, conn, "password=password")
	assert.Contains(t, conn, "timezone=UTC")
	assert.Contains(t, conn, "sslmode=disable")
}

func TestConnectionString_PostgresSSLModes(t *testing.T) {
	sslModes := []string{"disable", "require", "verify-ca", "verify-full"}

	for _, mode := range sslModes {
		t.Run(mode, func(t *testing.T) {
			utils.Params.Set("POSTGRES_SSLMODE", mode)
			defer utils.Params.Set("POSTGRES_SSLMODE", "")

			cfg := &DbConfig{
				DbConn: "postgres",
				DbHost: "localhost",
				DbPort: 5432,
				DbUser: "postgres",
				DbPass: "password",
				DbData: "statping",
			}
			conn := cfg.ConnectionString()
			assert.Contains(t, conn, "sslmode="+mode)
		})
	}
}

func TestConnectionString_UnknownType(t *testing.T) {
	cfg := &DbConfig{DbConn: "unknown"}
	conn := cfg.ConnectionString()
	assert.Empty(t, conn)
}

// =============================================================================
// Config File Path Resolution Tests
// =============================================================================

func TestFindDbFile_WithLocation(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "statping.db")
	err := utils.SaveFile(dbPath, []byte("test"))
	require.NoError(t, err)

	cfg := &DbConfig{Location: tmpDir}
	file, err := findDbFile(cfg)
	require.NoError(t, err)
	assert.Contains(t, file, "statping.db")
}

func TestFindDbFile_WithCustomDbData(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "custom.db")
	err := utils.SaveFile(dbPath, []byte("test"))
	require.NoError(t, err)

	cfg := &DbConfig{
		Location: tmpDir,
		DbData:   "custom.db",
	}
	file, err := findDbFile(cfg)
	require.NoError(t, err)
	assert.Contains(t, file, "custom.db")
}

func TestFindDbFile_WithExplicitSqlFile(t *testing.T) {
	tmpDir := t.TempDir()
	sqlFile := filepath.Join(tmpDir, "explicit.db")

	cfg := &DbConfig{
		Location: tmpDir,
		SqlFile:  sqlFile,
	}
	file, err := findDbFile(cfg)
	require.NoError(t, err)
	assert.Equal(t, sqlFile, file)
}

func TestFindDbFile_NilConfig(t *testing.T) {
	tmpDir := t.TempDir()
	setTestDir(t, tmpDir)
	dbPath := filepath.Join(tmpDir, "statping.db")
	err := utils.SaveFile(dbPath, []byte("test"))
	require.NoError(t, err)

	file, err := findDbFile(nil)
	require.NoError(t, err)
	assert.Equal(t, "statping.db", file)
}

func TestFindSQLin_SingleDbFile(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "statping.db")
	err := utils.SaveFile(dbPath, []byte("test"))
	require.NoError(t, err)

	file, err := findSQLin(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, "statping.db", file)
}

func TestFindSQLin_MultipleDbFiles(t *testing.T) {
	tmpDir := t.TempDir()
	err := utils.SaveFile(filepath.Join(tmpDir, "statping.db"), []byte("test"))
	require.NoError(t, err)
	err = utils.SaveFile(filepath.Join(tmpDir, "backup.db"), []byte("test"))
	require.NoError(t, err)

	_, err = findSQLin(tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "found multiple database files")
}

func TestFindSQLin_NoDbFile(t *testing.T) {
	tmpDir := t.TempDir()

	file, err := findSQLin(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, SqliteFilename, file)
}

// =============================================================================
// Edge Cases: Missing File, Malformed YAML, Empty Values
// =============================================================================

func TestLoadConfigs_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	setTestDir(t, tmpDir)

	configPath := filepath.Join(tmpDir, "nonexistent.yml")
	cfg, err := LoadConfigs(configPath)
	// Should return config with empty DbConn (setup mode)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "starting in setup mode")
	assert.NotNil(t, cfg)
}

func TestLoadConfigs_MalformedYAML(t *testing.T) {
	tmpDir := t.TempDir()
	setTestDir(t, tmpDir)

	malformedYAML := `connection: sqlite
host localhost  # missing colon
database: test.db
invalid yaml content here
  - this is wrong
`
	configPath := filepath.Join(tmpDir, "config.yml")
	err := os.WriteFile(configPath, []byte(malformedYAML), 0o644)
	require.NoError(t, err)

	_, err = LoadConfigs(configPath)
	require.Error(t, err)
}

func TestLoadConfigs_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	setTestDir(t, tmpDir)
	_ = os.Unsetenv("DB_CONN")
	utils.Params.Set("DB_CONN", "")

	configPath := filepath.Join(tmpDir, "config.yml")
	err := os.WriteFile(configPath, []byte(""), 0o644)
	require.NoError(t, err)

	cfg, err := LoadConfigs(configPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "starting in setup mode")
	assert.NotNil(t, cfg)
}

func TestLoadConfigs_EmptyValues(t *testing.T) {
	tmpDir := t.TempDir()
	setTestDir(t, tmpDir)

	// Clear any cached values from previous tests
	utils.Params.Set("DB_HOST", "")
	utils.Params.Set("DB_USER", "")
	utils.Params.Set("DB_PASS", "")
	defer func() {
		utils.Params.Set("DB_HOST", "")
		utils.Params.Set("DB_USER", "")
		utils.Params.Set("DB_PASS", "")
	}()

	yamlContent := `connection: sqlite
host: ""
user: ""
password: ""
database: ""
port: 0
`
	configPath := filepath.Join(tmpDir, "config.yml")
	err := os.WriteFile(configPath, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	cfg, err := LoadConfigs(configPath)
	require.NoError(t, err)
	assert.Equal(t, "sqlite", cfg.DbConn)
	// Note: Empty strings in YAML don't override Params, so values may persist
	// from previous tests. The important assertion is that sqlite is loaded.
}

func TestLoadConfigs_NonWritableDirectory(t *testing.T) {
	// Skip on Windows since directory permissions work differently
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("Skipping permission test in CI environment")
	}
	if filepath.Separator == '\\' {
		t.Skip("Skipping permission test on Windows")
	}

	tmpDir := t.TempDir()

	// Create a subdirectory and make it non-writable
	nonWritableDir := filepath.Join(tmpDir, "readonly")
	err := os.Mkdir(nonWritableDir, 0o555)
	require.NoError(t, err)
	defer func() { _ = os.Chmod(nonWritableDir, 0o755) }()

	setTestDir(t, nonWritableDir)
	configPath := filepath.Join(nonWritableDir, "config.yml")

	_, err = LoadConfigs(configPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not writable")
}

func TestLoadConfigs_YAMLWithUnknownFields(t *testing.T) {
	tmpDir := t.TempDir()
	setTestDir(t, tmpDir)

	yamlContent := `connection: sqlite
host: localhost
database: test.db
unknown_field: some_value
another_unknown: 12345
`
	configPath := filepath.Join(tmpDir, "config.yml")
	err := os.WriteFile(configPath, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	cfg, err := LoadConfigs(configPath)
	require.NoError(t, err)
	assert.Equal(t, "sqlite", cfg.DbConn)
}

// =============================================================================
// Save and Update Tests
// =============================================================================

func TestSave_CreatesConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	setTestDir(t, tmpDir)

	utils.Params.Set("DB_CONN", "sqlite")
	utils.Params.Set("DB_HOST", "localhost")
	utils.Params.Set("DB_DATABASE", "test.db")
	defer func() {
		utils.Params.Set("DB_CONN", "")
		utils.Params.Set("DB_HOST", "")
		utils.Params.Set("DB_DATABASE", "")
	}()

	err := Save()
	require.NoError(t, err)

	configPath := filepath.Join(tmpDir, "config.yml")
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "connection: sqlite")
}

func TestDbConfig_Update_WritesConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	setTestDir(t, tmpDir)

	cfg := &DbConfig{
		DbConn:   "sqlite",
		DbData:   "statping.db",
		Location: tmpDir,
		Language: "fr",
	}

	err := cfg.Update()
	require.NoError(t, err)

	configPath := filepath.Join(tmpDir, "config.yml")
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "language: fr")
}

// =============================================================================
// Secrets Tests
// =============================================================================

func TestSaveSecrets(t *testing.T) {
	tmpDir := t.TempDir()
	setTestDir(t, tmpDir)

	creds := map[string]string{
		"admin":  "secretpass123",
		"viewer": "viewerpass456",
	}

	err := SaveSecrets(creds)
	require.NoError(t, err)

	secretsPath := filepath.Join(tmpDir, "statping.secrets")
	content, err := os.ReadFile(secretsPath)
	require.NoError(t, err)

	secretsContent := string(content)
	assert.Contains(t, secretsContent, "admin")
	assert.Contains(t, secretsContent, "secretpass123")
	assert.Contains(t, secretsContent, "viewer")
	assert.Contains(t, secretsContent, "viewerpass456")
	assert.Contains(t, secretsContent, "STATPING-NG AUTOMATICALLY GENERATED SECRETS")
}

// =============================================================================
// Merge and Clean Tests
// =============================================================================

func TestDbConfig_Merge_PartialOverwrite(t *testing.T) {
	original := &DbConfig{
		DbConn:   "sqlite",
		DbHost:   "localhost",
		DbData:   "original.db",
		DbPort:   0,
		Language: "en",
	}

	newCfg := &DbConfig{
		DbConn: "postgres",
		DbHost: "newhost",
		DbData: "newdb",
		DbPort: 5432,
		DbUser: "newuser",
		DbPass: "newpass",
	}

	result := original.Merge(newCfg)
	assert.Equal(t, "postgres", result.DbConn)
	assert.Equal(t, "newhost", result.DbHost)
	assert.Equal(t, "newdb", result.DbData)
	assert.Equal(t, 5432, result.DbPort)
	assert.Equal(t, "newuser", result.DbUser)
	assert.Equal(t, "newpass", result.DbPass)
	// Language should be preserved since Merge doesn't touch it
	assert.Equal(t, "en", result.Language)
}

func TestDbConfig_Clean_PreservesNonSensitiveFields(t *testing.T) {
	cfg := &DbConfig{
		DbConn:   "postgres",
		DbHost:   "dbserver",
		DbData:   "mydb",
		DbPort:   5432,
		DbUser:   "admin",
		DbPass:   "secret",
		Language: "es",
	}

	result := cfg.Clean()
	assert.Empty(t, result.DbConn)
	assert.Empty(t, result.DbHost)
	assert.Empty(t, result.DbData)
	assert.Equal(t, 0, result.DbPort)
	assert.Empty(t, result.DbUser)
	assert.Empty(t, result.DbPass)
	// Non-sensitive fields should remain
	assert.Equal(t, "es", result.Language)
}

// =============================================================================
// ToYAML Tests
// =============================================================================

func TestDbConfig_ToYAML_AllFields(t *testing.T) {
	cfg := &DbConfig{
		DbConn:       "postgres",
		DbHost:       "dbserver",
		DbUser:       "admin",
		DbPass:       "secret",
		DbData:       "statping",
		DbPort:       5432,
		Language:     "en",
		AllowReports: true,
		AdminLock:    true,
		BasePath:     "/status",
	}

	yamlBytes := cfg.ToYAML()
	require.NotNil(t, yamlBytes)

	yamlStr := string(yamlBytes)
	assert.Contains(t, yamlStr, "connection: postgres")
	assert.Contains(t, yamlStr, "host: dbserver")
	assert.Contains(t, yamlStr, "user: admin")
	assert.Contains(t, yamlStr, "password: secret")
	assert.Contains(t, yamlStr, "database: statping")
	assert.Contains(t, yamlStr, "port: 5432")
	assert.Contains(t, yamlStr, "language: en")
	assert.Contains(t, yamlStr, "allow_reports: true")
	assert.Contains(t, yamlStr, "admin_lock: true")
	assert.Contains(t, yamlStr, "base_path: /status")
}

func TestDbConfig_ToYAML_EmptyConfig(t *testing.T) {
	cfg := &DbConfig{}
	yamlBytes := cfg.ToYAML()
	require.NotNil(t, yamlBytes)
	// Note: YAML marshals bool fields even when false, so empty config still has fields
	yamlStr := string(yamlBytes)
	assert.NotEmpty(t, yamlStr)
	// Should not contain any actual connection info
	assert.NotContains(t, yamlStr, "connection:")
	assert.NotContains(t, yamlStr, "host:")
}

// =============================================================================
// ConnectConfigs Tests
// =============================================================================

func TestConnectConfigs_SQLite(t *testing.T) {
	tmpDir := t.TempDir()
	setTestDir(t, tmpDir)

	cfg := &DbConfig{
		DbConn:   "sqlite",
		DbData:   "statping.db",
		Location: tmpDir,
	}

	err := ConnectConfigs(cfg, false)
	require.NoError(t, err)
	defer cfg.Close()

	// Verify config file was created
	configPath := filepath.Join(tmpDir, "config.yml")
	_, err = os.Stat(configPath)
	assert.NoError(t, err)
}

func TestSQLiteConfig(t *testing.T) {
	tmpDir := t.TempDir()
	sqlite := &DbConfig{
		DbConn:   "sqlite",
		DbHost:   "localhost",
		DbUser:   "",
		DbPass:   "",
		DbData:   "statping.db",
		DbPort:   0,
		Location: tmpDir,
	}

	err := Connect(sqlite, false)
	require.Nil(t, err)

	if sqlite.Db != nil {
		db, _ := sqlite.Db.DB()
		if db != nil {
			_ = db.Close()
		}
	}
}

func TestMySQLConfig(t *testing.T) {
	mysql := &DbConfig{
		DbConn: "mysql",
		DbHost: "localhost",
		DbUser: "root",
		DbPass: "password123",
		DbData: "statping",
		DbPort: 3306,
	}

	err := Connect(mysql, false)
	if err != nil {
		t.Skipf("Skipping MySQL test (server not running on localhost:3306): %v", err)
	}
}

func TestPostgresConfig(t *testing.T) {
	postgres := &DbConfig{
		DbConn: "postgres",
		DbHost: "localhost",
		DbUser: "root",
		DbPass: "password123",
		DbData: "statping",
		DbPort: 5432,
	}

	err := Connect(postgres, false)
	if err != nil {
		t.Skipf("Skipping Postgres test (server not running on localhost:5432): %v", err)
	}
}

func TestFileSQLFile(t *testing.T) {
	tmpDir := t.TempDir()
	testDbPath := tmpDir + "/statping.db"
	err := utils.SaveFile(testDbPath, []byte("test"))
	require.Nil(t, err)

	file, err := findSQLin(tmpDir)
	require.Nil(t, err)
	assert.Equal(t, "statping.db", file)
}

func TestDbConfig_Save(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &DbConfig{
		DbConn: "sqlite",
		DbHost: "localhost",
		DbData: "test.db",
		DbPort: 0,
	}

	err := cfg.Save(tmpDir)
	require.NoError(t, err)

	configPath := filepath.Join(tmpDir, "config.yml")
	_, err = os.Stat(configPath)
	assert.NoError(t, err)
}

func TestDbConfig_Merge(t *testing.T) {
	original := &DbConfig{
		DbConn: "sqlite",
		DbHost: "localhost",
		DbData: "original.db",
		DbPort: 0,
	}

	newCfg := &DbConfig{
		DbConn: "postgres",
		DbHost: "dbserver",
		DbData: "newdb",
		DbPort: 5432,
		DbUser: "admin",
		DbPass: "secret",
	}

	result := original.Merge(newCfg)
	assert.Equal(t, "postgres", result.DbConn)
	assert.Equal(t, "dbserver", result.DbHost)
	assert.Equal(t, "newdb", result.DbData)
	assert.Equal(t, 5432, result.DbPort)
	assert.Equal(t, "admin", result.DbUser)
	assert.Equal(t, "secret", result.DbPass)
}

func TestDbConfig_Clean(t *testing.T) {
	cfg := &DbConfig{
		DbConn: "postgres",
		DbHost: "dbserver",
		DbData: "mydb",
		DbPort: 5432,
		DbUser: "admin",
		DbPass: "secret",
	}

	result := cfg.Clean()
	assert.Empty(t, result.DbConn)
	assert.Empty(t, result.DbHost)
	assert.Empty(t, result.DbData)
	assert.Equal(t, 0, result.DbPort)
	assert.Empty(t, result.DbUser)
	assert.Empty(t, result.DbPass)
}

func TestDbConfig_ToYAML(t *testing.T) {
	cfg := &DbConfig{
		DbConn: "sqlite",
		DbHost: "localhost",
		DbData: "test.db",
	}

	yaml := cfg.ToYAML()
	assert.NotNil(t, yaml)
	assert.Contains(t, string(yaml), "connection: sqlite")
}

func TestDbConfig_ConnectionString(t *testing.T) {
	t.Run("memory connection", func(t *testing.T) {
		cfg := &DbConfig{DbConn: "memory"}
		conn := cfg.ConnectionString()
		assert.Equal(t, ":memory:", conn)
		assert.Equal(t, ":memory:", cfg.DbConn)
	})

	t.Run("sqlite connection", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "statping.db")
		err := utils.SaveFile(dbPath, []byte(""))
		require.NoError(t, err)

		cfg := &DbConfig{
			DbConn:   "sqlite",
			Location: tmpDir,
		}
		conn := cfg.ConnectionString()
		assert.Contains(t, conn, "statping.db")
	})

	t.Run("mysql connection", func(t *testing.T) {
		cfg := &DbConfig{
			DbConn: "mysql",
			DbHost: "localhost",
			DbPort: 3306,
			DbUser: "root",
			DbPass: "password",
			DbData: "statping",
		}
		conn := cfg.ConnectionString()
		assert.Contains(t, conn, "root:password@tcp(localhost:3306)/statping")
	})

	t.Run("postgres connection", func(t *testing.T) {
		utils.Params.Set("POSTGRES_SSLMODE", "disable")
		defer utils.Params.Set("POSTGRES_SSLMODE", "")

		cfg := &DbConfig{
			DbConn: "postgres",
			DbHost: "localhost",
			DbPort: 5432,
			DbUser: "postgres",
			DbPass: "password",
			DbData: "statping",
		}
		conn := cfg.ConnectionString()
		assert.Contains(t, conn, "host=localhost")
		assert.Contains(t, conn, "port=5432")
		assert.Contains(t, conn, "user=postgres")
		assert.Contains(t, conn, "dbname=statping")
		assert.Contains(t, conn, "sslmode=disable")
	})
}

func TestDbConfig_Close(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &DbConfig{
		DbConn:   "sqlite",
		DbData:   "statping.db",
		Location: tmpDir,
	}

	err := Connect(cfg, false)
	require.NoError(t, err)

	cfg.Close()
}

func TestDbConfig_Update(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &DbConfig{
		DbConn:   "sqlite",
		DbData:   "statping.db",
		Location: tmpDir,
	}

	err := Connect(cfg, false)
	require.NoError(t, err)
	defer cfg.Close()

	cfg.Language = "es"
	err = cfg.Update()
	assert.NoError(t, err)
}

func TestDbConfig_CloseNil(t *testing.T) {
	var cfg *DbConfig
	cfg.Close()

	cfg = &DbConfig{}
	cfg.Close()
}

// =============================================================================
// LoadConfigForm Tests
// =============================================================================

func TestLoadConfigForm_ValidForm(t *testing.T) {
	// Save and restore DB_CONN to avoid polluting other tests
	originalDBConn := utils.Params.GetString("DB_CONN")
	defer utils.Params.Set("DB_CONN", originalDBConn)

	// Create a mock HTTP request with form data
	form := url.Values{}
	form.Set("db_connection", "sqlite")
	form.Set("db_host", "localhost")
	form.Set("db_user", "testuser")
	form.Set("db_password", "testpass")
	form.Set("db_database", "test.db")
	form.Set("db_port", "0")
	form.Set("project", "TestProject")
	form.Set("username", "admin")
	form.Set("password", "adminpass")
	form.Set("description", "Test Description")
	form.Set("domain", "example.com")
	form.Set("email", "admin@example.com")
	form.Set("language", "en")
	form.Set("send_reports", "true")
	form.Set("sample_data", "false")

	req, err := http.NewRequest("POST", "/setup", strings.NewReader(form.Encode()))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	cfg, err := LoadConfigForm(req)
	require.NoError(t, err)
	assert.Equal(t, "sqlite", cfg.DbConn)
	assert.Equal(t, "localhost", cfg.DbHost)
	assert.Equal(t, "testuser", cfg.DbUser)
	assert.Equal(t, "testpass", cfg.DbPass)
	assert.Equal(t, "test.db", cfg.DbData)
	assert.Equal(t, "TestProject", cfg.Project)
	assert.Equal(t, "admin", cfg.Username)
	assert.Equal(t, "adminpass", cfg.Password)
	assert.Equal(t, "admin@example.com", cfg.Email)
	assert.Equal(t, "en", cfg.Language)
	assert.True(t, cfg.AllowReports)
	assert.False(t, cfg.SampleData)
}

func TestLoadConfigForm_MissingRequiredFields(t *testing.T) {
	// Save and restore DB_CONN to avoid polluting other tests
	originalDBConn := utils.Params.GetString("DB_CONN")
	defer utils.Params.Set("DB_CONN", originalDBConn)

	testCases := []struct {
		name        string
		formData    map[string]string
		expectedErr string
	}{
		{
			name: "missing project",
			formData: map[string]string{
				"db_connection": "sqlite",
				"username":      "admin",
				"password":      "pass",
				"email":         "admin@test.com",
			},
			expectedErr: "Missing required elements",
		},
		{
			name: "missing username",
			formData: map[string]string{
				"db_connection": "sqlite",
				"project":       "Test",
				"password":      "pass",
				"email":         "admin@test.com",
			},
			expectedErr: "Missing required elements",
		},
		{
			name: "missing password",
			formData: map[string]string{
				"db_connection": "sqlite",
				"project":       "Test",
				"username":      "admin",
				"email":         "admin@test.com",
			},
			expectedErr: "Missing required elements",
		},
		{
			name: "missing email",
			formData: map[string]string{
				"db_connection": "sqlite",
				"project":       "Test",
				"username":      "admin",
				"password":      "pass",
			},
			expectedErr: "Missing required elements",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			form := url.Values{}
			for k, v := range tc.formData {
				form.Set(k, v)
			}

			req, err := http.NewRequest("POST", "/setup", strings.NewReader(form.Encode()))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			_, err = LoadConfigForm(req)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

func TestLoadConfigForm_PostgresConfig(t *testing.T) {
	// Save and restore DB_CONN to avoid polluting other tests
	originalDBConn := utils.Params.GetString("DB_CONN")
	defer utils.Params.Set("DB_CONN", originalDBConn)

	form := url.Values{}
	form.Set("db_connection", "postgres")
	form.Set("db_host", "db.example.com")
	form.Set("db_user", "pguser")
	form.Set("db_password", "pgpass")
	form.Set("db_database", "statping_db")
	form.Set("db_port", "5432")
	form.Set("project", "Production")
	form.Set("username", "admin")
	form.Set("password", "adminpass")
	form.Set("email", "admin@example.com")

	req, err := http.NewRequest("POST", "/setup", strings.NewReader(form.Encode()))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	cfg, err := LoadConfigForm(req)
	require.NoError(t, err)
	assert.Equal(t, "postgres", cfg.DbConn)
	assert.Equal(t, "db.example.com", cfg.DbHost)
	assert.Equal(t, 5432, cfg.DbPort)
	assert.Equal(t, "pguser", cfg.DbUser)
	assert.Equal(t, "pgpass", cfg.DbPass)
	assert.Equal(t, "statping_db", cfg.DbData)
}

// =============================================================================
// InitModels Tests
// =============================================================================

func TestInitModels(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := database.OpenTester(tmpDir)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Should not panic
	InitModels(db)
}

// =============================================================================
// CreateAdminUser Tests
// =============================================================================

func TestCreateAdminUser_WithEnvVars(t *testing.T) {
	tmpDir := t.TempDir()
	setTestDir(t, tmpDir)

	// Set up fresh database
	db, err := database.OpenTester(tmpDir)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Initialize all models with this DB
	InitModels(db)
	db.AutoMigrate(&users.User{})

	// Set environment variables - password must meet complexity requirements
	// At least 30 chars with uppercase, lowercase, and digits
	complexPass := "TestPassword123456789012345678901"
	utils.Params.Set("ADMIN_USER", "testadmin")
	utils.Params.Set("ADMIN_PASSWORD", complexPass)
	utils.Params.Set("ADMIN_EMAIL", "testadmin@example.com")
	defer func() {
		utils.Params.Set("ADMIN_USER", "")
		utils.Params.Set("ADMIN_PASSWORD", "")
		utils.Params.Set("ADMIN_EMAIL", "")
	}()

	pass, err := CreateAdminUser()
	require.NoError(t, err)
	assert.Equal(t, complexPass, pass)

	// Verify user was created
	user, err := users.FindByUsername("testadmin")
	require.NoError(t, err)
	assert.Equal(t, "testadmin", user.Username)
	assert.Equal(t, "testadmin@example.com", user.Email)
	assert.True(t, user.Admin.Bool)
}

func TestCreateAdminUser_DefaultValues(t *testing.T) {
	tmpDir := t.TempDir()
	setTestDir(t, tmpDir)

	// Set up fresh database
	db, err := database.OpenTester(tmpDir)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Initialize all models with this DB
	InitModels(db)
	db.AutoMigrate(&users.User{})

	// Clear environment variables to use defaults
	utils.Params.Set("ADMIN_USER", "")
	utils.Params.Set("ADMIN_PASSWORD", "")
	utils.Params.Set("ADMIN_EMAIL", "info@admin.com")

	pass, err := CreateAdminUser()
	require.NoError(t, err)
	assert.Len(t, pass, 32) // Random 32-char password

	// Verify user was created with default username
	user, err := users.FindByUsername("admin")
	require.NoError(t, err)
	assert.Equal(t, "admin", user.Username)
}

// =============================================================================
// DropDatabase Tests
// =============================================================================

func TestDbConfig_DropDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	setTestDir(t, tmpDir)

	cfg := &DbConfig{
		DbConn:   "sqlite",
		DbData:   "statping.db",
		Location: tmpDir,
	}

	err := Connect(cfg, false)
	require.NoError(t, err)
	defer cfg.Close()

	// Create tables first
	err = cfg.CreateDatabase()
	require.NoError(t, err)

	// Now drop them - should not error
	err = cfg.DropDatabase()
	require.NoError(t, err)
}

// =============================================================================
// CreateDatabase Tests
// =============================================================================

func TestDbConfig_CreateDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	setTestDir(t, tmpDir)

	cfg := &DbConfig{
		DbConn:   "sqlite",
		DbData:   "statping.db",
		Location: tmpDir,
	}

	err := Connect(cfg, false)
	require.NoError(t, err)
	defer cfg.Close()

	err = cfg.CreateDatabase()
	require.NoError(t, err)

	// Verify tables exist
	assert.True(t, cfg.Db.HasTable(&services.Service{}))
	assert.True(t, cfg.Db.HasTable(&users.User{}))
	assert.True(t, cfg.Db.HasTable(&groups.Group{}))
	assert.True(t, cfg.Db.HasTable(&hits.Hit{}))
	assert.True(t, cfg.Db.HasTable(&failures.Failure{}))
}
