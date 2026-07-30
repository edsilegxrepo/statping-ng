package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnAuthenticatedMessageRoutes(t *testing.T) {
	ensureHandlerSetup(t)

	tests := []HTTPTest{
		{
			Name:           "No Authentication - New Message",
			URL:            "/api/messages",
			Method:         "POST",
			ExpectedStatus: 401, // Auth required
			NoAuth:         true,
		},
		{
			Name:           "No Authentication - Update Message",
			URL:            "/api/messages/1",
			Method:         "POST",
			ExpectedStatus: 401, // Auth required
			NoAuth:         true,
		},
		{
			Name:           "No Authentication - Delete Message",
			URL:            "/api/messages/1",
			Method:         "DELETE",
			ExpectedStatus: 401, // Auth required
			NoAuth:         true,
		},
		{
			Name:           "No Authentication - List Messages",
			URL:            "/api/messages",
			Method:         "GET",
			ExpectedStatus: 200, // GET is public
			NoAuth:         true,
		},
		{
			Name:           "No Authentication - View Message",
			URL:            "/api/messages/1",
			Method:         "GET",
			ExpectedStatus: 200, // GET is public
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

func TestMessagesApiRoutes(t *testing.T) {
	ensureHandlerSetup(t)

	tests := []HTTPTest{
		{
			Name:             "Statping Messages",
			URL:              "/api/messages",
			Method:           "GET",
			ExpectedStatus:   200,
			ExpectedContains: []string{`"title":"Routine Downtime"`},
		},
		{
			Name:   "Statping Create Message",
			URL:    "/api/messages",
			Method: "POST",
			Body: `{
					"title": "API Message",
					"description": "This is an example a upcoming message for a service!",
					"start_on": "2022-11-17T03:28:16.323797-08:00",
					"end_on": "2022-11-17T05:13:16.323798-08:00",
					"service": 1,
					"notify_users": true,
					"notify_method": "email",
					"notify_before": 6,
					"notify_before_scale": "hour"
				}`,
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, `"type":"message"`, `"method":"create"`, `"title":"API Message"`},
			BeforeTest:       SetTestENV,
			AfterTest:        UnsetTestENV,
			SecureRoute:      true,
		},
		{
			Name:             "Statping View Message",
			URL:              "/api/messages/1",
			Method:           "GET",
			ExpectedStatus:   200,
			ExpectedContains: []string{`"title":"Routine Downtime"`},
		},
		{
			Name:   "Statping Update Message",
			URL:    "/api/messages/1",
			Method: "POST",
			Body: `{
					"title": "Updated Message",
					"description": "This message was updated",
					"start_on": "2022-11-17T03:28:16.323797-08:00",
					"end_on": "2022-11-17T05:13:16.323798-08:00",
					"service": 1,
					"notify_users": true,
					"notify_method": "email",
					"notify_before": 3,
					"notify_before_scale": "hour"
				}`,
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, `"type":"message"`, MethodUpdate},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
		{
			Name:             "Statping Delete Message",
			URL:              "/api/messages/1",
			Method:           "DELETE",
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, MethodDelete},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
		{
			Name:           "Statping Missing Message",
			URL:            "/api/messages/999999",
			Method:         "GET",
			ExpectedStatus: 404,
		},
		{
			Name:             "Incorrect JSON POST",
			URL:              "/api/messages",
			Body:             BadJSON,
			ExpectedContains: []string{BadJSONResponse},
			BeforeTest:       SetTestENV,
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

func TestMessageUpdateHandler(t *testing.T) {
	ensureHandlerSetup(t)

	// First create a message to test updates on
	createTest := HTTPTest{
		Name:   "Create Message for Update Tests",
		URL:    "/api/messages",
		Method: "POST",
		Body: `{
			"title": "Message for Update Test",
			"description": "Testing update handler",
			"start_on": "2022-11-17T03:28:16.323797-08:00",
			"end_on": "2022-11-17T05:13:16.323798-08:00",
			"service": 1
		}`,
		ExpectedStatus:   200,
		ExpectedContains: []string{Success},
		BeforeTest:       SetTestENV,
		SecureRoute:      true,
	}

	_, t, err := RunHTTPTest(createTest, t)
	require.Nil(t, err)

	tests := []HTTPTest{
		{
			Name:   "Update Message - Valid",
			URL:    "/api/messages/2",
			Method: "POST",
			Body: `{
				"title": "Updated Title",
				"description": "Updated description",
				"start_on": "2022-12-01T10:00:00.000000-08:00",
				"end_on": "2022-12-01T12:00:00.000000-08:00",
				"service": 1,
				"notify_users": false
			}`,
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, MethodUpdate, `"title":"Updated Title"`},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
		{
			Name:             "Update Message - Invalid JSON",
			URL:              "/api/messages/2",
			Method:           "POST",
			Body:             BadJSON,
			ExpectedStatus:   422,
			ExpectedContains: []string{BadJSONResponse},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
		{
			Name:   "Update Message - Non-existent ID",
			URL:    "/api/messages/999999",
			Method: "POST",
			Body: `{
				"title": "Will Not Update",
				"description": "Non-existent message",
				"start_on": "2022-12-01T10:00:00.000000-08:00",
				"end_on": "2022-12-01T12:00:00.000000-08:00",
				"service": 1
			}`,
			ExpectedStatus:   404,
			ExpectedContains: []string{`"error"`},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
		{
			Name:   "Update Message - Invalid ID (Non-numeric)",
			URL:    "/api/messages/notanumber",
			Method: "POST",
			Body: `{
				"title": "Will Not Update",
				"description": "Invalid ID",
				"start_on": "2022-12-01T10:00:00.000000-08:00",
				"end_on": "2022-12-01T12:00:00.000000-08:00",
				"service": 1
			}`,
			ExpectedStatus:   422,
			ExpectedContains: []string{`"error"`},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
		{
			Name:   "Update Message - Partial Update",
			URL:    "/api/messages/2",
			Method: "POST",
			Body: `{
				"title": "Partial Update Title"
			}`,
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, MethodUpdate},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestMessageDeleteHandler(t *testing.T) {
	ensureHandlerSetup(t)

	// Create a message to delete
	createTest := HTTPTest{
		Name:   "Create Message for Delete Test",
		URL:    "/api/messages",
		Method: "POST",
		Body: `{
			"title": "Message to Delete",
			"description": "This will be deleted",
			"start_on": "2022-11-17T03:28:16.323797-08:00",
			"end_on": "2022-11-17T05:13:16.323798-08:00",
			"service": 1
		}`,
		ExpectedStatus:   200,
		ExpectedContains: []string{Success},
		BeforeTest:       SetTestENV,
		SecureRoute:      true,
	}

	_, t, err := RunHTTPTest(createTest, t)
	require.Nil(t, err)

	tests := []HTTPTest{
		{
			Name:             "Delete Message - Valid",
			URL:              "/api/messages/3",
			Method:           "DELETE",
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, MethodDelete},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
		{
			Name:             "Delete Message - Non-existent ID",
			URL:              "/api/messages/999999",
			Method:           "DELETE",
			ExpectedStatus:   404,
			ExpectedContains: []string{`"error"`},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
		{
			Name:             "Delete Message - Invalid ID (Non-numeric)",
			URL:              "/api/messages/invalid",
			Method:           "DELETE",
			ExpectedStatus:   422,
			ExpectedContains: []string{`"error"`},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
		{
			Name:             "Delete Message - Already Deleted",
			URL:              "/api/messages/3",
			Method:           "DELETE",
			ExpectedStatus:   404,
			ExpectedContains: []string{`"error"`},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestMessageCreateHandler(t *testing.T) {
	ensureHandlerSetup(t)

	tests := []HTTPTest{
		{
			Name:   "Create Message - Valid Full",
			URL:    "/api/messages",
			Method: "POST",
			Body: `{
				"title": "Full Valid Message",
				"description": "Complete message with all fields",
				"start_on": "2022-11-17T03:28:16.323797-08:00",
				"end_on": "2022-11-17T05:13:16.323798-08:00",
				"service": 1,
				"notify_users": true,
				"notify_method": "email",
				"notify_before": 6,
				"notify_before_scale": "hour"
			}`,
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, MethodCreate, `"title":"Full Valid Message"`},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
		{
			Name:   "Create Message - Minimal Fields",
			URL:    "/api/messages",
			Method: "POST",
			Body: `{
				"title": "Minimal Message",
				"service": 1
			}`,
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, MethodCreate},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
		{
			Name:             "Create Message - Invalid JSON",
			URL:              "/api/messages",
			Method:           "POST",
			Body:             BadJSON,
			ExpectedStatus:   422,
			ExpectedContains: []string{BadJSONResponse},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
		{
			Name:             "Create Message - Empty Body (Missing Title)",
			URL:              "/api/messages",
			Method:           "POST",
			Body:             `{}`,
			ExpectedStatus:   200, // Returns error as 200 with error body
			ExpectedContains: []string{`"error":"missing message title"`},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
		{
			Name:   "Create Message - With Slack Notify Method",
			URL:    "/api/messages",
			Method: "POST",
			Body: `{
				"title": "Slack Notification Message",
				"description": "Notify via Slack",
				"start_on": "2022-11-17T03:28:16.323797-08:00",
				"end_on": "2022-11-17T05:13:16.323798-08:00",
				"service": 2,
				"notify_users": true,
				"notify_method": "slack",
				"notify_before": 1,
				"notify_before_scale": "day"
			}`,
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, MethodCreate, `"notify_method":"slack"`},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
		{
			Name:   "Create Message - Different Service",
			URL:    "/api/messages",
			Method: "POST",
			Body: `{
				"title": "Service 2 Message",
				"description": "Message for second service",
				"start_on": "2022-11-17T03:28:16.323797-08:00",
				"end_on": "2022-11-17T05:13:16.323798-08:00",
				"service": 2
			}`,
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, MethodCreate, `"service":2`},
			BeforeTest:       SetTestENV,
			SecureRoute:      true,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestMessageGetHandler(t *testing.T) {
	ensureHandlerSetup(t)

	// First create a message to ensure we have one to get
	createTest := HTTPTest{
		Name:   "Create Message for Get Test",
		URL:    "/api/messages",
		Method: "POST",
		Body: `{
			"title": "Message for Get Test",
			"description": "Testing get handler",
			"start_on": "2022-11-17T03:28:16.323797-08:00",
			"end_on": "2022-11-17T05:13:16.323798-08:00",
			"service": 1
		}`,
		ExpectedStatus:   200,
		ExpectedContains: []string{Success, MethodCreate},
		BeforeTest:       SetTestENV,
		SecureRoute:      true,
	}

	_, t, err := RunHTTPTest(createTest, t)
	require.Nil(t, err)

	// Get a list of messages to find a valid ID
	listTest := HTTPTest{
		Name:           "List Messages to Find Valid ID",
		URL:            "/api/messages",
		Method:         "GET",
		ExpectedStatus: 200,
		GreaterThan:    0,
	}
	_, t, err = RunHTTPTest(listTest, t)
	require.Nil(t, err)

	tests := []HTTPTest{
		{
			Name:             "Get Single Message - Valid (Recently Created)",
			URL:              "/api/messages/2", // Message created earlier in this test run
			Method:           "GET",
			ExpectedStatus:   200,
			ExpectedContains: []string{`"id":`, `"title"`},
		},
		{
			Name:             "Get Single Message - Non-existent",
			URL:              "/api/messages/999999",
			Method:           "GET",
			ExpectedStatus:   404,
			ExpectedContains: []string{`"error"`},
		},
		{
			Name:             "Get Single Message - Invalid ID",
			URL:              "/api/messages/notanumber",
			Method:           "GET",
			ExpectedStatus:   422,
			ExpectedContains: []string{`"error"`},
		},
		{
			Name:             "Get Single Message - Negative ID",
			URL:              "/api/messages/-1",
			Method:           "GET",
			ExpectedStatus:   404,
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

func TestMessageListHandler(t *testing.T) {
	ensureHandlerSetup(t)

	tests := []HTTPTest{
		{
			Name:           "List All Messages",
			URL:            "/api/messages",
			Method:         "GET",
			ExpectedStatus: 200,
			GreaterThan:    0, // Should have at least one message from sample data
		},
		{
			Name:             "List Messages - Returns Array",
			URL:              "/api/messages",
			Method:           "GET",
			ExpectedStatus:   200,
			ExpectedContains: []string{`[`}, // Should be an array
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestFindMessageFunction(t *testing.T) {
	ensureHandlerSetup(t)

	// Create a message first to have a valid ID to find
	createTest := HTTPTest{
		Name:   "Create Message for Find Test",
		URL:    "/api/messages",
		Method: "POST",
		Body: `{
			"title": "Message for Find Test",
			"description": "Testing find function",
			"start_on": "2022-11-17T03:28:16.323797-08:00",
			"end_on": "2022-11-17T05:13:16.323798-08:00",
			"service": 1
		}`,
		ExpectedStatus: 200,
		BeforeTest:     SetTestENV,
		SecureRoute:    true,
	}
	_, t, err := RunHTTPTest(createTest, t)
	require.Nil(t, err)

	// Test the findMessage function indirectly through the handlers
	tests := []HTTPTest{
		{
			Name:             "Find Message - Valid ID (Recently Created)",
			URL:              "/api/messages/2", // Use known existing message ID
			Method:           "GET",
			ExpectedStatus:   200,
			ExpectedContains: []string{`"id":`, `"title"`},
		},
		{
			Name:             "Find Message - Zero ID",
			URL:              "/api/messages/0",
			Method:           "GET",
			ExpectedStatus:   404,
			ExpectedContains: []string{`"error"`},
		},
		{
			Name:             "Find Message - Large ID",
			URL:              "/api/messages/9223372036854775807",
			Method:           "GET",
			ExpectedStatus:   404,
			ExpectedContains: []string{`"error"`},
		},
		{
			Name:             "Find Message - Special Characters",
			URL:              "/api/messages/1a2b",
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
