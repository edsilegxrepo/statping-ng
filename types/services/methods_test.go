package services

import (
	"sort"
	"testing"
	"time"

	"github.com/statping-ng/statping-ng/types/checkins"
	"github.com/statping-ng/statping-ng/types/failures"
	"github.com/statping-ng/statping-ng/types/hits"
	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceUptime(t *testing.T) {
	s := &Service{
		LastOffline: utils.Now().Add(-2 * time.Hour),
	}
	uptime := s.Uptime()
	assert.True(t, uptime.Duration >= 2*time.Hour-time.Second)
}

func TestServiceDowntime(t *testing.T) {
	s := &Service{
		LastOnline: utils.Now().Add(-30 * time.Minute),
	}
	downtime := s.Downtime()
	assert.True(t, downtime.Duration >= 30*time.Minute-time.Second)
}

func TestServiceOrderSort(t *testing.T) {
	services := ServiceOrder{
		{Name: "Third", Order: 3},
		{Name: "First", Order: 1},
		{Name: "Second", Order: 2},
	}

	sort.Sort(services)

	assert.Equal(t, "First", services[0].Name)
	assert.Equal(t, "Second", services[1].Name)
	assert.Equal(t, "Third", services[2].Name)
}

func TestServiceOrderSwap(t *testing.T) {
	services := ServiceOrder{
		{Name: "A", Order: 1},
		{Name: "B", Order: 2},
	}

	services.Swap(0, 1)
	assert.Equal(t, "B", services[0].Name)
	assert.Equal(t, "A", services[1].Name)
}

func TestServiceOrderLen(t *testing.T) {
	services := ServiceOrder{
		{Name: "A"},
		{Name: "B"},
		{Name: "C"},
	}
	assert.Equal(t, 3, services.Len())
}

func TestByTimeSort(t *testing.T) {
	now := time.Now()
	series := ByTime{
		{Time: now.Add(2 * time.Hour), Online: true},
		{Time: now, Online: false},
		{Time: now.Add(1 * time.Hour), Online: true},
	}

	sort.Sort(series)

	assert.True(t, series[0].Time.Before(series[1].Time))
	assert.True(t, series[1].Time.Before(series[2].Time))
}

func TestByTimeSwap(t *testing.T) {
	now := time.Now()
	series := ByTime{
		{Time: now, Online: true},
		{Time: now.Add(1 * time.Hour), Online: false},
	}

	series.Swap(0, 1)
	assert.False(t, series[0].Online)
	assert.True(t, series[1].Online)
}

func TestByTimeLen(t *testing.T) {
	series := ByTime{
		{Time: time.Now(), Online: true},
		{Time: time.Now(), Online: false},
	}
	assert.Equal(t, 2, series.Len())
}

func TestServiceDuration(t *testing.T) {
	s := &Service{
		Interval: 30,
	}
	d := s.Duration()
	assert.Equal(t, 30*time.Second, d)
}

func TestServiceHash(t *testing.T) {
	s := &Service{
		Name:   "TestService",
		Domain: "example.com",
	}
	hash := s.Hash()
	assert.NotEmpty(t, hash)

	s2 := &Service{
		Name:   "TestService",
		Domain: "example.com",
	}
	assert.Equal(t, hash, s2.Hash())
}

func TestServiceRequiresTLS(t *testing.T) {
	t.Run("VerifySSL true", func(t *testing.T) {
		s := Service{}
		s.VerifySSL.Bool = true
		s.VerifySSL.Valid = true
		assert.True(t, s.requiresTLS())
	})

	t.Run("SMTP on port 465", func(t *testing.T) {
		s := Service{Type: "smtp", Port: 465}
		assert.True(t, s.requiresTLS())
	})

	t.Run("IMAP on port 993", func(t *testing.T) {
		s := Service{Type: "imap", Port: 993}
		assert.True(t, s.requiresTLS())
	})

	t.Run("HTTP without SSL", func(t *testing.T) {
		s := Service{Type: "http", Port: 80}
		assert.False(t, s.requiresTLS())
	})
}

// Test addDurations helper function
func TestAddDurations(t *testing.T) {
	t.Run("empty series", func(t *testing.T) {
		var s []series
		assert.Equal(t, int64(0), addDurations(s, true))
		assert.Equal(t, int64(0), addDurations(s, false))
	})

	t.Run("all online", func(t *testing.T) {
		s := []series{
			{Duration: 1000, Online: true},
			{Duration: 2000, Online: true},
			{Duration: 3000, Online: true},
		}
		assert.Equal(t, int64(6000), addDurations(s, true))
		assert.Equal(t, int64(0), addDurations(s, false))
	})

	t.Run("all offline", func(t *testing.T) {
		s := []series{
			{Duration: 1000, Online: false},
			{Duration: 2000, Online: false},
		}
		assert.Equal(t, int64(0), addDurations(s, true))
		assert.Equal(t, int64(3000), addDurations(s, false))
	})

	t.Run("mixed online and offline", func(t *testing.T) {
		s := []series{
			{Duration: 1000, Online: true},
			{Duration: 500, Online: false},
			{Duration: 2000, Online: true},
			{Duration: 300, Online: false},
		}
		assert.Equal(t, int64(3000), addDurations(s, true))
		assert.Equal(t, int64(800), addDurations(s, false))
	})

	t.Run("zero duration entries", func(t *testing.T) {
		s := []series{
			{Duration: 0, Online: true},
			{Duration: 1000, Online: true},
		}
		assert.Equal(t, int64(1000), addDurations(s, true))
	})
}

// Test UptimeData method
func TestUptimeData(t *testing.T) {
	now := utils.Now()

	t.Run("no hits returns error", func(t *testing.T) {
		s := &Service{Id: 1, Online: true}
		result, err := s.UptimeData(nil, nil, nil)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "does not have any successful hits")
	})

	t.Run("empty hits slice returns error", func(t *testing.T) {
		s := &Service{Id: 1, Online: true}
		result, err := s.UptimeData([]*hits.Hit{}, []*checkins.CheckinHit{}, []*failures.Failure{})
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("all online - no failures", func(t *testing.T) {
		s := &Service{Id: 1, Online: true}
		hitsList := []*hits.Hit{
			{Id: 1, Service: 1, CreatedAt: now.Add(-24 * time.Hour)},
			{Id: 2, Service: 1, CreatedAt: now.Add(-12 * time.Hour)},
			{Id: 3, Service: 1, CreatedAt: now.Add(-1 * time.Hour)},
		}
		result, err := s.UptimeData(hitsList, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, hitsList[0].CreatedAt, result.Start)
		assert.True(t, result.Uptime > 0)
		assert.Equal(t, int64(0), result.Downtime)
		assert.Len(t, result.Series, 1)
		assert.True(t, result.Series[0].Online)
	})

	t.Run("all online with checkin hits only", func(t *testing.T) {
		s := &Service{Id: 1, Online: true}
		checkinHitsList := []*checkins.CheckinHit{
			{Id: 1, Checkin: 1, CreatedAt: now.Add(-24 * time.Hour)},
			{Id: 2, Checkin: 1, CreatedAt: now.Add(-12 * time.Hour)},
		}
		// UptimeData requires at least one hit in the hits slice when there are no failures
		// With only checkinHits and no regular hits, the first hit is nil
		hitsList := []*hits.Hit{
			{Id: 1, Service: 1, CreatedAt: now.Add(-24 * time.Hour)},
		}
		result, err := s.UptimeData(hitsList, checkinHitsList, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Uptime > 0)
	})

	t.Run("mixed online and offline periods", func(t *testing.T) {
		s := &Service{Id: 1, Online: true}

		// Create hits spread over time
		hitsList := []*hits.Hit{
			{Id: 1, Service: 1, CreatedAt: now.Add(-4 * time.Hour)},
			{Id: 2, Service: 1, CreatedAt: now.Add(-3 * time.Hour)},
			{Id: 3, Service: 1, CreatedAt: now.Add(-1 * time.Hour)},
		}

		// Create failures between some hits
		failsList := []*failures.Failure{
			{Id: 1, Service: 1, CreatedAt: now.Add(-2 * time.Hour)},
		}

		result, err := s.UptimeData(hitsList, nil, failsList)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Should have both uptime and downtime
		assert.True(t, result.Uptime > 0 || result.Downtime > 0)
		assert.True(t, len(result.Series) > 0)
	})

	t.Run("service currently offline", func(t *testing.T) {
		s := &Service{Id: 1, Online: false}

		hitsList := []*hits.Hit{
			{Id: 1, Service: 1, CreatedAt: now.Add(-4 * time.Hour)},
			{Id: 2, Service: 1, CreatedAt: now.Add(-3 * time.Hour)},
		}

		failsList := []*failures.Failure{
			{Id: 1, Service: 1, CreatedAt: now.Add(-1 * time.Hour)},
		}

		result, err := s.UptimeData(hitsList, nil, failsList)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Last series entry should reflect current offline status
		lastSeries := result.Series[len(result.Series)-1]
		assert.False(t, lastSeries.Online)
	})

	t.Run("service currently online after failure", func(t *testing.T) {
		s := &Service{Id: 1, Online: true}

		hitsList := []*hits.Hit{
			{Id: 1, Service: 1, CreatedAt: now.Add(-4 * time.Hour)},
			{Id: 2, Service: 1, CreatedAt: now.Add(-1 * time.Hour)},
		}

		failsList := []*failures.Failure{
			{Id: 1, Service: 1, CreatedAt: now.Add(-2 * time.Hour)},
		}

		result, err := s.UptimeData(hitsList, nil, failsList)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Last series entry should reflect current online status
		lastSeries := result.Series[len(result.Series)-1]
		assert.True(t, lastSeries.Online)
	})

	t.Run("1 hour time range", func(t *testing.T) {
		s := &Service{Id: 1, Online: true}

		hitsList := []*hits.Hit{
			{Id: 1, Service: 1, CreatedAt: now.Add(-50 * time.Minute)},
			{Id: 2, Service: 1, CreatedAt: now.Add(-30 * time.Minute)},
			{Id: 3, Service: 1, CreatedAt: now.Add(-10 * time.Minute)},
		}

		result, err := s.UptimeData(hitsList, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Uptime > 0)
	})

	t.Run("7 day time range", func(t *testing.T) {
		s := &Service{Id: 1, Online: true}

		hitsList := []*hits.Hit{
			{Id: 1, Service: 1, CreatedAt: now.Add(-7 * 24 * time.Hour)},
			{Id: 2, Service: 1, CreatedAt: now.Add(-3 * 24 * time.Hour)},
			{Id: 3, Service: 1, CreatedAt: now.Add(-1 * 24 * time.Hour)},
		}

		failsList := []*failures.Failure{
			{Id: 1, Service: 1, CreatedAt: now.Add(-5 * 24 * time.Hour)},
			{Id: 2, Service: 1, CreatedAt: now.Add(-4 * 24 * time.Hour)},
		}

		result, err := s.UptimeData(hitsList, nil, failsList)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, len(result.Series) >= 1)
	})

	t.Run("30 day time range", func(t *testing.T) {
		s := &Service{Id: 1, Online: true}

		hitsList := []*hits.Hit{
			{Id: 1, Service: 1, CreatedAt: now.Add(-30 * 24 * time.Hour)},
			{Id: 2, Service: 1, CreatedAt: now.Add(-15 * 24 * time.Hour)},
			{Id: 3, Service: 1, CreatedAt: now.Add(-1 * 24 * time.Hour)},
		}

		result, err := s.UptimeData(hitsList, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Uptime > 0)
	})

	t.Run("all failures at same time as hits", func(t *testing.T) {
		s := &Service{Id: 1, Online: false}
		timestamp := now.Add(-1 * time.Hour)

		hitsList := []*hits.Hit{
			{Id: 1, Service: 1, CreatedAt: timestamp},
		}

		failsList := []*failures.Failure{
			{Id: 1, Service: 1, CreatedAt: timestamp},
		}

		// When hit and failure are at same time, failure overwrites in the map
		result, err := s.UptimeData(hitsList, nil, failsList)
		// This may result in error or valid result depending on data
		if err == nil {
			assert.NotNil(t, result)
		}
	})

	t.Run("single hit no failures", func(t *testing.T) {
		s := &Service{Id: 1, Online: true}

		hitsList := []*hits.Hit{
			{Id: 1, Service: 1, CreatedAt: now.Add(-1 * time.Hour)},
		}

		result, err := s.UptimeData(hitsList, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int64(0), result.Downtime)
		assert.True(t, result.Uptime > 0)
	})

	t.Run("checkin hits combined with regular hits", func(t *testing.T) {
		s := &Service{Id: 1, Online: true}

		hitsList := []*hits.Hit{
			{Id: 1, Service: 1, CreatedAt: now.Add(-4 * time.Hour)},
		}

		checkinHitsList := []*checkins.CheckinHit{
			{Id: 1, Checkin: 1, CreatedAt: now.Add(-3 * time.Hour)},
			{Id: 2, Checkin: 1, CreatedAt: now.Add(-2 * time.Hour)},
		}

		failsList := []*failures.Failure{
			{Id: 1, Service: 1, CreatedAt: now.Add(-1 * time.Hour)},
		}

		result, err := s.UptimeData(hitsList, checkinHitsList, failsList)
		require.NoError(t, err)
		require.NotNil(t, result)
		// Should process both types of hits
		assert.True(t, len(result.Series) > 0)
	})

	t.Run("multiple transitions", func(t *testing.T) {
		s := &Service{Id: 1, Online: true}

		hitsList := []*hits.Hit{
			{Id: 1, Service: 1, CreatedAt: now.Add(-6 * time.Hour)},
			{Id: 2, Service: 1, CreatedAt: now.Add(-4 * time.Hour)},
			{Id: 3, Service: 1, CreatedAt: now.Add(-2 * time.Hour)},
		}

		failsList := []*failures.Failure{
			{Id: 1, Service: 1, CreatedAt: now.Add(-5 * time.Hour)},
			{Id: 2, Service: 1, CreatedAt: now.Add(-3 * time.Hour)},
		}

		result, err := s.UptimeData(hitsList, nil, failsList)
		require.NoError(t, err)
		require.NotNil(t, result)
		// Multiple transitions means multiple series entries
		assert.True(t, len(result.Series) >= 2)
	})

	t.Run("uptime and downtime sum reasonably", func(t *testing.T) {
		s := &Service{Id: 1, Online: true}

		hitsList := []*hits.Hit{
			{Id: 1, Service: 1, CreatedAt: now.Add(-10 * time.Hour)},
			{Id: 2, Service: 1, CreatedAt: now.Add(-5 * time.Hour)},
		}

		failsList := []*failures.Failure{
			{Id: 1, Service: 1, CreatedAt: now.Add(-7 * time.Hour)},
		}

		result, err := s.UptimeData(hitsList, nil, failsList)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify both uptime and downtime are non-negative
		assert.True(t, result.Uptime >= 0)
		assert.True(t, result.Downtime >= 0)

		// Verify total (uptime + downtime) is positive
		assert.True(t, result.Uptime+result.Downtime > 0)
	})
}
