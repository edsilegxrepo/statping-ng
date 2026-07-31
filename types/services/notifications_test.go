package services

import (
	"errors"
	"testing"
	"time"

	"github.com/statping-ng/statping-ng/types/failures"
	"github.com/statping-ng/statping-ng/types/notifications"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockNotifier is a configurable mock for testing notification behavior
type mockNotifier struct {
	*notifications.Notification
	successCalls   int
	failureCalls   int
	saveCalls      int
	testCalls      int
	onSuccessError error
	onFailureError error
	onSuccessMsg   string
	onFailureMsg   string
}

func (m *mockNotifier) OnSuccess(s *Service) (string, error) {
	m.successCalls++
	return m.onSuccessMsg, m.onSuccessError
}

func (m *mockNotifier) OnFailure(s *Service, f failures.Failure) (string, error) {
	m.failureCalls++
	return m.onFailureMsg, m.onFailureError
}

func (m *mockNotifier) OnSave() (string, error) {
	m.saveCalls++
	return "", nil
}

func (m *mockNotifier) Select() *notifications.Notification {
	return m.Notification
}

func (m *mockNotifier) OnTest() (string, error) {
	m.testCalls++
	return "", nil
}

func (m *mockNotifier) Valid(form notifications.Values) error {
	return nil
}

func (m *mockNotifier) reset() {
	m.successCalls = 0
	m.failureCalls = 0
	m.saveCalls = 0
	m.testCalls = 0
	m.Notification.LastSentCount = 0
	m.Notification.Logs = nil
}

func newMockNotifier(method string, enabled bool) *mockNotifier {
	return &mockNotifier{
		Notification: &notifications.Notification{
			Method:    method,
			CreatedAt: utils.Now().Add(-5 * time.Second),
			Limits:    60,
			Enabled:   null.NewNullBool(enabled),
		},
	}
}

// clearAllNotifiers removes all notifiers from the registry for test isolation
func clearAllNotifiers() {
	notifiersLock.Lock()
	allNotifiers = make(map[string]ServiceNotifier)
	notifiersLock.Unlock()
}

func TestAddNotifier(t *testing.T) {
	t.Run("adds notifier to registry", func(t *testing.T) {
		clearAllNotifiers()
		mock := newMockNotifier("test_add", true)

		AddNotifier(mock)

		// Verify notifier was added
		notifiersLock.RLock()
		n, exists := allNotifiers["test_add"]
		notifiersLock.RUnlock()

		assert.True(t, exists, "notifier should exist in registry")
		assert.Equal(t, mock, n, "notifier should be the same instance")
	})

	t.Run("overwrites existing notifier with same method", func(t *testing.T) {
		clearAllNotifiers()
		mock1 := newMockNotifier("test_overwrite", true)
		mock2 := newMockNotifier("test_overwrite", false)

		AddNotifier(mock1)
		AddNotifier(mock2)

		notifiersLock.RLock()
		n := allNotifiers["test_overwrite"]
		notifiersLock.RUnlock()

		assert.Equal(t, mock2, n, "should contain the second notifier")
	})

	t.Run("multiple notifiers with different methods", func(t *testing.T) {
		clearAllNotifiers()
		mock1 := newMockNotifier("notifier_a", true)
		mock2 := newMockNotifier("notifier_b", true)
		mock3 := newMockNotifier("notifier_c", true)

		AddNotifier(mock1)
		AddNotifier(mock2)
		AddNotifier(mock3)

		notifiers := AllNotifiers()
		assert.Len(t, notifiers, 3, "should have 3 notifiers")
		assert.Contains(t, notifiers, "notifier_a")
		assert.Contains(t, notifiers, "notifier_b")
		assert.Contains(t, notifiers, "notifier_c")
	})
}

func TestUpdateNotifiers(t *testing.T) {
	t.Run("updates notifier fields from database", func(t *testing.T) {
		clearAllNotifiers()

		// Create notifier in database
		dbNotif := &notifications.Notification{
			Method:  "test_update",
			Enabled: null.NewNullBool(true),
			Limits:  100,
			Host:    null.NewNullString("updated-host"),
			Port:    null.NewNullInt64(8080),
		}
		err := db.Create(dbNotif).Error()
		require.NoError(t, err)

		// Add notifier to registry with different values
		mock := newMockNotifier("test_update", false)
		mock.Notification.Limits = 10
		AddNotifier(mock)

		// Run update
		UpdateNotifiers()

		// Verify fields were updated from database
		notif := mock.Select()
		assert.Equal(t, int64(100), int64(notif.Limits))
		assert.Equal(t, "updated-host", notif.Host.String)
		assert.Equal(t, int64(8080), notif.Port.Int64)
	})

	t.Run("handles nil notifier gracefully", func(t *testing.T) {
		clearAllNotifiers()

		// Create notification in database without corresponding notifier in registry
		dbNotif := &notifications.Notification{
			Method:  "orphan_notif",
			Enabled: null.NewNullBool(true),
		}
		err := db.Create(dbNotif).Error()
		require.NoError(t, err)

		// Should not panic when notifier not found in registry
		assert.NotPanics(t, func() {
			UpdateNotifiers()
		})
	})

	t.Run("updates multiple notifiers", func(t *testing.T) {
		clearAllNotifiers()

		// Create notifiers in database
		for i, method := range []string{"multi_a", "multi_b"} {
			dbNotif := &notifications.Notification{
				Method:  method,
				Enabled: null.NewNullBool(true),
				Limits:  100 + i,
			}
			_ = db.Create(dbNotif).Error()

			mock := newMockNotifier(method, false)
			mock.Notification.Limits = 1
			AddNotifier(mock)
		}

		UpdateNotifiers()

		// Verify both were updated
		for i, method := range []string{"multi_a", "multi_b"} {
			n := ReturnNotifier(method)
			require.NotNil(t, n)
			assert.Equal(t, 100+i, n.Select().Limits)
		}
	})
}

func TestLogMessage(t *testing.T) {
	t.Run("logs success message", func(t *testing.T) {
		clearAllNotifiers()
		mock := newMockNotifier("test_log_success", true)
		AddNotifier(mock)

		// Create notification in database
		dbNotif := &notifications.Notification{Method: "test_log_success", Enabled: null.NewNullBool(true)}
		_ = db.Create(dbNotif).Error()

		logMessage("test_log_success", "success message", nil, true, 123)

		notif := mock.Select()
		require.Len(t, notif.Logs, 1)
		assert.Equal(t, "success message", notif.Logs[0].Message)
		assert.True(t, notif.Logs[0].Success)
		assert.Nil(t, notif.Logs[0].Error)
		assert.Equal(t, int64(123), notif.Logs[0].Service)
	})

	t.Run("logs failure message with error", func(t *testing.T) {
		clearAllNotifiers()
		mock := newMockNotifier("test_log_failure", true)
		AddNotifier(mock)

		dbNotif := &notifications.Notification{Method: "test_log_failure", Enabled: null.NewNullBool(true)}
		_ = db.Create(dbNotif).Error()

		testErr := errors.New("notification failed")
		logMessage("test_log_failure", "", testErr, false, 456)

		notif := mock.Select()
		require.Len(t, notif.Logs, 1)
		assert.Equal(t, "", notif.Logs[0].Message)
		assert.False(t, notif.Logs[0].Success)
		assert.Equal(t, testErr, notif.Logs[0].Error)
		assert.Equal(t, int64(456), notif.Logs[0].Service)
	})

	t.Run("handles nonexistent notifier method", func(t *testing.T) {
		clearAllNotifiers()

		// Should not panic when notifier doesn't exist
		assert.NotPanics(t, func() {
			logMessage("nonexistent_method", "message", nil, true, 1)
		})
	})

	t.Run("truncates logs beyond 32 entries", func(t *testing.T) {
		clearAllNotifiers()
		mock := newMockNotifier("test_log_truncate", true)
		AddNotifier(mock)

		dbNotif := &notifications.Notification{Method: "test_log_truncate", Enabled: null.NewNullBool(true)}
		_ = db.Create(dbNotif).Error()

		// Add 35 log entries
		for i := 0; i < 35; i++ {
			logMessage("test_log_truncate", "msg", nil, true, int64(i))
		}

		notif := mock.Select()
		assert.Len(t, notif.Logs, 32, "logs should be truncated to 32")
		// First log should be for service 3 (0,1,2 were dropped)
		assert.Equal(t, int64(3), notif.Logs[0].Service)
	})
}

func TestSendSuccess(t *testing.T) {
	t.Run("sends notification when service comes online", func(t *testing.T) {
		clearAllNotifiers()
		mock := newMockNotifier("test_send_success", true)
		AddNotifier(mock)

		dbNotif := &notifications.Notification{Method: "test_send_success", Enabled: null.NewNullBool(true)}
		_ = db.Create(dbNotif).Error()

		service := Example(true)
		service.AllowNotifications = null.NewNullBool(true)
		service.prevOnline = false // Was offline
		service.Online = true

		sendSuccess(service)

		assert.Equal(t, 1, mock.successCalls, "OnSuccess should be called once")
		assert.True(t, service.prevOnline, "prevOnline should be updated")
	})

	t.Run("skips when notifications disabled", func(t *testing.T) {
		clearAllNotifiers()
		mock := newMockNotifier("test_notif_disabled", true)
		AddNotifier(mock)

		service := Example(true)
		service.AllowNotifications = null.NewNullBool(false)
		service.prevOnline = false
		service.Online = true

		sendSuccess(service)

		assert.Equal(t, 0, mock.successCalls, "OnSuccess should not be called")
	})

	t.Run("skips when already online (no state change)", func(t *testing.T) {
		clearAllNotifiers()
		mock := newMockNotifier("test_already_online", true)
		AddNotifier(mock)

		service := Example(true)
		service.AllowNotifications = null.NewNullBool(true)
		service.prevOnline = true
		service.Online = true

		sendSuccess(service)

		assert.Equal(t, 0, mock.successCalls, "OnSuccess should not be called when already online")
	})

	t.Run("resets notifyAfterCount on success", func(t *testing.T) {
		clearAllNotifiers()
		mock := newMockNotifier("test_reset_count", true)
		AddNotifier(mock)

		service := Example(true)
		service.AllowNotifications = null.NewNullBool(true)
		service.notifyAfterCount = 5

		sendSuccess(service)

		assert.Equal(t, int64(0), service.notifyAfterCount, "notifyAfterCount should be reset")
	})

	t.Run("handles OnSuccess error", func(t *testing.T) {
		clearAllNotifiers()
		mock := newMockNotifier("test_success_error", true)
		mock.onSuccessError = errors.New("notification error")
		AddNotifier(mock)

		dbNotif := &notifications.Notification{Method: "test_success_error", Enabled: null.NewNullBool(true)}
		_ = db.Create(dbNotif).Error()

		service := Example(true)
		service.AllowNotifications = null.NewNullBool(true)
		service.prevOnline = false
		service.Online = true

		sendSuccess(service)

		assert.Equal(t, 1, mock.successCalls, "OnSuccess should still be called")
		// Error should be logged
		notif := mock.Select()
		assert.Len(t, notif.Logs, 1)
		assert.NotNil(t, notif.Logs[0].Error)
	})

	t.Run("skips disabled notifier", func(t *testing.T) {
		clearAllNotifiers()
		enabledMock := newMockNotifier("enabled_notifier", true)
		disabledMock := newMockNotifier("disabled_notifier", false)
		AddNotifier(enabledMock)
		AddNotifier(disabledMock)

		dbNotif1 := &notifications.Notification{Method: "enabled_notifier", Enabled: null.NewNullBool(true)}
		dbNotif2 := &notifications.Notification{Method: "disabled_notifier", Enabled: null.NewNullBool(false)}
		_ = db.Create(dbNotif1).Error()
		_ = db.Create(dbNotif2).Error()

		service := Example(true)
		service.AllowNotifications = null.NewNullBool(true)
		service.prevOnline = false
		service.Online = true

		sendSuccess(service)

		assert.Equal(t, 1, enabledMock.successCalls, "enabled notifier should be called")
		assert.Equal(t, 0, disabledMock.successCalls, "disabled notifier should not be called")
	})
}

func TestSendFailure(t *testing.T) {
	t.Run("sends notification when service goes offline", func(t *testing.T) {
		clearAllNotifiers()
		mock := newMockNotifier("test_send_failure", true)
		AddNotifier(mock)

		dbNotif := &notifications.Notification{Method: "test_send_failure", Enabled: null.NewNullBool(true)}
		_ = db.Create(dbNotif).Error()

		service := Example(true)
		service.AllowNotifications = null.NewNullBool(true)
		service.prevOnline = true
		service.Online = false
		service.NotifyAfter = 0
		service.UpdateNotify = null.NewNullBool(false)

		failure := failures.Example()

		sendFailure(service, &failure)

		assert.Equal(t, 1, mock.failureCalls, "OnFailure should be called once")
		assert.False(t, service.prevOnline, "prevOnline should be updated to false")
	})

	t.Run("skips when notifications disabled", func(t *testing.T) {
		clearAllNotifiers()
		mock := newMockNotifier("test_fail_notif_disabled", true)
		AddNotifier(mock)

		service := Example(false)
		service.AllowNotifications = null.NewNullBool(false)
		service.prevOnline = true
		service.Online = false

		failure := failures.Example()

		sendFailure(service, &failure)

		assert.Equal(t, 0, mock.failureCalls, "OnFailure should not be called")
	})

	t.Run("skips when already offline and update notify disabled", func(t *testing.T) {
		clearAllNotifiers()
		mock := newMockNotifier("test_already_offline", true)
		AddNotifier(mock)

		service := Example(false)
		service.AllowNotifications = null.NewNullBool(true)
		service.prevOnline = false
		service.Online = false
		service.UpdateNotify = null.NewNullBool(false)

		failure := failures.Example()

		sendFailure(service, &failure)

		assert.Equal(t, 0, mock.failureCalls, "OnFailure should not be called when already offline")
	})

	t.Run("notifies when already offline but update notify enabled", func(t *testing.T) {
		clearAllNotifiers()
		mock := newMockNotifier("test_update_notify", true)
		AddNotifier(mock)

		dbNotif := &notifications.Notification{Method: "test_update_notify", Enabled: null.NewNullBool(true)}
		_ = db.Create(dbNotif).Error()

		service := Example(false)
		service.AllowNotifications = null.NewNullBool(true)
		service.prevOnline = false
		service.Online = false
		service.UpdateNotify = null.NewNullBool(true)
		service.NotifyAfter = 0

		failure := failures.Example()

		sendFailure(service, &failure)

		assert.Equal(t, 1, mock.failureCalls, "OnFailure should be called with UpdateNotify enabled")
	})

	t.Run("respects NotifyAfter threshold", func(t *testing.T) {
		clearAllNotifiers()
		mock := newMockNotifier("test_notify_after", true)
		AddNotifier(mock)

		service := Example(true)
		service.AllowNotifications = null.NewNullBool(true)
		service.prevOnline = true
		service.Online = false
		service.NotifyAfter = 3
		service.notifyAfterCount = 0

		failure := failures.Example()

		// First 3 failures should not trigger notification
		for i := 0; i < 3; i++ {
			service.prevOnline = true // Reset for test
			sendFailure(service, &failure)
			assert.Equal(t, 0, mock.failureCalls, "should not notify until threshold reached")
		}

		// Fourth failure should trigger notification
		service.prevOnline = true
		sendFailure(service, &failure)
		assert.Equal(t, 1, mock.failureCalls, "should notify after threshold reached")
	})

	t.Run("handles OnFailure error", func(t *testing.T) {
		clearAllNotifiers()
		mock := newMockNotifier("test_failure_error", true)
		mock.onFailureError = errors.New("failure notification error")
		AddNotifier(mock)

		dbNotif := &notifications.Notification{Method: "test_failure_error", Enabled: null.NewNullBool(true)}
		_ = db.Create(dbNotif).Error()

		service := Example(true)
		service.AllowNotifications = null.NewNullBool(true)
		service.prevOnline = true
		service.Online = false
		service.NotifyAfter = 0

		failure := failures.Example()

		sendFailure(service, &failure)

		assert.Equal(t, 1, mock.failureCalls, "OnFailure should still be called")
		// Error should be logged
		notif := mock.Select()
		assert.GreaterOrEqual(t, len(notif.Logs), 1)
	})

	t.Run("skips disabled notifier", func(t *testing.T) {
		clearAllNotifiers()
		enabledMock := newMockNotifier("fail_enabled", true)
		disabledMock := newMockNotifier("fail_disabled", false)
		AddNotifier(enabledMock)
		AddNotifier(disabledMock)

		dbNotif1 := &notifications.Notification{Method: "fail_enabled", Enabled: null.NewNullBool(true)}
		dbNotif2 := &notifications.Notification{Method: "fail_disabled", Enabled: null.NewNullBool(false)}
		_ = db.Create(dbNotif1).Error()
		_ = db.Create(dbNotif2).Error()

		service := Example(true)
		service.AllowNotifications = null.NewNullBool(true)
		service.prevOnline = true
		service.Online = false
		service.NotifyAfter = 0

		failure := failures.Example()

		sendFailure(service, &failure)

		assert.Equal(t, 1, enabledMock.failureCalls, "enabled notifier should be called")
		assert.Equal(t, 0, disabledMock.failureCalls, "disabled notifier should not be called")
	})
}

func TestAllNotifiers(t *testing.T) {
	t.Run("returns copy of notifiers map", func(t *testing.T) {
		clearAllNotifiers()
		mock := newMockNotifier("test_copy", true)
		AddNotifier(mock)

		notifiers := AllNotifiers()

		// Modify the returned map
		notifiers["modified"] = nil

		// Original should be unchanged
		notifiersLock.RLock()
		_, exists := allNotifiers["modified"]
		notifiersLock.RUnlock()

		assert.False(t, exists, "original map should not be modified")
	})

	t.Run("returns empty map when no notifiers", func(t *testing.T) {
		clearAllNotifiers()

		notifiers := AllNotifiers()

		assert.NotNil(t, notifiers)
		assert.Len(t, notifiers, 0)
	})
}

func TestReturnNotifier(t *testing.T) {
	t.Run("returns notifier by method", func(t *testing.T) {
		clearAllNotifiers()
		mock := newMockNotifier("test_return", true)
		AddNotifier(mock)

		result := ReturnNotifier("test_return")

		assert.Equal(t, mock, result)
	})

	t.Run("returns nil for unknown method", func(t *testing.T) {
		clearAllNotifiers()

		result := ReturnNotifier("unknown")

		assert.Nil(t, result)
	})
}

func TestFindNotifier(t *testing.T) {
	t.Run("finds notifier with database lookup", func(t *testing.T) {
		clearAllNotifiers()
		mock := newMockNotifier("test_find", true)
		AddNotifier(mock)

		dbNotif := &notifications.Notification{
			Method:  "test_find",
			Enabled: null.NewNullBool(true),
			Limits:  50,
		}
		_ = db.Create(dbNotif).Error()

		result := FindNotifier("test_find")

		require.NotNil(t, result)
		assert.Equal(t, "test_find", result.Method)
	})

	t.Run("returns nil for unknown method", func(t *testing.T) {
		clearAllNotifiers()

		result := FindNotifier("unknown_method")

		assert.Nil(t, result)
	})
}

func TestNotificationLimits(t *testing.T) {
	t.Run("respects rate limit", func(t *testing.T) {
		clearAllNotifiers()
		mock := newMockNotifier("test_rate_limit", true)
		mock.Notification.Limits = 2
		mock.Notification.LastSentCount = 0
		AddNotifier(mock)

		dbNotif := &notifications.Notification{Method: "test_rate_limit", Enabled: null.NewNullBool(true), Limits: 2}
		_ = db.Create(dbNotif).Error()

		service := Example(true)
		service.AllowNotifications = null.NewNullBool(true)

		// First notification - should send
		service.prevOnline = false
		service.Online = true
		sendSuccess(service)
		assert.Equal(t, 1, mock.successCalls)

		// Second notification - should send (at limit)
		service.prevOnline = false
		sendSuccess(service)
		assert.Equal(t, 2, mock.successCalls)

		// Third notification - should NOT send (over limit)
		service.prevOnline = false
		sendSuccess(service)
		assert.Equal(t, 2, mock.successCalls, "should not exceed limit")
	})
}
