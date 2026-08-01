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
	telegramToken   string
	telegramChannel string
)

func TestTelegramNotifier(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)

	t.Parallel()

	telegramToken = utils.Params.GetString("TELEGRAM_TOKEN")
	telegramChannel = utils.Params.GetString("TELEGRAM_CHANNEL")
	if telegramToken == "" || telegramChannel == "" {
		t.Log("Telegram notifier testing skipped, missing TELEGRAM_TOKEN and TELEGRAM_CHANNEL environment variable")
		t.SkipNow()
	}

	Telegram.ApiSecret = null.NewNullString(telegramToken)
	Telegram.Var1 = null.NewNullString(telegramChannel)

	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)
	core.Example()

	t.Run("Load Telegram", func(t *testing.T) {
		Telegram.ApiSecret = null.NewNullString(telegramToken)
		Telegram.Var1 = null.NewNullString(telegramChannel)
		Telegram.Delay = time.Duration(1 * time.Second)
		Telegram.Enabled = null.NewNullBool(true)

		Add(Telegram)

		assert.Equal(t, "Hunter Long", Telegram.Author)
		assert.Equal(t, telegramToken, Telegram.ApiSecret.String)
		assert.Equal(t, telegramChannel, Telegram.Var1.String)
	})

	t.Run("Telegram Within Limits", func(t *testing.T) {
		assert.True(t, Telegram.CanSend())
	})

	t.Run("Telegram OnSave", func(t *testing.T) {
		_, err := Telegram.OnSave()
		assert.Nil(t, err)
	})

	t.Run("Telegram OnFailure", func(t *testing.T) {
		_, err := Telegram.OnFailure(services.Example(false), failures.Example())
		assert.Nil(t, err)
	})

	t.Run("Telegram OnSuccess", func(t *testing.T) {
		_, err := Telegram.OnSuccess(services.Example(true))
		assert.Nil(t, err)
	})

	t.Run("Telegram Test", func(t *testing.T) {
		_, err := Telegram.OnTest()
		assert.Nil(t, err)
	})
}

func TestTelegramNotifierMock(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)

	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)
	core.Example()

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":123}}`))
	}))
	defer mockServer.Close()

	testTelegram := &telegram{&notifications.Notification{
		Method:      "telegram",
		Title:       "Telegram",
		Description: "Test telegram",
		Author:      "Hunter Long",
		Delay:       time.Duration(100 * time.Millisecond),
		Limits:      99,
		Host:        null.NewNullString(mockServer.URL),
		ApiSecret:   null.NewNullString("test_token"),
		Var1:        null.NewNullString("@testchannel"),
		SuccessData: null.NewNullString("Service {{.Service.Name}} is online"),
		FailureData: null.NewNullString("Service {{.Service.Name}} is offline"),
		Enabled:     null.NewNullBool(true),
	}}

	t.Run("telegram Select", func(t *testing.T) {
		notif := testTelegram.Select()
		assert.NotNil(t, notif)
		assert.Equal(t, "telegram", notif.Method)
	})

	t.Run("telegram Valid", func(t *testing.T) {
		err := testTelegram.Valid(notifications.Values{})
		assert.Nil(t, err)
	})
}
