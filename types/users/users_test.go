package users

import (
	"strings"
	"testing"

	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPassword = "Password123456789012345678901234567890"

// setupTestDB creates an isolated in-memory SQLite database for testing.
// Each workflow gets its own database - no shared state between workflows.
// The database is automatically closed when the test completes.
// Note: This modifies the package-level db, so workflows must run sequentially.
func setupTestDB(t *testing.T) {
	t.Helper()

	err := utils.InitLogs()
	require.Nil(t, err)

	testDb, err := database.OpenTester()
	require.Nil(t, err)

	testDb.CreateTable(&User{})

	// Save original db and restore on cleanup
	origDb := db
	SetDB(testDb)

	t.Cleanup(func() {
		_ = testDb.Close()
		// Restore original db (may be nil)
		db = origDb
	})
}

// ============================================================================
// Workflow 1: Core CRUD Chain
// Tests basic Create, Read, Update, Delete operations in sequence
// ============================================================================

func TestWorkflow1_CoreCRUD(t *testing.T) {
	setupTestDB(t)

	var exampleUserID int64

	t.Run("Init_CreateUser", func(t *testing.T) {
		user := &User{
			Username: "example_user",
			Email:    "info@example.com",
			Password: testPassword,
			Admin:    null.NewNullBool(true),
		}
		err := user.Create()
		require.Nil(t, err)
		require.NotZero(t, user.Id)
		exampleUserID = user.Id
	})

	t.Run("Find", func(t *testing.T) {
		item, err := Find(exampleUserID)
		require.Nil(t, err)
		assert.Equal(t, "example_user", item.Username)
		assert.NotEmpty(t, item.ApiKey)
		assert.NotEqual(t, testPassword, item.Password)
		assert.True(t, item.Admin.Bool)
	})

	t.Run("FindByUsername", func(t *testing.T) {
		item, err := FindByUsername("example_user")
		require.Nil(t, err)
		assert.Equal(t, "example_user", item.Username)
		assert.NotEmpty(t, item.ApiKey)
		assert.NotEqual(t, testPassword, item.Password)
		assert.True(t, item.Admin.Bool)
	})

	t.Run("All", func(t *testing.T) {
		items := All()
		assert.Len(t, items, 1)
	})

	t.Run("Create_SecondUser", func(t *testing.T) {
		user := &User{
			Username: "exampleuser2",
			Password: testPassword,
			Email:    "info@yahoo.com",
		}
		err := user.Create()
		require.Nil(t, err)
		assert.NotZero(t, user.Id)
		assert.Equal(t, "exampleuser2", user.Username)
		assert.NotEqual(t, testPassword, user.Password)
		assert.NotZero(t, user.CreatedAt)
		assert.NotEmpty(t, user.ApiKey)
	})

	t.Run("Update", func(t *testing.T) {
		item, err := Find(exampleUserID)
		require.Nil(t, err)
		item.Username = "updated_user"
		err = item.Update()
		require.Nil(t, err)
		assert.Equal(t, "updated_user", item.Username)
	})

	t.Run("Delete", func(t *testing.T) {
		all := All()
		assert.Len(t, all, 2)

		item, err := Find(exampleUserID)
		require.Nil(t, err)

		err = item.Delete()
		require.Nil(t, err)

		all = All()
		assert.Len(t, all, 1)
	})

	t.Run("Samples", func(t *testing.T) {
		_, err := Samples()
		require.Nil(t, err)
		assert.Len(t, All(), 3)
	})
}

// ============================================================================
// Workflow 2: Database Functions
// Comprehensive tests for all database.go functions
// ============================================================================

func TestWorkflow2_DatabaseFunctions(t *testing.T) {
	setupTestDB(t)

	t.Run("SetDB_Verify", func(t *testing.T) {
		users := All()
		assert.NotNil(t, users)
		assert.Len(t, users, 0)
	})

	t.Run("CreateUser", func(t *testing.T) {
		user := &User{
			Username: "testuser1",
			Password: testPassword,
			Email:    "test1@example.com",
			Admin:    null.NewNullBool(false),
		}
		err := user.Create()
		require.Nil(t, err)
		assert.NotZero(t, user.Id)
		assert.NotEmpty(t, user.ApiKey)
		assert.True(t, strings.HasPrefix(user.Password, "$2a$") || strings.HasPrefix(user.Password, "$2b$"))
	})

	t.Run("CreateAdminUser", func(t *testing.T) {
		user := &User{
			Username: "adminuser",
			Password: testPassword,
			Email:    "admin@example.com",
			Admin:    null.NewNullBool(true),
			Scopes:   "admin",
		}
		err := user.Create()
		require.Nil(t, err)
		assert.True(t, user.Admin.Bool)
		assert.Equal(t, "admin", user.Scopes)
	})

	t.Run("FindById", func(t *testing.T) {
		user, err := Find(1)
		require.Nil(t, err)
		assert.Equal(t, "testuser1", user.Username)
		assert.Equal(t, "test1@example.com", user.Email)
	})

	t.Run("FindByIdNotFound", func(t *testing.T) {
		_, err := Find(99999)
		assert.NotNil(t, err)
	})

	t.Run("FindByUsername", func(t *testing.T) {
		user, err := FindByUsername("testuser1")
		require.Nil(t, err)
		assert.Equal(t, int64(1), user.Id)
	})

	t.Run("FindByUsernameNotFound", func(t *testing.T) {
		_, err := FindByUsername("nonexistent")
		assert.NotNil(t, err)
	})

	t.Run("FindByAPIKey", func(t *testing.T) {
		user, err := Find(1)
		require.Nil(t, err)
		apiKey := user.ApiKey

		foundUser, err := FindByAPIKey(apiKey)
		require.Nil(t, err)
		assert.Equal(t, user.Id, foundUser.Id)
		assert.Equal(t, user.Username, foundUser.Username)
	})

	t.Run("FindByAPIKeyNotFound", func(t *testing.T) {
		_, err := FindByAPIKey("invalid-api-key-12345")
		assert.NotNil(t, err)
	})

	t.Run("FindByEmail", func(t *testing.T) {
		user, err := FindByEmail("test1@example.com")
		require.Nil(t, err)
		assert.Equal(t, "testuser1", user.Username)
	})

	t.Run("FindByEmailNotFound", func(t *testing.T) {
		_, err := FindByEmail("notfound@example.com")
		assert.NotNil(t, err)
	})

	t.Run("AllUsers", func(t *testing.T) {
		users := All()
		assert.Len(t, users, 2)
	})

	t.Run("UpdateUser", func(t *testing.T) {
		user, err := Find(1)
		require.Nil(t, err)
		originalApiKey := user.ApiKey

		user.Email = "updated@example.com"
		err = user.Update()
		require.Nil(t, err)

		updatedUser, err := Find(1)
		require.Nil(t, err)
		assert.Equal(t, "updated@example.com", updatedUser.Email)
		assert.Equal(t, originalApiKey, updatedUser.ApiKey)
	})

	t.Run("DeleteUser", func(t *testing.T) {
		user := &User{
			Username: "todelete",
			Password: testPassword,
			Email:    "delete@example.com",
		}
		err := user.Create()
		require.Nil(t, err)

		countBefore := len(All())
		err = user.Delete()
		require.Nil(t, err)
		countAfter := len(All())
		assert.Equal(t, countBefore-1, countAfter)
	})
}

// ============================================================================
// Workflow 3: Validation
// Tests for hooks.go - Validation and BeforeCreate/BeforeUpdate hooks
// ============================================================================

func TestWorkflow3_Validation(t *testing.T) {
	setupTestDB(t)

	t.Run("ValidateEmptyUsername", func(t *testing.T) {
		user := &User{
			Username: "",
			Password: testPassword,
			Email:    "test@example.com",
		}
		err := user.Validate()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "username is empty")
	})

	t.Run("ValidateEmptyPassword", func(t *testing.T) {
		user := &User{
			Username: "testuser",
			Password: "",
			Email:    "test@example.com",
		}
		err := user.Validate()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "password is empty")
	})

	t.Run("ValidateEmptyEmail", func(t *testing.T) {
		user := &User{
			Username: "testuser",
			Password: testPassword,
			Email:    "",
		}
		err := user.Validate()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "email is empty")
	})

	t.Run("ValidateShortPassword", func(t *testing.T) {
		user := &User{
			Username: "testuser",
			Password: "Short1",
			Email:    "test@example.com",
		}
		err := user.Validate()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "password must be at least 30 characters")
	})

	t.Run("ValidatePasswordNoUppercase", func(t *testing.T) {
		user := &User{
			Username: "testuser",
			Password: "alllowercase12345678901234567890",
			Email:    "test@example.com",
		}
		err := user.Validate()
		assert.NotNil(t, err)
	})

	t.Run("ValidatePasswordNoLowercase", func(t *testing.T) {
		user := &User{
			Username: "testuser",
			Password: "ALLUPPERCASE12345678901234567890",
			Email:    "test@example.com",
		}
		err := user.Validate()
		assert.NotNil(t, err)
	})

	t.Run("ValidatePasswordNoDigits", func(t *testing.T) {
		user := &User{
			Username: "testuser",
			Password: "NoDigitsHereAtAllInPassword!!!!!!",
			Email:    "test@example.com",
		}
		err := user.Validate()
		assert.NotNil(t, err)
	})

	t.Run("ValidateValidUser", func(t *testing.T) {
		user := &User{
			Username: "validuser",
			Password: testPassword,
			Email:    "valid@example.com",
		}
		err := user.Validate()
		assert.Nil(t, err)
	})

	t.Run("ValidateHashedPasswordSkipsComplexityCheck", func(t *testing.T) {
		hashedPassword := utils.HashPassword(testPassword)
		user := &User{
			Username: "hasheduser",
			Password: hashedPassword,
			Email:    "hashed@example.com",
		}
		err := user.Validate()
		assert.Nil(t, err)
	})

	t.Run("CreateFailsValidation", func(t *testing.T) {
		user := &User{
			Username: "invaliduser",
			Password: "short",
			Email:    "invalid@example.com",
		}
		err := user.Create()
		assert.NotNil(t, err)
	})
}

// ============================================================================
// Workflow 4: Password Hashing
// Tests for password hashing (BeforeCreate hook)
// ============================================================================

func TestWorkflow4_PasswordHashing(t *testing.T) {
	setupTestDB(t)

	t.Run("PasswordIsHashedOnCreate", func(t *testing.T) {
		user := &User{
			Username: "hashtest",
			Password: testPassword,
			Email:    "hash@example.com",
		}
		err := user.Create()
		require.Nil(t, err)

		assert.NotEqual(t, testPassword, user.Password)
		assert.True(t, strings.HasPrefix(user.Password, "$2a$") || strings.HasPrefix(user.Password, "$2b$"))
	})

	t.Run("CanVerifyPasswordWithCheckHash", func(t *testing.T) {
		user, err := FindByUsername("hashtest")
		require.Nil(t, err)

		isValid := utils.CheckHash(testPassword, user.Password)
		assert.True(t, isValid)

		isInvalid := utils.CheckHash("wrongpassword1234567890123456", user.Password)
		assert.False(t, isInvalid)
	})
}

// ============================================================================
// Workflow 5: API Key Generation
// Tests for API key generation (BeforeCreate hook)
// ============================================================================

func TestWorkflow5_APIKeyGeneration(t *testing.T) {
	setupTestDB(t)

	var user1ApiKey, user2ApiKey string

	t.Run("APIKeyGeneratedOnCreate", func(t *testing.T) {
		user := &User{
			Username: "apikeytest",
			Password: testPassword,
			Email:    "apikey@example.com",
		}
		assert.Empty(t, user.ApiKey)

		err := user.Create()
		require.Nil(t, err)

		assert.NotEmpty(t, user.ApiKey)
		assert.Len(t, user.ApiKey, 64) // SHA256 hex is 64 chars
		user1ApiKey = user.ApiKey
	})

	t.Run("APIKeyIsUnique", func(t *testing.T) {
		user := &User{
			Username: "apiuser2",
			Password: testPassword,
			Email:    "apiuser2@example.com",
		}
		err := user.Create()
		require.Nil(t, err)
		user2ApiKey = user.ApiKey

		assert.NotEqual(t, user1ApiKey, user2ApiKey)
	})

	t.Run("CanFindUserByAPIKey", func(t *testing.T) {
		found, err := FindByAPIKey(user1ApiKey)
		require.Nil(t, err)
		assert.Equal(t, "apikeytest", found.Username)
	})
}

// ============================================================================
// Workflow 6: Auth User Function
// Tests for auth.go - AuthUser function
// ============================================================================

func TestWorkflow6_AuthUserFunction(t *testing.T) {
	setupTestDB(t)

	// Create test user first
	user := &User{
		Username: "authuser",
		Password: testPassword,
		Email:    "authuser@example.com",
		Admin:    null.NewNullBool(true),
	}
	err := user.Create()
	require.Nil(t, err)

	t.Run("AuthUserSuccess", func(t *testing.T) {
		authedUser, ok := AuthUser("authuser", testPassword)
		assert.True(t, ok)
		assert.NotNil(t, authedUser)
		assert.Equal(t, "authuser", authedUser.Username)
	})

	t.Run("AuthUserWrongPassword", func(t *testing.T) {
		authedUser, ok := AuthUser("authuser", "WrongPassword12345678901234567890")
		assert.False(t, ok)
		assert.Nil(t, authedUser)
	})

	t.Run("AuthUserNonexistentUser", func(t *testing.T) {
		authedUser, ok := AuthUser("nonexistent", testPassword)
		assert.False(t, ok)
		assert.Nil(t, authedUser)
	})

	t.Run("AuthUserEmptyUsername", func(t *testing.T) {
		authedUser, ok := AuthUser("", testPassword)
		assert.False(t, ok)
		assert.Nil(t, authedUser)
	})

	t.Run("AuthUserEmptyPassword", func(t *testing.T) {
		authedUser, ok := AuthUser("authuser", "")
		assert.False(t, ok)
		assert.Nil(t, authedUser)
	})
}

// ============================================================================
// Workflow 7: User Scopes
// Tests for scopes.go - User scopes (no DB needed)
// ============================================================================

func TestWorkflow7_UserScopes(t *testing.T) {
	t.Run("AllScopesEmpty", func(t *testing.T) {
		user := &User{Scopes: ""}
		scopes := user.AllScopes()
		assert.Len(t, scopes, 1)
		assert.Equal(t, EmptyUser, scopes[0])
	})

	t.Run("AllScopesAdmin", func(t *testing.T) {
		user := &User{Scopes: "admin"}
		scopes := user.AllScopes()
		assert.Len(t, scopes, 1)
		assert.Equal(t, FullAdmin, scopes[0])
	})

	t.Run("AllScopesReadOnly", func(t *testing.T) {
		user := &User{Scopes: "readonly"}
		scopes := user.AllScopes()
		assert.Len(t, scopes, 1)
		assert.Equal(t, ReadOnly, scopes[0])
	})

	t.Run("AllScopesMultiple", func(t *testing.T) {
		user := &User{Scopes: "read:services,write:services"}
		scopes := user.AllScopes()
		assert.Len(t, scopes, 2)
		assert.Contains(t, scopes, RServices)
		assert.Contains(t, scopes, RWServices)
	})

	t.Run("AllScopesIncidents", func(t *testing.T) {
		user := &User{Scopes: "read:incidents,write:incidents"}
		scopes := user.AllScopes()
		assert.Len(t, scopes, 2)
		assert.Contains(t, scopes, RIncidents)
		assert.Contains(t, scopes, RWIncidents)
	})

	t.Run("AllScopesUnknown", func(t *testing.T) {
		user := &User{Scopes: "unknown"}
		scopes := user.AllScopes()
		assert.Len(t, scopes, 1)
		assert.Equal(t, EmptyUser, scopes[0])
	})
}

// ============================================================================
// Workflow 8: Admin User Operations
// Tests for admin user create, promote, demote
// ============================================================================

func TestWorkflow8_AdminUserOperations(t *testing.T) {
	setupTestDB(t)

	t.Run("CreateAdminUser", func(t *testing.T) {
		admin := &User{
			Username: "superadmin",
			Password: testPassword,
			Email:    "superadmin@example.com",
			Admin:    null.NewNullBool(true),
			Scopes:   "admin",
		}
		err := admin.Create()
		require.Nil(t, err)
		assert.True(t, admin.Admin.Bool)
	})

	t.Run("CreateNonAdminUser", func(t *testing.T) {
		regular := &User{
			Username: "regularuser",
			Password: testPassword,
			Email:    "regular@example.com",
			Admin:    null.NewNullBool(false),
			Scopes:   "readonly",
		}
		err := regular.Create()
		require.Nil(t, err)
		assert.False(t, regular.Admin.Bool)
	})

	t.Run("PromoteToAdmin", func(t *testing.T) {
		user, err := FindByUsername("regularuser")
		require.Nil(t, err)
		assert.False(t, user.Admin.Bool)

		user.Admin = null.NewNullBool(true)
		user.Scopes = "admin"
		err = user.Update()
		require.Nil(t, err)

		updated, err := FindByUsername("regularuser")
		require.Nil(t, err)
		assert.True(t, updated.Admin.Bool)
	})

	t.Run("DemoteFromAdmin", func(t *testing.T) {
		user, err := FindByUsername("superadmin")
		require.Nil(t, err)
		assert.True(t, user.Admin.Bool)

		user.Admin = null.NewNullBool(false)
		user.Scopes = "readonly"
		err = user.Update()
		require.Nil(t, err)

		updated, err := FindByUsername("superadmin")
		require.Nil(t, err)
		assert.False(t, updated.Admin.Bool)
	})
}

// ============================================================================
// Workflow 9: GORM Hooks
// Tests for AfterFind, AfterCreate, AfterUpdate, AfterDelete hooks
// ============================================================================

func TestWorkflow9_GormHooks(t *testing.T) {
	setupTestDB(t)

	var hookTestUserID int64

	t.Run("AfterCreateHook", func(t *testing.T) {
		user := &User{
			Username: "hooktest",
			Password: testPassword,
			Email:    "hook@example.com",
		}
		err := user.Create()
		require.Nil(t, err)
		assert.NotZero(t, user.Id)
		hookTestUserID = user.Id
	})

	t.Run("AfterFindHook", func(t *testing.T) {
		found, err := Find(hookTestUserID)
		require.Nil(t, err)
		assert.NotNil(t, found)
	})

	t.Run("AfterUpdateHook", func(t *testing.T) {
		user, err := Find(hookTestUserID)
		require.Nil(t, err)

		user.Email = "updated_hook@example.com"
		err = user.Update()
		require.Nil(t, err)
	})

	t.Run("AfterDeleteHook", func(t *testing.T) {
		// Create a user to delete
		user := &User{
			Username: "createhook",
			Password: testPassword,
			Email:    "createhook@example.com",
		}
		err := user.Create()
		require.Nil(t, err)

		err = user.Delete()
		require.Nil(t, err)
	})
}

// ============================================================================
// Workflow 10: Edge Cases
// Tests for edge cases and error handling
// ============================================================================

func TestWorkflow10_EdgeCases(t *testing.T) {
	setupTestDB(t)

	t.Run("CreateUserWithSpecialCharactersInUsername", func(t *testing.T) {
		user := &User{
			Username: "user_test-123",
			Password: testPassword,
			Email:    "special@example.com",
		}
		err := user.Create()
		require.Nil(t, err)
		assert.NotZero(t, user.Id)
	})

	t.Run("CreateUserWithLongEmail", func(t *testing.T) {
		longEmail := "verylongemailaddress@subdomain.example.com"
		user := &User{
			Username: "longemail",
			Password: testPassword,
			Email:    longEmail,
		}
		err := user.Create()
		require.Nil(t, err)
		assert.Equal(t, longEmail, user.Email)
	})

	t.Run("UpdateNonExistentUser", func(t *testing.T) {
		user := &User{
			Id:       99999,
			Username: "nonexistent",
			Password: testPassword,
			Email:    "nonexistent@example.com",
		}
		_ = user.Update()
		_, findErr := Find(99999)
		assert.NotNil(t, findErr)
	})

	t.Run("AllUsersEmptyDatabase", func(t *testing.T) {
		users := All()
		for _, u := range users {
			_ = u.Delete()
		}

		allUsers := All()
		assert.Len(t, allUsers, 0)
	})
}

// ============================================================================
// Workflow 11: Samples Function
// Tests for sample.go - Samples function
// ============================================================================

func TestWorkflow11_SamplesFunction(t *testing.T) {
	setupTestDB(t)

	t.Run("SamplesCreatesUsers", func(t *testing.T) {
		passwords, err := Samples()
		require.Nil(t, err)
		assert.NotNil(t, passwords)
		assert.Len(t, passwords, 2)
		assert.Contains(t, passwords, "testadmin")
		assert.Contains(t, passwords, "testadmin2")
	})

	t.Run("SampleUsersAreAdmins", func(t *testing.T) {
		user1, err := FindByUsername("testadmin")
		require.Nil(t, err)
		assert.True(t, user1.Admin.Bool)

		user2, err := FindByUsername("testadmin2")
		require.Nil(t, err)
		assert.True(t, user2.Admin.Bool)
	})

	t.Run("SamplePasswordsMeetComplexity", func(t *testing.T) {
		passwords, _ := Samples()

		for _, pass := range passwords {
			assert.True(t, len(pass) >= 30)
			assert.True(t, utils.ComplexityCheck(pass))
		}
	})
}

// ============================================================================
// Workflow 12: Timestamps
// Tests for CreatedAt, UpdatedAt timestamps
// ============================================================================

func TestWorkflow12_Timestamps(t *testing.T) {
	setupTestDB(t)

	var timestampUserID int64

	t.Run("CreatedAtIsSet", func(t *testing.T) {
		user := &User{
			Username: "timestamps",
			Password: testPassword,
			Email:    "timestamps@example.com",
		}
		err := user.Create()
		require.Nil(t, err)
		timestampUserID = user.Id

		found, err := Find(user.Id)
		require.Nil(t, err)
		assert.False(t, found.CreatedAt.IsZero())
	})

	t.Run("UpdatedAtIsSet", func(t *testing.T) {
		user, err := Find(timestampUserID)
		require.Nil(t, err)

		user.Email = "newtimestamps@example.com"
		err = user.Update()
		require.Nil(t, err)

		updated, err := Find(timestampUserID)
		require.Nil(t, err)
		assert.False(t, updated.UpdatedAt.IsZero())
	})
}

// ============================================================================
// Workflow 13: Case Insensitive Lookups
// Tests for case-insensitive username/email lookups
// ============================================================================

func TestWorkflow13_CaseInsensitiveLookups(t *testing.T) {
	setupTestDB(t)

	t.Run("FindByUsernameCaseInsensitive", func(t *testing.T) {
		user := &User{
			Username: "casetestuser",
			Password: testPassword,
			Email:    "casetest@example.com",
		}
		err := user.Create()
		require.Nil(t, err)

		// Should find with exact case
		found, err := FindByUsername("casetestuser")
		require.Nil(t, err)
		assert.Equal(t, "casetestuser", found.Username)

		// Should find with different case
		found, err = FindByUsername("CaseTestUser")
		require.Nil(t, err)
		assert.Equal(t, "casetestuser", found.Username)

		// Should find with all caps
		found, err = FindByUsername("CASETESTUSER")
		require.Nil(t, err)
		assert.Equal(t, "casetestuser", found.Username)
	})

	t.Run("FindByEmailCaseInsensitive", func(t *testing.T) {
		// Should find with exact case
		found, err := FindByEmail("casetest@example.com")
		require.Nil(t, err)
		assert.Equal(t, "casetest@example.com", found.Email)

		// Should find with different case
		found, err = FindByEmail("CaseTest@Example.COM")
		require.Nil(t, err)
		assert.Equal(t, "casetest@example.com", found.Email)
	})

	t.Run("UsernameNormalizedOnCreate", func(t *testing.T) {
		user := &User{
			Username: "MixedCaseUser",
			Password: testPassword,
			Email:    "mixedcase@example.com",
		}
		err := user.Create()
		require.Nil(t, err)

		found, err := Find(user.Id)
		require.Nil(t, err)
		assert.Equal(t, "mixedcaseuser", found.Username)
		assert.Equal(t, "mixedcase@example.com", found.Email)
	})
}

// ============================================================================
// Workflow 14: Last Admin Protection
// Tests for last admin protection logic
// ============================================================================

func TestWorkflow14_LastAdminProtection(t *testing.T) {
	setupTestDB(t)

	t.Run("CountEnabledAdminsWithNewAdmins", func(t *testing.T) {
		initialCount := CountEnabledAdmins()

		admin := &User{
			Username: "countadmintest",
			Password: testPassword,
			Email:    "countadmin@example.com",
			Admin:    null.NewNullBool(true),
			Enabled:  null.NewNullBool(true),
		}
		err := admin.Create()
		require.Nil(t, err)

		newCount := CountEnabledAdmins()
		assert.Equal(t, initialCount+1, newCount)

		admin.Enabled = null.NewNullBool(false)
		err = admin.Update()
		require.Nil(t, err)

		afterDisable := CountEnabledAdmins()
		assert.Equal(t, initialCount, afterDisable)
	})

	t.Run("IsLastEnabledAdminWithMultipleAdmins", func(t *testing.T) {
		admin1 := &User{
			Username: "lastadmintest1",
			Password: testPassword,
			Email:    "lastadmin1@example.com",
			Admin:    null.NewNullBool(true),
			Enabled:  null.NewNullBool(true),
		}
		err := admin1.Create()
		require.Nil(t, err)

		admin2 := &User{
			Username: "lastadmintest2",
			Password: testPassword,
			Email:    "lastadmin2@example.com",
			Admin:    null.NewNullBool(true),
			Enabled:  null.NewNullBool(true),
		}
		err = admin2.Create()
		require.Nil(t, err)

		assert.False(t, IsLastEnabledAdmin(admin1.Id))
		assert.False(t, IsLastEnabledAdmin(admin2.Id))
	})

	t.Run("IsLastEnabledAdminNonAdmin", func(t *testing.T) {
		user := &User{
			Username: "nonadminuser",
			Password: testPassword,
			Email:    "nonadmin@example.com",
			Admin:    null.NewNullBool(false),
			Enabled:  null.NewNullBool(true),
		}
		err := user.Create()
		require.Nil(t, err)

		assert.False(t, IsLastEnabledAdmin(user.Id))
	})

	t.Run("IsLastEnabledAdminDisabledAdmin", func(t *testing.T) {
		user := &User{
			Username: "disabledadmin",
			Password: testPassword,
			Email:    "disabledadmin@example.com",
			Admin:    null.NewNullBool(true),
			Enabled:  null.NewNullBool(false),
		}
		err := user.Create()
		require.Nil(t, err)

		assert.False(t, IsLastEnabledAdmin(user.Id))
	})
}
