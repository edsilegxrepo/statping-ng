package handlers

import (
	"os"
	"testing"

	"github.com/statping-ng/statping-ng/source"
	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/utils"
)

func TestMain(m *testing.M) {
	// Stop any running service goroutines and clear cache from previous packages
	services.StopAll()
	services.ClearCache()

	// Create a temp directory for all handler tests
	tmpDir, err := os.MkdirTemp("", "statping-handlers-test")
	if err != nil {
		os.Exit(1)
	}

	// Set the directory for all tests
	utils.Directory = tmpDir
	utils.InitEnvs()
	utils.Params.Set("STATPING_DIR", tmpDir)
	_ = utils.InitLogs()

	// Initialize assets and core (moved from init() in api_test.go)
	_ = source.Assets()
	core.New("test", "testcommithere")

	// Run all tests
	code := m.Run()

	// Stop services before cleanup
	services.StopAll()

	// Cleanup temp directory
	_ = os.RemoveAll(tmpDir)

	os.Exit(code)
}
