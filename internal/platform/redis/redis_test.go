package redis

import (
	"context"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

func TestRedisBlacklist(t *testing.T) {
	ctx := context.Background()

	t.Run("nil client returns graceful no-op on BlacklistUser", func(t *testing.T) {
		err := BlacklistUser(ctx, nil, "user-123", 900*time.Second, "ban")
		if err != nil {
			t.Errorf("expected nil error on nil client; got %v", err)
		}
	})

	t.Run("nil client returns false on IsUserBlacklisted", func(t *testing.T) {
		blacklisted := IsUserBlacklisted(ctx, nil, "user-123")
		if blacklisted {
			t.Errorf("expected false for nil client; got %v", blacklisted)
		}
	})

	t.Run("unreachable client fails open on IsUserBlacklisted", func(t *testing.T) {
		offlineClient := goredis.NewClient(&goredis.Options{
			Addr: "127.0.0.1:59998",
		})
		blacklisted := IsUserBlacklisted(ctx, offlineClient, "user-123")
		if blacklisted {
			t.Errorf("expected false (fail-open) for offline client; got %v", blacklisted)
		}
	})
}
