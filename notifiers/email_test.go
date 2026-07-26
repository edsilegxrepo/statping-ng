package notifiers

import (
	"bufio"
	"net"
	"strings"
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
	EMAIL_HOST     string
	EMAIL_USER     string
	EMAIL_PASS     string
	EMAIL_OUTGOING string
	EMAIL_SEND_TO  string
	EMAIL_PORT     int64
)

func TestEmailNotifier(t *testing.T) {
	t.Parallel()
	err := utils.InitLogs()
	require.Nil(t, err)

	EMAIL_HOST = utils.Params.GetString("EMAIL_HOST")
	EMAIL_USER = utils.Params.GetString("EMAIL_USER")
	EMAIL_PASS = utils.Params.GetString("EMAIL_PASS")
	EMAIL_OUTGOING = utils.Params.GetString("EMAIL_OUTGOING")
	EMAIL_SEND_TO = utils.Params.GetString("EMAIL_SEND_TO")
	EMAIL_PORT = utils.ToInt(utils.Params.GetString("EMAIL_PORT"))

	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)
	core.Example()

	if EMAIL_HOST == "" || EMAIL_USER == "" || EMAIL_PASS == "" {
		t.Log("email notifier testing skipped, missing EMAIL_ environment variables")
		t.SkipNow()
	}

	t.Run("New email", func(t *testing.T) {
		email.Host = null.NewNullString(EMAIL_HOST)
		email.Username = null.NewNullString(EMAIL_USER)
		email.Password = null.NewNullString(EMAIL_PASS)
		email.Var1 = null.NewNullString(EMAIL_OUTGOING)
		email.Var2 = null.NewNullString(EMAIL_SEND_TO)
		email.Port = null.NewNullInt64(EMAIL_PORT)
		email.Delay = time.Duration(100 * time.Millisecond)
		email.Enabled = null.NewNullBool(true)

		Add(email)
		assert.Equal(t, "Hunter Long", email.Author)
		assert.Equal(t, EMAIL_HOST, email.Host.String)
	})

	t.Run("email Within Limits", func(t *testing.T) {
		ok := email.CanSend()
		assert.True(t, ok)
	})

	t.Run("email OnSave", func(t *testing.T) {
		_, err := email.OnSave()
		assert.Nil(t, err)
	})

	t.Run("email OnFailure", func(t *testing.T) {
		_, err := email.OnFailure(services.Example(false), failures.Example())
		assert.Nil(t, err)
	})

	t.Run("email OnSuccess", func(t *testing.T) {
		_, err := email.OnSuccess(services.Example(false))
		assert.Nil(t, err)
	})

	t.Run("email Check Back Online", func(t *testing.T) {
		assert.True(t, services.Example(true).Online)
	})

	t.Run("email OnSuccess Again", func(t *testing.T) {
		_, err := email.OnSuccess(services.Example(true))
		assert.Nil(t, err)
	})

	t.Run("email Test", func(t *testing.T) {
		_, err := email.OnTest()
		assert.Nil(t, err)
	})
}

func TestEmailNotifierMock(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)

	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)
	core.Example()

	testEmail := &emailer{&notifications.Notification{
		Method:      "email",
		Title:       "SMTP Mail",
		Description: "Test email notifier",
		Author:      "Hunter Long",
		Delay:       time.Duration(100 * time.Millisecond),
		Limits:      99,
		Host:        null.NewNullString("smtp.example.com"),
		Username:    null.NewNullString("user@example.com"),
		Password:    null.NewNullString("secret"),
		Port:        null.NewNullInt64(587),
		Var1:        null.NewNullString("from@example.com"),
		Var2:        null.NewNullString("to@example.com"),
		Enabled:     null.NewNullBool(true),
	}}

	t.Run("email Select", func(t *testing.T) {
		notif := testEmail.Select()
		assert.NotNil(t, notif)
		assert.Equal(t, "email", notif.Method)
	})

	t.Run("email Valid", func(t *testing.T) {
		err := testEmail.Valid(notifications.Values{})
		assert.Nil(t, err)
	})

	t.Run("email OnSave", func(t *testing.T) {
		_, err := testEmail.OnSave()
		assert.Nil(t, err)
	})

	t.Run("email CanSend", func(t *testing.T) {
		assert.True(t, testEmail.CanSend())
	})

	_ = services.Example(true)
	_ = failures.Example()
}

func TestEmailWithMockSMTPServer(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)

	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)
	core.Example()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.Nil(t, err)
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	var receivedData strings.Builder
	serverReady := make(chan bool, 1)
	serverDone := make(chan bool, 1)

	go func() {
		serverReady <- true
		conn, err := listener.Accept()
		if err != nil {
			serverDone <- true
			return
		}
		defer conn.Close()

		conn.Write([]byte("220 mock.smtp.server ESMTP\r\n"))
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			line := scanner.Text()
			receivedData.WriteString(line + "\n")
			if strings.HasPrefix(line, "EHLO") || strings.HasPrefix(line, "HELO") {
				conn.Write([]byte("250-mock.smtp.server Hello\r\n"))
				conn.Write([]byte("250 OK\r\n"))
			} else if strings.HasPrefix(line, "MAIL FROM:") {
				conn.Write([]byte("250 OK\r\n"))
			} else if strings.HasPrefix(line, "RCPT TO:") {
				conn.Write([]byte("250 OK\r\n"))
			} else if strings.HasPrefix(line, "DATA") {
				conn.Write([]byte("354 Start mail input\r\n"))
			} else if line == "." {
				conn.Write([]byte("250 OK\r\n"))
			} else if strings.HasPrefix(line, "QUIT") {
				conn.Write([]byte("221 Bye\r\n"))
				break
			}
		}
		serverDone <- true
	}()

	<-serverReady

	testEmail := &emailer{&notifications.Notification{
		Method:      "email",
		Title:       "SMTP Mail",
		Description: "Test email notifier with mock SMTP",
		Author:      "Hunter Long",
		Delay:       time.Duration(100 * time.Millisecond),
		Limits:      99,
		Host:        null.NewNullString("127.0.0.1"),
		Username:    null.NewNullString(""),
		Password:    null.NewNullString(""),
		Port:        null.NewNullInt64(int64(addr.Port)),
		Var1:        null.NewNullString("from@example.com"),
		Var2:        null.NewNullString("to@example.com"),
		ApiKey:      null.NewNullString("true"),
		Enabled:     null.NewNullBool(true),
	}}

	t.Run("email OnFailure with mock SMTP", func(t *testing.T) {
		_, err := testEmail.OnFailure(services.Example(false), failures.Example())
		assert.Nil(t, err)
	})

	select {
	case <-serverDone:
	case <-time.After(2 * time.Second):
		t.Log("SMTP mock server timed out")
	}

	data := receivedData.String()
	assert.Contains(t, data, "MAIL FROM:")
	assert.Contains(t, data, "RCPT TO:")
}
