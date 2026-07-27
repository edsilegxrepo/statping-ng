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

func TestCreateHitsAt(t *testing.T) {
	t.Run("creates hits for specified duration", func(t *testing.T) {
		// Create hits for 1 hour with 10 minute intervals
		records := createHitsAt(100, -1*time.Hour, 10*time.Minute)

		// Should have approximately 6-7 hits (60 mins / 10 min intervals)
		assert.True(t, len(records) >= 6, "expected at least 6 records, got %d", len(records))

		// All records should have the correct service ID
		for _, hit := range records {
			assert.Equal(t, int64(100), hit.Service)
			// Perlin noise can generate negative values, so latency may be negative
			// Just verify the fields are set (non-zero or zero is fine based on noise)
			assert.False(t, hit.CreatedAt.IsZero(), "CreatedAt should be set")
		}
	})

	t.Run("creates hits in chronological order", func(t *testing.T) {
		records := createHitsAt(101, -30*time.Minute, 5*time.Minute)

		// Verify hits are in chronological order (oldest first)
		for i := 1; i < len(records); i++ {
			assert.True(t, records[i].CreatedAt.After(records[i-1].CreatedAt) ||
				records[i].CreatedAt.Equal(records[i-1].CreatedAt),
				"hits should be in chronological order")
		}
	})

	t.Run("respects SampleHits limit", func(t *testing.T) {
		originalLimit := SampleHits
		SampleHits = 10
		defer func() { SampleHits = originalLimit }()

		// Create hits for a very long duration that would exceed limit
		records := createHitsAt(102, -24*time.Hour, 1*time.Minute)

		// Should be capped at SampleHits + 1
		assert.True(t, len(records) <= 11, "should respect SampleHits limit")
	})

	t.Run("stops when reaching current time", func(t *testing.T) {
		// Start from 5 minutes ago with 1 minute intervals
		records := createHitsAt(103, -5*time.Minute, 1*time.Minute)

		// Should have created some records
		assert.True(t, len(records) > 0, "should create at least one record")

		// First record should be from approximately 5 minutes ago
		firstRecord := records[0]
		assert.True(t, firstRecord.CreatedAt.Before(time.Now()),
			"first record should be in the past")
	})
}

func TestSamples(t *testing.T) {
	testDb := setupTestDB(t)
	_ = testDb

	// Reduce sample size for testing
	originalLimit := SampleHits
	SampleHits = 10
	defer func() { SampleHits = originalLimit }()

	t.Run("creates sample hits for multiple services", func(t *testing.T) {
		err := Samples()
		assert.NoError(t, err)

		// Verify hits were created for services 1-6
		for serviceID := int64(1); serviceID <= 6; serviceID++ {
			mock := mockColumnID{column: "service", id: serviceID}
			hitters := AllHits(mock)
			count := hitters.Count()
			assert.True(t, count > 0, "service %d should have hits", serviceID)
		}
	})
}

func TestHitters_TimeBasedQueries(t *testing.T) {
	testDb := setupTestDB(t)

	baseTime := time.Now()

	// Create hits at specific times for service 10
	times := []time.Duration{
		-5 * time.Hour,
		-4 * time.Hour,
		-3 * time.Hour,
		-2 * time.Hour,
		-1 * time.Hour,
		-30 * time.Minute,
	}

	for i, offset := range times {
		h := &Hit{
			Service:   10,
			Latency:   int64((i + 1) * 100),
			PingTime:  int64((i + 1) * 50),
			CreatedAt: baseTime.Add(offset),
		}
		require.NoError(t, h.Create())
	}

	t.Run("Since filters by time correctly", func(t *testing.T) {
		hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 10)}

		// Get hits from last 2 hours
		since := hitters.Since(baseTime.Add(-2 * time.Hour))
		assert.True(t, len(since) >= 2, "should have at least 2 recent hits")

		// All returned hits should be after the since time
		for _, hit := range since {
			assert.True(t, hit.CreatedAt.After(baseTime.Add(-2*time.Hour).Add(-time.Second)),
				"hit should be after since time")
		}
	})

	t.Run("First returns oldest hit", func(t *testing.T) {
		hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 10)}
		first := hitters.First()

		// First hit should be the one created 5 hours ago (lowest ID)
		assert.NotNil(t, first)
		assert.Equal(t, int64(100), first.Latency)
	})

	t.Run("Last returns newest hit", func(t *testing.T) {
		hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 10)}
		last := hitters.Last()

		// Last hit should be the one created 30 minutes ago (highest ID)
		assert.NotNil(t, last)
		assert.Equal(t, int64(600), last.Latency)
	})
}

func TestHitters_Calculations(t *testing.T) {
	testDb := setupTestDB(t)

	// Create hits with known latencies for service 20
	latencies := []int64{100, 200, 300, 400, 500}
	for _, lat := range latencies {
		h := &Hit{
			Service:   20,
			Latency:   lat,
			PingTime:  lat / 2,
			CreatedAt: time.Now(),
		}
		require.NoError(t, h.Create())
	}

	hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 20)}

	t.Run("Sum calculates total latency", func(t *testing.T) {
		sum := hitters.Sum()
		expectedSum := int64(100 + 200 + 300 + 400 + 500)
		assert.Equal(t, expectedSum, sum)
	})

	t.Run("Avg calculates average latency", func(t *testing.T) {
		avg := hitters.Avg()
		expectedAvg := int64((100 + 200 + 300 + 400 + 500) / 5)
		assert.Equal(t, expectedAvg, avg)
	})

	t.Run("Count returns correct count", func(t *testing.T) {
		count := hitters.Count()
		assert.Equal(t, 5, count)
	})
}

func TestHitters_EdgeCases(t *testing.T) {
	testDb := setupTestDB(t)

	t.Run("First on empty set returns zero hit", func(t *testing.T) {
		hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 999)}
		first := hitters.First()
		assert.NotNil(t, first)
		assert.Equal(t, int64(0), first.Id)
	})

	t.Run("Last on empty set returns zero hit", func(t *testing.T) {
		hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 999)}
		last := hitters.Last()
		assert.NotNil(t, last)
		assert.Equal(t, int64(0), last.Id)
	})

	t.Run("List on empty set returns empty slice", func(t *testing.T) {
		hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 999)}
		list := hitters.List()
		assert.Empty(t, list)
	})

	t.Run("Count on empty set returns zero", func(t *testing.T) {
		hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 999)}
		count := hitters.Count()
		assert.Equal(t, 0, count)
	})

	t.Run("Sum on empty set returns zero", func(t *testing.T) {
		hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 999)}
		sum := hitters.Sum()
		assert.Equal(t, int64(0), sum)
	})

	t.Run("Avg on empty set returns zero", func(t *testing.T) {
		hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 999)}
		avg := hitters.Avg()
		assert.Equal(t, int64(0), avg)
	})

	t.Run("DeleteAll on empty set succeeds", func(t *testing.T) {
		hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 999)}
		err := hitters.DeleteAll()
		assert.NoError(t, err)
	})

	t.Run("LastAmount with zero returns empty", func(t *testing.T) {
		// Create some hits first
		h := &Hit{Service: 30, Latency: 100, PingTime: 50}
		require.NoError(t, h.Create())

		hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 30)}
		limited := hitters.LastAmount(0)
		assert.Empty(t, limited)
	})

	t.Run("Since with future time returns empty", func(t *testing.T) {
		h := &Hit{Service: 31, Latency: 100, PingTime: 50, CreatedAt: time.Now()}
		require.NoError(t, h.Create())

		hitters := Hitters{db: testDb.Model(&Hit{}).Where("service = ?", 31)}
		since := hitters.Since(time.Now().Add(time.Hour))
		assert.Empty(t, since)
	})
}

func TestAllHits_MultipleServices(t *testing.T) {
	testDb := setupTestDB(t)
	_ = testDb

	// Create hits for multiple services
	for serviceID := int64(40); serviceID <= 42; serviceID++ {
		for i := 0; i < int(serviceID-39); i++ {
			h := &Hit{
				Service:  serviceID,
				Latency:  int64(100 * serviceID),
				PingTime: int64(50 * serviceID),
			}
			require.NoError(t, h.Create())
		}
	}

	t.Run("returns only hits for specified service", func(t *testing.T) {
		mock40 := mockColumnID{column: "service", id: 40}
		hitters40 := AllHits(mock40)
		assert.Equal(t, 1, hitters40.Count())

		mock41 := mockColumnID{column: "service", id: 41}
		hitters41 := AllHits(mock41)
		assert.Equal(t, 2, hitters41.Count())

		mock42 := mockColumnID{column: "service", id: 42}
		hitters42 := AllHits(mock42)
		assert.Equal(t, 3, hitters42.Count())
	})
}

func TestSinceFunction(t *testing.T) {
	testDb := setupTestDB(t)
	_ = testDb

	baseTime := time.Now()

	// Create hits at different times for service 50
	for i := 0; i < 5; i++ {
		h := &Hit{
			Service:   50,
			Latency:   int64(100 + i*50),
			PingTime:  int64(50 + i*25),
			CreatedAt: baseTime.Add(-time.Duration(i) * time.Hour),
		}
		require.NoError(t, h.Create())
	}

	t.Run("returns hitters filtered by service and time", func(t *testing.T) {
		mock := mockColumnID{column: "service", id: 50}
		hitters := Since(baseTime.Add(-2*time.Hour), mock)

		// Should return hits from the last 2 hours
		list := hitters.List()
		assert.True(t, len(list) > 0, "should have hits since 2 hours ago")

		for _, hit := range list {
			assert.Equal(t, int64(50), hit.Service)
		}
	})

	t.Run("returns empty for non-existent service", func(t *testing.T) {
		mock := mockColumnID{column: "service", id: 999}
		hitters := Since(baseTime.Add(-24*time.Hour), mock)
		assert.Equal(t, 0, hitters.Count())
	})
}

func TestHit_GormHooks(t *testing.T) {
	testDb := setupTestDB(t)
	_ = testDb

	t.Run("AfterUpdate hook executes", func(t *testing.T) {
		h := &Hit{Service: 60, Latency: 100, PingTime: 50}
		require.NoError(t, h.Create())

		h.Latency = 200
		err := h.AfterUpdate(&gorm.DB{})
		assert.NoError(t, err)
	})

	t.Run("AfterDelete hook executes", func(t *testing.T) {
		h := &Hit{Service: 61, Latency: 100, PingTime: 50}
		require.NoError(t, h.Create())

		err := h.AfterDelete(&gorm.DB{})
		assert.NoError(t, err)
	})

	t.Run("AfterCreate hook executes", func(t *testing.T) {
		h := &Hit{Service: 62, Latency: 100, PingTime: 50}
		err := h.AfterCreate(&gorm.DB{})
		assert.NoError(t, err)
	})
}
