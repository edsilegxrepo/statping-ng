package handlers

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello, World!"))
	})

	t.Run("With gzip Accept-Encoding", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rr := httptest.NewRecorder()

		Gzip(handler).ServeHTTP(rr, req)

		assert.Equal(t, "gzip", rr.Header().Get("Content-Encoding"))

		reader, err := gzip.NewReader(rr.Body)
		require.Nil(t, err)
		defer func() { _ = reader.Close() }()

		body, err := io.ReadAll(reader)
		require.Nil(t, err)
		assert.Equal(t, "Hello, World!", string(body))
	})

	t.Run("Without gzip Accept-Encoding", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		Gzip(handler).ServeHTTP(rr, req)

		assert.Empty(t, rr.Header().Get("Content-Encoding"))
		assert.Equal(t, "Hello, World!", rr.Body.String())
	})
}

func TestGzipResponseWriter(t *testing.T) {
	var buf strings.Builder
	gz := gzip.NewWriter(&buf)
	rr := httptest.NewRecorder()

	gzw := gzipResponseWriter{Writer: gz, ResponseWriter: rr}
	n, err := gzw.Write([]byte("test data"))

	assert.Nil(t, err)
	assert.Equal(t, 9, n)
	_ = gz.Close()
}
