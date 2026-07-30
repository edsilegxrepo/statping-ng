package handlers

import (
	"testing"

	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/groups"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnAuthenticatedGroupRoutes(t *testing.T) {
	ensureHandlerSetup(t)
	tests := []HTTPTest{
		{
			Name:           "No Authentication - New Group",
			URL:            "/api/groups",
			Method:         "POST",
			ExpectedStatus: 401, // Auth required
			NoAuth:         true,
		},
		{
			Name:           "No Authentication - Update Group",
			URL:            "/api/groups/1",
			Method:         "POST",
			ExpectedStatus: 401, // Auth required
			NoAuth:         true,
		},
		{
			Name:           "No Authentication - Delete Group",
			URL:            "/api/groups/1",
			Method:         "DELETE",
			ExpectedStatus: 401, // Auth required
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

func TestGroupAPIRoutes(t *testing.T) {
	ensureHandlerSetup(t)
	tests := []HTTPTest{
		{
			Name:           "Statping Public Groups",
			URL:            "/api/groups",
			Method:         "GET",
			ExpectedStatus: 200,
			ResponseLen:    3,
			BeforeTest:     SetTestENV,
			AfterTest:      UnsetTestENV,
		},
		{
			Name:           "Statping View Public Group",
			URL:            "/api/groups/1",
			Method:         "GET",
			ExpectedStatus: 200,
			BeforeTest:     SetTestENV,
			AfterTest:      UnsetTestENV,
		},
		{
			Name:        "Statping Create Public Group",
			URL:         "/api/groups",
			HttpHeaders: []string{"Content-Type=application/json"},
			Body: `{
					"name": "New Group",
					"public": true
				}`,
			Method:         "POST",
			ExpectedStatus: 200,
			BeforeTest:     SetTestENV,
		},
		{
			Name:        "Statping Create Private Group",
			URL:         "/api/groups",
			HttpHeaders: []string{"Content-Type=application/json"},
			Body: `{
					"name": "New Private Group",
					"public": false
				}`,
			Method:         "POST",
			ExpectedStatus: 200,
		},
		{
			Name:             "Incorrect JSON POST",
			URL:              "/api/groups",
			Body:             BadJSON,
			ExpectedContains: []string{BadJSONResponse},
			Method:           "POST",
			ExpectedStatus:   422,
		},
		{
			Name:           "Statping Public and Private Groups",
			URL:            "/api/groups",
			Method:         "GET",
			ExpectedStatus: 200,
			GreaterThan:    3, // At least 3 groups (sample data), possibly more after creates
		},
		{
			Name:           "Statping View Private Group",
			URL:            "/api/groups/2",
			Method:         "GET",
			ExpectedStatus: 401, // GET doesn't need CSRF, so auth check runs
			NoAuth:         true,
		},
		{
			Name:           "Statping View Private Group Allowed",
			URL:            "/api/groups/2",
			Method:         "GET",
			ExpectedStatus: 200,
			BeforeTest:     SetTestENV,
		},
		{
			Name:           "Statping View Private Group with API Key",
			URL:            "/api/groups/2?api=" + core.App.ApiSecret,
			Method:         "GET",
			ExpectedStatus: 200,
			BeforeTest:     UnsetTestENV,
		},
		{
			Name:           "Statping View Private Group with API Header",
			URL:            "/api/groups/2",
			Method:         "GET",
			HttpHeaders:    []string{"Authorization=" + core.App.ApiSecret},
			ExpectedStatus: 200,
			BeforeTest:     UnsetTestENV,
		},
		{
			Name:           "Statping Reorder Groups",
			URL:            "/api/reorder/groups",
			Method:         "POST",
			Body:           `[{"group":1,"order":2},{"group":2,"order":1}]`,
			ExpectedStatus: 200,
			HttpHeaders:    []string{"Content-Type=application/json"},
			BeforeTest:     SetTestENV,
			SecureRoute:    true,
		},
		{
			Name:           "Statping View Unknown Group",
			URL:            "/api/groups/38383",
			Method:         "GET",
			BeforeTest:     SetTestENV,
			ExpectedStatus: 404,
		},
		{
			Name:   "Statping Update Group",
			URL:    "/api/groups/1",
			Method: "POST",
			Body: `{
					"name": "Updated Group",
					"public": false
					}`,
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, MethodUpdate},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
			AfterTest: func(t *testing.T) error {
				g, err := groups.Find(1)
				require.Nil(t, err)
				assert.Equal(t, "Updated Group", g.Name)
				return nil
			},
		},
		{
			Name:             "Statping Delete Group",
			URL:              "/api/groups/1",
			Method:           "DELETE",
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, MethodDelete},
			AfterTest:        UnsetTestENV,
			SecureRoute:      true,
		},
		// Group Update edge cases
		{
			Name:             "Update Group - Invalid JSON",
			URL:              "/api/groups/2",
			Method:           "POST",
			Body:             BadJSON,
			ExpectedStatus:   422,
			ExpectedContains: []string{BadJSONResponse},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
		{
			Name:           "Update Group - Non-existent Group",
			URL:            "/api/groups/99999",
			Method:         "POST",
			HttpHeaders:    []string{"Content-Type=application/json"},
			Body:           `{"name": "Test", "public": true}`,
			ExpectedStatus: 404,
			BeforeTest:     SetTestENV,
			SecureRoute:    true,
		},
		{
			Name:             "Update Group - Empty Name Validation",
			URL:              "/api/groups/2",
			Method:           "POST",
			HttpHeaders:      []string{"Content-Type=application/json"},
			Body:             `{"name": "", "public": true}`,
			ExpectedStatus:   200, // GORM validation error returned via sendErrorJson with default code
			ExpectedContains: []string{"group name is empty"},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
		{
			Name:           "Update Group - Invalid ID Format",
			URL:            "/api/groups/notanumber",
			Method:         "POST",
			HttpHeaders:    []string{"Content-Type=application/json"},
			Body:           `{"name": "Test", "public": true}`,
			ExpectedStatus: 422,
			BeforeTest:     SetTestENV,
			SecureRoute:    true,
		},
		// Group Delete edge cases
		{
			Name:           "Delete Group - Non-existent Group",
			URL:            "/api/groups/99999",
			Method:         "DELETE",
			ExpectedStatus: 404,
			BeforeTest:     SetTestENV,
			SecureRoute:    true,
		},
		{
			Name:           "Delete Group - Invalid ID Format",
			URL:            "/api/groups/notanumber",
			Method:         "DELETE",
			ExpectedStatus: 422,
			BeforeTest:     SetTestENV,
			SecureRoute:    true,
		},
		{
			Name:           "Delete Group - ID Zero",
			URL:            "/api/groups/0",
			Method:         "DELETE",
			ExpectedStatus: 422,
			BeforeTest:     SetTestENV,
			SecureRoute:    true,
		},
		// Group Reorder edge cases
		{
			Name:             "Reorder Groups - Invalid JSON",
			URL:              "/api/reorder/groups",
			Method:           "POST",
			Body:             BadJSON,
			ExpectedStatus:   422,
			ExpectedContains: []string{BadJSONResponse},
			HttpHeaders:      []string{"Content-Type=application/json"},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
		{
			Name:           "Reorder Groups - Empty Array",
			URL:            "/api/reorder/groups",
			Method:         "POST",
			Body:           `[]`,
			ExpectedStatus: 200,
			HttpHeaders:    []string{"Content-Type=application/json"},
			BeforeTest:     SetTestENV,
			SecureRoute:    true,
		},
		{
			Name:           "Reorder Groups - Non-existent Group in Array",
			URL:            "/api/reorder/groups",
			Method:         "POST",
			Body:           `[{"group":99999,"order":1}]`,
			ExpectedStatus: 404,
			HttpHeaders:    []string{"Content-Type=application/json"},
			BeforeTest:     SetTestENV,
			SecureRoute:    true,
		},
		// Group Create validation errors
		{
			Name:             "Create Group - Empty Name",
			URL:              "/api/groups",
			Method:           "POST",
			HttpHeaders:      []string{"Content-Type=application/json"},
			Body:             `{"name": "", "public": true}`,
			ExpectedStatus:   200, // GORM validation error returned via sendErrorJson with default code
			ExpectedContains: []string{"group name is empty"},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
		{
			Name:             "Create Group - Missing Name Field",
			URL:              "/api/groups",
			Method:           "POST",
			HttpHeaders:      []string{"Content-Type=application/json"},
			Body:             `{"public": true}`,
			ExpectedStatus:   200, // GORM validation error returned via sendErrorJson with default code
			ExpectedContains: []string{"group name is empty"},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
		// findGroup edge cases via GET endpoint
		{
			Name:           "Find Group - Invalid ID Format via GET",
			URL:            "/api/groups/abc",
			Method:         "GET",
			ExpectedStatus: 422,
			BeforeTest:     SetTestENV,
		},
		{
			Name:           "Find Group - ID Zero via GET",
			URL:            "/api/groups/0",
			Method:         "GET",
			ExpectedStatus: 422,
			BeforeTest:     SetTestENV,
		},
		{
			Name:           "Find Group - Negative ID via GET",
			URL:            "/api/groups/-1",
			Method:         "GET",
			ExpectedStatus: 404, // Negative is a valid integer, passes to DB lookup which returns not found
			BeforeTest:     SetTestENV,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}
