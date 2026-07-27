package database

import (
	"errors"
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

// TestTransactionRollbackOnPartialFailure tests that transactions rollback properly on partial failures
func TestTransactionRollbackOnPartialFailure(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	t.Run("RollbackOnError", func(t *testing.T) {
		// Start transaction
		tx := db.Begin()

		// First insert succeeds
		tx.Create(&TestModel{Name: "partial1", Value: 100})

		// Simulate failure by rolling back
		tx.Rollback()

		// Verify nothing was committed
		var count int64
		db.Model(&TestModel{}).Where("name = ?", "partial1").Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("RollbackAfterMultipleInserts", func(t *testing.T) {
		tx := db.Begin()

		// Multiple inserts
		for i := 0; i < 5; i++ {
			tx.Create(&TestModel{Name: "partial_multi", Value: i})
		}

		// Rollback
		tx.Rollback()

		// Verify none were committed
		var count int64
		db.Model(&TestModel{}).Where("name = ?", "partial_multi").Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("CommitAfterSuccessfulTransaction", func(t *testing.T) {
		tx := db.Begin()

		// Multiple inserts
		for i := 0; i < 3; i++ {
			tx.Create(&TestModel{Name: "committed", Value: i})
		}

		// Commit
		tx.Commit()

		// Verify all were committed
		var count int64
		db.Model(&TestModel{}).Where("name = ?", "committed").Count(&count)
		assert.Equal(t, int64(3), count)
	})

	t.Run("NestedTransactionBehavior", func(t *testing.T) {
		// Create initial record
		db.Create(&TestModel{Name: "initial", Value: 1})

		tx := db.Begin()
		tx.Create(&TestModel{Name: "nested_test", Value: 100})

		// Update in same transaction
		tx.Model(&TestModel{}).Where("name = ?", "nested_test").Updates(map[string]interface{}{"value": 200})

		tx.Rollback()

		// Verify the record doesn't exist
		var count int64
		db.Model(&TestModel{}).Where("name = ?", "nested_test").Count(&count)
		assert.Equal(t, int64(0), count)
	})
}

// TestConnectionErrorHandling tests database connection error scenarios
func TestConnectionErrorHandling(t *testing.T) {
	t.Run("InvalidDSN", func(t *testing.T) {
		// Try to open with invalid connection string for MySQL
		_, err := Openw("mysql", "invalid:invalid@tcp(nonexistent:9999)/invalid")
		assert.Error(t, err)
	})

	t.Run("InvalidPostgresDSN", func(t *testing.T) {
		_, err := Openw("postgres", "host=nonexistent port=9999 user=invalid dbname=invalid sslmode=disable connect_timeout=1")
		assert.Error(t, err)
	})

	t.Run("AvailableWithNilDB", func(t *testing.T) {
		result := Available(nil)
		assert.False(t, result)
	})

	t.Run("AvailableWithValidDB", func(t *testing.T) {
		db, err := Openw("sqlite", ":memory:")
		require.NoError(t, err)
		defer func() {
			sqlDB, _ := db.DB()
			_ = sqlDB.Close()
		}()

		result := Available(db)
		assert.True(t, result)
	})

	t.Run("AvailableAfterClose", func(t *testing.T) {
		db, err := Openw("sqlite", ":memory:")
		require.NoError(t, err)

		sqlDB, _ := db.DB()
		_ = sqlDB.Close()

		result := Available(db)
		assert.False(t, result)
	})

	t.Run("CloseNilDB", func(t *testing.T) {
		err := Close(nil)
		assert.NoError(t, err)
	})
}

// TestMigrationEdgeCases tests edge cases in database migration
func TestMigrationEdgeCases(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	t.Run("AutoMigrateMultipleTimes", func(t *testing.T) {
		// Run migration multiple times - should be idempotent
		err := db.AutoMigrate(&TestModel{}).Error()
		require.NoError(t, err)

		err = db.AutoMigrate(&TestModel{}).Error()
		require.NoError(t, err)

		assert.True(t, db.HasTable(&TestModel{}))
	})

	t.Run("CreateTable", func(t *testing.T) {
		type TestModel2 struct {
			ID   int64  `gorm:"primaryKey"`
			Data string `gorm:"size:100"`
		}

		err := db.CreateTable(&TestModel2{}).Error()
		require.NoError(t, err)
		assert.True(t, db.HasTable(&TestModel2{}))
	})

	t.Run("DropTableIfExists", func(t *testing.T) {
		type TempModel struct {
			ID int64 `gorm:"primaryKey"`
		}

		// Create and verify
		err := db.AutoMigrate(&TempModel{}).Error()
		require.NoError(t, err)
		assert.True(t, db.HasTable(&TempModel{}))

		// Drop
		err = db.DropTableIfExists(&TempModel{}).Error()
		require.NoError(t, err)
		assert.False(t, db.HasTable(&TempModel{}))

		// Drop again - should not error
		err = db.DropTableIfExists(&TempModel{}).Error()
		// SQLite may or may not error on dropping non-existent table
	})

	t.Run("DropTable", func(t *testing.T) {
		type DropTestModel struct {
			ID int64 `gorm:"primaryKey"`
		}

		err := db.AutoMigrate(&DropTestModel{}).Error()
		require.NoError(t, err)

		err = db.DropTable(&DropTestModel{}).Error()
		require.NoError(t, err)
		assert.False(t, db.HasTable(&DropTestModel{}))
	})
}

// TestDatabaseTypeDetection tests database type detection for different dialects
func TestDatabaseTypeDetection(t *testing.T) {
	t.Run("SQLiteType", func(t *testing.T) {
		db, err := Openw("sqlite", ":memory:")
		require.NoError(t, err)
		defer func() {
			sqlDB, _ := db.DB()
			_ = sqlDB.Close()
		}()

		assert.Equal(t, "sqlite", db.DbType())
	})

	t.Run("SQLite3Alias", func(t *testing.T) {
		db, err := Openw("sqlite3", ":memory:")
		require.NoError(t, err)
		defer func() {
			sqlDB, _ := db.DB()
			_ = sqlDB.Close()
		}()

		assert.Equal(t, "sqlite", db.DbType())
	})

	t.Run("EmptyDialectDefaultsToSQLite", func(t *testing.T) {
		db, err := Openw("", ":memory:")
		require.NoError(t, err)
		defer func() {
			sqlDB, _ := db.DB()
			_ = sqlDB.Close()
		}()

		assert.Equal(t, "sqlite", db.DbType())
	})

	t.Run("WrapPreservesType", func(t *testing.T) {
		db, err := Openw("sqlite", ":memory:")
		require.NoError(t, err)
		defer func() {
			sqlDB, _ := db.DB()
			_ = sqlDB.Close()
		}()

		gormDB := db.GormDB()
		wrapped := Wrap(gormDB)

		assert.Equal(t, "sqlite", wrapped.DbType())
	})

	t.Run("DbTypeDirectAccess", func(t *testing.T) {
		mysqlDb := &Db{Type: "mysql"}
		assert.Equal(t, "mysql", mysqlDb.DbType())

		postgresDb := &Db{Type: "postgres"}
		assert.Equal(t, "postgres", postgresDb.DbType())
	})
}

// TestBatchOperationsWithChunkSize tests batch operations and chunk size behavior
func TestBatchOperationsWithChunkSize(t *testing.T) {
	t.Run("SQLiteChunkSize", func(t *testing.T) {
		db := &Db{Type: "sqlite3"}
		assert.Equal(t, 100, db.ChunkSize())
	})

	t.Run("MySQLChunkSize", func(t *testing.T) {
		db := &Db{Type: "mysql"}
		assert.Equal(t, 3000, db.ChunkSize())
	})

	t.Run("PostgresChunkSize", func(t *testing.T) {
		db := &Db{Type: "postgres"}
		assert.Equal(t, 3000, db.ChunkSize())
	})

	t.Run("UnknownTypeChunkSize", func(t *testing.T) {
		db := &Db{Type: "mssql"}
		assert.Equal(t, 100, db.ChunkSize())
	})

	t.Run("BatchInsertWithinChunkSize", func(t *testing.T) {
		db, err := Openw("sqlite", ":memory:")
		require.NoError(t, err)
		defer func() {
			sqlDB, _ := db.DB()
			_ = sqlDB.Close()
		}()

		err = db.AutoMigrate(&TestModel{}).Error()
		require.NoError(t, err)

		chunkSize := db.ChunkSize()

		// Insert exactly chunk size records
		for i := 0; i < chunkSize; i++ {
			db.Create(&TestModel{Name: "batch", Value: i})
		}

		var count int64
		db.Model(&TestModel{}).Where("name = ?", "batch").Count(&count)
		assert.Equal(t, int64(chunkSize), count)
	})
}

// TestQueryBuilderMethods tests query builder methods comprehensively
func TestQueryBuilderMethods(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	// Insert test data
	for i := 1; i <= 20; i++ {
		category := "A"
		if i > 10 {
			category = "B"
		}
		db.Create(&TestModel{Name: category, Value: i})
	}

	t.Run("WhereWithMultipleConditions", func(t *testing.T) {
		var models []TestModel
		err := db.Where("name = ?", "A").Where("value > ?", 5).Find(&models).Error()
		require.NoError(t, err)
		assert.Len(t, models, 5) // Values 6-10 with name A
	})

	t.Run("OrCondition", func(t *testing.T) {
		var models []TestModel
		err := db.Where("value = ?", 1).Or("value = ?", 20).Find(&models).Error()
		require.NoError(t, err)
		assert.Len(t, models, 2)
	})

	t.Run("NotCondition", func(t *testing.T) {
		var models []TestModel
		err := db.Not("name = ?", "B").Find(&models).Error()
		require.NoError(t, err)
		assert.Len(t, models, 10) // Only category A
	})

	t.Run("OrderAscending", func(t *testing.T) {
		var models []TestModel
		err := db.Order("value ASC").Find(&models).Error()
		require.NoError(t, err)
		assert.Equal(t, 1, models[0].Value)
	})

	t.Run("OrderDescending", func(t *testing.T) {
		var models []TestModel
		err := db.Order("value DESC").Find(&models).Error()
		require.NoError(t, err)
		assert.Equal(t, 20, models[0].Value)
	})

	t.Run("LimitAndOffset", func(t *testing.T) {
		var models []TestModel
		err := db.Order("value ASC").Limit(5).Offset(5).Find(&models).Error()
		require.NoError(t, err)
		assert.Len(t, models, 5)
		assert.Equal(t, 6, models[0].Value)
	})

	t.Run("GroupAndHaving", func(t *testing.T) {
		var results []struct {
			Name  string
			Total int64
		}
		err := db.Model(&TestModel{}).
			Select("name, SUM(value) as total").
			Group("name").
			Having("SUM(value) > ?", 50).
			Scan(&results).Error()
		require.NoError(t, err)
		assert.Len(t, results, 2) // Both A and B have sum > 50
	})

	t.Run("MultipleSelects", func(t *testing.T) {
		var model TestModel
		err := db.MultipleSelects("name", "value").First(&model).Error()
		require.NoError(t, err)
		assert.NotEmpty(t, model.Name)
	})

	t.Run("First", func(t *testing.T) {
		var model TestModel
		err := db.Order("value ASC").First(&model).Error()
		require.NoError(t, err)
		assert.Equal(t, 1, model.Value)
	})

	t.Run("Last", func(t *testing.T) {
		var model TestModel
		// Last returns last record by primary key when no order specified
		err := db.Last(&model).Error()
		require.NoError(t, err)
		assert.NotZero(t, model.ID)
	})

	t.Run("Unscoped", func(t *testing.T) {
		// Unscoped should work without soft deletes
		var models []TestModel
		err := db.Unscoped().Find(&models).Error()
		require.NoError(t, err)
		assert.Len(t, models, 20)
	})

	t.Run("Table", func(t *testing.T) {
		var count int64
		err := db.Table("test_models").Count(&count).Error()
		require.NoError(t, err)
		assert.Equal(t, int64(20), count)
	})

	t.Run("Model", func(t *testing.T) {
		var count int64
		err := db.Model(&TestModel{}).Count(&count).Error()
		require.NoError(t, err)
		assert.Equal(t, int64(20), count)
	})
}

// TestRawSQLExecutionErrorHandling tests raw SQL execution and error handling
func TestRawSQLExecutionErrorHandling(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	t.Run("ValidRawQuery", func(t *testing.T) {
		db.Create(&TestModel{Name: "raw_test", Value: 42})

		var model TestModel
		err := db.Raw("SELECT * FROM test_models WHERE name = ?", "raw_test").Scan(&model).Error()
		require.NoError(t, err)
		assert.Equal(t, 42, model.Value)
	})

	t.Run("RawQueryWithNoResults", func(t *testing.T) {
		var model TestModel
		result := db.Raw("SELECT * FROM test_models WHERE name = ?", "nonexistent").Scan(&model)
		// Scan doesn't return ErrRecordNotFound for Raw queries with no results
		assert.NoError(t, result.Error())
	})

	t.Run("InvalidRawQuery", func(t *testing.T) {
		var models []TestModel
		result := db.Raw("SELECT * FROM nonexistent_table").Scan(&models)
		assert.Error(t, result.Error())
	})

	t.Run("ValidExec", func(t *testing.T) {
		result := db.Exec("INSERT INTO test_models (name, value) VALUES (?, ?)", "exec_test", 99)
		require.NoError(t, result.Error())
		assert.Equal(t, int64(1), result.RowsAffected())
	})

	t.Run("ExecUpdate", func(t *testing.T) {
		db.Create(&TestModel{Name: "update_test", Value: 1})

		result := db.Exec("UPDATE test_models SET value = ? WHERE name = ?", 100, "update_test")
		require.NoError(t, result.Error())
		assert.Equal(t, int64(1), result.RowsAffected())
	})

	t.Run("ExecDelete", func(t *testing.T) {
		db.Create(&TestModel{Name: "delete_test", Value: 1})

		result := db.Exec("DELETE FROM test_models WHERE name = ?", "delete_test")
		require.NoError(t, result.Error())
		assert.Equal(t, int64(1), result.RowsAffected())
	})

	t.Run("InvalidExec", func(t *testing.T) {
		result := db.Exec("INSERT INTO nonexistent_table (col) VALUES (?)", "value")
		assert.Error(t, result.Error())
	})

	t.Run("RawWithRows", func(t *testing.T) {
		db.Create(&TestModel{Name: "rows_test1", Value: 1})
		db.Create(&TestModel{Name: "rows_test2", Value: 2})

		rows, err := db.Raw("SELECT name, value FROM test_models WHERE name LIKE ?", "rows_test%").Rows()
		require.NoError(t, err)
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		assert.Equal(t, 2, count)
	})

	t.Run("RawWithRow", func(t *testing.T) {
		db.Create(&TestModel{Name: "row_test", Value: 123})

		row := db.Raw("SELECT value FROM test_models WHERE name = ?", "row_test").Row()
		var value int
		err := row.Scan(&value)
		require.NoError(t, err)
		assert.Equal(t, 123, value)
	})

	t.Run("ScanRows", func(t *testing.T) {
		db.Create(&TestModel{Name: "scanrows_test", Value: 456})

		rows, err := db.Raw("SELECT * FROM test_models WHERE name = ?", "scanrows_test").Rows()
		require.NoError(t, err)
		defer rows.Close()

		for rows.Next() {
			var model TestModel
			err := db.ScanRows(rows, &model)
			require.NoError(t, err)
			assert.Equal(t, 456, model.Value)
		}
	})
}

// TestDbStatus tests the Status method for different scenarios
func TestDbStatus(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	t.Run("StatusOK", func(t *testing.T) {
		db.Create(&TestModel{Name: "status_test", Value: 1})

		var model TestModel
		result := db.First(&model)
		assert.Equal(t, 200, result.Status())
	})

	t.Run("StatusNotFound", func(t *testing.T) {
		var model TestModel
		result := db.Where("id = ?", 99999).First(&model)
		assert.Equal(t, 404, result.Status())
	})

	t.Run("StatusError", func(t *testing.T) {
		var model TestModel
		result := db.Table("nonexistent").First(&model)
		assert.Equal(t, 500, result.Status())
	})
}

// TestReadOnlyMode tests that write operations are blocked in read-only mode
func TestReadOnlyMode(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	// Create a read-only wrapper
	gormDB := db.GormDB()
	readOnlyDb := &Db{
		Database: gormDB,
		Type:     "sqlite",
		ReadOnly: true,
	}

	t.Run("CreateBlocked", func(t *testing.T) {
		result := readOnlyDb.Create(&TestModel{Name: "readonly_test", Value: 1})
		// Should not actually create
		var count int64
		db.Model(&TestModel{}).Where("name = ?", "readonly_test").Count(&count)
		assert.Equal(t, int64(0), count)
		assert.Equal(t, readOnlyDb, result) // Returns self without action
	})

	t.Run("UpdateBlocked", func(t *testing.T) {
		// First insert via writable db
		db.Create(&TestModel{Name: "readonly_update", Value: 1})

		// ReadOnly blocks at the Update call level
		readOnlyDb.Model(&TestModel{}).Where("name = ?", "readonly_update").Update("value", 100)

		// Value should remain unchanged because ReadOnly mode blocks writes
		var model TestModel
		db.Where("name = ?", "readonly_update").First(&model)
		assert.Equal(t, 1, model.Value)
	})

	t.Run("DeleteBlocked", func(t *testing.T) {
		db.Create(&TestModel{Name: "readonly_delete", Value: 1})

		readOnlyDb.Where("name = ?", "readonly_delete").Delete(&TestModel{})

		// Record should still exist because ReadOnly mode blocks deletes
		var count int64
		db.Model(&TestModel{}).Where("name = ?", "readonly_delete").Count(&count)
		assert.Equal(t, int64(1), count)
	})

	t.Run("SaveBlocked", func(t *testing.T) {
		model := &TestModel{Name: "readonly_save", Value: 1}
		result := readOnlyDb.Save(model)
		assert.Equal(t, readOnlyDb, result)
	})

	t.Run("BeginBlockedInReadOnly", func(t *testing.T) {
		result := readOnlyDb.Begin()
		assert.Equal(t, readOnlyDb, result)
	})

	t.Run("CommitBlockedInReadOnly", func(t *testing.T) {
		result := readOnlyDb.Commit()
		assert.Equal(t, readOnlyDb, result)
	})

	t.Run("RollbackBlockedInReadOnly", func(t *testing.T) {
		result := readOnlyDb.Rollback()
		assert.Equal(t, readOnlyDb, result)
	})

	t.Run("AutoMigrateBlockedInReadOnly", func(t *testing.T) {
		type NewModel struct {
			ID int64 `gorm:"primaryKey"`
		}
		result := readOnlyDb.AutoMigrate(&NewModel{})
		assert.Equal(t, readOnlyDb, result)
		assert.False(t, db.HasTable(&NewModel{}))
	})
}

// TestTimeFormatting tests time formatting for different database types
func TestTimeFormatting(t *testing.T) {
	now := time.Date(2023, 6, 15, 10, 30, 45, 123456789, time.UTC)

	t.Run("SQLiteFormat", func(t *testing.T) {
		db := &Db{Type: "sqlite3"}
		formatted := db.FormatTime(now)
		assert.Equal(t, "2023-06-15 10:30:45", formatted)
	})

	t.Run("PostgresFormat", func(t *testing.T) {
		db := &Db{Type: "postgres"}
		formatted := db.FormatTime(now)
		assert.Equal(t, "2023-06-15 10:30:45.123456789", formatted)
	})

	t.Run("MySQLFormat", func(t *testing.T) {
		db := &Db{Type: "mysql"}
		formatted := db.FormatTime(now)
		assert.Equal(t, "2023-06-15 10:30:45", formatted)
	})
}

// TestTimeParsing tests time parsing for different database types
func TestTimeParsing(t *testing.T) {
	t.Run("SQLiteParse", func(t *testing.T) {
		db := &Db{Type: "sqlite3"}
		parsed, err := db.ParseTime("2023-06-15 10:30:45")
		require.NoError(t, err)
		assert.Equal(t, 2023, parsed.Year())
		assert.Equal(t, time.June, parsed.Month())
		assert.Equal(t, 15, parsed.Day())
	})

	t.Run("PostgresParse", func(t *testing.T) {
		db := &Db{Type: "postgres"}
		parsed, err := db.ParseTime("2023-06-15T10:30:45Z")
		require.NoError(t, err)
		assert.Equal(t, 2023, parsed.Year())
	})

	t.Run("MySQLParse", func(t *testing.T) {
		db := &Db{Type: "mysql"}
		parsed, err := db.ParseTime("2023-06-15T10:30:45Z")
		require.NoError(t, err)
		assert.Equal(t, 2023, parsed.Year())
	})

	t.Run("InvalidTimeParse", func(t *testing.T) {
		db := &Db{Type: "sqlite3"}
		_, err := db.ParseTime("invalid-time")
		assert.Error(t, err)
	})
}

// TestSelectByTime tests the SelectByTime SQL generation
func TestSelectByTime(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) {
		db := &Db{Type: "sqlite3"}
		sql := db.SelectByTime(5 * time.Minute)
		assert.Contains(t, sql, "strftime")
		assert.Contains(t, sql, "300") // 5 minutes = 300 seconds
	})

	t.Run("MySQL", func(t *testing.T) {
		db := &Db{Type: "mysql"}
		sql := db.SelectByTime(5 * time.Minute)
		assert.Contains(t, sql, "FROM_UNIXTIME")
		assert.Contains(t, sql, "300")
	})

	t.Run("Postgres", func(t *testing.T) {
		db := &Db{Type: "postgres"}
		sql := db.SelectByTime(5 * time.Minute)
		assert.Contains(t, sql, "date_trunc")
	})
}

// TestBeginFunction tests the Begin helper function
func TestBeginFunction(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	t.Run("BeginWithMigration", func(t *testing.T) {
		tx := Begin(db, "migration")
		assert.NotNil(t, tx)
		tx.Rollback()
	})

	t.Run("BeginWithModel", func(t *testing.T) {
		tx := Begin(db, &TestModel{})
		assert.NotNil(t, tx)
		tx.Rollback()
	})
}

// TestLogModeFunction tests the LogMode helper function
func TestLogModeFunction(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	// Should not panic
	result := LogMode(db, true)
	assert.NotNil(t, result)

	result = LogMode(db, false)
	assert.NotNil(t, result)
}

// TestDbNew tests creating a new session
func TestDbNew(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	newDb := db.New()
	assert.NotNil(t, newDb)
	assert.Equal(t, db.DbType(), newDb.DbType())
}

// TestDbDebug tests debug mode
func TestDbDebug(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	debugDb := db.Debug()
	assert.NotNil(t, debugDb)
}

// TestDbDb tests the Db() method
func TestDbDb(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	dbInterface := db.Db()
	assert.NotNil(t, dbInterface)
	assert.Equal(t, db, dbInterface)
}

// TestDbSetAndGet tests Set and Get methods
func TestDbSetAndGet(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	dbWithSetting := db.Set("test_key", "test_value")
	assert.NotNil(t, dbWithSetting)

	value, ok := dbWithSetting.Get("test_key")
	assert.True(t, ok)
	assert.Equal(t, "test_value", value)
}

// TestDbOmit tests the Omit method
func TestDbOmit(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	// Insert test data
	db.Create(&TestModel{Name: "omit_test", Value: 100})

	// Omit should work without error
	result := db.Model(&TestModel{}).Omit("value").Where("name = ?", "omit_test").Updates(map[string]interface{}{"name": "omit_updated"})
	require.NoError(t, result.Error())
}

// TestDbJoins tests the Joins method
func TestDbJoins(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	// Joins method should return a valid Database object
	result := db.Joins("LEFT JOIN other_table ON other_table.id = test_models.id")
	assert.NotNil(t, result)
}

// TestDbPreload tests the Preload method
func TestDbPreload(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	// Preload method should return a valid Database object
	result := db.Preload("RelatedModel")
	assert.NotNil(t, result)
}

// TestDbAddError tests the AddError method
func TestDbAddError(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	testErr := errors.New("test error")
	err = db.AddError(testErr)
	assert.Error(t, err)
	assert.Equal(t, testErr, err)
}

// TestUpdateMethods tests various Update methods
func TestUpdateMethods(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	t.Run("UpdateWithSingleArg", func(t *testing.T) {
		db.Create(&TestModel{Name: "update_single", Value: 1})

		result := db.Model(&TestModel{}).Where("name = ?", "update_single").Update(map[string]interface{}{"value": 50})
		require.NoError(t, result.Error())

		var model TestModel
		db.Where("name = ?", "update_single").First(&model)
		assert.Equal(t, 50, model.Value)
	})

	t.Run("UpdateWithTwoArgs", func(t *testing.T) {
		db.Create(&TestModel{Name: "update_two", Value: 1})

		result := db.Model(&TestModel{}).Where("name = ?", "update_two").Update("value", 75)
		require.NoError(t, result.Error())

		var model TestModel
		db.Where("name = ?", "update_two").First(&model)
		assert.Equal(t, 75, model.Value)
	})

	t.Run("UpdateColumn", func(t *testing.T) {
		db.Create(&TestModel{Name: "update_column", Value: 1})

		result := db.Model(&TestModel{}).Where("name = ?", "update_column").UpdateColumn("value", 200)
		require.NoError(t, result.Error())

		var model TestModel
		db.Where("name = ?", "update_column").First(&model)
		assert.Equal(t, 200, model.Value)
	})

	t.Run("UpdateColumns", func(t *testing.T) {
		db.Create(&TestModel{Name: "update_columns", Value: 1})

		result := db.Model(&TestModel{}).Where("name = ?", "update_columns").UpdateColumns(map[string]interface{}{"value": 300})
		require.NoError(t, result.Error())

		var model TestModel
		db.Where("name = ?", "update_columns").First(&model)
		assert.Equal(t, 300, model.Value)
	})
}

// TestWrapNilDialector tests Wrap with nil dialector
func TestWrapNilDialector(t *testing.T) {
	// Wrap with nil should default to sqlite3
	wrapped := Wrap(nil)
	assert.NotNil(t, wrapped)
	assert.Equal(t, "sqlite3", wrapped.DbType())
}

// TestDbTypeGlobalFunction tests the global DbType function
func TestDbTypeGlobalFunction(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	Set(db)
	assert.Equal(t, "sqlite", DbType())
}
