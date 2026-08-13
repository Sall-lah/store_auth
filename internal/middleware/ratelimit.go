package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter holds Redis client and rate limiting parameters to throttle public endpoints.
type RateLimiter struct {
	rdb         *redis.Client
	maxRequests int
	window      time.Duration
}

// NewRateLimiter constructs a RateLimiter middleware instance.
func NewRateLimiter(rdb *redis.Client, maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		rdb:         rdb,
		maxRequests: maxRequests,
		window:      window,
	}
}

// RateLimit returns a HTTP middleware handler that throttles requests based on client IP and endpoint path using Redis.
func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		key := fmt.Sprintf("ratelimit:%s:%s", clientIP, r.URL.Path)

		ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
		defer cancel()

		count, err := rl.rdb.Incr(ctx, key).Result()
		if err != nil {
			log.Printf("[RATE LIMITER WARNING] Redis increment failed: %v. Allowing request through.", err)
			next.ServeHTTP(w, r)
			return
		}

		if count == 1 {
			_ = rl.rdb.Expire(ctx, key, rl.window)
		}

		if count > int64(rl.maxRequests) {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(rl.window.Seconds())))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(errorResponse{
				Error: "Too many requests. Please try again later.",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	parts := strings.Split(r.RemoteAddr, ":")
	return parts[0]
}
