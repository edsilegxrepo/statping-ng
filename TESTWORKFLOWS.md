# Test Audit: types/users/users_test.go

## Workflow 1: Core CRUD Chain (lines 23-110)
Uses shared db from TestMain. MUST run in order.

| Test | Depends On | Creates State |
|------|------------|---------------|
| TestInit | TestMain db | Creates example user (id=1) |
| TestFind | example user exists | - |
| TestFindByUsername | example user exists | - |
| TestAll | 1 user exists | - |
| TestCreate | - | Creates exampleuser2 (id=2) |
| TestAuthUser | exampleuser2 exists | - (skipped) |
| TestUpdate | example user exists | Updates username to "updated_user" |
| TestDelete | 2 users exist | Deletes example, leaves 1 user |
| TestSamples | 1 user exists | Creates sample users, now 3 total |
| TestClose | - | CLOSES the shared db ← **PROBLEM** |

## Workflow 2: TestDatabaseFunctions (lines 116-246)
Creates own testDb, calls SetDB(testDb). Self-contained but leaks via SetDB.

**Sub-tests:** SetDB, CreateUser, CreateAdminUser, FindById, FindByIdNotFound, FindByUsername, FindByUsernameNotFound, FindByAPIKey, FindByAPIKeyNotFound, FindByEmail, FindByEmailNotFound, AllUsers, UpdateUser, DeleteUser

## Workflow 3: TestValidation (lines 252-380)
Creates own testDb, calls SetDB(testDb). Self-contained but leaks.

**Sub-tests:** ValidateEmptyUsername, ValidateEmptyPassword, ValidateEmptyEmail, ValidateShortPassword, ValidateValidUser, ValidatePreservesExistingHash, BeforeCreateSetsApiKey, BeforeCreateHashesPassword, BeforeCreateNormalizesEmail, BeforeUpdatePreservesApiKey

## Workflow 4: TestPasswordHashing (lines 386-420)
Creates own testDb, calls SetDB(testDb). Self-contained but leaks.

**Sub-tests:** PasswordIsHashedOnCreate, CanVerifyPasswordWithCheckHash

## Workflow 5: TestAPIKeyGeneration (lines 426-488)
Creates own testDb, calls SetDB(testDb). Self-contained but leaks.

**Sub-tests:** APIKeyGeneratedOnCreate, APIKeyIsUnique, APIKeyPersistsOnUpdate

## Workflow 6: TestAuthProviders (lines 494-592)
Creates own testDb, calls SetDB(testDb). Self-contained but leaks.

**Sub-tests:** DefaultAuthProviderIsLocal, AuthProviderPersists, GetAuthProvidersReturnsAllTypes

## Workflow 7: TestScopes (lines 598-662)
Creates own testDb, calls SetDB(testDb). Self-contained but leaks.

**Sub-tests:** ScopesCanBeSet, AdminWithScopes, EmptyScopes

## Workflow 8: TestEnabledFlag (lines 668-722)
Creates own testDb, calls SetDB(testDb). Self-contained but leaks.

**Sub-tests:** DefaultEnabledIsTrue, DisabledUserPersists, EnabledUserPersists

## Workflow 9: TestTimestamps (lines 728-845)
Creates own testDb, calls SetDB(testDb). Self-contained but leaks.

**Sub-tests:** CreatedAtIsSet, UpdatedAtIsSet

## Workflow 10: TestCaseInsensitiveLookups (lines 847-901)
NO own db - uses whatever db is set. Depends on Workflow 1's db being open!

**Sub-tests:** FindByUsernameCaseInsensitive, FindByEmailCaseInsensitive, UsernameNormalizedOnCreate

## Workflow 11: TestLastAdminProtection (lines 903-989)
NO own db - uses whatever db is set. Depends on Workflow 1's db being open!

**Sub-tests:** CountEnabledAdminsWithNewAdmins, IsLastEnabledAdminWithMultipleAdmins, IsLastEnabledAdminNonAdmin, IsLastEnabledAdminDisabledAdmin

---

## Problems Identified

1. **TestClose closes the shared db** - Workflows 10 & 11 run after and fail with "database is closed"

2. **Workflows 2-9 call SetDB(testDb)** - This overwrites the package-level db, then closes it. If Go runs tests in different order, chaos ensues.

3. **Workflows 10 & 11 have no isolation** - They assume shared db is still open

---

## Root Cause Analysis

The test file mixes two incompatible patterns:

1. **Shared db for sequential CRUD tests** (Workflow 1) - Tests run in order, each building on previous state
2. **Isolated dbs for parallel-safe tests** (Workflows 2-9) - Each creates own db for independence

The isolation is broken by `SetDB()` side effects. When Workflow 2-9 tests call `SetDB(testDb)`, they overwrite the package-level db pointer. When they close their local db, the package-level pointer now points to a closed connection.

**The core problem:** Tests share a package-level database connection, but:

1. **TestClose** (end of Workflow 1) closes the shared db
2. **Workflows 10 & 11** run after but expect the db to still be open → they fail with "database is closed"
3. **Workflows 2-9** each create their own db but call `SetDB(testDb)` which overwrites the package-level db, then close their local db - if test order varies, other tests break

---

## Design Principle

**Isolation between workflows, sequential within workflows.**

Each workflow:
- Has its own isolated database (no shared state between workflows)
- Runs its tests sequentially to respect dependency chains
- Cleans up its own db when done

This allows workflows to run in parallel with each other (isolated), while preserving test order within each workflow (sequential chains).

---

## Recommended Fix

Convert each workflow to a single parent test with sequential subtests:

```go
func TestWorkflow1_CoreCRUD(t *testing.T) {
    // Own isolated db for this workflow
    db := setupTestDB(t)
    defer db.Close()
    
    // Sequential subtests - NOT parallel
    t.Run("Init", func(t *testing.T) { ... })
    t.Run("Find", func(t *testing.T) { ... })
    t.Run("Create", func(t *testing.T) { ... })
    // ... etc
}

func TestWorkflow2_DatabaseFunctions(t *testing.T) {
    // Own isolated db for this workflow
    db := setupTestDB(t)
    defer db.Close()
    
    // Sequential subtests
    t.Run("CreateUser", func(t *testing.T) { ... })
    t.Run("FindById", func(t *testing.T) { ... })
    // ... etc
}
```

Key points:
- Each `TestWorkflowN_*` function is a workflow with its own db
- Subtests within a workflow run sequentially (no `t.Parallel()`)
- Workflows can run in parallel with each other (isolated dbs)
- No `SetDB()` calls - pass db explicitly or use closure
- No `TestClose` - each workflow cleans up its own db via `defer`

---

## Helper Function Template

```go
// setupTestDB creates an isolated in-memory SQLite database for testing.
// The database is automatically closed when the test completes.
func setupTestDB(t *testing.T) *gorm.DB {
    t.Helper()
    
    // Unique name prevents collisions if running parallel
    dbName := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
    
    db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    if err != nil {
        t.Fatalf("failed to open test db: %v", err)
    }
    
    // Auto-migrate the User model
    if err := db.AutoMigrate(&User{}); err != nil {
        t.Fatalf("failed to migrate: %v", err)
    }
    
    // Register cleanup
    t.Cleanup(func() {
        sqlDB, _ := db.DB()
        if sqlDB != nil {
            sqlDB.Close()
        }
    })
    
    return db
}
```

---

## Handling Package-Level DB

The `users` package has functions like `Find()`, `Create()`, etc. that use a package-level `db` variable. Two options:

### Option A: Temporary SetDB with restore (minimal changes)
```go
func TestWorkflow2_DatabaseFunctions(t *testing.T) {
    testDb := setupTestDB(t)
    
    // Save and restore package db
    origDb := db
    SetDB(testDb)
    t.Cleanup(func() { SetDB(origDb) })
    
    t.Run("CreateUser", func(t *testing.T) {
        // Uses package-level Create() which calls db internally
        user := &User{Username: "test"}
        err := user.Create()
        require.NoError(t, err)
    })
}
```

### Option B: Pass db explicitly (cleaner but more refactoring)
```go
// Add methods that accept db parameter
func (u *User) CreateWithDB(db *gorm.DB) error {
    return db.Create(u).Error
}

func TestWorkflow2_DatabaseFunctions(t *testing.T) {
    db := setupTestDB(t)
    
    t.Run("CreateUser", func(t *testing.T) {
        user := &User{Username: "test"}
        err := user.CreateWithDB(db)
        require.NoError(t, err)
    })
}
```

**Recommendation:** Start with Option A (faster), migrate to Option B later if needed.

---

## Concrete Example: Converting Workflow 1

**Before (current broken code):**
```go
func TestInit(t *testing.T) {
    // depends on TestMain db
    user := &User{Username: "example", ...}
    err := user.Create()
    require.NoError(t, err)
}

func TestFind(t *testing.T) {
    // depends on TestInit having run
    user, err := Find(1)
    require.NoError(t, err)
    assert.Equal(t, "example", user.Username)
}

func TestClose(t *testing.T) {
    // BREAKS everything after
    db.Close()
}
```

**After (isolated workflow):**
```go
func TestWorkflow1_CoreCRUD(t *testing.T) {
    // Isolated db for this workflow
    testDb := setupTestDB(t)
    origDb := db
    SetDB(testDb)
    t.Cleanup(func() { SetDB(origDb) })
    
    var createdUserID int64
    
    t.Run("Init", func(t *testing.T) {
        user := &User{Username: "example", Email: "example@test.com", Password: "password123456789012345678901"}
        err := user.Create()
        require.NoError(t, err)
        require.NotZero(t, user.Id)
        createdUserID = user.Id
    })
    
    t.Run("Find", func(t *testing.T) {
        user, err := Find(createdUserID)
        require.NoError(t, err)
        assert.Equal(t, "example", user.Username)
    })
    
    t.Run("FindByUsername", func(t *testing.T) {
        user, err := FindByUsername("example")
        require.NoError(t, err)
        assert.Equal(t, createdUserID, user.Id)
    })
    
    t.Run("All", func(t *testing.T) {
        users := All()
        assert.Len(t, users, 1)
    })
    
    t.Run("Create_Second", func(t *testing.T) {
        user := &User{Username: "exampleuser2", Email: "user2@test.com", Password: "password123456789012345678901"}
        err := user.Create()
        require.NoError(t, err)
    })
    
    t.Run("Update", func(t *testing.T) {
        user, _ := Find(createdUserID)
        user.Username = "updated_user"
        err := user.Update()
        require.NoError(t, err)
        
        updated, _ := Find(createdUserID)
        assert.Equal(t, "updated_user", updated.Username)
    })
    
    t.Run("Delete", func(t *testing.T) {
        user, _ := Find(createdUserID)
        err := user.Delete()
        require.NoError(t, err)
        
        users := All()
        assert.Len(t, users, 1) // Only exampleuser2 remains
    })
    
    // NO TestClose - cleanup handled by t.Cleanup()
}
```

---

## Workflow Dependencies Map

Some subtests need data from previous subtests. Track with variables in parent scope:

| Workflow | Shared State Between Subtests |
|----------|-------------------------------|
| 1 | `createdUserID`, user count |
| 2 | `createdUser`, `adminUser` |
| 3 | None (each validates independently) |
| 4 | `hashedUser` for verification |
| 5 | `user1`, `user2` for uniqueness check |
| 6 | `user` with auth provider |
| 7 | `user` with scopes |
| 8 | `enabledUser`, `disabledUser` |
| 9 | `user` for timestamp checks |
| 10 | Users with mixed-case names |
| 11 | Multiple admin/non-admin users |

---

## Implementation Checklist

- [ ] Create `setupTestDB(t *testing.T) *gorm.DB` helper that returns isolated db
- [ ] Convert Workflow 1 (lines 23-110) to `TestWorkflow1_CoreCRUD` with sequential subtests
- [ ] Convert Workflow 2 (lines 116-246) to `TestWorkflow2_DatabaseFunctions`
- [ ] Convert Workflow 3 (lines 252-380) to `TestWorkflow3_Validation`
- [ ] Convert Workflow 4 (lines 386-420) to `TestWorkflow4_PasswordHashing`
- [ ] Convert Workflow 5 (lines 426-488) to `TestWorkflow5_APIKeyGeneration`
- [ ] Convert Workflow 6 (lines 494-592) to `TestWorkflow6_AuthProviders`
- [ ] Convert Workflow 7 (lines 598-662) to `TestWorkflow7_Scopes`
- [ ] Convert Workflow 8 (lines 668-722) to `TestWorkflow8_EnabledFlag`
- [ ] Convert Workflow 9 (lines 728-845) to `TestWorkflow9_Timestamps`
- [ ] Convert Workflow 10 (lines 847-901) to `TestWorkflow10_CaseInsensitiveLookups`
- [ ] Convert Workflow 11 (lines 903-989) to `TestWorkflow11_LastAdminProtection`
- [ ] Remove `TestClose` function
- [ ] Remove `TestMain` shared db setup (each workflow is self-contained)
- [ ] Verify: `go test -race -count=3`
- [ ] Verify: `go test -shuffle=on`
- [ ] Verify: `go test -parallel=4` (workflows run in parallel)
