package docs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDocsHandler(t *testing.T) {
	handler := NewHandler()

	t.Run("ServeUI returns Swagger UI HTML", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/docs", nil)
		rec := httptest.NewRecorder()

		handler.ServeUI(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200; got %d", rec.Code)
		}
		if contentType := rec.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
			t.Fatalf("expected text/html charset=utf-8; got %s", contentType)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "SwaggerUIBundle") {
			t.Fatalf("expected body to contain SwaggerUIBundle initialization")
		}
	})

	t.Run("ServeSpecYAML returns valid OpenAPI 3.1 YAML", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/docs/openapi.yaml", nil)
		rec := httptest.NewRecorder()

		handler.ServeSpecYAML(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200; got %d", rec.Code)
		}
		if contentType := rec.Header().Get("Content-Type"); contentType != "application/yaml; charset=utf-8" {
			t.Fatalf("expected application/yaml charset=utf-8; got %s", contentType)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "openapi: 3.1.0") {
			t.Fatalf("expected body to contain 'openapi: 3.1.0'")
		}
		if !strings.Contains(body, "/.well-known/jwks.json") {
			t.Fatalf("expected body to document '/.well-known/jwks.json'")
		}
	})
}
