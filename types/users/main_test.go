package users

import (
	"os"
	"testing"

	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/utils"
)

func TestMain(m *testing.M) {
	// Initialize logging
	_ = utils.InitLogs()

	// Stop any running service goroutines from previous packages
	services.StopAll()
	services.ClearCache()

	// Each workflow creates its own isolated database via setupTestDB()
	// No shared database needed here

	code := m.Run()
	os.Exit(code)
}
