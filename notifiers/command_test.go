package notifiers

import (
	"runtime"
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

func TestCommandNotifier(t *testing.T) {
	t.Parallel()
	err := utils.InitLogs()
	require.Nil(t, err)
	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)
	core.Example()

	// Use platform-appropriate echo command
	var echoCmd string
	if runtime.GOOS == "windows" {
		echoCmd = "cmd"
	} else {
		echoCmd = "/bin/echo"
	}

	// Create a fresh command notifier for testing
	testCommand := &commandLine{&notifications.Notification{
		Method:      "command",
		Title:       "Command",
		Description: "Test command notifier",
		Author:      "Hunter Long",
		Delay:       time.Duration(100 * time.Millisecond),
		Limits:      99,
		Host:        null.NewNullString(echoCmd),
		Var1:        null.NewNullString("service {{.Service.Domain}} is online"),
		Var2:        null.NewNullString("service {{.Service.Domain}} is offline"),
		Enabled:     null.NewNullBool(true),
	}}

	t.Run("Load Command", func(t *testing.T) {
		Add(testCommand)
		assert.Equal(t, "Hunter Long", testCommand.Author)
		assert.Equal(t, echoCmd, testCommand.Host.String)
	})

	t.Run("Command Notifier CanSend", func(t *testing.T) {
		assert.True(t, testCommand.CanSend())
	})

	t.Run("Command Valid", func(t *testing.T) {
		err := testCommand.Valid(notifications.Values{})
		assert.Nil(t, err)
	})

	t.Run("Command Select", func(t *testing.T) {
		notif := testCommand.Select()
		assert.NotNil(t, notif)
		assert.Equal(t, "command", notif.Method)
	})

	t.Run("Command OnSave", func(t *testing.T) {
		_, err := testCommand.OnSave()
		assert.Nil(t, err)
	})

	// Skip actual command execution tests on Windows (different command syntax)
	if runtime.GOOS != "windows" {
		t.Run("Command OnFailure", func(t *testing.T) {
			_, err := testCommand.OnFailure(services.Example(false), failures.Example())
			assert.Nil(t, err)
		})

		t.Run("Command OnSuccess", func(t *testing.T) {
			_, err := testCommand.OnSuccess(services.Example(true))
			assert.Nil(t, err)
		})

		t.Run("Command OnTest", func(t *testing.T) {
			_, err := testCommand.OnTest()
			assert.Nil(t, err)
		})
	}
}

func TestRunCommand(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)

	// Skip on Windows since command syntax differs
	if runtime.GOOS == "windows" {
		t.Skip("Skipping runCommand test on Windows")
	}

	t.Run("valid command", func(t *testing.T) {
		out, _, err := runCommand("echo hello", nil)
		assert.Nil(t, err)
		assert.Contains(t, out, "hello")
	})

	t.Run("empty command returns error", func(t *testing.T) {
		_, _, err := runCommand("", nil)
		assert.Error(t, err)
	})
}
