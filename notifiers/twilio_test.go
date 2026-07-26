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
	TWILIO_SID    string
	TWILIO_SECRET string
)

func TestTwilioNotifier(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)

	t.Parallel()

	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)
	core.Example()

	TWILIO_SID = utils.Params.GetString("TWILIO_SID")
	TWILIO_SECRET = utils.Params.GetString("TWILIO_SECRET")

	if TWILIO_SID == "" || TWILIO_SECRET == "" {
		t.Log("twilio notifier testing skipped, missing TWILIO_SID and TWILIO_SECRET environment variable")
		t.SkipNow()
	}

	t.Run("Load Twilio", func(t *testing.T) {
		Twilio.ApiKey = null.NewNullString(TWILIO_SID)
		Twilio.ApiSecret = null.NewNullString(TWILIO_SECRET)
		Twilio.Var1 = null.NewNullString("15005550006")
		Twilio.Var2 = null.NewNullString("15005550006")
		Twilio.Delay = 100 * time.Millisecond
		Twilio.Enabled = null.NewNullBool(true)

		Add(Twilio)

		assert.Nil(t, err)
		assert.Equal(t, "Hunter Long", Twilio.Author)
		assert.Equal(t, TWILIO_SID, Twilio.ApiKey.String)
	})

	t.Run("Twilio Within Limits", func(t *testing.T) {
		assert.True(t, Twilio.CanSend())
	})

	t.Run("Twilio OnSave", func(t *testing.T) {
		_, err := Twilio.OnSave()
		assert.Nil(t, err)
	})

	t.Run("Twilio OnFailure", func(t *testing.T) {
		_, err := Twilio.OnFailure(services.Example(false), failures.Example())
		assert.Nil(t, err)
	})

	t.Run("Twilio OnSuccess", func(t *testing.T) {
		_, err := Twilio.OnSuccess(services.Example(true))
		assert.Nil(t, err)
	})

	t.Run("Twilio Test", func(t *testing.T) {
		_, err := Twilio.OnTest()
		assert.Nil(t, err)
	})
}

func TestTwilioNotifierMock(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)

	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)
	core.Example()

	var receivedBody string
	var receivedAuth string
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"sid":"SM123","status":"sent"}`))
	}))
	defer mockServer.Close()

	testTwilio := &twilio{&notifications.Notification{
		Method:      "twilio",
		Title:       "Twilio",
		Description: "Test twilio",
		Author:      "Hunter Long",
		Delay:       time.Duration(100 * time.Millisecond),
		Limits:      99,
		Host:        null.NewNullString(mockServer.URL),
		ApiKey:      null.NewNullString("test_account_sid"),
		ApiSecret:   null.NewNullString("test_token"),
		Var1:        null.NewNullString("18005551234"),
		Var2:        null.NewNullString("18005554321"),
		SuccessData: null.NewNullString("Service {{.Service.Name}} is online"),
		FailureData: null.NewNullString("Service {{.Service.Name}} is offline"),
		DataType:    "text",
		Enabled:     null.NewNullBool(true),
	}}

	t.Run("twilio Select", func(t *testing.T) {
		notif := testTwilio.Select()
		assert.NotNil(t, notif)
		assert.Equal(t, "twilio", notif.Method)
	})

	t.Run("twilio Valid", func(t *testing.T) {
		err := testTwilio.Valid(notifications.Values{})
		assert.Nil(t, err)
	})

	t.Run("twilio OnSave", func(t *testing.T) {
		_, err := testTwilio.OnSave()
		assert.Nil(t, err)
	})

	_ = receivedBody
	_ = receivedAuth
	_ = services.Example(true)
	_ = failures.Example()
}
