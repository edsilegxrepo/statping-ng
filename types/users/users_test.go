package users

import (
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
