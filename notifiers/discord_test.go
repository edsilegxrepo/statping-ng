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

var DISCORD_URL string

func TestDiscordNotifier(t *testing.T) {
	t.Parallel()
	err := utils.InitLogs()
	require.Nil(t, err)
	DISCORD_URL = utils.Params.GetString("DISCORD_URL")

	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)
	core.Example()

	if DISCORD_URL == "" {
		t.Log("discord notifier testing skipped, missing DISCORD_URL environment variable")
		t.SkipNow()
	}

	t.Run("Load discord", func(t *testing.T) {
		Discorder.Host = null.NewNullString(DISCORD_URL)
		Discorder.Delay = time.Duration(100 * time.Millisecond)
		Discorder.Enabled = null.NewNullBool(true)

		Add(Discorder)

		assert.Equal(t, "Hunter Long", Discorder.Author)
		assert.Equal(t, DISCORD_URL, Discorder.Host.String)
	})

	t.Run("discord Notifier Tester", func(t *testing.T) {
		assert.True(t, Discorder.CanSend())
	})

	t.Run("discord Notifier Tester OnSave", func(t *testing.T) {
		_, err := Discorder.OnSave()
		assert.Nil(t, err)
	})

	t.Run("discord OnFailure", func(t *testing.T) {
		_, err := Discorder.OnFailure(services.Example(false), failures.Example())
		assert.Nil(t, err)
	})

	t.Run("discord OnSuccess", func(t *testing.T) {
		_, err := Discorder.OnSuccess(services.Example(true))
		assert.Nil(t, err)
	})

	t.Run("discord Test", func(t *testing.T) {
		_, err := Discorder.OnTest()
		assert.Nil(t, err)
	})
}

func TestDiscordNotifierMock(t *testing.T) {
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
		w.WriteHeader(http.StatusNoContent)
	}))
	defer mockServer.Close()

	testDiscord := &discord{&notifications.Notification{
		Method:      "discord",
		Title:       "Discord",
		Description: "Test discord",
		Author:      "Hunter Long",
		Delay:       time.Duration(100 * time.Millisecond),
		Limits:      99,
		Host:        null.NewNullString(mockServer.URL),
		SuccessData: null.NewNullString(`{"content":"Service {{.Service.Name}} is online"}`),
		FailureData: null.NewNullString(`{"content":"Service {{.Service.Name}} is offline"}`),
		DataType:    "json",
		Enabled:     null.NewNullBool(true),
	}}

	t.Run("discord Select", func(t *testing.T) {
		notif := testDiscord.Select()
		assert.NotNil(t, notif)
		assert.Equal(t, "discord", notif.Method)
	})

	t.Run("discord Valid", func(t *testing.T) {
		err := testDiscord.Valid(notifications.Values{})
		assert.Nil(t, err)
	})

	t.Run("discord OnSave", func(t *testing.T) {
		_, err := testDiscord.OnSave()
		assert.Nil(t, err)
	})

	t.Run("discord OnFailure with mock", func(t *testing.T) {
		receivedBody = ""
		_, err := testDiscord.OnFailure(services.Example(false), failures.Example())
		assert.Nil(t, err)
		assert.Contains(t, receivedBody, "Statping Example")
		assert.Contains(t, receivedBody, "offline")
	})

	t.Run("discord OnSuccess with mock", func(t *testing.T) {
		receivedBody = ""
		_, err := testDiscord.OnSuccess(services.Example(true))
		assert.Nil(t, err)
		assert.Contains(t, receivedBody, "Statping Example")
		assert.Contains(t, receivedBody, "online")
	})

	t.Run("discord OnTest with mock", func(t *testing.T) {
		receivedBody = ""
		_, err := testDiscord.OnTest()
		assert.Nil(t, err)
	})
}
