package notifiers

import (
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

var mobileToken string

func TestMobileNotifierMock(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)

	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)
	core.Example()

	testMobile := &mobilePush{&notifications.Notification{
		Method:      "mobile",
		Title:       "Mobile",
		Description: "Test mobile notifier",
		Author:      "Hunter Long",
		Delay:       time.Duration(100 * time.Millisecond),
		Limits:      99,
		Var1:        null.NewNullString("test_device_id_123"),
		Enabled:     null.NewNullBool(true),
	}}

	t.Run("mobile Select", func(t *testing.T) {
		notif := testMobile.Select()
		assert.NotNil(t, notif)
		assert.Equal(t, "mobile", notif.Method)
	})

	t.Run("mobile Valid", func(t *testing.T) {
		err := testMobile.Valid(notifications.Values{})
		assert.Nil(t, err)
	})

	t.Run("mobile OnSave", func(t *testing.T) {
		_, err := testMobile.OnSave()
		assert.Nil(t, err)
	})

	_ = services.Example(true)
	_ = failures.Example()
}

func TestMobileNotifier(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)

	t.Parallel()

	mobileToken = utils.Params.GetString("MOBILE_TOKEN")
	if mobileToken == "" {
		t.Log("Mobile notifier testing skipped, missing MOBILE_ID environment variable")
		t.SkipNow()
	}

	Mobile.Var1 = null.NewNullString(mobileToken)

	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)
	core.Example()

	t.Run("Load Mobile", func(t *testing.T) {
		Mobile.Var1 = null.NewNullString(mobileToken)
		Mobile.Delay = time.Duration(100 * time.Millisecond)
		Mobile.Limits = 10
		Mobile.Enabled = null.NewNullBool(true)

		Add(Mobile)

		assert.Equal(t, "Hunter Long", Mobile.Author)
		assert.Equal(t, mobileToken, Mobile.Var1.String)
	})

	t.Run("Mobile Notifier Tester", func(t *testing.T) {
		assert.True(t, Mobile.CanSend())
	})

	t.Run("Mobile OnSave", func(t *testing.T) {
		_, err := Mobile.OnSave()
		assert.Nil(t, err)
	})

	t.Run("Mobile OnFailure", func(t *testing.T) {
		_, err := Mobile.OnFailure(services.Example(false), failures.Example())
		assert.Nil(t, err)
	})

	t.Run("Mobile OnSuccess", func(t *testing.T) {
		_, err := Mobile.OnSuccess(services.Example(true))
		assert.Nil(t, err)
	})

	t.Run("Mobile Test", func(t *testing.T) {
		_, err := Mobile.OnTest()
		assert.Nil(t, err)
	})
}
