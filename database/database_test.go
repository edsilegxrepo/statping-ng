package database

import (
	"testing"
	"time"

	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func init() {
	_ = utils.InitLogs()
}

// TestModel is a simple model for testing database operations
type TestModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"size:255"`
	Value     int       `gorm:"default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func TestOpenTester(t *testing.T) {
	db, err := OpenTester()
	require.NoError(t, err)
	require.NotNil(t, db)

	// Verify connection
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NotNil(t, sqlDB)

	err = sqlDB.Ping()
	assert.NoError(t, err)

	// Cleanup
	_ = sqlDB.Close()
}

func TestOpenwSQLite(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	require.NotNil(t, db)

	assert.Equal(t, "sqlite", db.DbType())

	sqlDB, _ := db.DB()
	_ = sqlDB.Close()
}

func TestOpenwInvalidDialect(t *testing.T) {
	// Empty dialect defaults to sqlite
	db, err := Openw("", ":memory:")
	require.NoError(t, err)
	assert.Equal(t, "sqlite", db.DbType())

	sqlDB, _ := db.DB()
	_ = sqlDB.Close()
}

func TestWrap(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)

	// Get the underlying gorm.DB and wrap it
	gormDB := db.GormDB()
	require.NotNil(t, gormDB)

	wrapped := Wrap(gormDB)
	assert.NotNil(t, wrapped)
	assert.Equal(t, "sqlite", wrapped.DbType())

	sqlDB, _ := db.DB()
	_ = sqlDB.Close()
}

func TestGetSet(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)

	Set(db)
	got := Get()
	assert.Equal(t, db, got)

	sqlDB, _ := db.DB()
	_ = sqlDB.Close()
}

func TestDbCRUDOperations(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	// Create table
	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	// CREATE
	t.Run("Create", func(t *testing.T) {
		model := &TestModel{Name: "test1", Value: 100}
		err := db.Create(model).Error()
		require.NoError(t, err)
		assert.NotZero(t, model.ID)
	})

	// READ
	t.Run("Find", func(t *testing.T) {
		var model TestModel
		err := db.First(&model).Error()
		require.NoError(t, err)
		assert.Equal(t, "test1", model.Name)
		assert.Equal(t, 100, model.Value)
	})

	// UPDATE
	t.Run("Update", func(t *testing.T) {
		err := db.Model(&TestModel{}).Where("name = ?", "test1").Updates(map[string]interface{}{"value": 200}).Error()
		require.NoError(t, err)

		var model TestModel
		err = db.First(&model).Error()
		require.NoError(t, err)
		assert.Equal(t, 200, model.Value)
	})

	// DELETE
	t.Run("Delete", func(t *testing.T) {
		err := db.Where("name = ?", "test1").Delete(&TestModel{}).Error()
		require.NoError(t, err)

		var count int64
		db.Model(&TestModel{}).Count(&count)
		assert.Equal(t, int64(0), count)
	})
}

func TestDbQueryMethods(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	// Insert test data
	for i := 1; i <= 10; i++ {
		db.Create(&TestModel{Name: "item", Value: i * 10})
	}

	t.Run("Where", func(t *testing.T) {
		var models []TestModel
		err := db.Where("value > ?", 50).Find(&models).Error()
		require.NoError(t, err)
		assert.Len(t, models, 5)
	})

	t.Run("Limit", func(t *testing.T) {
		var models []TestModel
		err := db.Limit(3).Find(&models).Error()
		require.NoError(t, err)
		assert.Len(t, models, 3)
	})

	t.Run("Offset", func(t *testing.T) {
		var models []TestModel
		err := db.Offset(5).Find(&models).Error()
		require.NoError(t, err)
		assert.Len(t, models, 5)
	})

	t.Run("Order", func(t *testing.T) {
		var models []TestModel
		err := db.Order("value DESC").Find(&models).Error()
		require.NoError(t, err)
		assert.Equal(t, 100, models[0].Value)
	})

	t.Run("Count", func(t *testing.T) {
		var count int64
		err := db.Model(&TestModel{}).Count(&count).Error()
		require.NoError(t, err)
		assert.Equal(t, int64(10), count)
	})

	t.Run("Pluck", func(t *testing.T) {
		var values []int
		err := db.Model(&TestModel{}).Pluck("value", &values).Error()
		require.NoError(t, err)
		assert.Len(t, values, 10)
	})

	t.Run("Select", func(t *testing.T) {
		var model TestModel
		err := db.Select("name", "value").First(&model).Error()
		require.NoError(t, err)
		assert.NotEmpty(t, model.Name)
	})
}

func TestDbChunkSize(t *testing.T) {
	tests := []struct {
		dialect  string
		expected int
	}{
		{"sqlite", 100},
		{"mysql", 3000},
		{"postgres", 3000},
	}

	for _, tt := range tests {
		t.Run(tt.dialect, func(t *testing.T) {
			db := &Db{Type: tt.dialect}
			assert.Equal(t, tt.expected, db.ChunkSize())
		})
	}
}

func TestDbHasTable(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	// Table doesn't exist yet
	assert.False(t, db.HasTable(&TestModel{}))

	// Create table
	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	// Now it exists
	assert.True(t, db.HasTable(&TestModel{}))
}

func TestDbHasIndex(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	// Index doesn't exist yet
	assert.False(t, db.HasIndex(&TestModel{}, "idx_test_name"))

	// Create index
	err = db.Model(&TestModel{}).AddIndex("idx_test_name", "name").Error()
	require.NoError(t, err)

	// Now it exists
	assert.True(t, db.HasIndex(&TestModel{}, "idx_test_name"))
}

func TestDbTransaction(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	t.Run("Commit", func(t *testing.T) {
		tx := db.Begin()
		tx.Create(&TestModel{Name: "txtest", Value: 999})
		tx.Commit()

		var count int64
		db.Model(&TestModel{}).Where("name = ?", "txtest").Count(&count)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Rollback", func(t *testing.T) {
		tx := db.Begin()
		tx.Create(&TestModel{Name: "rollbacktest", Value: 888})
		tx.Rollback()

		var count int64
		db.Model(&TestModel{}).Where("name = ?", "rollbacktest").Count(&count)
		assert.Equal(t, int64(0), count)
	})
}

func TestDbTimeQueries(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	// Insert with specific times
	now := time.Now()
	oldTime := now.Add(-48 * time.Hour)

	db.Create(&TestModel{Name: "recent", Value: 1, CreatedAt: now})
	db.Create(&TestModel{Name: "old", Value: 2, CreatedAt: oldTime})

	t.Run("Since", func(t *testing.T) {
		var models []TestModel
		err := db.Model(&TestModel{}).Since(now.Add(-1 * time.Hour)).Find(&models).Error()
		require.NoError(t, err)
		assert.Len(t, models, 1)
		assert.Equal(t, "recent", models[0].Name)
	})

	t.Run("Between", func(t *testing.T) {
		var models []TestModel
		err := db.Model(&TestModel{}).Between(oldTime.Add(-1*time.Hour), now.Add(1*time.Hour)).Find(&models).Error()
		require.NoError(t, err)
		assert.Len(t, models, 2)
	})
}

func TestDbRawAndExec(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	// Insert via Exec
	err = db.Exec("INSERT INTO test_models (name, value) VALUES (?, ?)", "rawtest", 42).Error()
	require.NoError(t, err)

	// Query via Raw
	var model TestModel
	err = db.Raw("SELECT * FROM test_models WHERE name = ?", "rawtest").Scan(&model).Error()
	require.NoError(t, err)
	assert.Equal(t, 42, model.Value)
}

func TestDbRowsAffected(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	// Insert multiple
	for i := 0; i < 5; i++ {
		db.Create(&TestModel{Name: "bulk", Value: i})
	}

	// Update all
	result := db.Model(&TestModel{}).Where("name = ?", "bulk").Updates(map[string]interface{}{"value": 999})
	assert.Equal(t, int64(5), result.RowsAffected())
}

func TestDbError(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	// Query non-existent table
	var model TestModel
	result := db.Table("nonexistent").First(&model)
	assert.Error(t, result.Error())
}

func TestDbRecordNotFound(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	var model TestModel
	result := db.Where("id = ?", 99999).First(&model)
	assert.True(t, result.RecordNotFound())
}

func TestDbLogMode(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	// Should not panic
	db.LogMode(true)
	db.LogMode(false)
}

func TestDbScopes(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		db.Create(&TestModel{Name: "scope", Value: i * 10})
	}

	// Define a scope
	highValue := func(db *gorm.DB) *gorm.DB {
		return db.Where("value >= ?", 50)
	}

	var models []TestModel
	err = db.Scopes(highValue).Find(&models).Error()
	require.NoError(t, err)
	assert.Len(t, models, 5)
}
