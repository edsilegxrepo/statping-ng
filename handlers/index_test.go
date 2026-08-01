package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexRoutes(t *testing.T) {
	ensureHandlerSetup(t)

	tests := []HTTPTest{
		{
			Name:           "Base Handler - 404 page",
			URL:            "/nonexistent-page-that-does-not-exist",
			Method:         "GET",
			ExpectedStatus: 200,
			NoAuth:         true,
		},
		{
			Name:             "Health Check Handler",
			URL:              "/health",
			Method:           "GET",
			ExpectedStatus:   200,
			ExpectedContains: []string{"online", "services", "setup"},
			NoAuth:           true,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}
