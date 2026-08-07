package database

import (
	"sync"
	"testing"
)

// TestEnv holds an isolated test environment.
// Each test workflow should call SetupTestEnv() to get its own database.
type TestEnv struct {
	DB      Database
	mu      sync.Mutex
	cleanup func()
}

// SetupTestEnv creates an isolated test environment with its own database.
// The database is automatically closed when the test completes.
// Usage:
//
//	func TestWorkflow(t *testing.T) {
//	    env := database.SetupTestEnv(t)
//	    mypackage.SetDB(env.DB)
//	    // ... run tests
//	}
func SetupTestEnv(t *testing.T) *TestEnv {
	t.Helper()

	db, err := OpenTester()
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	env := &TestEnv{
		DB: db,
		cleanup: func() {
			_ = db.Close()
		},
	}

	t.Cleanup(func() {
		env.mu.Lock()
		defer env.mu.Unlock()
		if env.cleanup != nil {
			env.cleanup()
			env.cleanup = nil
		}
	})

	return env
}

// SetupTestEnvWithTables creates an isolated test environment and auto-migrates the given tables.
// Usage:
//
//	func TestWorkflow(t *testing.T) {
//	    env := database.SetupTestEnvWithTables(t, &User{}, &Service{})
//	    users.SetDB(env.DB)
//	    services.SetDB(env.DB)
//	    // ... run tests
//	}
func SetupTestEnvWithTables(t *testing.T, tables ...interface{}) *TestEnv {
	t.Helper()

	env := SetupTestEnv(t)

	for _, table := range tables {
		env.DB.CreateTable(table)
	}

	return env
}
