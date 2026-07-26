package notifiers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

var (
	GOTIFY_URL   string
	GOTIFY_TOKEN string
)

func TestGotifyNotifier(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)

	t.Parallel()
	GOTIFY_URL = utils.Params.GetString("GOTIFY_URL")
	GOTIFY_TOKEN = utils.Params.GetString("GOTIFY_TOKEN")

	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)
	core.Example()

	if GOTIFY_URL == "" {
		t.Log("gotify notifier testing skipped, missing GOTIFY_URL environment variable")
		t.SkipNow()
	}
	if GOTIFY_TOKEN == "" {
		t.Log("gotify notifier testing skipped, missing GOTIFY_TOKEN environment variable")
		t.SkipNow()
	}

	t.Run("Load gotify", func(t *testing.T) {
		Gotify.Host = null.NewNullString(GOTIFY_URL)
		Gotify.Delay = time.Duration(100 * time.Millisecond)
		Gotify.Enabled = null.NewNullBool(true)

		Add(Gotify)

		assert.Equal(t, "Hugo van Rijswijk", Gotify.Author)
		assert.Equal(t, GOTIFY_URL, Gotify.Host.String)
	})

	t.Run("gotify Notifier Tester", func(t *testing.T) {
		assert.True(t, Gotify.CanSend())
	})

	t.Run("gotify Notifier Tester OnSave", func(t *testing.T) {
		_, err := Gotify.OnSave()
		assert.Nil(t, err)
	})

	t.Run("gotify OnFailure", func(t *testing.T) {
		_, err := Gotify.OnFailure(services.Example(false), failures.Example())
		assert.Nil(t, err)
	})

	t.Run("gotify OnSuccess", func(t *testing.T) {
		_, err := Gotify.OnSuccess(services.Example(true))
		assert.Nil(t, err)
	})

	t.Run("gotify Test", func(t *testing.T) {
		_, err := Gotify.OnTest()
		assert.Nil(t, err)
	})
}

func TestGotifyNotifierMock(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)

	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)
	core.Example()

	var receivedBody string
	var receivedToken string
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		receivedToken = r.Header.Get("X-Gotify-Key")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	defer mockServer.Close()

	testGotify := &gotify{&notifications.Notification{
		Method:      "gotify",
		Title:       "Gotify",
		Description: "Test gotify",
		Author:      "Hugo van Rijswijk",
		Delay:       time.Duration(100 * time.Millisecond),
		Limits:      99,
		Host:        null.NewNullString(mockServer.URL),
		ApiKey:      null.NewNullString("test_token_123"),
		SuccessData: null.NewNullString(`{"title":"{{.Service.Name}}","message":"online"}`),
		FailureData: null.NewNullString(`{"title":"{{.Service.Name}}","message":"offline"}`),
		DataType:    "json",
		Enabled:     null.NewNullBool(true),
	}}

	t.Run("gotify Select", func(t *testing.T) {
		notif := testGotify.Select()
		assert.NotNil(t, notif)
		assert.Equal(t, "gotify", notif.Method)
	})

	t.Run("gotify Valid", func(t *testing.T) {
		err := testGotify.Valid(notifications.Values{})
		assert.Nil(t, err)
	})

	t.Run("gotify OnSave", func(t *testing.T) {
		_, err := testGotify.OnSave()
		assert.Nil(t, err)
	})

	t.Run("gotify OnFailure with mock", func(t *testing.T) {
		receivedBody = ""
		_, err := testGotify.OnFailure(services.Example(false), failures.Example())
		assert.Nil(t, err)
		assert.Contains(t, receivedBody, "Statping Example")
		assert.Contains(t, receivedBody, "offline")
		assert.Equal(t, "test_token_123", receivedToken)
	})

	t.Run("gotify OnSuccess with mock", func(t *testing.T) {
		receivedBody = ""
		_, err := testGotify.OnSuccess(services.Example(true))
		assert.Nil(t, err)
		assert.Contains(t, receivedBody, "Statping Example")
		assert.Contains(t, receivedBody, "online")
	})

	t.Run("gotify OnTest with mock", func(t *testing.T) {
		receivedBody = ""
		_, err := testGotify.OnTest()
		assert.Nil(t, err)
		assert.Contains(t, receivedBody, "Test")
	})
}
