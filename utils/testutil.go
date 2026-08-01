package utils

import (
	"os"
	"path/filepath"
	"sync"
)

var (
	testFilesDir     string
	testFilesDirOnce sync.Once
)

// TestFilesDir returns the path to the testfiles directory for persistent test artifacts.
// The directory is created if it doesn't exist.
// Use this for test outputs that should survive test runs (for debugging).
// For ephemeral test files, use t.TempDir() instead.
func TestFilesDir() string {
	testFilesDirOnce.Do(func() {
		// Find repo root by looking for go.mod
		dir, err := os.Getwd()
		if err != nil {
			testFilesDir = filepath.Join(os.TempDir(), "statping-testfiles")
		} else {
			// Walk up to find go.mod
			for {
				if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
					testFilesDir = filepath.Join(dir, "testfiles")
					break
				}
				parent := filepath.Dir(dir)
				if parent == dir {
					// Reached root, use temp dir as fallback
					testFilesDir = filepath.Join(os.TempDir(), "statping-testfiles")
					break
				}
				dir = parent
			}
		}
		// Ensure directory exists
		_ = os.MkdirAll(testFilesDir, 0o750)
	})
	return testFilesDir
}
