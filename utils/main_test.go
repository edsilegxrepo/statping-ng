package utils

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Create a temp directory for all utils tests
	tmpDir, err := os.MkdirTemp("", "statping-utils-test")
	if err != nil {
		os.Exit(1)
	}

	// Set the directory before any tests run
	Directory = tmpDir
	if Params != nil {
		Params.Set("STATPING_DIR", tmpDir)
	}

	// Run all tests
	code := m.Run()

	// Cleanup temp directory
	_ = os.RemoveAll(tmpDir)

	os.Exit(code)
}
