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

	t.Run("defaults for JWT access, refresh expiry, and Kafka settings", func(t *testing.T) {
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
		if cfg.KafkaBrokers != "localhost:9092" {
			t.Errorf("expected default KafkaBrokers 'localhost:9092'; got %s", cfg.KafkaBrokers)
		}
		if cfg.KafkaTopicAuthEvents != "auth.events" {
			t.Errorf("expected default KafkaTopicAuthEvents 'auth.events'; got %s", cfg.KafkaTopicAuthEvents)
		}
		if cfg.OTPProvider != "kafka" {
			t.Errorf("expected default OTPProvider 'kafka'; got %s", cfg.OTPProvider)
		}
	})

	t.Run("loads custom REDIS_PASSWORD and custom KAFKA settings when provided", func(t *testing.T) {
		_ = os.Setenv("DATABASE_URL", "postgresql://localhost:5432/test")
		_ = os.Setenv("REDIS_PASSWORD", "custom_redis_pass")
		_ = os.Setenv("KAFKA_BROKERS", "broker1:9092,broker2:9092")
		_ = os.Setenv("KAFKA_TOPIC_AUTH_EVENTS", "custom.auth.events")
		_ = os.Setenv("OTP_PROVIDER", "mock")
		defer os.Unsetenv("DATABASE_URL")
		defer os.Unsetenv("REDIS_PASSWORD")
		defer os.Unsetenv("KAFKA_BROKERS")
		defer os.Unsetenv("KAFKA_TOPIC_AUTH_EVENTS")
		defer os.Unsetenv("OTP_PROVIDER")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if cfg.RedisPassword != "custom_redis_pass" {
			t.Errorf("expected RedisPassword 'custom_redis_pass'; got %s", cfg.RedisPassword)
		}
		if cfg.KafkaBrokers != "broker1:9092,broker2:9092" {
			t.Errorf("expected KafkaBrokers 'broker1:9092,broker2:9092'; got %s", cfg.KafkaBrokers)
		}
		if cfg.KafkaTopicAuthEvents != "custom.auth.events" {
			t.Errorf("expected KafkaTopicAuthEvents 'custom.auth.events'; got %s", cfg.KafkaTopicAuthEvents)
		}
		if cfg.OTPProvider != "mock" {
			t.Errorf("expected OTPProvider 'mock'; got %s", cfg.OTPProvider)
		}
	})
}
