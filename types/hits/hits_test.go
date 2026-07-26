package hits

import (
	"testing"
	"time"

	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	_ = utils.InitLogs()
	m.Run()
}

type mockColumnID struct {
	column string
	id     int64
}

func (m mockColumnID) HitsColumnID() (string, int64) {
	return m.column, m.id
}

func setupTestDB(t *testing.T) database.Database {
	testDb, err := database.OpenTester()
	require.NoError(t, err)
	require.NoError(t, testDb.AutoMigrate(&Hit{}).Error())
	SetDB(testDb)
	return testDb
}

func TestHit_BeforeCreate(t *testing.T) {
	t.Run("sets CreatedAt when zero", func(t *testing.T) {
		h := &Hit{
			Service:  1,
			Latency:  100,
			PingTime: 50,
		}
		err := h.BeforeCreate(&gorm.DB{})
		assert.Nil(t, err)
		assert.False(t, h.CreatedAt.IsZero())
		assert.True(t, time.Since(h.CreatedAt) < time.Second)
	})

	t.Run("preserves existing CreatedAt", func(t *testing.T) {
		existingTime := time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)
		h := &Hit{
			Service:   1,
			Latency:   100,
			PingTime:  50,
			CreatedAt: existingTime,
		}
		err := h.BeforeCreate(&gorm.DB{})
		assert.Nil(t, err)
		assert.Equal(t, existingTime, h.CreatedAt)
	})
}

func TestHit_AfterFind(t *testing.T) {
	h := &Hit{}
	err := h.AfterFind(&gorm.DB{})
	assert.NoError(t, err)
}

func TestHit_CRUD(t *testing.T) {
	testDb := setupTestDB(t)
	_ = testDb

	t.Run("Create hit", func(t *testing.T) {
		h := &Hit{
			Service:  1,
			Latency:  150,
			PingTime: 75,
		}
		err := h.Create()
		assert.Nil(t, err)
		assert.True(t, h.Id > 0)
	})

	t.Run("Update hit", func(t *testing.T) {
		h := &Hit{
			Service:  1,
			Latency:  200,
			PingTime: 100,
		}
		err := h.Create()
		require.Nil(t, err)

		h.Latency = 250
		err = h.Update()
		assert.Nil(t, err)
	})

	t.Run("Delete hit", func(t *testing.T) {
		h := &Hit{
			Service:  1,
			Latency:  300,
			PingTime: 150,
		}
		err := h.Create()
		require.Nil(t, err)

		err = h.Delete()
		assert.Nil(t, err)
	})
}

func TestHitters(t *testing.T) {
	testDb := setupTestDB(t)

	for i := 0; i < 5; i++ {
		h := &Hit{
			Service:   1,
			Latency:   int64(100 + i*10),
			PingTime:  int64(50 + i*5),
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
		}
		require.NoError(t, h.Create())
	}

	for i := 0; i < 3; i++ {
		h := &Hit{
			Service:   2,
			Latency:   int64(200 + i*20),
			PingTime:  int64(100 + i*10),
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
		}
		require.NoError(t, h.Create())
	}

	t.Run("Db returns database", func(t *testing.T) {
		hitters := Hitters{db: testDb.Model(&Hit{})}
		assert.NotNil(t, hitters.Db())
	})

	t.Run("First returns first hit", func(t *testing.T) {
		hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 1)}
		first := hitters.First()
		assert.NotNil(t, first)
		assert.Equal(t, int64(1), first.Service)
	})

	t.Run("Last returns last hit", func(t *testing.T) {
		hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 1)}
		last := hitters.Last()
		assert.NotNil(t, last)
		assert.Equal(t, int64(1), last.Service)
	})

	t.Run("List returns all hits", func(t *testing.T) {
		hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 1)}
		list := hitters.List()
		assert.Len(t, list, 5)
	})

	t.Run("Count returns hit count", func(t *testing.T) {
		hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 1)}
		count := hitters.Count()
		assert.Equal(t, 5, count)
	})

	t.Run("LastAmount returns limited hits", func(t *testing.T) {
		hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 1)}
		limited := hitters.LastAmount(3)
		assert.Len(t, limited, 3)
	})

	t.Run("Sum returns sum of latency", func(t *testing.T) {
		hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 1)}
		sum := hitters.Sum()
		assert.True(t, sum > 0)
	})

	t.Run("Avg returns average latency", func(t *testing.T) {
		hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 1)}
		avg := hitters.Avg()
		assert.True(t, avg > 0)
	})

	t.Run("Since returns hits since time", func(t *testing.T) {
		hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 1)}
		since := hitters.Since(time.Now().Add(-2 * time.Hour))
		assert.True(t, len(since) > 0)
	})

	t.Run("DeleteAll removes all hits", func(t *testing.T) {
		hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 2)}
		err := hitters.DeleteAll()
		assert.NoError(t, err)

		count := hitters.Count()
		assert.Equal(t, 0, count)
	})
}

func TestAllHits(t *testing.T) {
	testDb := setupTestDB(t)
	_ = testDb

	for i := 0; i < 3; i++ {
		h := &Hit{
			Service:  3,
			Latency:  int64(100 + i*10),
			PingTime: int64(50 + i*5),
		}
		require.NoError(t, h.Create())
	}

	mock := mockColumnID{column: "service", id: 3}
	hitters := AllHits(mock)
	assert.NotNil(t, hitters)
	assert.Equal(t, 3, hitters.Count())
}

func TestSince(t *testing.T) {
	testDb := setupTestDB(t)
	_ = testDb

	for i := 0; i < 3; i++ {
		h := &Hit{
			Service:   4,
			Latency:   int64(100 + i*10),
			PingTime:  int64(50 + i*5),
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
		}
		require.NoError(t, h.Create())
	}

	mock := mockColumnID{column: "service", id: 4}
	hitters := Since(time.Now().Add(-2*time.Hour), mock)
	assert.NotNil(t, hitters)
}
