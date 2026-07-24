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
	sqlite := &DbConfig{
		DbConn: "sqlite",
		DbHost: "localhost",
		DbUser: "",
		DbPass: "",
		DbData: "",
		DbPort: 0,
	}

	err := Connect(sqlite, false)
	require.Nil(t, err)
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
	file, err := findSQLin(utils.Directory)
	require.Nil(t, err)
	assert.Equal(t, "statping.db", file)
}
