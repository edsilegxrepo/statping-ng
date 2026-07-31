package main

import (
	"os"
	"testing"

	"github.com/statping-ng/statping-ng/utils"
)

func TestMain(m *testing.M) {
	// Create a temp directory for all cmd tests
	tmpDir, err := os.MkdirTemp("", "statping-cmd-test")
	if err != nil {
		os.Exit(1)
	}

	// Set the directory for all tests
	utils.Directory = tmpDir
	utils.Params.Set("STATPING_DIR", tmpDir)

	// Run all tests
	code := m.Run()

	// Cleanup temp directory
	_ = os.RemoveAll(tmpDir)

	os.Exit(code)
}
