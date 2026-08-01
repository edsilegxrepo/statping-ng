package database

import (
	"os"
	"testing"

	"github.com/statping-ng/statping-ng/utils"
)

func TestMain(m *testing.M) {
	// Create a temp directory for all database tests
	tmpDir, err := os.MkdirTemp("", "statping-database-test")
	if err != nil {
		os.Exit(1)
	}

	// Set the directory before any tests run
	utils.Directory = tmpDir
	utils.InitEnvs()
	utils.Params.Set("STATPING_DIR", tmpDir)

	// Run all tests
	code := m.Run()

	// Cleanup temp directory
	_ = os.RemoveAll(tmpDir)

	os.Exit(code)
}
