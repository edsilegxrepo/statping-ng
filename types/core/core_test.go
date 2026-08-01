package core

import (
	"testing"

	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	_ = utils.InitLogs()

	t.Run("creates new core with version and commit", func(t *testing.T) {
		New("1.0.0", "abc123")

		assert.NotNil(t, App)
		assert.Equal(t, "1.0.0", App.Version)
		assert.Equal(t, "abc123", App.Commit)
		assert.False(t, App.Started.IsZero())
	})
}

func TestCore_TableName(t *testing.T) {
	c := Core{}
	assert.Equal(t, "core", c.TableName())
}

func TestExample(t *testing.T) {
	_ = utils.InitLogs()

	t.Run("creates example core", func(t *testing.T) {
		core := Example()

		assert.NotNil(t, App)
		assert.NotNil(t, core)
		assert.Equal(t, "Statping Testing", App.Name)
		assert.Equal(t, "exampleapisecret", App.ApiSecret)
		assert.Equal(t, "http://localhost:8080", App.Domain)
	})
}

func TestSetDB(t *testing.T) {
	_ = utils.InitLogs()

	t.Run("sets database and loads core", func(t *testing.T) {
		testDB, err := database.OpenTester()
		require.NoError(t, err)
		q := testDB.AutoMigrate(&Core{})
		require.NoError(t, q.Error())

		core := &Core{
			Name:      "Test Core",
			ApiSecret: "testsecret123",
			Domain:    "http://test.local",
		}
		err = testDB.Create(core).Error()
		require.NoError(t, err)

		SetDB(testDB)
		assert.NotNil(t, db)
	})
}

func TestCore_Create(t *testing.T) {
	_ = utils.InitLogs()

	t.Run("creates core with generated API secret", func(t *testing.T) {
		testDB, err := database.OpenTester()
		require.NoError(t, err)
		q := testDB.AutoMigrate(&Core{})
		require.NoError(t, q.Error())
		SetDB(testDB)

		core := &Core{
			Name:   "New Core",
			Domain: "http://new.local",
		}
		err = core.Create()
		require.NoError(t, err)
		assert.NotEmpty(t, core.ApiSecret)
		assert.Len(t, core.ApiSecret, 32)
	})

	t.Run("creates core with provided API secret", func(t *testing.T) {
		testDB, err := database.OpenTester()
		require.NoError(t, err)
		q := testDB.AutoMigrate(&Core{})
		require.NoError(t, q.Error())
		SetDB(testDB)

		core := &Core{
			Name:      "New Core",
			Domain:    "http://new.local",
			ApiSecret: "myownsecret",
		}
		err = core.Create()
		require.NoError(t, err)
		assert.Equal(t, "myownsecret", core.ApiSecret)
	})
}

func TestCore_Update(t *testing.T) {
	_ = utils.InitLogs()

	t.Run("updates core fields", func(t *testing.T) {
		testDB, err := database.OpenTester()
		require.NoError(t, err)
		q := testDB.AutoMigrate(&Core{})
		require.NoError(t, q.Error())
		SetDB(testDB)

		core := &Core{
			Name:      "Original Name",
			Domain:    "http://original.local",
			ApiSecret: "secret123",
		}
		err = core.Create()
		require.NoError(t, err)

		core.Name = "Updated Name"
		core.Domain = "http://updated.local"
		err = core.Update()
		require.NoError(t, err)
	})
}

func TestCore_Delete(t *testing.T) {
	_ = utils.InitLogs()

	t.Run("delete returns nil", func(t *testing.T) {
		core := &Core{}
		err := core.Delete()
		assert.NoError(t, err)
	})
}

func TestSelect(t *testing.T) {
	_ = utils.InitLogs()

	t.Run("selects core from database", func(t *testing.T) {
		testDB, err := database.OpenTester()
		require.NoError(t, err)
		q := testDB.AutoMigrate(&Core{})
		require.NoError(t, q.Error())
		SetDB(testDB)

		original := &Core{
			Name:      "Select Test",
			Domain:    "http://select.local",
			ApiSecret: "selectsecret",
			Language:  "en",
		}
		err = original.Create()
		require.NoError(t, err)

		selected, err := Select()
		require.NoError(t, err)
		assert.NotNil(t, selected)
		assert.Equal(t, "Select Test", selected.Name)
		assert.Equal(t, "http://select.local", selected.Domain)
	})
}

func TestSamples(t *testing.T) {
	_ = utils.InitLogs()

	t.Run("creates sample core with env values", func(t *testing.T) {
		testDB, err := database.OpenTester()
		require.NoError(t, err)
		q := testDB.AutoMigrate(&Core{})
		require.NoError(t, q.Error())
		SetDB(testDB)

		utils.Params.Set("NAME", "Sample App")
		utils.Params.Set("DESCRIPTION", "Sample Description")
		utils.Params.Set("DOMAIN", "http://sample.local")
		utils.Params.Set("LANGUAGE", "en")
		defer func() {
			utils.Params.Set("NAME", "")
			utils.Params.Set("DESCRIPTION", "")
			utils.Params.Set("DOMAIN", "")
			utils.Params.Set("LANGUAGE", "")
		}()

		err = Samples()
		require.NoError(t, err)

		selected, err := Select()
		require.NoError(t, err)
		assert.Equal(t, "Sample App", selected.Name)
		assert.Equal(t, "Sample Description", selected.Description)
	})
}

func TestCore_AfterFind(t *testing.T) {
	_ = utils.InitLogs()

	t.Run("AfterFind hook executes without error", func(t *testing.T) {
		core := &Core{}
		err := core.AfterFind(nil)
		assert.NoError(t, err)
	})
}
