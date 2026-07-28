package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError(t *testing.T) {
	t.Run("Error returns error string", func(t *testing.T) {
		e := &appError{Err: "test error", Code: 500}
		assert.Equal(t, "test error", e.Error())
	})

	t.Run("Status returns code", func(t *testing.T) {
		e := appError{Err: "test", Code: 404}
		assert.Equal(t, 404, e.Status())
	})

	t.Run("Status returns 200 when code is 0", func(t *testing.T) {
		e := appError{Err: "test", Code: 0}
		assert.Equal(t, 200, e.Status())
	})
}

func TestNew(t *testing.T) {
	t.Run("creates error with message", func(t *testing.T) {
		err := New("new error")
		assert.Equal(t, "new error", err.Error())
		assert.Equal(t, 200, err.Status())
	})
}

func TestErr(t *testing.T) {
	t.Run("wraps existing error", func(t *testing.T) {
		original := &appError{Err: "original", Code: 500}
		wrapped := Err(original)
		assert.Equal(t, "original", wrapped.Error())
		assert.Equal(t, 500, wrapped.Status())
	})
}

func TestWrap(t *testing.T) {
	t.Run("wraps error with message", func(t *testing.T) {
		original := errors.New("base error")
		wrapped := Wrap(original, "context")
		assert.Contains(t, wrapped.Error(), "base error")
		assert.Contains(t, wrapped.Error(), "context")
	})
}

func TestMissing(t *testing.T) {
	t.Run("creates 404 error for missing object", func(t *testing.T) {
		type Service struct{}
		err := Missing(Service{}, 123)
		assert.Contains(t, err.Error(), "service")
		assert.Contains(t, err.Error(), "123")
		assert.Contains(t, err.Error(), "not found")

		appErr, ok := err.(*appError)
		assert.True(t, ok)
		assert.Equal(t, 404, appErr.Code)
	})
}

func TestSplitVar(t *testing.T) {
	t.Run("extracts type name", func(t *testing.T) {
		type MyStruct struct{}
		result := splitVar(MyStruct{})
		assert.Equal(t, "mystruct", result)
	})

	t.Run("handles pointer types", func(t *testing.T) {
		type MyStruct struct{}
		result := splitVar(&MyStruct{})
		assert.Equal(t, "mystruct", result)
	})
}

func TestPredefinedErrors(t *testing.T) {
	t.Run("NotAuthenticated has correct code", func(t *testing.T) {
		assert.Equal(t, 401, NotAuthenticated.Code)
		assert.Equal(t, "user not authenticated", NotAuthenticated.Err)
	})

	t.Run("DecodeJSON has correct code", func(t *testing.T) {
		assert.Equal(t, 422, DecodeJSON.Code)
	})

	t.Run("IDMissing has correct code", func(t *testing.T) {
		assert.Equal(t, 422, IDMissing.Code)
	})

	t.Run("NotNumber has correct code", func(t *testing.T) {
		assert.Equal(t, 422, NotNumber.Code)
	})

	t.Run("ServiceNameMissing has correct code", func(t *testing.T) {
		assert.Equal(t, 422, ServiceNameMissing.Code)
	})
}
