package services

import (
	"os"
	"testing"

	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/types/checkins"
	"github.com/statping-ng/statping-ng/types/failures"
	"github.com/statping-ng/statping-ng/types/hits"
	"github.com/statping-ng/statping-ng/types/incidents"
	"github.com/statping-ng/statping-ng/types/messages"
	"github.com/statping-ng/statping-ng/types/notifications"
	"github.com/statping-ng/statping-ng/utils"
)

func TestMain(m *testing.M) {
	_ = utils.InitLogs()

	// Create temp directory for file-based SQLite database
	tmpDir, err := os.MkdirTemp("", "statping-services-test")
	if err != nil {
		os.Exit(1)
	}

	// Set up file-based test database (better isolation than in-memory)
	testDb, err := database.OpenTester(tmpDir)
	if err != nil {
		os.Exit(1)
	}

	// Set package-level db for all packages that need it
	db = testDb
	SetDB(testDb)
	hits.SetDB(testDb)
	failures.SetDB(testDb)
	checkins.SetDB(testDb)
	incidents.SetDB(testDb)
	messages.SetDB(testDb)
	notifications.SetDB(testDb)

	// Run migrations for all required tables
	_ = db.AutoMigrate(&Service{})
	_ = db.AutoMigrate(&hits.Hit{})
	_ = db.AutoMigrate(&failures.Failure{})
	_ = db.AutoMigrate(&checkins.Checkin{})
	_ = db.AutoMigrate(&checkins.CheckinHit{})
	_ = db.AutoMigrate(&incidents.Incident{})
	_ = db.AutoMigrate(&incidents.IncidentUpdate{})
	_ = db.AutoMigrate(&messages.Message{})
	_ = db.AutoMigrate(&notifications.Notification{})

	// Initialize allServices map
	allServices = make(map[int64]*Service)

	// Run tests
	code := m.Run()

	// Cleanup
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
	_ = os.RemoveAll(tmpDir)

	os.Exit(code)
}
