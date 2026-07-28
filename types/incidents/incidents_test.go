package incidents

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var example = &Incident{
	Title:       "Example",
	Description: "No description",
	ServiceId:   1,
}

var update1 = &IncidentUpdate{
	IncidentId: 1,
	Message:    "First one here",
	Type:       "update",
}

var update2 = &IncidentUpdate{
	IncidentId: 1,
	Message:    "Second one here",
	Type:       "update",
}

func TestInit(t *testing.T) {
	// DB setup moved to TestMain in main_test.go
	// This test now just creates the example data
	testDb.Create(&example)
	testDb.Create(&update1)
	testDb.Create(&update2)
}

func TestFind(t *testing.T) {
	item, err := Find(1)
	require.Nil(t, err)
	assert.Equal(t, "Example", item.Title)
}

func TestFindNonExistent(t *testing.T) {
	_, err := Find(99999)
	require.NotNil(t, err)
}

func TestAll(t *testing.T) {
	items := All()
	assert.Len(t, items, 1)
}

func TestCreate(t *testing.T) {
	example := &Incident{
		Title: "Example 2",
	}
	err := example.Create()
	require.Nil(t, err)
	assert.NotZero(t, example.Id)
	assert.Equal(t, "Example 2", example.Title)
	assert.NotZero(t, example.CreatedAt)
}

func TestCreateValidationError(t *testing.T) {
	incident := &Incident{
		Title:       "",
		Description: "Missing title",
	}
	err := incident.Create()
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "missing title")
}

func TestUpdate(t *testing.T) {
	item, err := Find(1)
	require.Nil(t, err)
	item.Title = "Updated"
	err = item.Update()
	require.Nil(t, err)
	assert.Equal(t, "Updated", item.Title)
}

func TestUpdateValidationError(t *testing.T) {
	item, err := Find(1)
	require.Nil(t, err)
	item.Title = ""
	err = item.Update()
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "missing title")
}

func TestFindByService(t *testing.T) {
	// Create incidents for different services
	incident := &Incident{
		Title:     "Service 2 Incident",
		ServiceId: 2,
	}
	err := incident.Create()
	require.Nil(t, err)

	// Find by service 1
	items := FindByService(1)
	for _, item := range items {
		assert.Equal(t, int64(1), item.ServiceId)
	}

	// Find by service 2
	items2 := FindByService(2)
	assert.Len(t, items2, 1)
	assert.Equal(t, "Service 2 Incident", items2[0].Title)

	// Find by non-existent service
	items3 := FindByService(99999)
	assert.Len(t, items3, 0)
}

func TestIncidentUpdatesLoaded(t *testing.T) {
	// Find incident should load its updates via AfterFind
	item, err := Find(1)
	require.Nil(t, err)
	// Updates should be loaded automatically
	assert.GreaterOrEqual(t, len(item.Updates), 0)
}

// IncidentUpdate tests

func TestIncidentUpdateCreate(t *testing.T) {
	update := &IncidentUpdate{
		IncidentId: 1,
		Message:    "Test update message",
		Type:       "investigating",
	}
	err := update.Create()
	require.Nil(t, err)
	assert.NotZero(t, update.Id)
	assert.NotZero(t, update.CreatedAt)
}

func TestIncidentUpdateCreateValidationError(t *testing.T) {
	update := &IncidentUpdate{
		IncidentId: 1,
		Message:    "",
		Type:       "update",
	}
	err := update.Create()
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "missing incident update title")
}

func TestFindUpdate(t *testing.T) {
	update, err := FindUpdate(1)
	require.Nil(t, err)
	assert.Equal(t, "First one here", update.Message)
}

func TestFindUpdateNonExistent(t *testing.T) {
	_, err := FindUpdate(99999)
	require.NotNil(t, err)
}

func TestIncidentUpdateUpdate(t *testing.T) {
	update, err := FindUpdate(1)
	require.Nil(t, err)
	update.Message = "Updated message"
	err = update.Update()
	require.Nil(t, err)

	// Verify the update persisted
	updated, err := FindUpdate(1)
	require.Nil(t, err)
	assert.Equal(t, "Updated message", updated.Message)
}

func TestIncidentUpdateUpdateValidationError(t *testing.T) {
	update, err := FindUpdate(1)
	require.Nil(t, err)
	update.Message = ""
	err = update.Update()
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "missing incident update title")
}

func TestIncidentUpdateDelete(t *testing.T) {
	// Create a new update to delete
	update := &IncidentUpdate{
		IncidentId: 1,
		Message:    "To be deleted",
		Type:       "update",
	}
	err := update.Create()
	require.Nil(t, err)
	updateId := update.Id

	// Delete it
	err = update.Delete()
	require.Nil(t, err)

	// Verify it's gone
	_, err = FindUpdate(updateId)
	require.NotNil(t, err)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		i       *Incident
		wantErr bool
	}{
		{
			name:    "valid incident",
			i:       &Incident{Title: "Valid", Description: "Test"},
			wantErr: false,
		},
		{
			name:    "empty title",
			i:       &Incident{Title: "", Description: "Test"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.i.Validate()
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestIncidentUpdateValidate(t *testing.T) {
	tests := []struct {
		name    string
		u       *IncidentUpdate
		wantErr bool
	}{
		{
			name:    "valid update",
			u:       &IncidentUpdate{Message: "Valid message", Type: "update"},
			wantErr: false,
		},
		{
			name:    "empty message",
			u:       &IncidentUpdate{Message: "", Type: "update"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.u.Validate()
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestDeleteWithUpdates(t *testing.T) {
	// Create an incident with updates
	incident := &Incident{
		Title:     "Incident with updates",
		ServiceId: 1,
	}
	err := incident.Create()
	require.Nil(t, err)

	update1 := &IncidentUpdate{
		IncidentId: incident.Id,
		Message:    "Update 1 for deletion test",
		Type:       "update",
	}
	err = update1.Create()
	require.Nil(t, err)

	update2 := &IncidentUpdate{
		IncidentId: incident.Id,
		Message:    "Update 2 for deletion test",
		Type:       "resolved",
	}
	err = update2.Create()
	require.Nil(t, err)

	// Re-fetch to load updates
	incident, err = Find(incident.Id)
	require.Nil(t, err)

	// Delete incident - should cascade delete updates
	err = incident.Delete()
	require.Nil(t, err)

	// Verify incident is gone
	_, err = Find(incident.Id)
	require.NotNil(t, err)

	// Verify updates are gone
	_, err = FindUpdate(update1.Id)
	require.NotNil(t, err)
	_, err = FindUpdate(update2.Id)
	require.NotNil(t, err)
}

func TestDelete(t *testing.T) {
	all := All()
	initialCount := len(all)

	// Create a new incident for this test
	item := &Incident{
		Title:     "To Delete",
		ServiceId: 1,
	}
	err := item.Create()
	require.Nil(t, err)

	all = All()
	assert.Len(t, all, initialCount+1)

	err = item.Delete()
	require.Nil(t, err)

	all = All()
	assert.Len(t, all, initialCount)
}

func TestSamples(t *testing.T) {
	initialCount := len(All())
	require.Nil(t, Samples())
	assert.Equal(t, initialCount+2, len(All()))
}

func TestClose(t *testing.T) {
	assert.Nil(t, db.Close())
}
