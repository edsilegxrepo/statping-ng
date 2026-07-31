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

var MATTERMOST_URL string

func TestMattermostNotifier(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)

	t.Parallel()
	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)
	core.Example()

	MATTERMOST_URL = utils.Params.GetString("MATTERMOST_URL")
	if MATTERMOST_URL == "" {
		t.Log("mattermost notifier testing skipped, missing MATTERMOST_URL environment variable")
		t.SkipNow()
	}

	mattermoster.Host = null.NewNullString(MATTERMOST_URL)
	mattermoster.Enabled = null.NewNullBool(true)

	t.Run("Load mattermost", func(t *testing.T) {
		mattermoster.Host = null.NewNullString(MATTERMOST_URL)
		mattermoster.Delay = 100 * time.Millisecond
		mattermoster.Limits = 3
		Add(mattermoster)
		assert.Equal(t, "Adam Boutcher", mattermoster.Author)
		assert.Equal(t, MATTERMOST_URL, mattermoster.Host.String)
	})

	t.Run("mattermost Within Limits", func(t *testing.T) {
		ok := mattermoster.CanSend()
		assert.True(t, ok)
	})

	t.Run("mattermost OnSave", func(t *testing.T) {
		_, err := mattermoster.OnSave()
		assert.Nil(t, err)
	})

	t.Run("mattermost OnFailure", func(t *testing.T) {
		_, err := mattermoster.OnFailure(services.Example(false), failures.Example())
		assert.Nil(t, err)
	})

	t.Run("mattermost OnSuccess", func(t *testing.T) {
		_, err := mattermoster.OnSuccess(services.Example(true))
		assert.Nil(t, err)
	})
}

func TestMattermostNotifierMock(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)

	// Allow internal URLs for httptest mock server
	utils.Params.Set("ALLOW_INTERNAL_URLS", true)
	defer utils.Params.Set("ALLOW_INTERNAL_URLS", false)

	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)
	core.Example()

	var receivedBody string
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer mockServer.Close()

	testMattermost := &mattermost{&notifications.Notification{
		Method:      mattermostMethod,
		Title:       "Mattermost",
		Description: "Test mattermost",
		Author:      "Adam Boutcher",
		Delay:       time.Duration(100 * time.Millisecond),
		Limits:      99,
		Host:        null.NewNullString(mockServer.URL),
		SuccessData: null.NewNullString(`{"text":"Service {{.Service.Name}} is online"}`),
		FailureData: null.NewNullString(`{"text":"Service {{.Service.Name}} is offline"}`),
		DataType:    "json",
		Enabled:     null.NewNullBool(true),
	}}

	t.Run("mattermost Select", func(t *testing.T) {
		notif := testMattermost.Select()
		assert.NotNil(t, notif)
		assert.Equal(t, mattermostMethod, notif.Method)
	})

	t.Run("mattermost Valid", func(t *testing.T) {
		err := testMattermost.Valid(notifications.Values{})
		assert.Nil(t, err)
	})

	t.Run("mattermost OnSave", func(t *testing.T) {
		_, err := testMattermost.OnSave()
		assert.Nil(t, err)
	})

	t.Run("mattermost OnFailure with mock", func(t *testing.T) {
		receivedBody = ""
		_, err := testMattermost.OnFailure(services.Example(false), failures.Example())
		assert.Nil(t, err)
		assert.Contains(t, receivedBody, "Statping Example")
		assert.Contains(t, receivedBody, "offline")
	})

	t.Run("mattermost OnSuccess with mock", func(t *testing.T) {
		receivedBody = ""
		_, err := testMattermost.OnSuccess(services.Example(true))
		assert.Nil(t, err)
		assert.Contains(t, receivedBody, "Statping Example")
		assert.Contains(t, receivedBody, "online")
	})

	t.Run("mattermost OnTest with mock", func(t *testing.T) {
		receivedBody = ""
		msg, err := testMattermost.OnTest()
		assert.Nil(t, err)
		assert.Equal(t, "ok", msg)
	})
}
