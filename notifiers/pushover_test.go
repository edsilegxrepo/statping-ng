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
	PUSHOVER_TOKEN string
	PUSHOVER_API   string
)

func TestPushoverNotifier(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)

	t.Parallel()
	PUSHOVER_TOKEN = utils.Params.GetString("PUSHOVER_TOKEN")
	PUSHOVER_API = utils.Params.GetString("PUSHOVER_API")

	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)
	core.Example()

	if PUSHOVER_TOKEN == "" || PUSHOVER_API == "" {
		t.Log("Pushover notifier testing skipped, missing PUSHOVER_TOKEN and PUSHOVER_API environment variable")
		t.SkipNow()
	}

	t.Run("Load Pushover", func(t *testing.T) {
		Pushover.ApiKey = null.NewNullString(PUSHOVER_TOKEN)
		Pushover.ApiSecret = null.NewNullString(PUSHOVER_API)
		Pushover.Var1 = null.NewNullString("Normal")
		Pushover.Var2 = null.NewNullString("vibrate")
		Pushover.Enabled = null.NewNullBool(true)

		Add(Pushover)

		assert.Nil(t, err)
		assert.Equal(t, "Hunter Long", Pushover.Author)
		assert.Equal(t, PUSHOVER_TOKEN, Pushover.ApiKey.String)
	})

	t.Run("Pushover Within Limits", func(t *testing.T) {
		assert.True(t, Pushover.CanSend())
	})

	t.Run("Pushover OnSave", func(t *testing.T) {
		_, err := Pushover.OnSave()
		assert.Nil(t, err)
	})

	t.Run("Pushover OnFailure", func(t *testing.T) {
		_, err := Pushover.OnFailure(services.Example(false), failures.Example())
		assert.Nil(t, err)
	})

	t.Run("Pushover OnSuccess", func(t *testing.T) {
		_, err := Pushover.OnSuccess(services.Example(true))
		assert.Nil(t, err)
	})

	t.Run("Pushover Test", func(t *testing.T) {
		_, err := Pushover.OnTest()
		assert.Nil(t, err)
	})
}

func TestPushoverNotifierMock(t *testing.T) {
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
		_, _ = w.Write([]byte(`{"status":1}`))
	}))
	defer mockServer.Close()

	testPushover := &pushover{&notifications.Notification{
		Method:      "pushover",
		Title:       "Pushover",
		Description: "Test pushover",
		Author:      "Hunter Long",
		Delay:       time.Duration(100 * time.Millisecond),
		Limits:      99,
		Host:        null.NewNullString(mockServer.URL),
		ApiKey:      null.NewNullString("user_token"),
		ApiSecret:   null.NewNullString("app_token"),
		Var1:        null.NewNullString("normal"),
		Var2:        null.NewNullString("pushover"),
		SuccessData: null.NewNullString("Service {{.Service.Name}} is online"),
		FailureData: null.NewNullString("Service {{.Service.Name}} is offline"),
		DataType:    "text",
		Enabled:     null.NewNullBool(true),
	}}

	t.Run("pushover Select", func(t *testing.T) {
		notif := testPushover.Select()
		assert.NotNil(t, notif)
		assert.Equal(t, "pushover", notif.Method)
	})

	t.Run("pushover Valid", func(t *testing.T) {
		err := testPushover.Valid(notifications.Values{})
		assert.Nil(t, err)
	})

	t.Run("pushover OnSave", func(t *testing.T) {
		_, err := testPushover.OnSave()
		assert.Nil(t, err)
	})
}
