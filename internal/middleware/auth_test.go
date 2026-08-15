package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/redis/go-redis/v9"
	"store_auth/internal/jwt"
)

func TestAuthenticateMiddleware(t *testing.T) {
	privPath := filepath.Join("..", "..", "keys", "private.pem")
	pubPath := filepath.Join("..", "..", "keys", "public.pem")

	jwtSvc, err := jwt.NewService(privPath, pubPath, 15)
	if err != nil {
		t.Fatalf("failed to initialize JWT service: %v", err)
	}

	validToken, err := jwtSvc.GenerateToken("user-active-1", "user@example.com", "CUSTOMER")
	if err != nil {
		t.Fatalf("failed to generate valid token: %v", err)
	}

	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetUserClaimsFromContext(r.Context())
		if !ok || claims == nil {
			http.Error(w, "no claims", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(claims.Subject))
	})

	t.Run("Missing access_token cookie returns 401", func(t *testing.T) {
		authMiddleware := Authenticate(jwtSvc, nil)
		handler := authMiddleware(dummyHandler)

		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d; got %d", http.StatusUnauthorized, rec.Code)
		}

		var errResp errorResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
		if errResp.Error != "Authentication cookie missing or empty" {
			t.Errorf("unexpected error message: %v", errResp.Error)
		}
	})

	t.Run("Invalid access_token cookie returns 401", func(t *testing.T) {
		authMiddleware := Authenticate(jwtSvc, nil)
		handler := authMiddleware(dummyHandler)

		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		req.AddCookie(&http.Cookie{
			Name:  "access_token",
			Value: "invalid.jwt.token",
		})
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d; got %d", http.StatusUnauthorized, rec.Code)
		}

		var errResp errorResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
		if errResp.Error != "Invalid or expired authentication token" {
			t.Errorf("unexpected error message: %v", errResp.Error)
		}
	})

	t.Run("Valid token with nil Redis succeeds", func(t *testing.T) {
		authMiddleware := Authenticate(jwtSvc, nil)
		handler := authMiddleware(dummyHandler)

		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		req.AddCookie(&http.Cookie{
			Name:  "access_token",
			Value: validToken,
		})
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d; got %d", http.StatusOK, rec.Code)
		}
		if rec.Body.String() != "user-active-1" {
			t.Errorf("expected body %q; got %q", "user-active-1", rec.Body.String())
		}
	})

	t.Run("Valid token with offline Redis fails open", func(t *testing.T) {
		offlineRdb := redis.NewClient(&redis.Options{
			Addr: "127.0.0.1:59999", // Unreachable port
		})
		authMiddleware := Authenticate(jwtSvc, offlineRdb)
		handler := authMiddleware(dummyHandler)

		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		req.AddCookie(&http.Cookie{
			Name:  "access_token",
			Value: validToken,
		})
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d on fail-open; got %d", http.StatusOK, rec.Code)
		}
	})
}
