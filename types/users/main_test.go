package users

import (
	"os"
	"testing"

	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/utils"
)

var testDb database.Database

func TestMain(m *testing.M) {
	// Initialize logging
	_ = utils.InitLogs()

	// Stop any running service goroutines from previous packages
	services.StopAll()
	services.ClearCache()

	// Create isolated database for this test chain
	var err error
	testDb, err = database.OpenTester()
	if err != nil {
		os.Exit(1)
	}

	// Create tables and set DB
	testDb.CreateTable(&User{})
	SetDB(testDb)

	code := m.Run()

	// Cleanup
	testDb.Close()
	os.Exit(code)
}
