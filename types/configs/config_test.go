package configs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = utils.InitLogs()
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
