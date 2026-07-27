package notifications

import (
	"sort"
	"testing"
	"time"

	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = utils.InitLogs()
	utils.InitEnvs()
}

func TestNotification_Name(t *testing.T) {
	tests := []struct {
		method   string
		expected string
	}{
		{"Slack", "slack"},
		{"Discord Bot", "discord_bot"},
		{"EMAIL", "email"},
		{"Push Over", "push_over"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			n := Notification{Method: tt.method}
			assert.Equal(t, tt.expected, n.Name())
		})
	}
}

func TestNotification_LastSentDur(t *testing.T) {
	n := Notification{
		LastSent: time.Now().Add(-5 * time.Minute),
	}
	dur := n.LastSentDur()
	assert.True(t, dur >= 5*time.Minute && dur < 6*time.Minute,
		"Expected duration around 5 minutes, got %v", dur)
}

func TestNotification_CanSend(t *testing.T) {
	t.Run("disabled notifier cannot send", func(t *testing.T) {
		n := &Notification{
			Enabled: null.NewNullBool(false),
		}
		assert.False(t, n.CanSend())
	})

	t.Run("enabled notifier with no limits can send", func(t *testing.T) {
		n := &Notification{
			Enabled:       null.NewNullBool(true),
			Limits:        100,
			LastSent:      time.Now().Add(-2 * time.Hour),
			LastSentCount: 0,
		}
		assert.True(t, n.CanSend())
	})

	t.Run("enabled notifier at limit cannot send", func(t *testing.T) {
		n := &Notification{
			Enabled:       null.NewNullBool(true),
			Limits:        5,
			LastSent:      time.Now(),
			LastSentCount: 5,
		}
		assert.False(t, n.CanSend())
	})

	t.Run("enabled notifier resets count after timeout", func(t *testing.T) {
		n := &Notification{
			Enabled:       null.NewNullBool(true),
			Limits:        5,
			LastSent:      time.Now().Add(-2 * time.Hour),
			LastSentCount: 3,
		}
		assert.True(t, n.CanSend())
		// Count should be decremented
		assert.Equal(t, 2, n.LastSentCount)
	})
}

func TestNotification_GetValue(t *testing.T) {
	n := &Notification{
		Host:      null.NewNullString("smtp.example.com"),
		Port:      null.NewNullInt64(587),
		Username:  null.NewNullString("user@example.com"),
		Password:  null.NewNullString("secret123"),
		Var1:      null.NewNullString("variable1"),
		Var2:      null.NewNullString("variable2"),
		ApiKey:    null.NewNullString("api-key-123"),
		ApiSecret: null.NewNullString("api-secret-456"),
		Limits:    60,
	}

	tests := []struct {
		field    string
		expected string
	}{
		{"host", "smtp.example.com"},
		{"HOST", "smtp.example.com"}, // case insensitive
		{"port", "587"},
		{"username", "user@example.com"},
		{"password", "secret123"},
		{"var1", "variable1"},
		{"var2", "variable2"},
		{"api_key", "api-key-123"},
		{"api_secret", "api-secret-456"},
		{"limits", "60"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			assert.Equal(t, tt.expected, n.GetValue(tt.field))
		})
	}
}

func TestNotification_CanSend_EnabledDisabledState(t *testing.T) {
	t.Run("freshly disabled notifier cannot send", func(t *testing.T) {
		n := &Notification{
			Enabled:       null.NewNullBool(false),
			Limits:        100,
			LastSent:      time.Time{}, // zero time
			LastSentCount: 0,
		}
		assert.False(t, n.CanSend())
	})

	t.Run("enabled notifier with zero LastSent can send", func(t *testing.T) {
		n := &Notification{
			Enabled:       null.NewNullBool(true),
			Limits:        10,
			LastSent:      time.Time{}, // zero time - never sent
			LastSentCount: 0,
		}
		assert.True(t, n.CanSend())
	})

	t.Run("toggle enabled state affects CanSend", func(t *testing.T) {
		n := &Notification{
			Enabled:       null.NewNullBool(true),
			Limits:        10,
			LastSent:      time.Now().Add(-2 * time.Hour),
			LastSentCount: 0,
		}
		assert.True(t, n.CanSend())

		// Disable
		n.Enabled = null.NewNullBool(false)
		assert.False(t, n.CanSend())

		// Re-enable
		n.Enabled = null.NewNullBool(true)
		assert.True(t, n.CanSend())
	})
}

func TestNotification_CanSend_LimitsHandling(t *testing.T) {
	t.Run("zero limits blocks all sends", func(t *testing.T) {
		n := &Notification{
			Enabled:       null.NewNullBool(true),
			Limits:        0,
			LastSent:      time.Now().Add(-2 * time.Hour),
			LastSentCount: 0,
		}
		assert.False(t, n.CanSend())
	})

	t.Run("negative limits blocks all sends", func(t *testing.T) {
		n := &Notification{
			Enabled:       null.NewNullBool(true),
			Limits:        -1,
			LastSent:      time.Now().Add(-2 * time.Hour),
			LastSentCount: 0,
		}
		assert.False(t, n.CanSend())
	})

	t.Run("high limits allow many sends", func(t *testing.T) {
		n := &Notification{
			Enabled:       null.NewNullBool(true),
			Limits:        1000,
			LastSent:      time.Now(),
			LastSentCount: 999,
		}
		assert.True(t, n.CanSend())
	})

	t.Run("count exactly at limit cannot send", func(t *testing.T) {
		n := &Notification{
			Enabled:       null.NewNullBool(true),
			Limits:        5,
			LastSent:      time.Now(),
			LastSentCount: 5,
		}
		assert.False(t, n.CanSend())
	})

	t.Run("count one below limit can send", func(t *testing.T) {
		n := &Notification{
			Enabled:       null.NewNullBool(true),
			Limits:        5,
			LastSent:      time.Now(),
			LastSentCount: 4,
		}
		assert.True(t, n.CanSend())
	})
}

func TestNotification_CanSend_RateLimitReset(t *testing.T) {
	t.Run("count decrements after 60 minute window", func(t *testing.T) {
		n := &Notification{
			Enabled:       null.NewNullBool(true),
			Limits:        5,
			LastSent:      time.Now().Add(-61 * time.Minute),
			LastSentCount: 5,
		}
		result := n.CanSend()
		assert.True(t, result)
		assert.Equal(t, 4, n.LastSentCount)
	})

	t.Run("count does not decrement within 60 minute window", func(t *testing.T) {
		n := &Notification{
			Enabled:       null.NewNullBool(true),
			Limits:        5,
			LastSent:      time.Now().Add(-59 * time.Minute),
			LastSentCount: 3,
		}
		_ = n.CanSend()
		assert.Equal(t, 3, n.LastSentCount)
	})

	t.Run("count does not go negative", func(t *testing.T) {
		n := &Notification{
			Enabled:       null.NewNullBool(true),
			Limits:        5,
			LastSent:      time.Now().Add(-61 * time.Minute),
			LastSentCount: 0,
		}
		result := n.CanSend()
		assert.True(t, result)
		assert.Equal(t, 0, n.LastSentCount)
	})

	t.Run("multiple calls decrement count each time after timeout", func(t *testing.T) {
		n := &Notification{
			Enabled:       null.NewNullBool(true),
			Limits:        10,
			LastSent:      time.Now().Add(-2 * time.Hour),
			LastSentCount: 3,
		}

		// First call decrements to 2
		assert.True(t, n.CanSend())
		assert.Equal(t, 2, n.LastSentCount)

		// Second call decrements to 1
		assert.True(t, n.CanSend())
		assert.Equal(t, 1, n.LastSentCount)

		// Third call decrements to 0
		assert.True(t, n.CanSend())
		assert.Equal(t, 0, n.LastSentCount)

		// Fourth call stays at 0
		assert.True(t, n.CanSend())
		assert.Equal(t, 0, n.LastSentCount)
	})
}

func TestNotification_LastSentDur_EdgeCases(t *testing.T) {
	t.Run("zero LastSent gives large duration", func(t *testing.T) {
		n := Notification{
			LastSent: time.Time{},
		}
		dur := n.LastSentDur()
		// Zero time is year 0001, so duration should be very large
		assert.True(t, dur > 24*time.Hour*365)
	})

	t.Run("future LastSent gives negative duration", func(t *testing.T) {
		n := Notification{
			LastSent: time.Now().Add(5 * time.Minute),
		}
		dur := n.LastSentDur()
		assert.True(t, dur < 0)
	})

	t.Run("recent LastSent gives small duration", func(t *testing.T) {
		n := Notification{
			LastSent: time.Now().Add(-10 * time.Second),
		}
		dur := n.LastSentDur()
		assert.True(t, dur >= 10*time.Second && dur < 15*time.Second)
	})
}

func TestNotification_Values(t *testing.T) {
	n := &Notification{
		Host:      null.NewNullString("mail.example.com"),
		Port:      null.NewNullInt64(465),
		Username:  null.NewNullString("admin"),
		Password:  null.NewNullString("pass123"),
		Var1:      null.NewNullString("var1value"),
		Var2:      null.NewNullString("var2value"),
		ApiKey:    null.NewNullString("key123"),
		ApiSecret: null.NewNullString("secret456"),
	}

	values := n.Values()

	assert.Equal(t, "mail.example.com", values.Host)
	assert.Equal(t, int64(465), values.Port)
	assert.Equal(t, "admin", values.Username)
	assert.Equal(t, "pass123", values.Password)
	assert.Equal(t, "var1value", values.Var1)
	assert.Equal(t, "var2value", values.Var2)
	assert.Equal(t, "key123", values.ApiKey)
	assert.Equal(t, "secret456", values.ApiSecret)
}

func TestNotification_Values_EmptyFields(t *testing.T) {
	n := &Notification{}

	values := n.Values()

	assert.Equal(t, "", values.Host)
	assert.Equal(t, int64(0), values.Port)
	assert.Equal(t, "", values.Username)
	assert.Equal(t, "", values.Password)
	assert.Equal(t, "", values.Var1)
	assert.Equal(t, "", values.Var2)
	assert.Equal(t, "", values.ApiKey)
	assert.Equal(t, "", values.ApiSecret)
}

func TestNotification_UpdateFields(t *testing.T) {
	t.Run("updates all fields from source", func(t *testing.T) {
		original := &Notification{
			Id:          1,
			Method:      "slack",
			Title:       "Slack Notifier",
			Description: "Sends to Slack",
		}

		update := &Notification{
			Id:          2,
			Limits:      50,
			Enabled:     null.NewNullBool(true),
			Host:        null.NewNullString("new-host.com"),
			Port:        null.NewNullInt64(8080),
			Username:    null.NewNullString("newuser"),
			Password:    null.NewNullString("newpass"),
			ApiKey:      null.NewNullString("newkey"),
			ApiSecret:   null.NewNullString("newsecret"),
			Var1:        null.NewNullString("newvar1"),
			Var2:        null.NewNullString("newvar2"),
			SuccessData: null.NewNullString("success template"),
			FailureData: null.NewNullString("failure template"),
		}

		result := original.UpdateFields(update)

		assert.Equal(t, int64(2), result.Id)
		assert.Equal(t, 50, result.Limits)
		assert.True(t, result.Enabled.Bool)
		assert.Equal(t, "new-host.com", result.Host.String)
		assert.Equal(t, int64(8080), result.Port.Int64)
		assert.Equal(t, "newuser", result.Username.String)
		assert.Equal(t, "newpass", result.Password.String)
		assert.Equal(t, "newkey", result.ApiKey.String)
		assert.Equal(t, "newsecret", result.ApiSecret.String)
		assert.Equal(t, "newvar1", result.Var1.String)
		assert.Equal(t, "newvar2", result.Var2.String)
		assert.Equal(t, "success template", result.SuccessData.String)
		assert.Equal(t, "failure template", result.FailureData.String)

		// Non-updated fields preserved
		assert.Equal(t, "slack", result.Method)
		assert.Equal(t, "Slack Notifier", result.Title)
		assert.Equal(t, "Sends to Slack", result.Description)
	})

	t.Run("returns self when update is nil", func(t *testing.T) {
		original := &Notification{
			Id:     1,
			Method: "discord",
			Limits: 25,
		}

		result := original.UpdateFields(nil)

		assert.Equal(t, original, result)
		assert.Equal(t, int64(1), result.Id)
		assert.Equal(t, "discord", result.Method)
		assert.Equal(t, 25, result.Limits)
	})

	t.Run("returns same pointer", func(t *testing.T) {
		original := &Notification{Id: 1}
		update := &Notification{Id: 2}

		result := original.UpdateFields(update)

		assert.Same(t, original, result)
	})
}

func TestNotificationOrder_Sorting(t *testing.T) {
	notifications := NotificationOrder{
		{Id: 5, Method: "slack"},
		{Id: 1, Method: "email"},
		{Id: 3, Method: "discord"},
		{Id: 2, Method: "telegram"},
		{Id: 4, Method: "pushover"},
	}

	sort.Sort(notifications)

	expected := []int64{1, 2, 3, 4, 5}
	for i, n := range notifications {
		assert.Equal(t, expected[i], n.Id)
	}
}

func TestNotificationOrder_Empty(t *testing.T) {
	notifications := NotificationOrder{}

	assert.Equal(t, 0, notifications.Len())
	sort.Sort(notifications) // Should not panic
}

func TestNotificationOrder_SingleElement(t *testing.T) {
	notifications := NotificationOrder{
		{Id: 1, Method: "email"},
	}

	assert.Equal(t, 1, notifications.Len())
	sort.Sort(notifications)
	assert.Equal(t, int64(1), notifications[0].Id)
}

func TestNotificationOrder_AlreadySorted(t *testing.T) {
	notifications := NotificationOrder{
		{Id: 1, Method: "a"},
		{Id: 2, Method: "b"},
		{Id: 3, Method: "c"},
	}

	sort.Sort(notifications)

	assert.Equal(t, int64(1), notifications[0].Id)
	assert.Equal(t, int64(2), notifications[1].Id)
	assert.Equal(t, int64(3), notifications[2].Id)
}

func TestNotificationLog(t *testing.T) {
	t.Run("successful log entry", func(t *testing.T) {
		log := &NotificationLog{
			Message:   "Notification sent successfully",
			Error:     nil,
			Success:   true,
			Service:   42,
			CreatedAt: time.Now(),
		}

		assert.Equal(t, "Notification sent successfully", log.Message)
		assert.Nil(t, log.Error)
		assert.True(t, log.Success)
		assert.Equal(t, int64(42), log.Service)
	})

	t.Run("failed log entry with error", func(t *testing.T) {
		testErr := assert.AnError
		log := &NotificationLog{
			Message:   "Failed to send notification",
			Error:     testErr,
			Success:   false,
			Service:   123,
			CreatedAt: time.Now(),
		}

		assert.Equal(t, "Failed to send notification", log.Message)
		assert.Equal(t, testErr, log.Error)
		assert.False(t, log.Success)
		assert.Equal(t, int64(123), log.Service)
	})
}

func TestNotification_Logs(t *testing.T) {
	now := time.Now()
	logs := []*NotificationLog{
		{Message: "sent 1", Success: true, Service: 1, CreatedAt: now.Add(-2 * time.Minute)},
		{Message: "failed 1", Success: false, Service: 1, CreatedAt: now.Add(-1 * time.Minute)},
		{Message: "sent 2", Success: true, Service: 2, CreatedAt: now},
	}

	n := &Notification{
		Id:     1,
		Method: "test",
		Logs:   logs,
	}

	assert.Len(t, n.Logs, 3)
	assert.True(t, n.Logs[0].Success)
	assert.False(t, n.Logs[1].Success)
	assert.True(t, n.Logs[2].Success)
}

func TestNotificationForm(t *testing.T) {
	form := NotificationForm{
		Type:        "password",
		Title:       "API Token",
		Placeholder: "Enter your API token",
		DbField:     "api_key",
		SmallText:   "Found in your account settings",
		Required:    true,
		IsHidden:    false,
		ListOptions: []string{"option1", "option2"},
	}

	assert.Equal(t, "password", form.Type)
	assert.Equal(t, "API Token", form.Title)
	assert.Equal(t, "Enter your API token", form.Placeholder)
	assert.Equal(t, "api_key", form.DbField)
	assert.Equal(t, "Found in your account settings", form.SmallText)
	assert.True(t, form.Required)
	assert.False(t, form.IsHidden)
	assert.Equal(t, []string{"option1", "option2"}, form.ListOptions)
}

func TestNotification_Form(t *testing.T) {
	forms := []NotificationForm{
		{Type: "text", Title: "Host", DbField: "host", Required: true},
		{Type: "number", Title: "Port", DbField: "port", Required: true},
		{Type: "password", Title: "Password", DbField: "password", Required: false},
	}

	n := &Notification{
		Id:     1,
		Method: "smtp",
		Form:   forms,
	}

	assert.Len(t, n.Form, 3)
	assert.Equal(t, "host", n.Form[0].DbField)
	assert.Equal(t, "port", n.Form[1].DbField)
	assert.Equal(t, "password", n.Form[2].DbField)
}

func TestNotification_Name_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		expected string
	}{
		{"empty method", "", ""},
		{"multiple spaces", "Push  Over", "push__over"},
		{"leading space", " Slack", "_slack"},
		{"trailing space", "Slack ", "slack_"},
		{"all uppercase with spaces", "MY NOTIFIER", "my_notifier"},
		{"mixed case no spaces", "SlackBot", "slackbot"},
		{"numbers in name", "SMS 2FA", "sms_2fa"},
		{"special chars preserved", "Slack-Bot", "slack-bot"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := Notification{Method: tt.method}
			assert.Equal(t, tt.expected, n.Name())
		})
	}
}

func TestNotification_Struct_Defaults(t *testing.T) {
	n := Notification{}

	assert.Equal(t, int64(0), n.Id)
	assert.Equal(t, "", n.Method)
	assert.False(t, n.Enabled.Bool)
	assert.Equal(t, 0, n.Limits)
	assert.False(t, n.Removable)
	assert.True(t, n.LastSent.IsZero())
	assert.Equal(t, 0, n.LastSentCount)
	assert.Nil(t, n.Logs)
	assert.Nil(t, n.Form)
	assert.Equal(t, time.Duration(0), n.Delay)
}

func TestNotification_CanSend_ComplexScenarios(t *testing.T) {
	t.Run("simulates burst then recovery", func(t *testing.T) {
		n := &Notification{
			Enabled:       null.NewNullBool(true),
			Limits:        3,
			LastSent:      time.Now(),
			LastSentCount: 0,
		}

		// Can send up to limit
		assert.True(t, n.CanSend())
		n.LastSentCount++
		assert.True(t, n.CanSend())
		n.LastSentCount++
		assert.True(t, n.CanSend())
		n.LastSentCount++

		// Now at limit
		assert.False(t, n.CanSend())

		// Simulate time passing - reset window
		n.LastSent = time.Now().Add(-61 * time.Minute)

		// Should be able to send again (count decrements)
		assert.True(t, n.CanSend())
		assert.Equal(t, 2, n.LastSentCount)
	})

	t.Run("disabled mid-session blocks sends", func(t *testing.T) {
		n := &Notification{
			Enabled:       null.NewNullBool(true),
			Limits:        10,
			LastSent:      time.Now(),
			LastSentCount: 2,
		}

		assert.True(t, n.CanSend())

		// Admin disables notifier mid-session
		n.Enabled = null.NewNullBool(false)
		assert.False(t, n.CanSend())

		// Even with count reset, still disabled
		n.LastSentCount = 0
		assert.False(t, n.CanSend())
	})
}

func TestNotification_GetValue_EmptyValues(t *testing.T) {
	n := &Notification{
		Limits: 0,
	}

	assert.Equal(t, "", n.GetValue("host"))
	assert.Equal(t, "0", n.GetValue("port"))
	assert.Equal(t, "", n.GetValue("username"))
	assert.Equal(t, "", n.GetValue("password"))
	assert.Equal(t, "", n.GetValue("var1"))
	assert.Equal(t, "", n.GetValue("var2"))
	assert.Equal(t, "", n.GetValue("api_key"))
	assert.Equal(t, "", n.GetValue("api_secret"))
	assert.Equal(t, "0", n.GetValue("limits"))
}

func TestNotification_Timestamps(t *testing.T) {
	now := time.Now()
	n := &Notification{
		Id:        1,
		Method:    "test",
		CreatedAt: now.Add(-24 * time.Hour),
		UpdatedAt: now,
	}

	require.False(t, n.CreatedAt.IsZero())
	require.False(t, n.UpdatedAt.IsZero())
	assert.True(t, n.UpdatedAt.After(n.CreatedAt))
}

func TestNotification_Delay(t *testing.T) {
	n := &Notification{
		Id:     1,
		Method: "test",
		Delay:  5 * time.Second,
	}

	assert.Equal(t, 5*time.Second, n.Delay)
}

func TestNotification_DataFields(t *testing.T) {
	n := &Notification{
		Id:          1,
		Method:      "webhook",
		SuccessData: null.NewNullString(`{"status": "up", "service": "{{.Name}}"}`),
		FailureData: null.NewNullString(`{"status": "down", "service": "{{.Name}}", "error": "{{.Message}}"}`),
		DataType:    "application/json",
		RequestInfo: "POST https://webhook.example.com/notify",
	}

	assert.Contains(t, n.SuccessData.String, "up")
	assert.Contains(t, n.FailureData.String, "down")
	assert.Equal(t, "application/json", n.DataType)
	assert.Contains(t, n.RequestInfo, "POST")
}

func TestNotification_AuthorInfo(t *testing.T) {
	n := &Notification{
		Id:          1,
		Method:      "custom_notifier",
		Title:       "Custom Notifier",
		Description: "A custom notification integration",
		Author:      "Developer Name",
		AuthorUrl:   "https://github.com/developer",
		Icon:        "bell",
	}

	assert.Equal(t, "Custom Notifier", n.Title)
	assert.Equal(t, "A custom notification integration", n.Description)
	assert.Equal(t, "Developer Name", n.Author)
	assert.Equal(t, "https://github.com/developer", n.AuthorUrl)
	assert.Equal(t, "bell", n.Icon)
}

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) database.Database {
	testDB, err := database.Openw("sqlite", ":memory:")
	require.NoError(t, err)
	require.NotNil(t, testDB)

	// Auto-migrate the Notification table
	err = testDB.GormDB().AutoMigrate(&Notification{})
	require.NoError(t, err)

	// Set the package-level db
	SetDB(testDB)

	return testDB
}

func TestSetDB(t *testing.T) {
	testDB := setupTestDB(t)
	defer func() {
		sqlDB, _ := testDB.DB()
		_ = sqlDB.Close()
	}()

	// SetDB should not panic and db should be set
	assert.NotNil(t, db)
}

func TestNotification_Create(t *testing.T) {
	testDB := setupTestDB(t)
	defer func() {
		sqlDB, _ := testDB.DB()
		_ = sqlDB.Close()
	}()

	t.Run("creates new notification", func(t *testing.T) {
		n := &Notification{
			Method:      "test_create",
			Host:        null.NewNullString("localhost"),
			Port:        null.NewNullInt64(8080),
			Enabled:     null.NewNullBool(true),
			Limits:      50,
			SuccessData: null.NewNullString("success"),
			FailureData: null.NewNullString("failure"),
		}

		err := n.Create()
		require.NoError(t, err)

		// Verify it was created
		found, err := Find("test_create")
		require.NoError(t, err)
		assert.Equal(t, "localhost", found.Host.String)
		assert.Equal(t, int64(8080), found.Port.Int64)
	})

	t.Run("updates existing notification on create", func(t *testing.T) {
		// Create initial
		n1 := &Notification{
			Method:      "test_update_on_create",
			Host:        null.NewNullString("host1"),
			SuccessData: null.NewNullString(""),
			FailureData: null.NewNullString(""),
		}
		err := n1.Create()
		require.NoError(t, err)

		// "Create" again with same method - should update defaults
		n2 := &Notification{
			Method:      "test_update_on_create",
			Host:        null.NewNullString("host2"),
			SuccessData: null.NewNullString("new success"),
			FailureData: null.NewNullString("new failure"),
		}
		err = n2.Create()
		require.NoError(t, err)

		// Should have updated the success/failure data
		found, err := Find("test_update_on_create")
		require.NoError(t, err)
		assert.Equal(t, "new success", found.SuccessData.String)
		assert.Equal(t, "new failure", found.FailureData.String)
	})
}

func TestNotification_Update(t *testing.T) {
	testDB := setupTestDB(t)
	defer func() {
		sqlDB, _ := testDB.DB()
		_ = sqlDB.Close()
	}()

	// Create a notification first
	n := &Notification{
		Method:  "test_update",
		Host:    null.NewNullString("original"),
		Enabled: null.NewNullBool(false),
		Limits:  10,
	}
	err := testDB.Create(n).Error()
	require.NoError(t, err)

	// Update it
	n.Host = null.NewNullString("updated")
	n.Enabled = null.NewNullBool(true)
	n.Limits = 100

	err = n.Update()
	require.NoError(t, err)

	// Verify update
	found, err := Find("test_update")
	require.NoError(t, err)
	assert.Equal(t, "updated", found.Host.String)
	assert.True(t, found.Enabled.Bool)
	assert.Equal(t, 100, found.Limits)
}

func TestFind(t *testing.T) {
	testDB := setupTestDB(t)
	defer func() {
		sqlDB, _ := testDB.DB()
		_ = sqlDB.Close()
	}()

	t.Run("finds existing notification", func(t *testing.T) {
		n := &Notification{
			Method:   "test_find",
			Username: null.NewNullString("testuser"),
		}
		err := testDB.Create(n).Error()
		require.NoError(t, err)

		found, err := Find("test_find")
		require.NoError(t, err)
		assert.Equal(t, "testuser", found.Username.String)
	})

	t.Run("returns error for non-existent notification", func(t *testing.T) {
		_, err := Find("nonexistent_method")
		assert.Error(t, err)
	})
}

func TestAll(t *testing.T) {
	testDB := setupTestDB(t)
	defer func() {
		sqlDB, _ := testDB.DB()
		_ = sqlDB.Close()
	}()

	t.Run("returns all notifications", func(t *testing.T) {
		// Create multiple notifications
		methods := []string{"slack", "email", "discord"}
		for _, method := range methods {
			n := &Notification{Method: method}
			err := testDB.Create(n).Error()
			require.NoError(t, err)
		}

		all := All()
		assert.Len(t, all, 3)
	})

	t.Run("returns nil for empty table", func(t *testing.T) {
		// Create fresh DB
		emptyDB, err := database.Openw("sqlite", ":memory:")
		require.NoError(t, err)
		defer func() {
			sqlDB, _ := emptyDB.DB()
			_ = sqlDB.Close()
		}()

		err = emptyDB.GormDB().AutoMigrate(&Notification{})
		require.NoError(t, err)
		SetDB(emptyDB)

		all := All()
		assert.Len(t, all, 0)
	})
}

func TestNotification_Logger(t *testing.T) {
	n := &Notification{
		Method: "test_logger",
	}

	logger := n.Logger()
	assert.NotNil(t, logger)
}

func TestNotification_Hooks(t *testing.T) {
	testDB := setupTestDB(t)
	defer func() {
		sqlDB, _ := testDB.DB()
		_ = sqlDB.Close()
	}()

	t.Run("AfterFind hook executes", func(t *testing.T) {
		n := &Notification{Method: "hook_find_test"}
		err := testDB.Create(n).Error()
		require.NoError(t, err)

		// Find triggers AfterFind hook
		found, err := Find("hook_find_test")
		require.NoError(t, err)
		assert.NotNil(t, found)
	})

	t.Run("AfterCreate hook executes", func(t *testing.T) {
		n := &Notification{Method: "hook_create_test"}
		err := testDB.Create(n).Error()
		require.NoError(t, err)
		// Hook executed without error
	})

	t.Run("AfterUpdate hook executes", func(t *testing.T) {
		n := &Notification{Method: "hook_update_test"}
		err := testDB.Create(n).Error()
		require.NoError(t, err)

		n.Host = null.NewNullString("updated")
		err = n.Update()
		require.NoError(t, err)
		// Hook executed without error
	})

	t.Run("AfterDelete hook executes", func(t *testing.T) {
		n := &Notification{Method: "hook_delete_test"}
		err := testDB.Create(n).Error()
		require.NoError(t, err)

		err = testDB.Delete(n).Error()
		require.NoError(t, err)
		// Hook executed without error
	})
}
