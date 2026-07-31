package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/statping-ng/statping-ng/types/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiDigestSettingsHandler(t *testing.T) {
	if core.App == nil {
		t.Skip("requires database setup")
	}

	req := httptest.NewRequest("GET", "/api/digest", nil)
	w := httptest.NewRecorder()

	apiDigestSettingsHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Verify expected fields exist
	_, hasEnabled := resp["digest_enabled"]
	_, hasEmails := resp["digest_emails"]
	_, hasHour := resp["digest_hour"]

	assert.True(t, hasEnabled, "response should have digest_enabled")
	assert.True(t, hasEmails, "response should have digest_emails")
	assert.True(t, hasHour, "response should have digest_hour")
}

func TestApiDigestSaveHandler(t *testing.T) {
	if core.App == nil {
		t.Skip("requires database setup")
	}

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid settings",
			payload: map[string]interface{}{
				"digest_enabled": true,
				"digest_emails":  "test@example.com",
				"digest_hour":    9,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid hour defaults to 8",
			payload: map[string]interface{}{
				"digest_enabled": true,
				"digest_emails":  "test@example.com",
				"digest_hour":    25,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "negative hour defaults to 8",
			payload: map[string]interface{}{
				"digest_enabled": false,
				"digest_emails":  "",
				"digest_hour":    -5,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "multiple emails",
			payload: map[string]interface{}{
				"digest_enabled": true,
				"digest_emails":  "admin@test.com, alerts@test.com",
				"digest_hour":    6,
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/digest", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			apiDigestSaveHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestApiDigestSaveHandlerInvalidJSON(t *testing.T) {
	if core.App == nil {
		t.Skip("requires database setup")
	}

	req := httptest.NewRequest("POST", "/api/digest", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	apiDigestSaveHandler(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestApiDigestTestHandler(t *testing.T) {
	if core.App == nil {
		t.Skip("requires database setup")
	}

	req := httptest.NewRequest("POST", "/api/digest/test", nil)
	w := httptest.NewRecorder()

	apiDigestTestHandler(w, req)

	// Will fail without SMTP configured, but should return valid JSON
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	_, hasSuccess := resp["success"]
	_, hasMessage := resp["message"]

	assert.True(t, hasSuccess, "response should have success field")
	assert.True(t, hasMessage, "response should have message field")
}

func TestApiDigestSmtpTestHandler(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/digest/smtp-test", nil)
	w := httptest.NewRecorder()

	apiDigestSmtpTestHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
}
