package redis

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// NewClient initializes a Redis client connected to the given URL string and validates connectivity via ping.
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
