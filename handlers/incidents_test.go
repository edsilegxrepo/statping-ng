package handlers

import (
	"fmt"
	"testing"

	"github.com/statping-ng/statping-ng/types/incidents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnAuthenticatedIncidentRoutes(t *testing.T) {
	ensureHandlerSetup(t)
	tests := []HTTPTest{
		{
			Name:           "No Authentication - New Incident",
			URL:            "/api/services/1/incidents",
			Method:         "POST",
			ExpectedStatus: 403, // CSRF rejects before auth check
			NoAuth:         true,
		},
		{
			Name:           "No Authentication - New Incident Update",
			URL:            "/api/incidents/updates",
			Method:         "POST",
			ExpectedStatus: 403, // CSRF rejects before auth check
			NoAuth:         true,
		},
		{
			Name:           "No Authentication - Update Incident",
			URL:            "/api/incidents/1",
			Method:         "POST",
			ExpectedStatus: 403, // CSRF rejects before auth check
			NoAuth:         true,
		},
		{
			Name:           "No Authentication - Delete Incident",
			URL:            "/api/incidents/1",
			Method:         "DELETE",
			ExpectedStatus: 403, // CSRF rejects before auth check
			NoAuth:         true,
		},
		{
			Name:           "No Authentication - Delete Incident Update",
			URL:            "/api/incidents/1/updates/1",
			Method:         "DELETE",
			ExpectedStatus: 403, // CSRF rejects before auth check
			NoAuth:         true,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			str, t, err := RunHTTPTest(v, t)
			t.Logf("Test %s: \n %v\n", v.Name, str)
			assert.Nil(t, err)
		})
	}
}

func TestIncidentsAPIRoutes(t *testing.T) {
	ensureHandlerSetup(t)
	tests := []HTTPTest{
		{
			Name:           "Statping Create Incident",
			URL:            "/api/services/1/incidents",
			Method:         "POST",
			ExpectedStatus: 200,
			BeforeTest:     SetTestENV,
			AfterTest:      UnsetTestENV,
			Body: `{
					"title": "New Incident",
					"description": "This is a test for incidents"
				    }`,
			ExpectedContains: []string{Success},
		},
		{
			Name:           "Statping Service 1 Incidents",
			URL:            "/api/services/1/incidents",
			Method:         "GET",
			ExpectedStatus: 200,
			GreaterThan:    0, // At least 1 incident after creation
		},
		{
			Name: "Statping Update Incident",
			URL:  "/api/incidents/1",
			Body: `{
					"title": "Updated Incident",
					"description": "This is an updated incidents"
				    }`,
			Method:           "POST",
			ExpectedStatus:   200,
			BeforeTest:       SetTestENV,
			ExpectedContains: []string{Success},
		},
		{
			Name:             "Statping View Incident Updates",
			URL:              "/api/incidents/1/updates",
			Method:           "GET",
			ExpectedStatus:   200,
			ResponseLen:      3,
			BeforeTest:       SetTestENV,
			ExpectedContains: []string{`"type":"investigating"`},
		},
		{
			Name:   "Statping Create Incident Update",
			URL:    "/api/incidents/1/updates",
			Method: "POST",
			Body: `{
								"message": "Test message here",
								"type": "Update"
								}`,
			ExpectedStatus:   200,
			BeforeTest:       SetTestENV,
			ExpectedContains: []string{Success},
		},
		{
			Name:             "Incorrect Checkin JSON POST",
			URL:              "/api/incidents/1/updates",
			Body:             BadJSON,
			ExpectedContains: []string{BadJSONResponse},
			BeforeTest:       SetTestENV,
			Method:           "POST",
			ExpectedStatus:   422,
		},
		{
			Name:             "Statping Delete Incident Update",
			URL:              "/api/incidents/1/updates/1",
			Method:           "DELETE",
			ExpectedStatus:   200,
			BeforeTest:       SetTestENV,
			ExpectedContains: []string{Success},
		},
		{
			Name:             "Statping Delete Incident",
			URL:              "/api/incidents/1",
			Method:           "DELETE",
			ExpectedStatus:   200,
			BeforeTest:       SetTestENV,
			ExpectedContains: []string{Success},
		},
		{
			Name:             "Incorrect JSON POST",
			URL:              "/api/services/1/incidents",
			Body:             BadJSON,
			ExpectedContains: []string{BadJSONResponse},
			Method:           "POST",
			ExpectedStatus:   422,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestDeleteIncidentHandler(t *testing.T) {
	ensureHandlerSetup(t)

	// Create a test incident first
	createTest := HTTPTest{
		Name:           "Setup - Create Incident for Delete Test",
		URL:            "/api/services/2/incidents",
		Method:         "POST",
		ExpectedStatus: 200,
		Body: `{
			"title": "Incident to Delete",
			"description": "This incident will be deleted"
		}`,
		BeforeTest: SetTestENV,
	}
	_, t, err := RunHTTPTest(createTest, t)
	require.Nil(t, err)

	tests := []HTTPTest{
		{
			Name:             "Delete Incident - Valid ID",
			URL:              "/api/incidents/2",
			Method:           "DELETE",
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, MethodDelete},
			BeforeTest:       SetTestENV,
		},
		{
			Name:             "Delete Incident - Non-Existent ID",
			URL:              "/api/incidents/99999",
			Method:           "DELETE",
			ExpectedStatus:   404,
			ExpectedContains: []string{`"error"`},
			BeforeTest:       SetTestENV,
		},
		{
			Name:             "Delete Incident - Invalid ID (non-numeric)",
			URL:              "/api/incidents/invalid",
			Method:           "DELETE",
			ExpectedStatus:   422,
			ExpectedContains: []string{`"error"`},
			BeforeTest:       SetTestENV,
		},
		{
			Name:             "Delete Incident - ID Zero",
			URL:              "/api/incidents/0",
			Method:           "DELETE",
			ExpectedStatus:   422,
			ExpectedContains: []string{`"error"`},
			BeforeTest:       SetTestENV,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestUpdateIncidentHandler(t *testing.T) {
	ensureHandlerSetup(t)

	// Create a test incident first
	createTest := HTTPTest{
		Name:           "Setup - Create Incident for Update Test",
		URL:            "/api/services/2/incidents",
		Method:         "POST",
		ExpectedStatus: 200,
		Body: `{
			"title": "Incident to Update",
			"description": "This incident will be updated"
		}`,
		BeforeTest: SetTestENV,
	}
	_, t, err := RunHTTPTest(createTest, t)
	require.Nil(t, err)

	tests := []HTTPTest{
		{
			Name:   "Update Incident - Valid Data",
			URL:    "/api/incidents/3",
			Method: "POST",
			Body: `{
				"title": "Updated Title",
				"description": "Updated description"
			}`,
			ExpectedStatus:   200,
			ExpectedContains: []string{Success},
			BeforeTest:       SetTestENV,
		},
		{
			Name:             "Update Incident - Invalid JSON",
			URL:              "/api/incidents/3",
			Method:           "POST",
			Body:             BadJSON,
			ExpectedStatus:   422,
			ExpectedContains: []string{BadJSONResponse},
			BeforeTest:       SetTestENV,
		},
		{
			Name:   "Update Incident - Non-Existent ID",
			URL:    "/api/incidents/99999",
			Method: "POST",
			Body: `{
				"title": "Test",
				"description": "Test"
			}`,
			ExpectedStatus:   404,
			ExpectedContains: []string{`"error"`},
			BeforeTest:       SetTestENV,
		},
		{
			Name:             "Update Incident - Invalid ID (non-numeric)",
			URL:              "/api/incidents/invalid",
			Method:           "POST",
			Body:             `{"title":"Test"}`,
			ExpectedStatus:   422,
			ExpectedContains: []string{`"error"`},
			BeforeTest:       SetTestENV,
		},
		{
			Name:             "Update Incident - ID Zero",
			URL:              "/api/incidents/0",
			Method:           "POST",
			Body:             `{"title":"Test"}`,
			ExpectedStatus:   422,
			ExpectedContains: []string{`"error"`},
			BeforeTest:       SetTestENV,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestDeleteIncidentUpdateHandler(t *testing.T) {
	ensureHandlerSetup(t)

	// Create a test incident and update first
	createIncidentTest := HTTPTest{
		Name:           "Setup - Create Incident for Update Delete Test",
		URL:            "/api/services/2/incidents",
		Method:         "POST",
		ExpectedStatus: 200,
		Body: `{
			"title": "Incident for Update Delete",
			"description": "This incident has updates to delete"
		}`,
		BeforeTest: SetTestENV,
	}
	_, t, err := RunHTTPTest(createIncidentTest, t)
	require.Nil(t, err)

	// Get the latest incident ID
	allIncidents := incidents.All()
	require.Greater(t, len(allIncidents), 0)
	latestIncident := allIncidents[len(allIncidents)-1]

	// Create an update for this incident
	createUpdateTest := HTTPTest{
		Name:           "Setup - Create Update for Delete Test",
		URL:            fmt.Sprintf("/api/incidents/%d/updates", latestIncident.Id),
		Method:         "POST",
		ExpectedStatus: 200,
		Body: `{
			"message": "Update to delete",
			"type": "investigating"
		}`,
		BeforeTest: SetTestENV,
	}
	_, t, _ = RunHTTPTest(createUpdateTest, t)

	tests := []HTTPTest{
		{
			Name:             "Delete Incident Update - Non-Existent Update ID",
			URL:              "/api/incidents/1/updates/99999",
			Method:           "DELETE",
			ExpectedStatus:   500, // Handler returns raw gorm error for non-existent update
			ExpectedContains: []string{`"error"`},
			BeforeTest:       SetTestENV,
		},
		{
			Name:             "Delete Incident Update - Invalid Update ID (non-numeric)",
			URL:              "/api/incidents/1/updates/invalid",
			Method:           "DELETE",
			ExpectedStatus:   500, // Handler converts non-numeric to 0, then returns gorm error
			ExpectedContains: []string{`"error"`},
			BeforeTest:       SetTestENV,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestServiceIncidentsHandler(t *testing.T) {
	ensureHandlerSetup(t)

	tests := []HTTPTest{
		{
			Name:           "List Incidents - Valid Service",
			URL:            "/api/services/2/incidents",
			Method:         "GET",
			ExpectedStatus: 200,
		},
		{
			Name:             "List Incidents - Non-Existent Service",
			URL:              "/api/services/99999/incidents",
			Method:           "GET",
			ExpectedStatus:   404,
			ExpectedContains: []string{`"error"`},
		},
		{
			Name:             "List Incidents - Invalid Service ID (non-numeric)",
			URL:              "/api/services/invalid/incidents",
			Method:           "GET",
			ExpectedStatus:   422,
			ExpectedContains: []string{`"error"`},
		},
		{
			Name:             "List Incidents - Service ID Zero",
			URL:              "/api/services/0/incidents",
			Method:           "GET",
			ExpectedStatus:   422,
			ExpectedContains: []string{`"error"`},
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestIncidentUpdatesHandler(t *testing.T) {
	ensureHandlerSetup(t)

	// Create a test incident first
	createTest := HTTPTest{
		Name:           "Setup - Create Incident for Updates Test",
		URL:            "/api/services/2/incidents",
		Method:         "POST",
		ExpectedStatus: 200,
		Body: `{
			"title": "Incident for Updates Listing",
			"description": "This incident will have updates listed"
		}`,
		BeforeTest: SetTestENV,
	}
	_, t, err := RunHTTPTest(createTest, t)
	require.Nil(t, err)

	tests := []HTTPTest{
		{
			Name:           "List Updates - Valid Incident (empty updates)",
			URL:            "/api/incidents/5/updates",
			Method:         "GET",
			ExpectedStatus: 200,
		},
		{
			Name:             "List Updates - Non-Existent Incident",
			URL:              "/api/incidents/99999/updates",
			Method:           "GET",
			ExpectedStatus:   404,
			ExpectedContains: []string{`"error"`},
		},
		{
			Name:             "List Updates - Invalid Incident ID (non-numeric)",
			URL:              "/api/incidents/invalid/updates",
			Method:           "GET",
			ExpectedStatus:   422,
			ExpectedContains: []string{`"error"`},
		},
		{
			Name:             "List Updates - Incident ID Zero",
			URL:              "/api/incidents/0/updates",
			Method:           "GET",
			ExpectedStatus:   422,
			ExpectedContains: []string{`"error"`},
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestCreateIncidentValidationHandler(t *testing.T) {
	ensureHandlerSetup(t)

	tests := []HTTPTest{
		{
			Name:   "Create Incident - Missing Title (validation error)",
			URL:    "/api/services/2/incidents",
			Method: "POST",
			Body: `{
				"title": "",
				"description": "This incident has no title"
			}`,
			ExpectedStatus:   200, // Validation error from gorm hook returns without proper status code
			ExpectedContains: []string{`"error":"missing title"`},
			BeforeTest:       SetTestENV,
		},
		{
			Name:             "Create Incident - Empty Body",
			URL:              "/api/services/2/incidents",
			Method:           "POST",
			Body:             `{}`,
			ExpectedStatus:   200, // Validation error from gorm hook returns without proper status code
			ExpectedContains: []string{`"error":"missing title"`},
			BeforeTest:       SetTestENV,
		},
		{
			Name:   "Create Incident - Non-Existent Service",
			URL:    "/api/services/99999/incidents",
			Method: "POST",
			Body: `{
				"title": "Test Incident",
				"description": "For non-existent service"
			}`,
			ExpectedStatus:   404,
			ExpectedContains: []string{`"error"`},
			BeforeTest:       SetTestENV,
		},
		{
			Name:             "Create Incident - Invalid Service ID (non-numeric)",
			URL:              "/api/services/invalid/incidents",
			Method:           "POST",
			Body:             `{"title":"Test","description":"Test"}`,
			ExpectedStatus:   422,
			ExpectedContains: []string{`"error"`},
			BeforeTest:       SetTestENV,
		},
		{
			Name:             "Create Incident - Service ID Zero",
			URL:              "/api/services/0/incidents",
			Method:           "POST",
			Body:             `{"title":"Test","description":"Test"}`,
			ExpectedStatus:   422,
			ExpectedContains: []string{`"error"`},
			BeforeTest:       SetTestENV,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestCreateIncidentUpdateValidationHandler(t *testing.T) {
	ensureHandlerSetup(t)

	// Create a test incident first
	createTest := HTTPTest{
		Name:           "Setup - Create Incident for Update Validation Test",
		URL:            "/api/services/2/incidents",
		Method:         "POST",
		ExpectedStatus: 200,
		Body: `{
			"title": "Incident for Update Validation",
			"description": "This incident tests update validation"
		}`,
		BeforeTest: SetTestENV,
	}
	_, t, err := RunHTTPTest(createTest, t)
	require.Nil(t, err)

	tests := []HTTPTest{
		{
			Name:   "Create Update - Missing Message (validation error)",
			URL:    "/api/incidents/6/updates",
			Method: "POST",
			Body: `{
				"message": "",
				"type": "investigating"
			}`,
			ExpectedStatus:   200, // Validation error from gorm hook returns without proper status code
			ExpectedContains: []string{`"error":"missing incident update title"`},
			BeforeTest:       SetTestENV,
		},
		{
			Name:             "Create Update - Empty Body",
			URL:              "/api/incidents/6/updates",
			Method:           "POST",
			Body:             `{}`,
			ExpectedStatus:   200, // Validation error from gorm hook returns without proper status code
			ExpectedContains: []string{`"error":"missing incident update title"`},
			BeforeTest:       SetTestENV,
		},
		{
			Name:   "Create Update - Non-Existent Incident",
			URL:    "/api/incidents/99999/updates",
			Method: "POST",
			Body: `{
				"message": "Test update",
				"type": "investigating"
			}`,
			ExpectedStatus:   404,
			ExpectedContains: []string{`"error"`},
			BeforeTest:       SetTestENV,
		},
		{
			Name:             "Create Update - Invalid Incident ID (non-numeric)",
			URL:              "/api/incidents/invalid/updates",
			Method:           "POST",
			Body:             `{"message":"Test","type":"update"}`,
			ExpectedStatus:   422,
			ExpectedContains: []string{`"error"`},
			BeforeTest:       SetTestENV,
		},
		{
			Name:             "Create Update - Incident ID Zero",
			URL:              "/api/incidents/0/updates",
			Method:           "POST",
			Body:             `{"message":"Test","type":"update"}`,
			ExpectedStatus:   422,
			ExpectedContains: []string{`"error"`},
			BeforeTest:       SetTestENV,
		},
		{
			Name:   "Create Update - Valid Update",
			URL:    "/api/incidents/6/updates",
			Method: "POST",
			Body: `{
				"message": "Valid update message",
				"type": "resolved"
			}`,
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, MethodCreate},
			BeforeTest:       SetTestENV,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}
