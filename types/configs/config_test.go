package configs

import (
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

	// Close the database connection so t.TempDir() can clean up
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
	// Create a test db file in the temp directory
	testDbPath := tmpDir + "/statping.db"
	err := utils.SaveFile(testDbPath, []byte("test"))
	require.Nil(t, err)

	file, err := findSQLin(tmpDir)
	require.Nil(t, err)
	assert.Equal(t, "statping.db", file)
}
