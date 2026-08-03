package handlers

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/users"
	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ = fmt.Sprintf // silence unused import when test is skipped

func TestUnAuthenticatedUserRoutes(t *testing.T) {
	ensureHandlerSetup(t)
	tests := []HTTPTest{
		{
			Name:           "No Authentication - New User",
			URL:            "/api/users",
			Method:         "POST",
			ExpectedStatus: 401, // Auth required
			NoAuth:         true,
		},
		{
			Name:           "No Authentication - Update User",
			URL:            "/api/users/1",
			Method:         "POST",
			ExpectedStatus: 401, // Auth required
			NoAuth:         true,
		},
		{
			Name:           "No Authentication - View User",
			URL:            "/api/users/1",
			Method:         "GET",
			ExpectedStatus: 401, // GET doesn't need CSRF, so auth check runs
			NoAuth:         true,
		},
		{
			Name:           "No Authentication - Delete User",
			URL:            "/api/users/1",
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

func TestApiCheckUserToken(t *testing.T) {
	ensureHandlerSetup(t)

	tokenForm := url.Values{}
	tokenForm.Add("token", "invalid_token")

	emptyForm := url.Values{}

	tests := []HTTPTest{
		{
			Name:             "Check User Token - Missing Token",
			URL:              "/api/users/token",
			Method:           "POST",
			Body:             emptyForm.Encode(),
			HttpHeaders:      []string{"Content-Type=application/x-www-form-urlencoded"},
			ExpectedStatus:   200,
			ExpectedContains: []string{"missing token"},
		},
		{
			Name:             "Check User Token - Invalid Token",
			URL:              "/api/users/token",
			Method:           "POST",
			Body:             tokenForm.Encode(),
			HttpHeaders:      []string{"Content-Type=application/x-www-form-urlencoded"},
			ExpectedStatus:   500,
			ExpectedContains: []string{"error"},
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestApiUserAPIKeyAuth(t *testing.T) {
	ensureHandlerSetup(t)

	allUsers := users.All()
	require.NotEmpty(t, allUsers, "need users to test API key auth")

	var testUser *users.User
	for _, u := range allUsers {
		if u.ApiKey != "" {
			testUser = u
			break
		}
	}
	require.NotNil(t, testUser, "need a user with API key")

	tests := []HTTPTest{
		{
			Name:           "Auth with User API Key in query",
			URL:            "/api/users?api=" + testUser.ApiKey,
			Method:         "GET",
			ExpectedStatus: 200,
			NoAuth:         true,
		},
		{
			Name:           "Auth with User API Key in header",
			URL:            "/api/users",
			Method:         "GET",
			HttpHeaders:    []string{"Authorization=Bearer " + testUser.ApiKey},
			ExpectedStatus: 200,
			NoAuth:         true,
		},
		{
			Name:           "Auth with invalid API Key",
			URL:            "/api/users?api=invalid_key_12345",
			Method:         "GET",
			ExpectedStatus: 401,
			NoAuth:         true,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			require.Nil(t, err)
		})
	}
}

func TestApiUserDeleteLastUserError(t *testing.T) {
	t.Skip("Skipping: destructive test that deletes all users - run in isolation with -run TestApiUserDeleteLastUserError")
	ensureHandlerSetup(t)

	allUsers := users.All()
	require.True(t, len(allUsers) > 0, "need users to test delete-last-user error")

	tempUser := &users.User{
		Username: "temp_delete_test_user",
		Email:    "temp_delete@example.com",
		Password: utils.HashPassword("Password123456789012345678901234567890"),
	}
	require.NoError(t, tempUser.Create())

	for _, u := range users.All() {
		if u.Id != tempUser.Id {
			_ = u.Delete()
		}
	}

	tests := []HTTPTest{
		{
			Name:             "Delete Last User Error",
			URL:              "/api/users/" + fmt.Sprintf("%d", tempUser.Id),
			Method:           "DELETE",
			HttpHeaders:      []string{"Authorization=" + core.App.ApiSecret},
			ExpectedStatus:   200,
			ExpectedContains: []string{`"error":"cannot delete the last user"`},
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestApiUsersRoutes(t *testing.T) {
	ensureHandlerSetup(t)

	form := url.Values{}
	form.Add("username", "adminupdated")
	form.Add("password", "Password123456789012345678901234567890")

	badForm := url.Values{}
	badForm.Add("username", "adminupdated")
	badForm.Add("password", "wrongpassword")

	tests := []HTTPTest{
		{
			Name:           "Check Basic Authentication",
			URL:            "/api",
			Method:         "GET",
			ExpectedStatus: 401,
			BeforeTest: func(t *testing.T) error {
				utils.Params.Set("AUTH_USERNAME", "admin")
				utils.Params.Set("AUTH_PASSWORD", "admin")
				return nil
			},
			AfterTest: func(t *testing.T) error {
				utils.Params.Set("AUTH_USERNAME", "")
				utils.Params.Set("AUTH_PASSWORD", "")
				return nil
			},
		},
		{
			Name:           "Statping All Users",
			URL:            "/api/users",
			Method:         "GET",
			HttpHeaders:    []string{"Authorization=" + core.App.ApiSecret},
			ExpectedStatus: 200,
			GreaterThan:    3, // At least 3 users from initial setup; may be more from other tests
			BeforeTest:     SetTestENV,
		},
		{
			Name:        "Statping Create User",
			URL:         "/api/users",
			HttpHeaders: []string{"Content-Type=application/json", "Authorization=" + core.App.ApiSecret},
			Method:      "POST",
			Body: `{
					"username": "adminuser2",
					"email": "info@adminemail.com",
					"password": "Password123456789012345678901234567890",
					"admin": true
				}`,
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, MethodCreate},
		},
		{
			Name:           "Statping View User",
			URL:            "/api/users/1",
			Method:         "GET",
			HttpHeaders:    []string{"Authorization=" + core.App.ApiSecret},
			ExpectedStatus: 200,
		},
		{
			Name:           "Statping Incorrect User ID",
			URL:            "/api/users/NOinteger",
			Method:         "GET",
			HttpHeaders:    []string{"Authorization=" + core.App.ApiSecret},
			ExpectedStatus: 422,
		},
		{
			Name:           "Statping Missing User",
			URL:            "/api/users/9393939393",
			Method:         "GET",
			HttpHeaders:    []string{"Authorization=" + core.App.ApiSecret},
			ExpectedStatus: 404,
		},
		{
			Name:        "Statping Update User",
			URL:         "/api/users/1",
			Method:      "POST",
			HttpHeaders: []string{"Authorization=" + core.App.ApiSecret},
			Body: `{
					"username": "adminupdated",
					"email": "info@email.com",
					"password": "Password123456789012345678901234567890",
					"admin": true
				}`,
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, MethodUpdate},
		},
		{
			Name:             "Statping Delete User",
			URL:              "/api/users/2",
			Method:           "DELETE",
			HttpHeaders:      []string{"Authorization=" + core.App.ApiSecret},
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, MethodDelete},
		},
		{
			Name:             "Statping Login User",
			URL:              "/api/login",
			Method:           "POST",
			Body:             form.Encode(),
			ExpectedContains: []string{`"token"`},
			ExpectedStatus:   200,
			HttpHeaders:      []string{"Content-Type=application/x-www-form-urlencoded"},
		},
		{
			Name:             "Statping Bad Login User",
			URL:              "/api/login",
			Method:           "POST",
			Body:             badForm.Encode(),
			ExpectedContains: []string{`incorrect authentication`},
			ExpectedStatus:   200,
			HttpHeaders:      []string{"Content-Type=application/x-www-form-urlencoded"},
		},
		{
			Name:           "Statping Logout",
			URL:            "/api/logout",
			Method:         "GET",
			HttpHeaders:    []string{"Authorization=" + core.App.ApiSecret},
			ExpectedStatus: 200,
		},
		{
			Name:             "Incorrect JSON POST",
			URL:              "/api/users",
			Body:             BadJSON,
			HttpHeaders:      []string{"Authorization=" + core.App.ApiSecret},
			ExpectedContains: []string{BadJSONResponse},
			BeforeTest:       SetTestENV,
			Method:           "POST",
			ExpectedStatus:   422,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			require.Nil(t, err)
		})
	}
}
