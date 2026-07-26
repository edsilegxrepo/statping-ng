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

func TestAmazonSNSNotifierMock(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)

	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)
	core.Example()

	testSNS := &amazonSNS{&notifications.Notification{
		Method:      "amazon_sns",
		Title:       "Amazon SNS",
		Description: "Test SNS notifier",
		Author:      "Hunter Long",
		Delay:       time.Duration(100 * time.Millisecond),
		Limits:      99,
		ApiKey:      null.NewNullString("test_access_key"),
		ApiSecret:   null.NewNullString("test_secret_key"),
		Var1:        null.NewNullString("us-east-1"),
		Host:        null.NewNullString("arn:aws:sns:us-east-1:123456789:test-topic"),
		Enabled:     null.NewNullBool(true),
	}}

	t.Run("SNS Select", func(t *testing.T) {
		notif := testSNS.Select()
		assert.NotNil(t, notif)
		assert.Equal(t, "amazon_sns", notif.Method)
	})

	t.Run("SNS Valid", func(t *testing.T) {
		err := testSNS.Valid(notifications.Values{})
		assert.Nil(t, err)
	})

	t.Run("SNS OnSave", func(t *testing.T) {
		_, err := testSNS.OnSave()
		assert.Nil(t, err)
	})

	_ = services.Example(true)
	_ = failures.Example()
}

func TestAmazonSNSNotifier(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)
	snsToken := utils.Params.GetString("SNS_TOKEN")
	snsSecret := utils.Params.GetString("SNS_SECRET")
	snsRegion := utils.Params.GetString("SNS_REGION")
	snsTopic := utils.Params.GetString("SNS_TOPIC")

	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)
	core.Example()

	if snsToken == "" || snsSecret == "" || snsRegion == "" || snsTopic == "" {
		t.Log("SNS notifier testing skipped, missing SNS_TOKEN, SNS_SECRET, SNS_REGION, SNS_TOPIC environment variables")
		t.SkipNow()
	}

	t.Run("Load SNS", func(t *testing.T) {
		AmazonSNS.ApiKey = null.NewNullString(snsToken)
		AmazonSNS.ApiSecret = null.NewNullString(snsSecret)
		AmazonSNS.Var1 = null.NewNullString(snsRegion)
		AmazonSNS.Host = null.NewNullString(snsTopic)
		AmazonSNS.Delay = 15 * time.Second
		AmazonSNS.Enabled = null.NewNullBool(true)

		Add(AmazonSNS)

		assert.Equal(t, "Hunter Long", AmazonSNS.Author)
		assert.Equal(t, snsToken, AmazonSNS.ApiKey.String)
		assert.Equal(t, snsSecret, AmazonSNS.ApiSecret.String)
	})

	t.Run("SNS Notifier Tester", func(t *testing.T) {
		assert.True(t, AmazonSNS.CanSend())
	})

	t.Run("SNS Notifier Tester OnSave", func(t *testing.T) {
		_, err := AmazonSNS.OnSave()
		assert.Nil(t, err)
	})

	t.Run("SNS OnFailure", func(t *testing.T) {
		_, err := AmazonSNS.OnFailure(services.Example(false), failures.Example())
		assert.Nil(t, err)
	})

	t.Run("SNS OnSuccess", func(t *testing.T) {
		_, err := AmazonSNS.OnSuccess(services.Example(true))
		assert.Nil(t, err)
	})

	t.Run("SNS Test", func(t *testing.T) {
		_, err := AmazonSNS.OnTest()
		assert.Nil(t, err)
	})
}
