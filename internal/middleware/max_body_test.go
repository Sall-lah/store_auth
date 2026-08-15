package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMaxBodyMiddleware(t *testing.T) {
	// 100 bytes limit for test
	limit := int64(100)
	handler := MaxBody(limit)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("payload within limit passes", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(make([]byte, 50)))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d; got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("payload exceeding limit fails", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(make([]byte, 150)))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("expected status %d; got %d", http.StatusRequestEntityTooLarge, rec.Code)
		}
	})
}
