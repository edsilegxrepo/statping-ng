package notifiers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/notifications"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLineNotifyMock(t *testing.T) {
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
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":200,"message":"ok"}`))
	}))
	defer mockServer.Close()

	testLine := &lineNotifier{&notifications.Notification{
		Method:      lineNotifyMethod,
		Title:       "LINE Notify",
		Description: "Test line notify",
		Author:      "Kanin Peanviriyakulkit",
		Delay:       time.Duration(100 * time.Millisecond),
		Limits:      99,
		Host:        null.NewNullString(mockServer.URL),
		ApiSecret:   null.NewNullString("test_token_123"),
		Enabled:     null.NewNullBool(true),
	}}

	t.Run("line Select", func(t *testing.T) {
		notif := testLine.Select()
		assert.NotNil(t, notif)
		assert.Equal(t, lineNotifyMethod, notif.Method)
	})

	t.Run("line Valid", func(t *testing.T) {
		err := testLine.Valid(notifications.Values{})
		assert.Nil(t, err)
	})

	t.Run("line OnSave", func(t *testing.T) {
		_, err := testLine.OnSave()
		assert.Nil(t, err)
	})

	_ = receivedBody
	_ = receivedAuth
}
