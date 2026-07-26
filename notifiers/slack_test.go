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

var SLACK_URL string

func TestSlackNotifier(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)

	t.Parallel()
	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)
	core.Example()

	SLACK_URL = utils.Params.GetString("SLACK_URL")
	if SLACK_URL == "" {
		t.Log("slack notifier testing skipped, missing SLACK_URL environment variable")
		t.SkipNow()
	}

	slacker.Host = null.NewNullString(SLACK_URL)
	slacker.Enabled = null.NewNullBool(true)

	t.Run("Load slack", func(t *testing.T) {
		slacker.Host = null.NewNullString(SLACK_URL)
		slacker.Delay = 100 * time.Millisecond
		slacker.Limits = 3
		Add(slacker)
		assert.Equal(t, "Hunter Long", slacker.Author)
		assert.Equal(t, SLACK_URL, slacker.Host.String)
	})

	t.Run("slack Within Limits", func(t *testing.T) {
		ok := slacker.CanSend()
		assert.True(t, ok)
	})

	t.Run("slack OnSave", func(t *testing.T) {
		_, err := slacker.OnSave()
		assert.Nil(t, err)
	})

	t.Run("slack OnFailure", func(t *testing.T) {
		_, err := slacker.OnFailure(services.Example(false), failures.Example())
		assert.Nil(t, err)
	})

	t.Run("slack OnSuccess", func(t *testing.T) {
		_, err := slacker.OnSuccess(services.Example(true))
		assert.Nil(t, err)
	})
}

func TestSlackNotifierMock(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)

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

	testSlack := &slack{&notifications.Notification{
		Method:      slackMethod,
		Title:       "Slack",
		Description: "Test slack",
		Author:      "Hunter Long",
		Delay:       time.Duration(100 * time.Millisecond),
		Limits:      99,
		Host:        null.NewNullString(mockServer.URL),
		SuccessData: null.NewNullString(`{"text":"Service {{.Service.Name}} is online"}`),
		FailureData: null.NewNullString(`{"text":"Service {{.Service.Name}} is offline"}`),
		DataType:    "json",
		Enabled:     null.NewNullBool(true),
	}}

	t.Run("slack Select", func(t *testing.T) {
		notif := testSlack.Select()
		assert.NotNil(t, notif)
		assert.Equal(t, slackMethod, notif.Method)
	})

	t.Run("slack Valid", func(t *testing.T) {
		err := testSlack.Valid(notifications.Values{})
		assert.Nil(t, err)
	})

	t.Run("slack OnSave", func(t *testing.T) {
		_, err := testSlack.OnSave()
		assert.Nil(t, err)
	})

	t.Run("slack OnFailure with mock", func(t *testing.T) {
		receivedBody = ""
		_, err := testSlack.OnFailure(services.Example(false), failures.Example())
		assert.Nil(t, err)
		assert.Contains(t, receivedBody, "Statping Example")
		assert.Contains(t, receivedBody, "offline")
	})

	t.Run("slack OnSuccess with mock", func(t *testing.T) {
		receivedBody = ""
		_, err := testSlack.OnSuccess(services.Example(true))
		assert.Nil(t, err)
		assert.Contains(t, receivedBody, "Statping Example")
		assert.Contains(t, receivedBody, "online")
	})

	t.Run("slack OnTest with mock", func(t *testing.T) {
		receivedBody = ""
		msg, err := testSlack.OnTest()
		assert.Nil(t, err)
		assert.Equal(t, "ok", msg)
	})
}
