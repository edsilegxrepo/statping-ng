package types

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	t.Run("Error method returns error string", func(t *testing.T) {
		e := returnErr("test error")
		assert.Equal(t, "test error", e.Error())
	})

	t.Run("String method returns error string", func(t *testing.T) {
		e := returnErr("test error")
		assert.Equal(t, "test error", e.String())
	})
}

func TestReturnErrCode(t *testing.T) {
	t.Run("returns error with code", func(t *testing.T) {
		err := returnErrCode("not found", http.StatusNotFound)
		assert.NotNil(t, err)
		assert.Equal(t, "not found", err.Error())

		e, ok := err.(Error)
		assert.True(t, ok)
		assert.Equal(t, http.StatusNotFound, e.code)
	})
}

func TestConvertError(t *testing.T) {
	t.Run("converts Error pointer to string", func(t *testing.T) {
		e := returnErr("pointer error")
		result := convertError(&e)
		assert.Equal(t, "pointer error", result)
	})

	t.Run("returns string as-is", func(t *testing.T) {
		result := convertError("plain string")
		assert.Equal(t, "plain string", result)
	})

	t.Run("returns empty for unknown types", func(t *testing.T) {
		result := convertError(123)
		assert.Equal(t, "", result)
	})
}

type testErrorer struct {
	msg string
}

func (e testErrorer) Error() string {
	return e.msg
}

func TestErrWrap(t *testing.T) {
	t.Run("wraps error with format string", func(t *testing.T) {
		baseErr := testErrorer{msg: "base error"}
		wrapped := ErrWrap(baseErr, "wrapped: %s", "details")
		assert.Contains(t, wrapped.Error(), "base error")
		assert.Contains(t, wrapped.Error(), "wrapped: details")
	})

	t.Run("wraps error with Error pointer format", func(t *testing.T) {
		baseErr := testErrorer{msg: "base error"}
		formatErr := returnErr("format error")
		wrapped := ErrWrap(baseErr, &formatErr)
		assert.Contains(t, wrapped.Error(), "base error")
	})
}

func TestErr(t *testing.T) {
	t.Run("wraps error with message", func(t *testing.T) {
		baseErr := testErrorer{msg: "original"}
		wrapped := Err(baseErr, "context")
		assert.Contains(t, wrapped.Error(), "original")
		assert.Contains(t, wrapped.Error(), "context")
	})
}

func TestPredefinedErrors(t *testing.T) {
	t.Run("ErrorNotFound has correct code", func(t *testing.T) {
		e, ok := ErrorNotFound.(Error)
		assert.True(t, ok)
		assert.Equal(t, http.StatusNotFound, e.code)
	})

	t.Run("ErrorJSONParse has correct code", func(t *testing.T) {
		e, ok := ErrorJSONParse.(Error)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, e.code)
	})

	t.Run("ErrorCreateService has correct message", func(t *testing.T) {
		assert.Contains(t, ErrorCreateService.Error(), "creating service")
	})
}

func TestErrorInterface(t *testing.T) {
	t.Run("Error implements error interface", func(t *testing.T) {
		var err error = returnErr("test")
		assert.NotNil(t, err)
		assert.Equal(t, "test", err.Error())
	})

	t.Run("Error can be used with errors.Is pattern", func(t *testing.T) {
		err := returnErr("custom error")
		var _ error = err
		assert.NotNil(t, err)
	})
}
