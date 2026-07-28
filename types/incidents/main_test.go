package incidents

import (
	"os"
	"testing"

	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/utils"
)

var testDb database.Database

func TestMain(m *testing.M) {
	// Initialize logging
	_ = utils.InitLogs()

	// Note: Cannot call services.StopAll() here due to import cycle
	// incidents is imported by services, so we can't import services here
	// This package's tests are isolated via fresh database anyway

	// Create isolated database for this test chain
	var err error
	testDb, err = database.OpenTester()
	if err != nil {
		os.Exit(1)
	}

	// Create tables and set DB
	testDb.AutoMigrate(&Incident{}, &IncidentUpdate{})
	SetDB(testDb)

	code := m.Run()

	// Cleanup
	testDb.Close()
	os.Exit(code)
}
