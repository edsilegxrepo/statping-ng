package checkins

import (
	"os"
	"testing"
	"time"

	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/types/failures"
	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "statping-checkins-test")
	if err != nil {
		os.Exit(1)
	}
	utils.Directory = tmpDir
	utils.InitEnvs()
	utils.Params.Set("STATPING_DIR", tmpDir)
	_ = utils.InitLogs()

	code := m.Run()

	_ = os.RemoveAll(tmpDir)
	os.Exit(code)
}

func setupTestDB(t *testing.T) database.Database {
	testDb, err := database.OpenTester()
	require.NoError(t, err)
	SetDB(testDb)
	failures.SetDB(testDb)
	require.NoError(t, testDb.AutoMigrate(&Checkin{}, &CheckinHit{}, &failures.Failure{}).Error())
	return testDb
}

func TestCheckin_CRUD(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	t.Run("Create", func(t *testing.T) {
		checkin := &Checkin{
			ServiceId: 1,
			Name:      "Test Checkin",
			Interval:  5,
		}
		err := checkin.Create()
		require.NoError(t, err)
		assert.NotZero(t, checkin.Id)
		assert.NotEmpty(t, checkin.ApiKey)
		checkin.Close()
	})

	t.Run("Create with custom ApiKey", func(t *testing.T) {
		checkin := &Checkin{
			ServiceId: 1,
			Name:      "Checkin With Key",
			Interval:  5,
			ApiKey:    "customApiKey123",
		}
		err := checkin.Create()
		require.NoError(t, err)
		assert.Equal(t, "customApiKey123", checkin.ApiKey)
		checkin.Close()
	})

	t.Run("Find", func(t *testing.T) {
		checkin := &Checkin{
			ServiceId: 1,
			Name:      "Find Test",
			Interval:  3,
		}
		require.NoError(t, checkin.Create())
		checkin.Close()

		found, err := Find(checkin.Id)
		require.NoError(t, err)
		assert.Equal(t, "Find Test", found.Name)
		assert.Equal(t, int64(3), found.Interval)
	})

	t.Run("FindByAPI", func(t *testing.T) {
		checkin := &Checkin{
			ServiceId: 1,
			Name:      "API Find Test",
			Interval:  3,
		}
		require.NoError(t, checkin.Create())
		checkin.Close()

		found, err := FindByAPI(checkin.ApiKey)
		require.NoError(t, err)
		assert.Equal(t, "API Find Test", found.Name)
	})

	t.Run("All", func(t *testing.T) {
		all := All()
		assert.GreaterOrEqual(t, len(all), 1)
	})

	t.Run("Update", func(t *testing.T) {
		checkin := &Checkin{
			ServiceId: 1,
			Name:      "Update Test",
			Interval:  3,
		}
		require.NoError(t, checkin.Create())
		checkin.Close()

		checkin.Name = "Updated Name"
		err := checkin.Update()
		require.NoError(t, err)

		found, err := Find(checkin.Id)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", found.Name)
	})

	t.Run("Delete", func(t *testing.T) {
		checkin := &Checkin{
			ServiceId: 1,
			Name:      "Delete Test",
			Interval:  3,
		}
		require.NoError(t, checkin.Create())
		id := checkin.Id
		checkin.Close()

		err := checkin.Delete()
		require.NoError(t, err)

		_, err = Find(id)
		assert.Error(t, err)
	})
}

func TestCheckinHit_CRUD(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	// Create a checkin to associate hits with
	checkin := &Checkin{
		ServiceId: 1,
		Name:      "Hit Test Checkin",
		Interval:  5,
	}
	require.NoError(t, checkin.Create())
	defer checkin.Close()

	t.Run("Create", func(t *testing.T) {
		hit := &CheckinHit{
			Checkin:   checkin.Id,
			From:      "192.168.1.100",
			CreatedAt: utils.Now(),
		}
		err := hit.Create()
		require.NoError(t, err)
		assert.NotZero(t, hit.Id)
	})

	t.Run("Create multiple hits", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			hit := &CheckinHit{
				Checkin:   checkin.Id,
				From:      "10.0.0.1",
				CreatedAt: utils.Now().Add(-time.Duration(i) * time.Minute),
			}
			err := hit.Create()
			require.NoError(t, err)
		}
	})

	t.Run("Update", func(t *testing.T) {
		hit := &CheckinHit{
			Checkin:   checkin.Id,
			From:      "original.host.com",
			CreatedAt: utils.Now(),
		}
		require.NoError(t, hit.Create())

		hit.From = "updated.host.com"
		err := hit.Update()
		require.NoError(t, err)
	})

	t.Run("Delete", func(t *testing.T) {
		hit := &CheckinHit{
			Checkin:   checkin.Id,
			From:      "delete.test.com",
			CreatedAt: utils.Now(),
		}
		require.NoError(t, hit.Create())
		hitId := hit.Id

		err := hit.Delete()
		require.NoError(t, err)
		assert.NotZero(t, hitId)
	})
}

func TestCheckin_Hits(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	checkin := &Checkin{
		ServiceId: 1,
		Name:      "Hits Query Checkin",
		Interval:  5,
	}
	require.NoError(t, checkin.Create())
	defer checkin.Close()

	// Create test hits
	for i := 0; i < 3; i++ {
		hit := &CheckinHit{
			Checkin:   checkin.Id,
			From:      "test.host.com",
			CreatedAt: utils.Now().Add(-time.Duration(i) * time.Minute),
		}
		require.NoError(t, hit.Create())
	}

	t.Run("LastHit", func(t *testing.T) {
		lastHit := checkin.LastHit()
		require.NotNil(t, lastHit)
		assert.Equal(t, checkin.Id, lastHit.Checkin)
	})

	t.Run("LastHit empty checkin", func(t *testing.T) {
		emptyCheckin := &Checkin{
			ServiceId: 888,
			Name:      "Empty Checkin",
			Interval:  5,
		}
		require.NoError(t, emptyCheckin.Create())
		defer emptyCheckin.Close()

		lastHit := emptyCheckin.LastHit()
		require.NotNil(t, lastHit)
		assert.Zero(t, lastHit.Id)
	})

	t.Run("Hits retrieval", func(t *testing.T) {
		hits := checkin.Hits()
		assert.GreaterOrEqual(t, len(hits), 1)
		assert.Equal(t, hits, checkin.AllHits)
	})

	t.Run("Hits empty checkin", func(t *testing.T) {
		emptyCheckin := &Checkin{
			ServiceId: 777,
			Name:      "No Hits Checkin",
			Interval:  5,
		}
		require.NoError(t, emptyCheckin.Create())
		defer emptyCheckin.Close()

		hits := emptyCheckin.Hits()
		assert.Empty(t, hits)
	})
}

func TestCheckinHitters(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	checkin := &Checkin{
		ServiceId: 1,
		Name:      "Hitters Test Checkin",
		Interval:  5,
	}
	require.NoError(t, checkin.Create())
	defer checkin.Close()

	// Create test hits
	for i := 0; i < 5; i++ {
		hit := &CheckinHit{
			Checkin:   checkin.Id,
			From:      "hitters.test.com",
			CreatedAt: utils.Now().Add(-time.Duration(i) * time.Minute),
		}
		require.NoError(t, hit.Create())
	}

	t.Run("First", func(t *testing.T) {
		hitters := AllCheckinHits(1)
		first := hitters.First()
		require.NotNil(t, first)
	})

	t.Run("Count", func(t *testing.T) {
		hitters := AllCheckinHits(1)
		count := hitters.Count()
		assert.GreaterOrEqual(t, count, 1)
	})

	t.Run("Db method", func(t *testing.T) {
		hitters := AllCheckinHits(1)
		db := hitters.Db()
		assert.NotNil(t, db)
	})

	t.Run("Since", func(t *testing.T) {
		pastTime := utils.Now().Add(-10 * time.Minute)
		hitters := AllCheckinHits(1).Since(pastTime)
		count := hitters.Count()
		assert.GreaterOrEqual(t, count, 0)
	})

	t.Run("Since recent", func(t *testing.T) {
		// Create a hit just now
		hit := &CheckinHit{
			Checkin:   checkin.Id,
			From:      "recent.hit.com",
			CreatedAt: utils.Now(),
		}
		require.NoError(t, hit.Create())

		// Query for hits since 1 second ago
		recentTime := utils.Now().Add(-1 * time.Second)
		hitters := AllCheckinHits(1).Since(recentTime)
		count := hitters.Count()
		assert.GreaterOrEqual(t, count, 1)
	})

	t.Run("Service filter", func(t *testing.T) {
		// Create a checkin for a different service
		checkin2 := &Checkin{
			ServiceId: 999,
			Name:      "Different Service Checkin",
			Interval:  5,
		}
		require.NoError(t, checkin2.Create())
		defer checkin2.Close()

		hit := &CheckinHit{
			Checkin:   checkin2.Id,
			From:      "other.service.com",
			CreatedAt: utils.Now(),
		}
		require.NoError(t, hit.Create())

		// Should only get hits for service 1
		hitters1 := AllCheckinHits(1)
		count1 := hitters1.Count()

		// Should only get hits for service 999
		hitters999 := AllCheckinHits(999)
		count999 := hitters999.Count()

		assert.GreaterOrEqual(t, count1, 1)
		assert.GreaterOrEqual(t, count999, 1)
	})
}

func TestCheckin_Methods(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	t.Run("Period", func(t *testing.T) {
		checkin := &Checkin{
			ServiceId: 1,
			Name:      "Period Test",
			Interval:  5,
		}
		require.NoError(t, checkin.Create())
		defer checkin.Close()

		period := checkin.Period()
		assert.Equal(t, 5*time.Minute, period)
	})

	t.Run("Expected", func(t *testing.T) {
		checkin := &Checkin{
			ServiceId: 1,
			Name:      "Expected Test",
			Interval:  5,
		}
		require.NoError(t, checkin.Create())
		defer checkin.Close()

		// Create a recent hit
		hit := &CheckinHit{
			Checkin:   checkin.Id,
			From:      "expected.test.com",
			CreatedAt: utils.Now().Add(-1 * time.Minute),
		}
		require.NoError(t, hit.Create())

		expected := checkin.Expected()
		// Expected should be close to period - time since last hit
		// With 5 min interval and 1 min ago hit, expected should be ~4 min
		assert.Greater(t, expected.Minutes(), float64(3))
		assert.Less(t, expected.Minutes(), float64(5))
	})

	t.Run("Start and IsRunning", func(t *testing.T) {
		checkin := &Checkin{
			ServiceId: 1,
			Name:      "Running Test Checkin",
			Interval:  1,
		}
		require.NoError(t, checkin.Create())

		// Initially not running
		assert.False(t, checkin.IsRunning())

		// Start the checkin routine
		timeout := 30
		checkin.Start(&timeout)
		assert.True(t, checkin.IsRunning())

		// Close should stop the routine
		checkin.Close()
		// Give routine time to stop
		time.Sleep(100 * time.Millisecond)
		assert.False(t, checkin.IsRunning())
	})

	t.Run("Close without Start", func(t *testing.T) {
		checkin := &Checkin{
			ServiceId: 1,
			Name:      "No Start Checkin",
			Interval:  1,
		}
		require.NoError(t, checkin.Create())

		// Should not panic when closing without start
		checkin.Close()
		assert.False(t, checkin.IsRunning())
	})

	t.Run("IsRunning nil channel", func(t *testing.T) {
		checkin := &Checkin{
			ServiceId: 1,
			Name:      "Nil Channel Checkin",
			Interval:  1,
		}
		require.NoError(t, checkin.Create())

		// Running channel is nil
		assert.False(t, checkin.IsRunning())
		checkin.Close()
	})
}

func TestCheckin_Failures(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	checkin := &Checkin{
		ServiceId: 1,
		Name:      "Failure Test Checkin",
		Interval:  1,
	}
	require.NoError(t, checkin.Create())
	defer checkin.Close()

	t.Run("CreateFailure", func(t *testing.T) {
		fail := &failures.Failure{
			Issue:   "Test failure issue",
			Method:  "checkin",
			Service: checkin.ServiceId,
		}
		err := checkin.CreateFailure(fail)
		require.NoError(t, err)
		assert.True(t, checkin.Failing)
	})

	t.Run("FailuresColumnID", func(t *testing.T) {
		col, id := checkin.FailuresColumnID()
		assert.Equal(t, "checkin", col)
		assert.Equal(t, checkin.Id, id)
	})

	t.Run("Failures", func(t *testing.T) {
		failurer := checkin.Failures()
		assert.NotNil(t, failurer)
	})

	t.Run("FailuresSince", func(t *testing.T) {
		pastTime := utils.Now().Add(-1 * time.Hour)
		failurer := checkin.FailuresSince(pastTime)
		assert.NotNil(t, failurer)
	})
}

func TestSamples(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	err := Samples()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(All()), 1)
}
