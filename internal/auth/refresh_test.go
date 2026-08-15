package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_AuthCookies(t *testing.T) {
	h := NewHandler(nil, true) // isProd = true

	t.Run("setAuthCookies sets access_token and refresh_token cookies with correct specifications", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.setAuthCookies(rec, "mock-access-token", "mock-refresh-token")

		cookies := rec.Result().Cookies()
		if len(cookies) != 2 {
			t.Fatalf("expected 2 cookies; got %d", len(cookies))
		}

		var accessCookie, refreshCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == "access_token" {
				accessCookie = c
			} else if c.Name == "refresh_token" {
				refreshCookie = c
			}
		}

		if accessCookie == nil {
			t.Fatal("access_token cookie missing")
		}
		if accessCookie.Value != "mock-access-token" {
			t.Errorf("expected access_token 'mock-access-token'; got %q", accessCookie.Value)
		}
		if accessCookie.Path != "/" {
			t.Errorf("expected access_token path '/'; got %q", accessCookie.Path)
		}
		if accessCookie.MaxAge != 900 {
			t.Errorf("expected access_token MaxAge 900 (15 min); got %d", accessCookie.MaxAge)
		}
		if !accessCookie.HttpOnly {
			t.Errorf("expected access_token HttpOnly=true")
		}
		if !accessCookie.Secure {
			t.Errorf("expected access_token Secure=true in production")
		}
		if accessCookie.SameSite != http.SameSiteLaxMode {
			t.Errorf("expected access_token SameSite=Lax; got %v", accessCookie.SameSite)
		}

		if refreshCookie == nil {
			t.Fatal("refresh_token cookie missing")
		}
		if refreshCookie.Value != "mock-refresh-token" {
			t.Errorf("expected refresh_token 'mock-refresh-token'; got %q", refreshCookie.Value)
		}
		if refreshCookie.Path != "/api/auth/refresh" {
			t.Errorf("expected refresh_token path '/api/auth/refresh'; got %q", refreshCookie.Path)
		}
		if refreshCookie.MaxAge != 604800 {
			t.Errorf("expected refresh_token MaxAge 604800 (7 days); got %d", refreshCookie.MaxAge)
		}
		if !refreshCookie.HttpOnly {
			t.Errorf("expected refresh_token HttpOnly=true")
		}
		if !refreshCookie.Secure {
			t.Errorf("expected refresh_token Secure=true in production")
		}
		if refreshCookie.SameSite != http.SameSiteLaxMode {
			t.Errorf("expected refresh_token SameSite=Lax; got %v", refreshCookie.SameSite)
		}
	})

	t.Run("clearAuthCookies invalidates both cookies with MaxAge=-1", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.clearAuthCookies(rec)

		cookies := rec.Result().Cookies()
		if len(cookies) != 2 {
			t.Fatalf("expected 2 cookies; got %d", len(cookies))
		}

		for _, c := range cookies {
			if c.Value != "" {
				t.Errorf("expected empty cookie value for %s; got %q", c.Name, c.Value)
			}
			if c.MaxAge != -1 {
				t.Errorf("expected MaxAge -1 for %s; got %d", c.Name, c.MaxAge)
			}
		}
	})

	t.Run("Logout handler clears cookies", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		req.AddCookie(&http.Cookie{
			Name:  "refresh_token",
			Value: "dummy-refresh-token",
			Path:  "/api/auth/refresh",
		})

		h.Logout(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200; got %d", rec.Code)
		}

		cookies := rec.Result().Cookies()
		for _, c := range cookies {
			if c.MaxAge != -1 {
				t.Errorf("expected cookie %s to have MaxAge=-1; got %d", c.Name, c.MaxAge)
			}
		}
	})

	t.Run("Refresh handler without cookie returns 401 and clears cookies", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)

		h.Refresh(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401; got %d", rec.Code)
		}

		cookies := rec.Result().Cookies()
		if len(cookies) != 2 {
			t.Fatalf("expected 2 clear cookies; got %d", len(cookies))
		}
	})
}

func TestTokenGenerationAndHashing(t *testing.T) {
	t.Run("generateRandomToken produces unique 32-byte base64url strings", func(t *testing.T) {
		t1, err := generateRandomToken(32)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}
		t2, err := generateRandomToken(32)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		if t1 == "" || t2 == "" {
			t.Fatal("generated token cannot be empty")
		}
		if t1 == t2 {
			t.Fatal("expected consecutively generated tokens to be unique")
		}
	})

	t.Run("hashToken generates consistent SHA-256 hex digest", func(t *testing.T) {
		token := "my-opaque-token-12345"
		h1 := hashToken(token)
		h2 := hashToken(token)

		if h1 != h2 {
			t.Fatalf("expected deterministic hash output; got %s != %s", h1, h2)
		}
		if len(h1) != 64 {
			t.Errorf("expected 64 hex character SHA-256 hash length; got %d", len(h1))
		}
	})
}
