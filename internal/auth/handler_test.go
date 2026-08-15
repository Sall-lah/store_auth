package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_PayloadSizeLimit(t *testing.T) {
	h := NewHandler(nil, false)

	// Create payload larger than DefaultMaxBodyBytes (1 MB)
	oversizedBody := strings.Repeat("A", int(middlewareDefaultMaxBodyBytesForTest()+1024))
	largePayload, _ := json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": "password123",
		"name":     oversizedBody,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(largePayload))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d; got %d (body: %s)", http.StatusRequestEntityTooLarge, rec.Code, rec.Body.String())
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Error != "Request payload exceeds maximum allowed size" {
		t.Errorf("expected error message %q; got %q", "Request payload exceeds maximum allowed size", errResp.Error)
	}
}

func TestHandler_Validation(t *testing.T) {
	h := NewHandler(nil, false)

	t.Run("invalid email and short password", func(t *testing.T) {
		payload, _ := json.Marshal(RegisterRequest{
			Email:    "invalid-email",
			Password: "short",
			Name:     "   ",
		})

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(payload))
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d; got %d", http.StatusBadRequest, rec.Code)
		}

		var errResp ErrorResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &errResp)

		if errResp.Details["email"] != "Invalid email format" {
			t.Errorf("expected email validation error; got %v", errResp.Details["email"])
		}
		if errResp.Details["password"] != "Password must be at least 8 characters long" {
			t.Errorf("expected password validation error; got %v", errResp.Details["password"])
		}
		if errResp.Details["name"] != "Name is required" {
			t.Errorf("expected name validation error; got %v", errResp.Details["name"])
		}
	})

	t.Run("invalid json body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte("{invalid-json")))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d; got %d", http.StatusBadRequest, rec.Code)
		}
	})
}

func middlewareDefaultMaxBodyBytesForTest() int64 {
	return 1 << 20
}
