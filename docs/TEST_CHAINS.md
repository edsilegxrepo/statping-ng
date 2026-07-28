# Test Chains and Dependencies

This document maps test dependencies to enable safe refactoring of the test architecture.

## Summary

| Package | Isolation | Internal Order | Cross-Package Deps | Can Run Standalone |
|---------|-----------|----------------|-------------------|-------------------|
| database/ | Full | None | None | Yes |
| notifiers/ | Full | None | None | Yes |
| types/checkins/ | Full | None | None | Yes |
| types/hits/ | Full | None | None | Yes |
| types/failures/ | Full | None | None | Yes |
| types/users/ | Shared DB | Strict | None | Yes |
| types/incidents/ | Shared DB | Strict | None | Yes |
| types/groups/ | Shared DB | Strict | services | Yes |
| types/services/ | Shared DB | Strict | hits, failures, checkins, incidents, messages, notifications | Yes |
| handlers/ | Shared DB | Strict | All type packages | Yes |
| integration/ | External | Mixed | None | Yes (with build tag) |

---

## Packages with FULL ISOLATION (Safe to Parallelize)

### database/
- **Setup**: Each test calls `OpenTester()` for fresh in-memory SQLite
- **Globals**: None shared between tests
- **Order**: Independent
- **Notes**: Ideal test architecture - each test is self-contained

### notifiers/
- **Setup**: Tests call `utils.InitLogs()` and `core.Example()` inline
- **Globals**: Uses `core.App` but only reads example data
- **Order**: Independent, uses `t.Parallel()`
- **Notes**: Good isolation via example fixtures

### types/checkins/
- **Setup**: Each test calls `setupTestDB()` for fresh DB
- **Globals**: Package `db` reset per test
- **Order**: Independent
- **Notes**: Good pattern to replicate

### types/hits/
- **Setup**: Each test calls `setupTestDB()` for fresh DB
- **Globals**: Package `db` reset per test
- **Order**: Independent

### types/failures/
- **Setup**: Each test calls `setupTestDB()` for fresh DB
- **Globals**: Package `db` reset per test
- **Order**: Independent

---

## Packages with STRICT INTERNAL ORDER (Test Chains)

### types/users/

**Chain:**
```
TestInit (creates user ID=1)
    ↓
TestFind (expects ID=1)
    ↓
TestFindByUsername (expects "exampleuser")
    ↓
TestAll (expects 1 user)
    ↓
TestCreate (creates user ID=2)
    ↓
TestUpdate (modifies user ID=1)
    ↓
TestDelete (deletes user ID=1, expects 2→1 users)
    ↓
TestSamples (expects 1 user, creates 2 more)
    ↓
TestClose (closes shared db)
```

**Globals Written**: `db` via `SetDB()`
**Hardcoded IDs**: `Find(1)`, `Find(2)`

---

### types/incidents/

**Chain:**
```
TestInit (creates incident ID=1, updates ID=1,2)
    ↓
TestFind (expects ID=1)
    ↓
TestAll (expects 1 incident)
    ↓
TestCreate (creates incident ID=2)
    ↓
TestUpdate (modifies ID=1, title→"Updated")
    ↓
TestDelete (deletes ID=1)
    ↓
TestSamples (creates 2 more)
    ↓
TestClose
```

**Globals Written**: `db` via `SetDB()`
**Hardcoded IDs**: `Find(1)`, `Find(2)`

---

### types/groups/

**Chain:**
```
TestInit (creates group ID=1, services s1/s2)
    ↓
TestFind (expects ID=1, name="Example Group")
    ↓
TestAll (expects 1 group)
    ↓
TestCreate (creates group ID=2)
    ↓
TestUpdate (changes name to "Updated")
    ↓
TestDelete (deletes ID=1, expects 2→1 groups)
    ↓
TestSamples (expects 1 group, creates 3 more)
    ↓
TestSelectGroupsPublicFiltering (expects 4 groups)
    ↓
TestClose
```

**Globals Written**: `db`, `services.SetDB()`
**Cross-Package**: Imports `types/services` for service creation

---

### types/services/

**Chain:**
```
TestMain (creates SQLite file, sets DB for 7 packages)
    ↓
TestStartExampleEndpoints / ensureEndpointsStarted()
    (starts HTTP servers on ports 15000-15004)
    (creates example service ID=1)
    ↓
TestServices/Test Find (expects ID=1)
    ↓
TestServices/Test All (expects 1 service)
    ↓
TestServices/Test Create (creates service ID=2)
    ↓
TestServices/Test Update (modifies ID=1)
    ↓
TestServices/Test Delete (DELETES ID=1 - breaks later tests!)
    ↓
TestServices/Test Samples (creates sample services)
    ↓
TestServices/Test Close
```

**Globals Written**: 
- Package `db` via `SetDB()`
- `hits.SetDB()`, `failures.SetDB()`, `checkins.SetDB()`
- `incidents.SetDB()`, `messages.SetDB()`, `notifications.SetDB()`
- `allServices` map

**Critical Issue**: `TestServices/Test Delete` removes service ID=1, which other packages may depend on.

---

### handlers/

**Setup Chain:**
```
TestMain (creates temp dir, sets utils.Directory)
    ↓
init() (InitLogs, source.Assets, core.New)
    ↓
TestSetupRoutes (MUST RUN FIRST)
    - Calls /api/setup endpoint
    - Creates database with sample data
    - Creates 6+ services, users, groups, etc.
    ↓
ensureHandlerSetup() (sync.Once wrapper)
    ↓
All other handler tests
```

**Globals Written**:
- `utils.Directory`, `utils.Params`
- `core.App` (Setup, ApiSecret, Name)
- All type package data (services, users, groups, etc.)

**Hardcoded URLs**:
- `/api/services/1`, `/api/services/1/incidents`
- `/api/users/1`, `/api/groups/1`
- `/api/incidents/1`, `/api/checkins/1`

---

## Cross-Package State Pollution

### The Problem

When running `go test ./...`, packages execute in this order:
```
cmd → database → handlers → integration → notifiers → source → types → types/* → utils
```

Package `types/services/main_test.go` sets DB for 7 packages:
```go
SetDB(db)
hits.SetDB(db)
failures.SetDB(db)
checkins.SetDB(db)
incidents.SetDB(db)
messages.SetDB(db)
notifications.SetDB(db)
```

If `handlers` runs before `types/services`, handlers creates its own DB state. Then when `types/services` runs, it overwrites the DB connections with its own, but other packages (hits, failures, etc.) now have stale references.

### The sync.Once Trap

Both `handlers/api_test.go` and `types/services/services_test.go` use `sync.Once`:
```go
var setupOnce sync.Once

func ensureSetup(t *testing.T) {
    setupOnce.Do(func() {
        // Setup runs once per process
    })
}
```

**Problem**: If the first package to run has a failed or partial setup, `sync.Once` never retries. All subsequent tests fail because setup appears "complete" but isn't.

---

## Current Implementation (Per-Chain Isolation)

Each test chain now has a `TestMain` that:
1. Calls `services.StopAll()` to stop any leaked goroutines
2. Calls `services.ClearCache()` to clear in-memory state  
3. Creates a fresh isolated database
4. Runs tests in order (chains preserved)
5. Cleans up on exit

**Packages with TestMain isolation:**
- `handlers/` - stops services, creates temp dir
- `types/services/` - stops services, creates isolated DB for 7 packages
- `types/users/` - creates isolated DB
- `types/groups/` - stops services, creates isolated DB
- `types/incidents/` - creates isolated DB (can't import services due to cycle)

**Key helper function:**
```go
// types/services/database.go
func StopAll() {
    // Stops all running service goroutines
    // Must be called before ClearCache()
}
```

---

## Testing Commands

```bash
# Recommended: Run with sequential execution, single iteration
go test ./... -p 1 -count=1

# Run each package in isolation
go test ./database/... -v
go test ./handlers/... -v
go test ./types/services/... -v

# AVOID: -count=N with N>1 (TestMain only runs once per process)
# Use separate invocations instead:
go test ./... -p 1 -count=1 && go test ./... -p 1 -count=1

# AVOID: Parallel execution without -p 1 (state pollution)
go test ./...

# Debug: See which test corrupted state
go test ./... -p 1 -v 2>&1 | grep -E "(=== RUN|FAIL)"
```
