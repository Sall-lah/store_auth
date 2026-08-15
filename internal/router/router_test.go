package router

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"store_auth/internal/auth"
	"store_auth/internal/jwt"
)

func TestRouterRoutes(t *testing.T) {
	privPath := filepath.Join("..", "..", "keys", "private.pem")
	pubPath := filepath.Join("..", "..", "keys", "public.pem")

	jwtSvc, err := jwt.NewService(privPath, pubPath, 15)
	if err != nil {
		t.Fatalf("failed to initialize JWT service: %v", err)
	}

	authHandler := auth.NewHandler(nil, false)
	jwksHandler := jwt.NewHandler(jwtSvc)

	r := SetupRouter(authHandler, jwksHandler, nil, jwtSvc, nil)

	t.Run("POST /api/auth/refresh route is mounted and responds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		// Without refresh cookie, it should return 401 Unauthorized
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized; got %d", rec.Code)
		}
	})

	t.Run("POST /api/auth/logout route is mounted and responds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK; got %d", rec.Code)
		}
	})

	t.Run("GET /.well-known/jwks.json is mounted", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK; got %d", rec.Code)
		}
	})

	t.Run("GET /docs serves Swagger UI HTML", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/docs", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK; got %d", rec.Code)
		}
		if contentType := rec.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
			t.Fatalf("expected text/html charset=utf-8; got %s", contentType)
		}
	})

	t.Run("GET /docs/openapi.yaml serves OpenAPI spec", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/docs/openapi.yaml", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK; got %d", rec.Code)
		}
		if contentType := rec.Header().Get("Content-Type"); contentType != "application/yaml; charset=utf-8" {
			t.Fatalf("expected application/yaml charset=utf-8; got %s", contentType)
		}
	})

	t.Run("GET /swagger serves Swagger UI HTML", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/swagger", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK; got %d", rec.Code)
		}
	})
}

