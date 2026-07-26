package services

import (
	"sort"
	"testing"
	"time"

	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
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
