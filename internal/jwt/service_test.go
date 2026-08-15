package jwt

import (
	"path/filepath"
	"testing"
	"time"
)

func TestJWTService(t *testing.T) {
	privPath := filepath.Join("..", "..", "keys", "private.pem")
	pubPath := filepath.Join("..", "..", "keys", "public.pem")

	svc, err := NewService(privPath, pubPath, 15)
	if err != nil {
		t.Fatalf("failed to create JWT service: %v", err)
	}

	t.Run("Generate and validate 15-minute token", func(t *testing.T) {
		tokenStr, err := svc.GenerateToken("user-123", "test@example.com", "CUSTOMER")
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		claims, err := svc.ValidateToken(tokenStr)
		if err != nil {
			t.Fatalf("failed to validate token: %v", err)
		}

		if claims.Subject != "user-123" {
			t.Errorf("expected subject user-123; got %s", claims.Subject)
		}
		if claims.Email != "test@example.com" {
			t.Errorf("expected email test@example.com; got %s", claims.Email)
		}
		if claims.Role != "CUSTOMER" {
			t.Errorf("expected role CUSTOMER; got %s", claims.Role)
		}

		expDuration := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time)
		if expDuration != 15*time.Minute {
			t.Errorf("expected 15m expiration window; got %v", expDuration)
		}
	})

	t.Run("GetJWKS exports public key set", func(t *testing.T) {
		jwks := svc.GetJWKS()
		if len(jwks.Keys) == 0 {
			t.Fatal("expected at least 1 key in JWKS")
		}
		if jwks.Keys[0].Alg != "RS256" {
			t.Errorf("expected alg RS256; got %s", jwks.Keys[0].Alg)
		}
		if jwks.Keys[0].Use != "sig" {
			t.Errorf("expected use sig; got %s", jwks.Keys[0].Use)
		}
	})
}
