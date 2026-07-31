package notifiers

import (
	"strings"
	"testing"
	"time"

	"github.com/statping-ng/statping-ng/types/notifier"
	"github.com/stretchr/testify/assert"
)

func TestParseEmails(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single email",
			input:    "test@example.com",
			expected: []string{"test@example.com"},
		},
		{
			name:     "multiple emails comma separated",
			input:    "a@test.com, b@test.com, c@test.com",
			expected: []string{"a@test.com", "b@test.com", "c@test.com"},
		},
		{
			name:     "emails with extra whitespace",
			input:    "  admin@test.com  ,  alerts@test.com  ",
			expected: []string{"admin@test.com", "alerts@test.com"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "only whitespace and commas",
			input:    "  ,  ,  ",
			expected: nil,
		},
		{
			name:     "single email with trailing comma",
			input:    "test@example.com,",
			expected: []string{"test@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseEmails(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "seconds only",
			duration: 45 * time.Second,
			expected: "45s",
		},
		{
			name:     "minutes only",
			duration: 15 * time.Minute,
			expected: "15m",
		},
		{
			name:     "hours and minutes",
			duration: 2*time.Hour + 30*time.Minute,
			expected: "2h 30m",
		},
		{
			name:     "less than a second",
			duration: 500 * time.Millisecond,
			expected: "0s",
		},
		{
			name:     "exactly one hour",
			duration: time.Hour,
			expected: "1h 0m",
		},
		{
			name:     "59 minutes",
			duration: 59 * time.Minute,
			expected: "59m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "no truncation needed",
			input:    "short",
			maxLen:   10,
			expected: "short",
		},
		{
			name:     "exact length",
			input:    "exact",
			maxLen:   5,
			expected: "exact",
		},
		{
			name:     "truncation with ellipsis",
			input:    "this is a very long string that needs truncation",
			maxLen:   20,
			expected: "this is a very long ...",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderDigestEmail(t *testing.T) {
	data := digestData{
		AppName:         "Test Statping",
		Domain:          "https://status.example.com",
		GeneratedAt:     "2026-07-30 08:00 UTC",
		Period:          "2026-07-29 08:00 to 08:00",
		TotalServices:   10,
		HealthyServices: 8,
		FailedServices:  2,
		HasFailures:     true,
		HasAppErrors:    false,
		ServiceSummary: []notifier.ServiceDigest{
			{
				Name:          "API Server",
				Status:        "Offline",
				FailureCount:  15,
				TotalDowntime: "2h 30m",
				LastFailure:   "07:45",
			},
		},
	}

	html := renderDigestEmail(data)

	// Verify HTML structure
	assert.Contains(t, html, "<!DOCTYPE html>")
	assert.Contains(t, html, "Test Statping")
	assert.Contains(t, html, "API Server")
	assert.Contains(t, html, "Offline")
	assert.Contains(t, html, "15")    // failure count
	assert.Contains(t, html, "07:45") // last failure time

	// Verify stats
	assert.Contains(t, html, "10") // total services
	assert.Contains(t, html, "8")  // healthy
	assert.Contains(t, html, "2")  // failed
}

func TestRenderDigestEmailNoFailures(t *testing.T) {
	data := digestData{
		AppName:         "Healthy System",
		Domain:          "https://status.example.com",
		GeneratedAt:     "2026-07-30 08:00 UTC",
		Period:          "2026-07-29 08:00 to 08:00",
		TotalServices:   5,
		HealthyServices: 5,
		FailedServices:  0,
		HasFailures:     false,
		HasAppErrors:    false,
	}

	html := renderDigestEmail(data)

	// Should show "all healthy" message
	assert.Contains(t, html, "All services healthy")
	assert.Contains(t, html, "no issues")
}

func TestRenderDigestEmailWithAppErrors(t *testing.T) {
	data := digestData{
		AppName:         "Test System",
		Domain:          "https://status.example.com",
		GeneratedAt:     "2026-07-30 08:00 UTC",
		Period:          "2026-07-29 08:00 to 08:00",
		TotalServices:   5,
		HealthyServices: 5,
		FailedServices:  0,
		HasFailures:     false,
		HasAppErrors:    true,
		AppErrors: []notifier.AppError{
			{
				Timestamp: "07:30",
				Message:   "Database connection timeout",
			},
		},
	}

	html := renderDigestEmail(data)

	assert.Contains(t, html, "Application Errors")
	assert.Contains(t, html, "Database connection timeout")
}

func TestDigestDataStructure(t *testing.T) {
	// Verify digestData can be created and marshaled
	data := digestData{
		AppName:         "Test",
		Domain:          "https://test.com",
		GeneratedAt:     time.Now().Format("2006-01-02 15:04 MST"),
		Period:          "24h",
		TotalServices:   10,
		HealthyServices: 9,
		FailedServices:  1,
		HasFailures:     true,
		HasAppErrors:    false,
		ServiceSummary: []notifier.ServiceDigest{
			{Name: "Test", Status: "Online", FailureCount: 1},
		},
	}

	// Verify it renders without error
	html := renderDigestEmail(data)
	assert.NotEmpty(t, html)
	assert.True(t, strings.HasPrefix(html, "<!DOCTYPE html>"))
}

func TestSchedulerOnce(t *testing.T) {
	// Verify StartDigestScheduler can be called multiple times safely
	// (sync.Once pattern)
	StopDigestScheduler()

	// These should not panic
	StartDigestScheduler()
	StartDigestScheduler()
	StartDigestScheduler()

	StopDigestScheduler()
}

func TestTestSMTPConnection(t *testing.T) {
	// Should return a result even without SMTP configured
	result := TestSMTPConnection()

	assert.NotNil(t, result)
	// Result is a struct - Host might be empty if not configured, that's OK
	// Just verify we got a valid result
	assert.IsType(t, &SMTPDiagResult{}, result)
}
