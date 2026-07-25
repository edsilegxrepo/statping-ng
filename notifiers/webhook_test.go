package notifiers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/failures"
	"github.com/statping-ng/statping-ng/types/notifications"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookNotifier(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)

	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)
	core.Example()

	// Create a mock HTTP server
	var receivedBody string
	var receivedMethod string
	var receivedContentType string
	var receivedHeaders http.Header
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedContentType = r.Header.Get("Content-Type")
		receivedHeaders = r.Header
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer mockServer.Close()

	// Create a fresh webhook instance for testing
	testWebhook := &webhooker{&notifications.Notification{
		Method:      "webhook",
		Title:       "Webhook",
		Description: "Test webhook",
		Author:      "Hunter Long",
		Delay:       0,
		SuccessData: null.NewNullString(`{"id": "{{.Service.Id}}", "online": true}`),
		FailureData: null.NewNullString(`{"id": "{{.Service.Id}}", "online": false}`),
		DataType:    "json",
		Limits:      180,
		Host:        null.NewNullString(mockServer.URL),
		Var1:        null.NewNullString("POST"),
		ApiKey:      null.NewNullString("application/json"),
		Enabled:     null.NewNullBool(true),
	}}

	t.Run("Load webhooker", func(t *testing.T) {
		Add(testWebhook)
		assert.Equal(t, "Hunter Long", testWebhook.Author)
		assert.Equal(t, mockServer.URL, testWebhook.Host.String)
	})

	t.Run("webhooker CanSend", func(t *testing.T) {
		assert.True(t, testWebhook.CanSend())
	})

	t.Run("webhooker Valid", func(t *testing.T) {
		err := testWebhook.Valid(notifications.Values{})
		assert.Nil(t, err)
	})

	t.Run("webhooker Select", func(t *testing.T) {
		notif := testWebhook.Select()
		assert.NotNil(t, notif)
		assert.Equal(t, "webhook", notif.Method)
	})

	t.Run("webhooker OnSave", func(t *testing.T) {
		_, err := testWebhook.OnSave()
		assert.Nil(t, err)
	})

	t.Run("webhooker Send with mock server", func(t *testing.T) {
		err := testWebhook.Send(`{"test": "message"}`)
		assert.Nil(t, err)
		assert.Equal(t, "POST", receivedMethod)
		assert.Equal(t, "application/json", receivedContentType)
		assert.Equal(t, `{"test": "message"}`, receivedBody)
	})

	t.Run("webhooker OnFailure", func(t *testing.T) {
		receivedBody = ""
		_, err := testWebhook.OnFailure(services.Example(false), failures.Example())
		assert.Nil(t, err)
		assert.Contains(t, receivedBody, "6283")
		assert.Contains(t, receivedBody, "false")
	})

	t.Run("webhooker OnSuccess", func(t *testing.T) {
		receivedBody = ""
		_, err := testWebhook.OnSuccess(services.Example(true))
		assert.Nil(t, err)
		assert.Contains(t, receivedBody, "6283")
		assert.Contains(t, receivedBody, "true")
	})

	t.Run("webhooker OnTest", func(t *testing.T) {
		receivedBody = ""
		msg, err := testWebhook.OnTest()
		assert.Nil(t, err)
		assert.NotEmpty(t, msg)
	})

	t.Run("webhooker with custom headers", func(t *testing.T) {
		testWebhook.ApiSecret = null.NewNullString("Authorization=Bearer token123,X-Custom=value")
		err := testWebhook.Send(`{"test": "headers"}`)
		assert.Nil(t, err)
		assert.Equal(t, "Bearer token123", receivedHeaders.Get("Authorization"))
		assert.Equal(t, "value", receivedHeaders.Get("X-Custom"))
	})

	t.Run("webhooker with GET method", func(t *testing.T) {
		testWebhook.Var1 = null.NewNullString("GET")
		err := testWebhook.Send(``)
		assert.Nil(t, err)
		assert.Equal(t, "GET", receivedMethod)
	})
}
