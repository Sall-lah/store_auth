package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewClient initializes a Redis client connected to the given URL string and validates connectivity via ping.
// Why: Establishes a shared Redis connection pool and logs connectivity status for caching and rate limiting.
func NewClient(redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis url: %w", err)
	}

	rdb := redis.NewClient(opts)

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Printf("[REDIS WARNING] Failed to ping Redis at startup: %v (fail-open mode active)", err)
	} else {
		log.Printf("[REDIS INFO] Connected to Redis successfully at %s", opts.Addr)
	}

	return rdb, nil
}

// BlacklistUser places a user ID into the Redis revocation blacklist with a given TTL and reason.
// Why: Enables instantaneous revocation of active access tokens without requiring database lookups on every request.
func BlacklistUser(ctx context.Context, rdb *redis.Client, userID string, ttl time.Duration, reason string) error {
	if rdb == nil {
		return nil
	}
	key := fmt.Sprintf("revoked:user:%s", userID)
	if err := rdb.SetEx(ctx, key, reason, ttl).Err(); err != nil {
		return fmt.Errorf("failed to blacklist user in redis: %w", err)
	}
	return nil
}

// IsUserBlacklisted checks whether the user's ID exists in the Redis revocation blacklist.
// Why: Provides fast O(1) token invalidation check in authentication middleware with fail-open resilience.
func IsUserBlacklisted(ctx context.Context, rdb *redis.Client, userID string) bool {
	if rdb == nil {
		return false
	}
	key := fmt.Sprintf("revoked:user:%s", userID)
	exists, err := rdb.Exists(ctx, key).Result()
	if err != nil {
		log.Printf("[REDIS WARNING] Failed to check blacklist for user %s: %v (fail-open)", userID, err)
		return false
	}
	return exists > 0
}

