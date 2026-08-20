package config

import (
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	t.Run("validate fails when DATABASE_URL is missing", func(t *testing.T) {
		cfg := &Config{
			DatabaseURL: "",
		}
		if err := cfg.validate(); err == nil {
			t.Error("expected error for empty DATABASE_URL; got nil")
		}
	})

	t.Run("validate succeeds when DATABASE_URL is provided", func(t *testing.T) {
		cfg := &Config{
			DatabaseURL: "postgresql://localhost:5432/test",
		}
		if err := cfg.validate(); err != nil {
			t.Errorf("expected nil error; got %v", err)
		}
	})

	t.Run("defaults for JWT access and refresh expiry", func(t *testing.T) {
		_ = os.Setenv("DATABASE_URL", "postgresql://localhost:5432/test")
		defer os.Unsetenv("DATABASE_URL")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if cfg.JWTAccessExpiryMinutes != 15 {
			t.Errorf("expected default JWTAccessExpiryMinutes 15; got %d", cfg.JWTAccessExpiryMinutes)
		}
		if cfg.JWTRefreshExpiryDays != 7 {
			t.Errorf("expected default JWTRefreshExpiryDays 7; got %d", cfg.JWTRefreshExpiryDays)
		}
	})

	t.Run("loads custom REDIS_PASSWORD when provided", func(t *testing.T) {
		_ = os.Setenv("DATABASE_URL", "postgresql://localhost:5432/test")
		_ = os.Setenv("REDIS_PASSWORD", "custom_redis_pass")
		defer os.Unsetenv("DATABASE_URL")
		defer os.Unsetenv("REDIS_PASSWORD")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if cfg.RedisPassword != "custom_redis_pass" {
			t.Errorf("expected RedisPassword 'custom_redis_pass'; got %s", cfg.RedisPassword)
		}
	})
}
