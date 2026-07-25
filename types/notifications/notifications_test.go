package notifications

import (
	"testing"
	"time"

	"github.com/statping-ng/statping-ng/types/null"
	"github.com/stretchr/testify/assert"
)

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
