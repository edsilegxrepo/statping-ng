package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvert(t *testing.T) {
	tests := []struct {
		name     string
		input    []any
		expected []string
	}{
		{
			name:     "empty input",
			input:    []any{},
			expected: nil,
		},
		{
			name:     "single string",
			input:    []any{"test"},
			expected: []string{"test"},
		},
		{
			name:     "single int",
			input:    []any{42},
			expected: []string{"42"},
		},
		{
			name:     "mixed types",
			input:    []any{"service", 1, 3.14, true},
			expected: []string{"service", "1", "3.14", "true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convert(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHisto(t *testing.T) {
	// Test that Histo doesn't panic with valid inputs
	t.Run("duration method", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Histo("duration", 0.5, "http://example.com", "GET")
		})
	})

	t.Run("bytes method", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Histo("bytes", 1024.0, "http://example.com", "GET")
		})
	})

	t.Run("unknown method does nothing", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Histo("unknown", 100.0, "test")
		})
	})
}

func TestGauge(t *testing.T) {
	t.Run("status_code method", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Gauge("status_code", 200.0, "test-service")
		})
	})

	t.Run("online method", func(t *testing.T) {
		assert.NotPanics(t, func() {
			// serviceOnline gauge requires 2 labels: service, type
			Gauge("online", 1.0, "test-service", "http")
		})
	})

	t.Run("unknown method does nothing", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Gauge("unknown", 100.0, "test")
		})
	})
}

func TestInc(t *testing.T) {
	t.Run("failure method", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Inc("failure", "test-service")
		})
	})

	t.Run("success method", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Inc("success", "test-service")
		})
	})

	t.Run("unknown method does nothing", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Inc("unknown", "test")
		})
	})
}

func TestAdd(t *testing.T) {
	t.Run("failure method", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Add("failure", 5.0, "test-service")
		})
	})

	t.Run("success method", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Add("success", 10.0, "test-service")
		})
	})

	t.Run("unknown method does nothing", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Add("unknown", 100.0, "test")
		})
	})
}

func TestTimer(t *testing.T) {
	t.Run("returns observer", func(t *testing.T) {
		observer := Timer("/api/services")
		assert.NotNil(t, observer)
	})
}

func TestServiceTimer(t *testing.T) {
	t.Run("returns observer", func(t *testing.T) {
		observer := ServiceTimer("test-service")
		assert.NotNil(t, observer)
	})
}
