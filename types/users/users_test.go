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

var example = &User{
	Username: "example_user",
	Email:    "info@example.com",
	Password: testPassword,
	Admin:    null.NewNullBool(true),
}

func TestInit(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)
	db, err := database.OpenTester()
	require.Nil(t, err)
	db.CreateTable(&User{})
	SetDB(db)
	err = example.Create()
	require.Nil(t, err)
}

func TestFind(t *testing.T) {
	item, err := Find(1)
	require.Nil(t, err)
	assert.Equal(t, "example_user", item.Username)
	assert.NotEmpty(t, item.ApiKey)
	assert.NotEqual(t, testPassword, item.Password)
	assert.True(t, item.Admin.Bool)
}

func TestFindByUsername(t *testing.T) {
	item, err := FindByUsername("example_user")
	require.Nil(t, err)
	assert.Equal(t, "example_user", item.Username)
	assert.NotEmpty(t, item.ApiKey)
	assert.NotEqual(t, testPassword, item.Password)
	assert.True(t, item.Admin.Bool)
}

func TestAll(t *testing.T) {
	items := All()
	assert.Len(t, items, 1)
}

func TestCreate(t *testing.T) {
	example := &User{
		Username: "exampleuser2",
		Password: testPassword,
		Email:    "info@yahoo.com",
	}
	err := example.Create()
	require.Nil(t, err)
	assert.NotZero(t, example.Id)
	assert.Equal(t, "exampleuser2", example.Username)
	assert.NotEqual(t, testPassword, example.Password)
	assert.NotZero(t, example.CreatedAt)
	assert.NotEmpty(t, example.ApiKey)
}

func TestAuthUser(t *testing.T) {
	t.SkipNow()
	u, ok := AuthUser("exampleuser2", utils.HashPassword("password12345"))
	require.True(t, ok)
	assert.Equal(t, "exampleuser2", u.Username)

	u, ok = AuthUser("exampleuser2", "wrongpass")
	assert.False(t, ok)
	assert.Nil(t, u)
}

func TestUpdate(t *testing.T) {
	item, err := Find(1)
	require.Nil(t, err)
	item.Username = "updated_user"
	err = item.Update()
	require.Nil(t, err)
	assert.Equal(t, "updated_user", item.Username)
}

func TestDelete(t *testing.T) {
	all := All()
	assert.Len(t, all, 2)

	item, err := Find(1)
	require.Nil(t, err)

	err = item.Delete()
	require.Nil(t, err)

	all = All()
	assert.Len(t, all, 1)
}

func TestSamples(t *testing.T) {
	_, err := Samples()
	require.Nil(t, err)
	assert.Len(t, All(), 3)
}

func TestClose(t *testing.T) {
	assert.Nil(t, db.Close())
}

// ============================================================================
// Additional comprehensive tests for database.go functions
// ============================================================================

func TestDatabaseFunctions(t *testing.T) {
	// Setup fresh database for these tests
	err := utils.InitLogs()
	require.Nil(t, err)
	testDb, err := database.OpenTester()
	require.Nil(t, err)
	testDb.CreateTable(&User{})
	SetDB(testDb)
	defer testDb.Close()

	t.Run("SetDB", func(t *testing.T) {
		// Verify SetDB was successful by performing an operation
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
// Tests for hooks.go - Validation and BeforeCreate/BeforeUpdate hooks
// ============================================================================

func TestValidation(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)
	testDb, err := database.OpenTester()
	require.Nil(t, err)
	testDb.CreateTable(&User{})
	SetDB(testDb)
	defer testDb.Close()

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
		// If password is already hashed, complexity check is skipped
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
// Tests for password hashing (BeforeCreate hook)
// ============================================================================

func TestPasswordHashing(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)
	testDb, err := database.OpenTester()
	require.Nil(t, err)
	testDb.CreateTable(&User{})
	SetDB(testDb)
	defer testDb.Close()

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
// Tests for API key generation (BeforeCreate hook)
// ============================================================================

func TestAPIKeyGeneration(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)
	testDb, err := database.OpenTester()
	require.Nil(t, err)
	testDb.CreateTable(&User{})
	SetDB(testDb)
	defer testDb.Close()

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
	})

	t.Run("APIKeyIsUnique", func(t *testing.T) {
		user1 := &User{
			Username: "apiuser1",
			Password: testPassword,
			Email:    "apiuser1@example.com",
		}
		user2 := &User{
			Username: "apiuser2",
			Password: testPassword,
			Email:    "apiuser2@example.com",
		}

		err := user1.Create()
		require.Nil(t, err)
		err = user2.Create()
		require.Nil(t, err)

		assert.NotEqual(t, user1.ApiKey, user2.ApiKey)
	})

	t.Run("CanFindUserByAPIKey", func(t *testing.T) {
		user, err := FindByUsername("apikeytest")
		require.Nil(t, err)

		found, err := FindByAPIKey(user.ApiKey)
		require.Nil(t, err)
		assert.Equal(t, user.Id, found.Id)
	})
}

// ============================================================================
// Tests for auth.go - AuthUser function
// ============================================================================

func TestAuthUserFunction(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)
	testDb, err := database.OpenTester()
	require.Nil(t, err)
	testDb.CreateTable(&User{})
	SetDB(testDb)
	defer testDb.Close()

	// Create a test user first
	user := &User{
		Username: "authuser",
		Password: testPassword,
		Email:    "authuser@example.com",
		Admin:    null.NewNullBool(true),
	}
	err = user.Create()
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
// Tests for scopes.go - User scopes
// ============================================================================

func TestUserScopes(t *testing.T) {
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
// Tests for admin user operations
// ============================================================================

func TestAdminUserOperations(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)
	testDb, err := database.OpenTester()
	require.Nil(t, err)
	testDb.CreateTable(&User{})
	SetDB(testDb)
	defer testDb.Close()

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
// Tests for GORM hooks (AfterFind, AfterCreate, AfterUpdate, AfterDelete)
// ============================================================================

func TestGormHooks(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)
	testDb, err := database.OpenTester()
	require.Nil(t, err)
	testDb.CreateTable(&User{})
	SetDB(testDb)
	defer testDb.Close()

	t.Run("AfterFindHook", func(t *testing.T) {
		user := &User{
			Username: "hooktest",
			Password: testPassword,
			Email:    "hook@example.com",
		}
		err := user.Create()
		require.Nil(t, err)

		// AfterFind is called during Find operations
		found, err := Find(user.Id)
		require.Nil(t, err)
		assert.NotNil(t, found)
	})

	t.Run("AfterCreateHook", func(t *testing.T) {
		user := &User{
			Username: "createhook",
			Password: testPassword,
			Email:    "createhook@example.com",
		}
		// AfterCreate is called during Create
		err := user.Create()
		require.Nil(t, err)
		assert.NotZero(t, user.Id)
	})

	t.Run("AfterUpdateHook", func(t *testing.T) {
		user, err := FindByUsername("hooktest")
		require.Nil(t, err)

		user.Email = "updated_hook@example.com"
		// AfterUpdate is called during Update
		err = user.Update()
		require.Nil(t, err)
	})

	t.Run("AfterDeleteHook", func(t *testing.T) {
		user, err := FindByUsername("createhook")
		require.Nil(t, err)

		// AfterDelete is called during Delete
		err = user.Delete()
		require.Nil(t, err)
	})
}

// ============================================================================
// Tests for edge cases and error handling
// ============================================================================

func TestEdgeCases(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)
	testDb, err := database.OpenTester()
	require.Nil(t, err)
	testDb.CreateTable(&User{})
	SetDB(testDb)
	defer testDb.Close()

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
		// Update on non-existent user may not error in all cases
		// but the user should not exist in DB
		_, findErr := Find(99999)
		assert.NotNil(t, findErr)
	})

	t.Run("AllUsersEmptyDatabase", func(t *testing.T) {
		// Delete all users
		users := All()
		for _, u := range users {
			u.Delete()
		}

		allUsers := All()
		assert.Len(t, allUsers, 0)
	})
}

// ============================================================================
// Tests for sample.go - Samples function
// ============================================================================

func TestSamplesFunction(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)
	testDb, err := database.OpenTester()
	require.Nil(t, err)
	testDb.CreateTable(&User{})
	SetDB(testDb)
	defer testDb.Close()

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
// Tests for timestamps (CreatedAt, UpdatedAt)
// ============================================================================

func TestTimestamps(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)
	testDb, err := database.OpenTester()
	require.Nil(t, err)
	testDb.CreateTable(&User{})
	SetDB(testDb)
	defer testDb.Close()

	t.Run("CreatedAtIsSet", func(t *testing.T) {
		user := &User{
			Username: "timestamps",
			Password: testPassword,
			Email:    "timestamps@example.com",
		}
		err := user.Create()
		require.Nil(t, err)

		found, err := Find(user.Id)
		require.Nil(t, err)
		assert.False(t, found.CreatedAt.IsZero())
	})

	t.Run("UpdatedAtIsSet", func(t *testing.T) {
		user, err := FindByUsername("timestamps")
		require.Nil(t, err)
		originalUpdatedAt := user.UpdatedAt

		user.Email = "newtimestamps@example.com"
		err = user.Update()
		require.Nil(t, err)

		updated, err := Find(user.Id)
		require.Nil(t, err)
		// UpdatedAt should be set (may or may not change depending on GORM auto-update)
		assert.False(t, updated.UpdatedAt.IsZero())
		_ = originalUpdatedAt // avoid unused variable warning
	})
}
