package services

import (
	"testing"

	"github.com/statping-ng/statping-ng/types/checkins"
	"github.com/statping-ng/statping-ng/types/incidents"
	"github.com/statping-ng/statping-ng/types/messages"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindWithRelations(t *testing.T) {
	// Create a test service
	testService := &Service{
		Name:           "FindWithRelations Test",
		Domain:         "https://example.com",
		ExpectedStatus: 200,
		Interval:       30,
		Type:           "http",
		Method:         "GET",
		Timeout:        5,
		Order:          99,
		VerifySSL:      null.NewNullBool(false),
		Public:         null.NewNullBool(true),
	}

	err := testService.Create()
	require.Nil(t, err)
	require.NotZero(t, testService.Id)

	// Create related data
	testIncident := &incidents.Incident{
		Title:       "Test Incident",
		Description: "Test incident for FindWithRelations",
		ServiceId:   testService.Id,
	}
	err = testIncident.Create()
	require.Nil(t, err)

	testMessage := &messages.Message{
		Title:       "Test Message",
		Description: "Test message for FindWithRelations",
		ServiceId:   testService.Id,
		StartOn:     utils.Now(),
		EndOn:       utils.Now().Add(60 * 60),
	}
	err = testMessage.Create()
	require.Nil(t, err)

	testCheckin := &checkins.Checkin{
		ServiceId: testService.Id,
		Name:      "Test Checkin",
		Interval:  60,
		ApiKey:    "testkey123",
	}
	err = testCheckin.Create()
	require.Nil(t, err)

	t.Run("FindWithRelations returns service with relations", func(t *testing.T) {
		service, err := FindWithRelations(testService.Id)
		require.Nil(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, "FindWithRelations Test", service.Name)
	})

	t.Run("FindWithRelations with non-existent ID returns error", func(t *testing.T) {
		_, err := FindWithRelations(999999)
		assert.NotNil(t, err)
	})

	// Cleanup
	_ = testService.Delete()
}

func TestAllWithRelations(t *testing.T) {
	services := allWithRelations()
	assert.NotNil(t, services)
}

func TestClearCache(t *testing.T) {
	// Add a service to cache
	testService := &Service{
		Name:           "Cache Test Service",
		Domain:         "https://cache-test.com",
		ExpectedStatus: 200,
		Interval:       30,
		Type:           "http",
		Method:         "GET",
		Timeout:        5,
	}
	err := testService.Create()
	require.Nil(t, err)

	// Verify it's in the cache
	servicesLock.RLock()
	_, exists := allServices[testService.Id]
	servicesLock.RUnlock()
	assert.True(t, exists)

	// Clear cache
	ClearCache()

	// Verify cache is empty
	servicesLock.RLock()
	count := len(allServices)
	servicesLock.RUnlock()
	assert.Zero(t, count)

	// Cleanup
	_ = testService.Delete()

	// Repopulate cache for subsequent tests
	_ = all()
}

func TestAll(t *testing.T) {
	// Reload services to populate cache
	_ = all()

	services := All()
	assert.NotNil(t, services)
}
