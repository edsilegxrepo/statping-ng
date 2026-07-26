package handlers

import (
	"testing"

	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
)

func TestDashboardAPIRoutes(t *testing.T) {
	ensureHandlerSetup(t)

	tests := []HTTPTest{
		{
			Name:           "Get Theme API",
			URL:            "/api/theme",
			Method:         "GET",
			ExpectedStatus: 200,
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

func TestConfigsAPIRoutes(t *testing.T) {
	// Skip if DB_CONN not set - configs endpoint requires a valid config.yml
	// with DB_CONN set, which doesn't exist in SQLite in-memory test mode
	if utils.Params.GetString("DB_CONN") == "" {
		t.Skip("Skipping configs API test - requires DB_CONN to be set")
	}

	ensureHandlerSetup(t)

	tests := []HTTPTest{
		{
			Name:           "Get Configs API - Settings Configs",
			URL:            "/api/settings/configs",
			Method:         "GET",
			ExpectedStatus: 200,
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

func TestUnAuthenticatedDashboardRoutes(t *testing.T) {
	tests := []HTTPTest{
		{
			Name:           "No Auth - Get Theme",
			URL:            "/api/theme",
			Method:         "GET",
			ExpectedStatus: 401,
			NoAuth:         true,
		},
		{
			Name:           "No Auth - Save Theme",
			URL:            "/api/theme",
			Method:         "POST",
			ExpectedStatus: 403,
			NoAuth:         true,
		},
		{
			Name:           "No Auth - Get Configs",
			URL:            "/api/settings/configs",
			Method:         "GET",
			ExpectedStatus: 401,
			NoAuth:         true,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestThemeAPIRoutes(t *testing.T) {
	ensureHandlerSetup(t)

	tests := []HTTPTest{
		{
			Name:           "Save Theme API - Invalid JSON",
			URL:            "/api/theme",
			Method:         "POST",
			Body:           `{invalid json}`,
			ExpectedStatus: 500,
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
