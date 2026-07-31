package notifiers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
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

// setupTestEnvironment initializes the test database and core for notifier tests.
func setupTestEnvironment(t *testing.T) {
	t.Helper()
	err := utils.InitLogs()
	require.Nil(t, err)

	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)
	core.Example()
}

// =============================================================================
// 1. Network Timeout Handling Tests
// =============================================================================

func TestWebhookNetworkTimeout(t *testing.T) {
	setupTestEnvironment(t)

	// Create a mock server that delays longer than the client timeout
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(15 * time.Second) // Longer than default 10s timeout
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer slowServer.Close()

	testWebhook := &webhooker{&notifications.Notification{
		Method:      "webhook",
		Title:       "Webhook Timeout Test",
		Description: "Test webhook timeout handling",
		Author:      "Test",
		Delay:       0,
		SuccessData: null.NewNullString(`{"id": "{{.Service.Id}}", "online": true}`),
		FailureData: null.NewNullString(`{"id": "{{.Service.Id}}", "online": false}`),
		DataType:    "json",
		Limits:      180,
		Host:        null.NewNullString(slowServer.URL),
		Var1:        null.NewNullString("POST"),
		ApiKey:      null.NewNullString("application/json"),
		Enabled:     null.NewNullBool(true),
	}}

	t.Run("webhook Send times out", func(t *testing.T) {
		start := time.Now()
		err := testWebhook.Send(`{"test": "timeout"}`)
		elapsed := time.Since(start)

		assert.NotNil(t, err, "Expected timeout error")
		assert.True(t, elapsed < 15*time.Second, "Should timeout before server responds")
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})

	t.Run("webhook OnFailure times out", func(t *testing.T) {
		start := time.Now()
		_, err := testWebhook.OnFailure(services.Example(false), failures.Example())
		elapsed := time.Since(start)

		assert.NotNil(t, err, "Expected timeout error")
		assert.True(t, elapsed < 15*time.Second, "Should timeout before server responds")
	})
}

func TestSlackNetworkTimeout(t *testing.T) {
	setupTestEnvironment(t)

	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(15 * time.Second)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer slowServer.Close()

	testSlack := &slack{&notifications.Notification{
		Method:      slackMethod,
		Title:       "Slack Timeout Test",
		Description: "Test slack timeout",
		Author:      "Test",
		Delay:       time.Duration(100 * time.Millisecond),
		Limits:      99,
		Host:        null.NewNullString(slowServer.URL),
		SuccessData: null.NewNullString(`{"text":"Service {{.Service.Name}} is online"}`),
		FailureData: null.NewNullString(`{"text":"Service {{.Service.Name}} is offline"}`),
		DataType:    "json",
		Enabled:     null.NewNullBool(true),
	}}

	t.Run("slack OnFailure times out", func(t *testing.T) {
		start := time.Now()
		_, err := testSlack.OnFailure(services.Example(false), failures.Example())
		elapsed := time.Since(start)

		assert.NotNil(t, err, "Expected timeout error")
		assert.True(t, elapsed < 15*time.Second, "Should timeout before server responds")
	})
}

// =============================================================================
// 2. Malformed API Response Tests
// =============================================================================

func TestWebhookMalformedResponses(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("invalid JSON response", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{not valid json`))
		}))
		defer mockServer.Close()

		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			Var1:        null.NewNullString("POST"),
			SuccessData: null.NewNullString(`{"online": true}`),
			FailureData: null.NewNullString(`{"online": false}`),
			Enabled:     null.NewNullBool(true),
		}}

		// Webhook notifier doesn't parse response JSON, so this should succeed
		err := testWebhook.Send(`{"test": "message"}`)
		assert.Nil(t, err, "Webhook should not fail on malformed response JSON")
	})

	t.Run("unexpected 500 status code", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "internal server error"}`))
		}))
		defer mockServer.Close()

		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			Var1:        null.NewNullString("POST"),
			SuccessData: null.NewNullString(`{"online": true}`),
			FailureData: null.NewNullString(`{"online": false}`),
			Enabled:     null.NewNullBool(true),
		}}

		// Webhook Send doesn't check status code - just that request succeeded
		err := testWebhook.Send(`{"test": "message"}`)
		assert.Nil(t, err, "Webhook should complete request even with 500 status")
	})

	t.Run("unexpected 404 status code", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error": "not found"}`))
		}))
		defer mockServer.Close()

		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			Var1:        null.NewNullString("POST"),
			SuccessData: null.NewNullString(`{"online": true}`),
			FailureData: null.NewNullString(`{"online": false}`),
			Enabled:     null.NewNullBool(true),
		}}

		err := testWebhook.Send(`{"test": "message"}`)
		assert.Nil(t, err, "Webhook should complete request even with 404 status")
	})

	t.Run("empty response body", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			// No body written
		}))
		defer mockServer.Close()

		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			Var1:        null.NewNullString("POST"),
			SuccessData: null.NewNullString(`{"online": true}`),
			FailureData: null.NewNullString(`{"online": false}`),
			Enabled:     null.NewNullBool(true),
		}}

		err := testWebhook.Send(`{"test": "message"}`)
		assert.Nil(t, err, "Webhook should handle empty response body")
	})

	t.Run("truncated response", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "1000")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"partial":`)) // Write less than declared
		}))
		defer mockServer.Close()

		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			Var1:        null.NewNullString("POST"),
			SuccessData: null.NewNullString(`{"online": true}`),
			FailureData: null.NewNullString(`{"online": false}`),
			Enabled:     null.NewNullBool(true),
		}}

		// The webhook just reads the response but doesn't validate completeness
		err := testWebhook.Send(`{"test": "message"}`)
		// This may or may not error depending on http client behavior
		_ = err // Just ensure no panic
	})
}

func TestSlackMalformedResponses(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("slack non-ok response", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("invalid_webhook"))
		}))
		defer mockServer.Close()

		testSlack := &slack{&notifications.Notification{
			Method:      slackMethod,
			Title:       "Slack",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			SuccessData: null.NewNullString(`{"text":"online"}`),
			FailureData: null.NewNullString(`{"text":"offline"}`),
			Enabled:     null.NewNullBool(true),
		}}

		// OnTest specifically checks for "ok" response
		_, err := testSlack.OnTest()
		assert.NotNil(t, err, "Slack OnTest should fail when response is not 'ok'")
		assert.Contains(t, err.Error(), "incorrect")
	})

	t.Run("slack 400 bad request", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("invalid_payload"))
		}))
		defer mockServer.Close()

		testSlack := &slack{&notifications.Notification{
			Method:      slackMethod,
			Title:       "Slack",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			SuccessData: null.NewNullString(`{"text":"online"}`),
			FailureData: null.NewNullString(`{"text":"offline"}`),
			Enabled:     null.NewNullBool(true),
		}}

		_, err := testSlack.OnFailure(services.Example(false), failures.Example())
		// utils.HttpRequest doesn't fail on non-2xx codes, just returns the body
		assert.Nil(t, err)
	})
}

// =============================================================================
// 3. Retry Backoff Behavior Tests
// =============================================================================

func TestWebhookRetryBehavior(t *testing.T) {
	setupTestEnvironment(t)

	var requestCount int32

	t.Run("multiple consecutive failures", func(t *testing.T) {
		atomic.StoreInt32(&requestCount, 0)

		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&requestCount, 1)
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error": "service unavailable"}`))
		}))
		defer mockServer.Close()

		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test retry",
			Author:      "Test",
			Delay:       time.Duration(100 * time.Millisecond),
			Limits:      99,
			Host:        null.NewNullString(mockServer.URL),
			Var1:        null.NewNullString("POST"),
			SuccessData: null.NewNullString(`{"online": true}`),
			FailureData: null.NewNullString(`{"online": false}`),
			Enabled:     null.NewNullBool(true),
		}}

		// Send multiple failure notifications
		for i := 0; i < 5; i++ {
			_, _ = testWebhook.OnFailure(services.Example(false), failures.Example())
		}

		// Verify each send was attempted
		count := atomic.LoadInt32(&requestCount)
		assert.Equal(t, int32(5), count, "Each failure should trigger a request")
	})
}

func TestWebhookTransientFailureRecovery(t *testing.T) {
	setupTestEnvironment(t)

	var requestCount int32

	t.Run("server recovers after failures", func(t *testing.T) {
		atomic.StoreInt32(&requestCount, 0)

		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&requestCount, 1)
			if count <= 2 {
				// First two requests fail
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte(`{"error": "temporarily unavailable"}`))
			} else {
				// Subsequent requests succeed
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}
		}))
		defer mockServer.Close()

		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test recovery",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			Var1:        null.NewNullString("POST"),
			SuccessData: null.NewNullString(`{"online": true}`),
			FailureData: null.NewNullString(`{"online": false}`),
			Enabled:     null.NewNullBool(true),
		}}

		// First two should "fail" (return 503)
		_, err1 := testWebhook.OnSuccess(services.Example(true))
		_, err2 := testWebhook.OnSuccess(services.Example(true))
		// Third should succeed
		_, err3 := testWebhook.OnSuccess(services.Example(true))

		// HTTP errors don't propagate as Go errors since the request itself succeeded
		assert.Nil(t, err1)
		assert.Nil(t, err2)
		assert.Nil(t, err3)
		assert.Equal(t, int32(3), atomic.LoadInt32(&requestCount))
	})
}

// =============================================================================
// 4. Rate Limiting (429) Response Tests
// =============================================================================

func TestWebhookRateLimitResponse(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("429 Too Many Requests response", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", "60")
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", "1234567890")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error": "rate_limited", "retry_after": 60}`))
		}))
		defer mockServer.Close()

		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test rate limit",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			Var1:        null.NewNullString("POST"),
			SuccessData: null.NewNullString(`{"online": true}`),
			FailureData: null.NewNullString(`{"online": false}`),
			Enabled:     null.NewNullBool(true),
		}}

		// Webhook doesn't have special 429 handling - request completes
		err := testWebhook.Send(`{"test": "rate limit"}`)
		assert.Nil(t, err, "Webhook completes even with 429 response")
	})

	t.Run("429 with JSON error body", func(t *testing.T) {
		var receivedBody string
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			receivedBody = string(body)
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{
				"error": {
					"code": "rate_limited",
					"message": "Too many requests. Please retry after 60 seconds.",
					"retry_after_ms": 60000
				}
			}`))
		}))
		defer mockServer.Close()

		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			Var1:        null.NewNullString("POST"),
			SuccessData: null.NewNullString(`{"online": true}`),
			FailureData: null.NewNullString(`{"online": false}`),
			Enabled:     null.NewNullBool(true),
		}}

		_, err := testWebhook.OnFailure(services.Example(false), failures.Example())
		assert.Nil(t, err)
		assert.Contains(t, receivedBody, "online")
	})
}

func TestDiscordRateLimitResponse(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("discord 429 rate limit", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset-After", "1.5")
			w.Header().Set("Retry-After", "1500")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message": "You are being rate limited.", "retry_after": 1.5, "global": false}`))
		}))
		defer mockServer.Close()

		testDiscord := &discord{&notifications.Notification{
			Method:      "discord",
			Title:       "Discord",
			Description: "Test rate limit",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			SuccessData: null.NewNullString(`{"content":"Service online"}`),
			FailureData: null.NewNullString(`{"content":"Service offline"}`),
			Enabled:     null.NewNullBool(true),
		}}

		// Discord uses utils.HttpRequest which doesn't fail on non-2xx
		_, err := testDiscord.OnFailure(services.Example(false), failures.Example())
		assert.Nil(t, err)
	})
}

// =============================================================================
// 5. Empty/Null Field Handling Tests
// =============================================================================

func TestWebhookEmptyNullFields(t *testing.T) {
	setupTestEnvironment(t)

	var receivedBody string
	var receivedHeaders http.Header
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	t.Run("empty Host field", func(t *testing.T) {
		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString(""),
			Var1:        null.NewNullString("POST"),
			SuccessData: null.NewNullString(`{"online": true}`),
			Enabled:     null.NewNullBool(true),
		}}

		err := testWebhook.Send(`{"test": "empty host"}`)
		assert.NotNil(t, err, "Should fail with empty host")
	})

	t.Run("null SuccessData field", func(t *testing.T) {
		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			Var1:        null.NewNullString("POST"),
			SuccessData: null.NullString{}, // Null/empty
			FailureData: null.NewNullString(`{"online": false}`),
			Enabled:     null.NewNullBool(true),
		}}

		_, err := testWebhook.OnSuccess(services.Example(true))
		assert.Nil(t, err, "Should handle null SuccessData")
		// Empty template results in empty body
		assert.Equal(t, "", receivedBody)
	})

	t.Run("null FailureData field", func(t *testing.T) {
		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			Var1:        null.NewNullString("POST"),
			SuccessData: null.NewNullString(`{"online": true}`),
			FailureData: null.NullString{}, // Null/empty
			Enabled:     null.NewNullBool(true),
		}}

		_, err := testWebhook.OnFailure(services.Example(false), failures.Example())
		assert.Nil(t, err, "Should handle null FailureData")
	})

	t.Run("empty ApiKey (Content-Type) uses default", func(t *testing.T) {
		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			Var1:        null.NewNullString("POST"),
			ApiKey:      null.NullString{}, // Empty
			SuccessData: null.NewNullString(`{"online": true}`),
			FailureData: null.NewNullString(`{"online": false}`),
			Enabled:     null.NewNullBool(true),
		}}

		_ = testWebhook.Send(`{"test": "default content-type"}`)
		assert.Equal(t, "application/json", receivedHeaders.Get("Content-Type"))
	})

	t.Run("empty custom headers (ApiSecret)", func(t *testing.T) {
		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			Var1:        null.NewNullString("POST"),
			ApiSecret:   null.NullString{}, // Empty
			SuccessData: null.NewNullString(`{"online": true}`),
			Enabled:     null.NewNullBool(true),
		}}

		err := testWebhook.Send(`{"test": "no custom headers"}`)
		assert.Nil(t, err)
	})

	t.Run("malformed custom headers", func(t *testing.T) {
		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			Var1:        null.NewNullString("POST"),
			ApiSecret:   null.NewNullString("invalidheaderwithoutequals,=,key=,=value"), // Malformed
			SuccessData: null.NewNullString(`{"online": true}`),
			Enabled:     null.NewNullBool(true),
		}}

		// Should not panic with malformed headers
		err := testWebhook.Send(`{"test": "malformed headers"}`)
		assert.Nil(t, err, "Should handle malformed headers gracefully")
	})
}

func TestSlackEmptyNullFields(t *testing.T) {
	setupTestEnvironment(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer mockServer.Close()

	t.Run("slack with empty SuccessData", func(t *testing.T) {
		testSlack := &slack{&notifications.Notification{
			Method:      slackMethod,
			Title:       "Slack",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			SuccessData: null.NullString{}, // Empty
			FailureData: null.NewNullString(`{"text":"offline"}`),
			Enabled:     null.NewNullBool(true),
		}}

		_, err := testSlack.OnSuccess(services.Example(true))
		assert.Nil(t, err, "Should handle empty SuccessData")
	})

	t.Run("slack with whitespace-only data", func(t *testing.T) {
		testSlack := &slack{&notifications.Notification{
			Method:      slackMethod,
			Title:       "Slack",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			SuccessData: null.NewNullString("   \n\t   "),
			FailureData: null.NewNullString(`{"text":"offline"}`),
			Enabled:     null.NewNullBool(true),
		}}

		_, err := testSlack.OnSuccess(services.Example(true))
		assert.Nil(t, err, "Should handle whitespace-only SuccessData")
	})
}

// =============================================================================
// 6. Template Rendering Error Tests
// =============================================================================

func TestTemplateRenderingErrors(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("invalid template syntax - unclosed braces", func(t *testing.T) {
		result := ReplaceTemplate(`{"id": {{.Service.Id}`, replacer{Service: services.Example(true)})
		assert.Contains(t, result, "template")
	})

	t.Run("invalid template syntax - missing closing brace", func(t *testing.T) {
		result := ReplaceTemplate(`{"name": "{{.Service.Name}"`, replacer{Service: services.Example(true)})
		assert.Contains(t, result, "template")
	})

	t.Run("non-existent field access", func(t *testing.T) {
		result := ReplaceTemplate(`{"field": "{{.Service.NonExistentField}}"}`, replacer{Service: services.Example(true)})
		assert.Contains(t, result, "template")
	})

	t.Run("deeply nested non-existent field", func(t *testing.T) {
		result := ReplaceTemplate(`{"field": "{{.Service.Nested.Deep.Field}}"}`, replacer{Service: services.Example(true)})
		assert.Contains(t, result, "template")
	})

	t.Run("nil Failure field access", func(t *testing.T) {
		result := ReplaceTemplate(`{"issue": "{{.Failure.Issue}}"}`, replacer{Service: services.Example(true)})
		// Failure is zero-valued, so Issue will be empty string
		assert.Contains(t, result, `"issue": ""`)
	})

	t.Run("valid template with special characters in data", func(t *testing.T) {
		svc := services.Example(true)
		svc.Name = `Service "with" <special> & chars`
		result := ReplaceTemplate(`{"name": "{{.Service.Name}}"}`, replacer{Service: svc})
		// html/template escapes special characters for safety
		assert.Contains(t, result, "&lt;special&gt;")
		assert.Contains(t, result, "&amp;")
	})

	t.Run("template with function that doesn't exist", func(t *testing.T) {
		result := ReplaceTemplate(`{"id": "{{nonExistentFunc .Service.Id}}"}`, replacer{Service: services.Example(true)})
		assert.Contains(t, result, "template")
	})

	t.Run("empty template string", func(t *testing.T) {
		result := ReplaceTemplate("", replacer{Service: services.Example(true)})
		assert.Equal(t, "", result)
	})

	t.Run("template with only whitespace", func(t *testing.T) {
		result := ReplaceTemplate("   \n\t   ", replacer{Service: services.Example(true)})
		assert.Equal(t, "   \n\t   ", result)
	})
}

func TestWebhookTemplateRenderingEdgeCases(t *testing.T) {
	setupTestEnvironment(t)

	var receivedBody string
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	t.Run("webhook with invalid SuccessData template", func(t *testing.T) {
		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			Var1:        null.NewNullString("POST"),
			SuccessData: null.NewNullString(`{"id": {{.Service.Id}`), // Invalid
			FailureData: null.NewNullString(`{"online": false}`),
			Enabled:     null.NewNullBool(true),
		}}

		_, err := testWebhook.OnSuccess(services.Example(true))
		assert.Nil(t, err) // Send doesn't fail, but template error goes in body
		assert.Contains(t, receivedBody, "template")
	})

	t.Run("webhook with non-existent field in template", func(t *testing.T) {
		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			Var1:        null.NewNullString("POST"),
			SuccessData: null.NewNullString(`{"value": "{{.Service.FakeField}}"}`),
			FailureData: null.NewNullString(`{"online": false}`),
			Enabled:     null.NewNullBool(true),
		}}

		_, err := testWebhook.OnSuccess(services.Example(true))
		assert.Nil(t, err)
		assert.Contains(t, receivedBody, "template")
	})

	t.Run("webhook template with newlines and special chars", func(t *testing.T) {
		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			Var1:        null.NewNullString("POST"),
			SuccessData: null.NewNullString("{\n\t\"name\": \"{{.Service.Name}}\",\n\t\"id\": {{.Service.Id}}\n}"),
			Enabled:     null.NewNullBool(true),
		}}

		_, err := testWebhook.OnSuccess(services.Example(true))
		assert.Nil(t, err)
		assert.Contains(t, receivedBody, "Statping Example")
	})
}

// =============================================================================
// Additional Edge Cases
// =============================================================================

func TestWebhookConnectionRefused(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("connection refused error", func(t *testing.T) {
		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString("http://127.0.0.1:59999"), // Unlikely to be in use
			Var1:        null.NewNullString("POST"),
			SuccessData: null.NewNullString(`{"online": true}`),
			Enabled:     null.NewNullBool(true),
		}}

		err := testWebhook.Send(`{"test": "connection refused"}`)
		assert.NotNil(t, err, "Should fail when connection is refused")
	})
}

func TestWebhookInvalidURL(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("malformed URL", func(t *testing.T) {
		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString("not-a-valid-url"),
			Var1:        null.NewNullString("POST"),
			SuccessData: null.NewNullString(`{"online": true}`),
			Enabled:     null.NewNullBool(true),
		}}

		err := testWebhook.Send(`{"test": "invalid url"}`)
		assert.NotNil(t, err, "Should fail with invalid URL")
	})

	t.Run("URL with unsupported protocol", func(t *testing.T) {
		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString("ftp://example.com/webhook"),
			Var1:        null.NewNullString("POST"),
			SuccessData: null.NewNullString(`{"online": true}`),
			Enabled:     null.NewNullBool(true),
		}}

		err := testWebhook.Send(`{"test": "ftp url"}`)
		assert.NotNil(t, err, "Should fail with unsupported protocol")
	})
}

func TestWebhookLargePayload(t *testing.T) {
	setupTestEnvironment(t)

	var receivedBodyLen int
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBodyLen = len(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	t.Run("large payload handling", func(t *testing.T) {
		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			Var1:        null.NewNullString("POST"),
			SuccessData: null.NewNullString(`{"online": true}`),
			Enabled:     null.NewNullBool(true),
		}}

		// Create a large payload (100KB)
		largePayload := `{"data": "` + strings.Repeat("x", 100*1024) + `"}`
		err := testWebhook.Send(largePayload)
		assert.Nil(t, err, "Should handle large payloads")
		assert.True(t, receivedBodyLen > 100*1024, "Server should receive full payload")
	})
}

func TestWebhookHTTPMethods(t *testing.T) {
	setupTestEnvironment(t)

	var receivedMethod string
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	for _, method := range methods {
		t.Run("HTTP method "+method, func(t *testing.T) {
			testWebhook := &webhooker{&notifications.Notification{
				Method:      "webhook",
				Title:       "Webhook",
				Description: "Test",
				Author:      "Test",
				Host:        null.NewNullString(mockServer.URL),
				Var1:        null.NewNullString(method),
				SuccessData: null.NewNullString(`{"online": true}`),
				Enabled:     null.NewNullBool(true),
			}}

			_ = testWebhook.Send(`{"test": "method"}`)
			assert.Equal(t, method, receivedMethod)
		})
	}
}

func TestWebhookHostHeaderOverride(t *testing.T) {
	setupTestEnvironment(t)

	var receivedHost string
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHost = r.Host
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	t.Run("custom Host header", func(t *testing.T) {
		testWebhook := &webhooker{&notifications.Notification{
			Method:      "webhook",
			Title:       "Webhook",
			Description: "Test",
			Author:      "Test",
			Host:        null.NewNullString(mockServer.URL),
			Var1:        null.NewNullString("POST"),
			ApiSecret:   null.NewNullString("Host=custom.example.com"),
			SuccessData: null.NewNullString(`{"online": true}`),
			Enabled:     null.NewNullBool(true),
		}}

		_ = testWebhook.Send(`{"test": "host override"}`)
		assert.Equal(t, "custom.example.com", receivedHost)
	})
}

func TestReplaceVarsEdgeCases(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("ReplaceVars with nil service fields", func(t *testing.T) {
		svc := &services.Service{} // Zero-valued service
		fail := failures.Failure{}
		result := ReplaceVars("Service: {{.Service.Name}}, Issue: {{.Failure.Issue}}", svc, fail)
		assert.Contains(t, result, "Service: ")
		assert.Contains(t, result, "Issue: ")
	})

	t.Run("ReplaceVars with Core fields", func(t *testing.T) {
		svc := services.Example(true)
		fail := failures.Example()
		result := ReplaceVars("Domain: {{.Core.Domain}}", svc, fail)
		// Core.Domain comes from core.App which is set in Example()
		assert.NotContains(t, result, "template")
	})
}
