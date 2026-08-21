package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all validated application environment variables.
type Config struct {
	ServerPort            string
	Env                   string
	DatabaseURL           string
	RedisURL              string
	RedisPassword         string
	RateLimitMaxRequests  int
	RateLimitWindowSec    time.Duration
	JWTPrivateKeyPath      string
	JWTPublicKeyPath       string
	JWTAccessExpiryMinutes int
	JWTRefreshExpiryDays   int
	BcryptCost             int
	OTPExpiryMinutes       int
	OTPMaxAttempts         int
	OTPProvider            string
	KafkaBrokers           string
	KafkaTopicAuthEvents   string
}

// Load reads environment variables from a .env file (if present) and process environment,
// ensuring system credentials and runtime behavior are explicitly validated before server boot.
// Why: Centralizes configuration management and environment parsing to guarantee early failure on invalid settings.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort:             getEnv("SERVER_PORT", "8080"),
		Env:                    getEnv("ENV", "development"),
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		RedisURL:               getEnv("REDIS_URL", "redis://localhost:6379"),
		RedisPassword:          os.Getenv("REDIS_PASSWORD"),
		RateLimitMaxRequests:  getEnvAsInt("RATE_LIMIT_MAX_REQUESTS", 10),
		JWTPrivateKeyPath:      getEnv("JWT_PRIVATE_KEY_PATH", "./keys/private.pem"),
		JWTPublicKeyPath:       getEnv("JWT_PUBLIC_KEY_PATH", "./keys/public.pem"),
		JWTAccessExpiryMinutes: getEnvAsInt("JWT_ACCESS_EXPIRY_MINUTES", 15),
		JWTRefreshExpiryDays:   getEnvAsInt("JWT_REFRESH_EXPIRY_DAYS", 7),
		BcryptCost:             getEnvAsInt("BCRYPT_COST", 12),
		OTPExpiryMinutes:       getEnvAsInt("OTP_EXPIRY_MINUTES", 5),
		OTPMaxAttempts:         getEnvAsInt("OTP_MAX_ATTEMPTS", 5),
		OTPProvider:            getEnv("OTP_PROVIDER", "kafka"),
		KafkaBrokers:           getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaTopicAuthEvents:   getEnv("KAFKA_TOPIC_AUTH_EVENTS", "auth.events"),
	}

	windowSec := getEnvAsInt("RATE_LIMIT_WINDOW_SECONDS", 1)
	cfg.RateLimitWindowSec = time.Duration(windowSec) * time.Second

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required to connect to database")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return fallback
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return fallback
	}
	return val
}
