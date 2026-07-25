package services

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = utils.InitLogs()
}

func TestCheckHTTP(t *testing.T) {
	t.Run("Successful HTTP check", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}))
		defer server.Close()

		s := &Service{
			Name:           "Test HTTP",
			Domain:         server.URL,
			Type:           "http",
			Method:         "GET",
			Timeout:        10,
			ExpectedStatus: 200,
		}

		result, err := CheckHttp(s, false)
		assert.NoError(t, err)
		assert.True(t, result.Online)
		assert.Equal(t, 200, result.LastStatusCode)
	})

	t.Run("HTTP returns expected status code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}))
		defer server.Close()

		s := &Service{
			Name:           "Test HTTP 201",
			Domain:         server.URL,
			Type:           "http",
			Method:         "GET",
			Timeout:        10,
			ExpectedStatus: 201,
		}

		result, err := CheckHttp(s, false)
		assert.NoError(t, err)
		assert.True(t, result.Online)
		assert.Equal(t, 201, result.LastStatusCode)
	})

	t.Run("HTTP unexpected status code fails", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		s := &Service{
			Name:           "Test HTTP 500",
			Domain:         server.URL,
			Type:           "http",
			Method:         "GET",
			Timeout:        10,
			ExpectedStatus: 200,
		}

		result, _ := CheckHttp(s, false)
		assert.False(t, result.Online)
		assert.Equal(t, 500, result.LastStatusCode)
	})

	t.Run("HTTP with expected body regex", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"status": "healthy", "version": "1.0.0"}`))
		}))
		defer server.Close()

		s := &Service{
			Name:           "Test HTTP Body",
			Domain:         server.URL,
			Type:           "http",
			Method:         "GET",
			Timeout:        10,
			ExpectedStatus: 200,
			Expected:       null.NewNullString(`"status":\s*"healthy"`),
		}

		result, err := CheckHttp(s, false)
		assert.NoError(t, err)
		assert.True(t, result.Online)
	})

	t.Run("HTTP body regex no match fails", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"status": "unhealthy"}`))
		}))
		defer server.Close()

		s := &Service{
			Name:           "Test HTTP Body Fail",
			Domain:         server.URL,
			Type:           "http",
			Method:         "GET",
			Timeout:        10,
			ExpectedStatus: 200,
			Expected:       null.NewNullString(`"status":\s*"healthy"`),
		}

		result, _ := CheckHttp(s, false)
		assert.False(t, result.Online)
	})

	t.Run("HTTP POST with data", func(t *testing.T) {
		var receivedBody string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			buf := make([]byte, 1024)
			n, _ := r.Body.Read(buf)
			receivedBody = string(buf[:n])
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		s := &Service{
			Name:           "Test HTTP POST",
			Domain:         server.URL,
			Type:           "http",
			Method:         "POST",
			Timeout:        10,
			ExpectedStatus: 200,
			PostData:       null.NewNullString(`{"key": "value"}`),
		}

		result, err := CheckHttp(s, false)
		assert.NoError(t, err)
		assert.True(t, result.Online)
		assert.Equal(t, `{"key": "value"}`, receivedBody)
	})

	t.Run("HTTP with custom headers", func(t *testing.T) {
		var receivedHeader string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHeader = r.Header.Get("X-Custom-Header")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		s := &Service{
			Name:           "Test HTTP Headers",
			Domain:         server.URL,
			Type:           "http",
			Method:         "GET",
			Timeout:        10,
			ExpectedStatus: 200,
			Headers:        null.NewNullString(`X-Custom-Header=test-value`),
		}

		result, err := CheckHttp(s, false)
		assert.NoError(t, err)
		assert.True(t, result.Online)
		assert.Equal(t, "test-value", receivedHeader)
	})

	t.Run("HTTP timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(3 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		s := &Service{
			Name:           "Test HTTP Timeout",
			Domain:         server.URL,
			Type:           "http",
			Method:         "GET",
			Timeout:        1,
			ExpectedStatus: 200,
		}

		result, err := CheckHttp(s, false)
		assert.Error(t, err)
		assert.False(t, result.Online)
	})

	t.Run("HTTP connection refused", func(t *testing.T) {
		s := &Service{
			Name:           "Test Connection Refused",
			Domain:         "http://127.0.0.1:59999",
			Type:           "http",
			Method:         "GET",
			Timeout:        2,
			ExpectedStatus: 200,
		}

		result, err := CheckHttp(s, false)
		assert.Error(t, err)
		assert.False(t, result.Online)
	})
}

func TestCheckTCP(t *testing.T) {
	t.Run("Successful TCP check", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defer listener.Close()

		port := listener.Addr().(*net.TCPAddr).Port

		go func() {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			defer conn.Close()
		}()

		s := &Service{
			Name:    "Test TCP",
			Domain:  "127.0.0.1",
			Type:    "tcp",
			Port:    port,
			Timeout: 10,
		}

		result, err := CheckTcp(s, false)
		assert.NoError(t, err)
		assert.True(t, result.Online)
	})

	t.Run("TCP connection refused", func(t *testing.T) {
		s := &Service{
			Name:    "Test TCP Refused",
			Domain:  "127.0.0.1",
			Type:    "tcp",
			Port:    59998,
			Timeout: 2,
		}

		result, err := CheckTcp(s, false)
		assert.Error(t, err)
		assert.False(t, result.Online)
	})

	t.Run("TCP with IPv6 address format", func(t *testing.T) {
		s := &Service{
			Name:    "Test TCP IPv6",
			Domain:  "::1",
			Type:    "tcp",
			Port:    59997,
			Timeout: 1,
		}

		// This will fail to connect but tests IPv6 address formatting
		result, _ := CheckTcp(s, false)
		assert.False(t, result.Online)
	})
}

func TestCheckCmd(t *testing.T) {
	t.Run("Successful command check", func(t *testing.T) {
		cmdConfig := CmdConfig{
			Cmd:  "echo",
			Args: []string{"hello"},
		}
		configJSON, _ := json.Marshal(cmdConfig)

		s := &Service{
			Name:           "Test Cmd",
			Type:           "cmd",
			Timeout:        10,
			ExpectedStatus: 0,
			PostData:       null.NewNullString(string(configJSON)),
		}

		result, err := CheckCmd(s, false)
		assert.NoError(t, err)
		assert.True(t, result.Online)
		assert.Equal(t, 0, result.LastStatusCode)
		assert.Contains(t, result.LastResponse, "hello")
	})

	t.Run("Command with non-zero exit code", func(t *testing.T) {
		var cmdConfig CmdConfig
		// Use a command that exits with code 1
		if isWindows() {
			cmdConfig = CmdConfig{Cmd: "cmd", Args: []string{"/c", "exit 1"}}
		} else {
			cmdConfig = CmdConfig{Cmd: "sh", Args: []string{"-c", "exit 1"}}
		}
		configJSON, _ := json.Marshal(cmdConfig)

		s := &Service{
			Name:           "Test Cmd Exit 1",
			Type:           "cmd",
			Timeout:        10,
			ExpectedStatus: 1,
			PostData:       null.NewNullString(string(configJSON)),
		}

		result, err := CheckCmd(s, false)
		assert.NoError(t, err)
		assert.True(t, result.Online)
		assert.Equal(t, 1, result.LastStatusCode)
	})

	t.Run("Command exit code mismatch fails", func(t *testing.T) {
		var cmdConfig CmdConfig
		if isWindows() {
			cmdConfig = CmdConfig{Cmd: "cmd", Args: []string{"/c", "exit 1"}}
		} else {
			cmdConfig = CmdConfig{Cmd: "sh", Args: []string{"-c", "exit 1"}}
		}
		configJSON, _ := json.Marshal(cmdConfig)

		s := &Service{
			Name:           "Test Cmd Exit Mismatch",
			Type:           "cmd",
			Timeout:        10,
			ExpectedStatus: 0,
			PostData:       null.NewNullString(string(configJSON)),
		}

		result, _ := CheckCmd(s, false)
		assert.False(t, result.Online)
	})

	t.Run("Command with stdout regex match", func(t *testing.T) {
		cmdConfig := CmdConfig{
			Cmd:    "echo",
			Args:   []string{"version: 1.2.3"},
			Stdout: `version:\s*\d+\.\d+\.\d+`,
		}
		configJSON, _ := json.Marshal(cmdConfig)

		s := &Service{
			Name:           "Test Cmd Stdout",
			Type:           "cmd",
			Timeout:        10,
			ExpectedStatus: 0,
			PostData:       null.NewNullString(string(configJSON)),
		}

		result, err := CheckCmd(s, false)
		assert.NoError(t, err)
		assert.True(t, result.Online)
	})

	t.Run("Command with invalid JSON config", func(t *testing.T) {
		s := &Service{
			Name:     "Test Cmd Invalid",
			Type:     "cmd",
			Timeout:  10,
			PostData: null.NewNullString("not json"),
		}

		_, err := CheckCmd(s, false)
		assert.Error(t, err)
	})

	t.Run("Command timeout", func(t *testing.T) {
		var cmdConfig CmdConfig
		if isWindows() {
			cmdConfig = CmdConfig{Cmd: "cmd", Args: []string{"/c", "timeout /t 5"}}
		} else {
			cmdConfig = CmdConfig{Cmd: "sleep", Args: []string{"5"}}
		}
		configJSON, _ := json.Marshal(cmdConfig)

		s := &Service{
			Name:           "Test Cmd Timeout",
			Type:           "cmd",
			Timeout:        1,
			ExpectedStatus: 0,
			PostData:       null.NewNullString(string(configJSON)),
		}

		result, _ := CheckCmd(s, false)
		assert.False(t, result.Online)
	})

	t.Run("Command with environment variables", func(t *testing.T) {
		cmdConfig := CmdConfig{
			Cmd:  "echo",
			Args: []string{"test"},
			Env:  map[string]string{"TEST_VAR": "test_value"},
		}
		configJSON, _ := json.Marshal(cmdConfig)

		s := &Service{
			Name:           "Test Cmd Env",
			Type:           "cmd",
			Timeout:        10,
			ExpectedStatus: 0,
			PostData:       null.NewNullString(string(configJSON)),
		}

		result, err := CheckCmd(s, false)
		assert.NoError(t, err)
		assert.True(t, result.Online)
	})
}

func TestCmdConfig(t *testing.T) {
	t.Run("Parse CmdConfig JSON", func(t *testing.T) {
		jsonStr := `{"cmd": "echo", "args": ["hello", "world"], "dir": "/tmp", "env": {"KEY": "VALUE"}}`
		var config CmdConfig
		err := json.Unmarshal([]byte(jsonStr), &config)

		assert.NoError(t, err)
		assert.Equal(t, "echo", config.Cmd)
		assert.Equal(t, []string{"hello", "world"}, config.Args)
		assert.Equal(t, "/tmp", config.Dir)
		assert.Equal(t, "VALUE", config.Env["KEY"])
	})

	t.Run("CmdConfig with stdin and stdout regex", func(t *testing.T) {
		jsonStr := `{"cmd": "cat", "stdin": "input data", "stdout": "input.*"}`
		var config CmdConfig
		err := json.Unmarshal([]byte(jsonStr), &config)

		assert.NoError(t, err)
		assert.Equal(t, "input data", config.Stdin)
		assert.Equal(t, "input.*", config.Stdout)
	})
}

func TestParseHost(t *testing.T) {
	tests := []struct {
		svcType  string
		domain   string
		expected string
	}{
		{"tcp", "example.com", "example.com"},
		{"udp", "example.com", "example.com"},
		{"grpc", "example.com", "example.com"},
		{"http", "https://example.com/path", "example.com"},
		{"http", "https://example.com:8080/path", "example.com"},
		{"http", "http://127.0.0.1:3000", "127.0.0.1"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.svcType, tt.domain), func(t *testing.T) {
			s := &Service{Type: tt.svcType, Domain: tt.domain}
			result := parseHost(s)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsIPv6(t *testing.T) {
	tests := []struct {
		address  string
		expected bool
	}{
		{"192.168.1.1", false},
		{"127.0.0.1", false},
		{"::1", true},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"fe80::1", true},
		{"example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.address, func(t *testing.T) {
			result := isIPv6(tt.address)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMakeCmdEnv(t *testing.T) {
	oldEnv := []string{"PATH=/usr/bin", "HOME=/home/user"}
	newEnv := map[string]string{"TEST": "value", "DEBUG": "true"}

	result := makeCmdEnv(oldEnv, newEnv)

	assert.Contains(t, result, "PATH=/usr/bin")
	assert.Contains(t, result, "HOME=/home/user")

	hasTest := false
	hasDebug := false
	for _, env := range result {
		if strings.HasPrefix(env, "TEST=") {
			hasTest = true
		}
		if strings.HasPrefix(env, "DEBUG=") {
			hasDebug = true
		}
	}
	assert.True(t, hasTest)
	assert.True(t, hasDebug)
}

func TestHTTPRedirects(t *testing.T) {
	t.Run("Follow redirects", func(t *testing.T) {
		finalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("final destination"))
		}))
		defer finalServer.Close()

		redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, finalServer.URL, http.StatusFound)
		}))
		defer redirectServer.Close()

		s := &Service{
			Name:           "Test Redirect",
			Domain:         redirectServer.URL,
			Type:           "http",
			Method:         "GET",
			Timeout:        10,
			ExpectedStatus: 200,
			Redirect:       null.NewNullBool(true),
		}

		result, err := CheckHttp(s, false)
		assert.NoError(t, err)
		assert.True(t, result.Online)
		assert.Equal(t, 200, result.LastStatusCode)
	})

	t.Run("No follow redirects", func(t *testing.T) {
		redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "http://example.com", http.StatusFound)
		}))
		defer redirectServer.Close()

		s := &Service{
			Name:           "Test No Redirect",
			Domain:         redirectServer.URL,
			Type:           "http",
			Method:         "GET",
			Timeout:        10,
			ExpectedStatus: 302,
			Redirect:       null.NewNullBool(false),
		}

		result, err := CheckHttp(s, false)
		assert.NoError(t, err)
		assert.True(t, result.Online)
		assert.Equal(t, 302, result.LastStatusCode)
	})
}

func TestServiceLatencyMeasurement(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := &Service{
		Name:           "Test Latency",
		Domain:         server.URL,
		Type:           "http",
		Method:         "GET",
		Timeout:        10,
		ExpectedStatus: 200,
	}

	result, err := CheckHttp(s, false)
	assert.NoError(t, err)
	assert.True(t, result.Online)
	// Latency should be at least 50ms (50000 microseconds)
	assert.Greater(t, result.Latency, int64(40000))
}

func isWindows() bool {
	return strings.Contains(strings.ToLower(utils.Directory), "\\") ||
		strings.Contains(strings.ToLower(utils.Directory), "c:") ||
		strings.Contains(strings.ToLower(utils.Directory), "d:")
}
