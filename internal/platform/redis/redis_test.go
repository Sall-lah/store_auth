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

// TestNewClient validates that NewClient correctly parses credentials, addresses, and database indices from URLs.
func TestNewClient(t *testing.T) {
	t.Run("parses valid URL without password", func(t *testing.T) {
		client, err := NewClient("redis://localhost:6379/0")
		if err != nil {
			t.Fatalf("expected nil error; got %v", err)
		}
		if client.Options().Addr != "localhost:6379" {
			t.Errorf("expected addr 'localhost:6379'; got %s", client.Options().Addr)
		}
		if client.Options().Password != "" {
			t.Errorf("expected empty password; got %s", client.Options().Password)
		}
	})

	t.Run("parses URL with password", func(t *testing.T) {
		client, err := NewClient("redis://:mysecretpassword@localhost:6379/1")
		if err != nil {
			t.Fatalf("expected nil error; got %v", err)
		}
		if client.Options().Password != "mysecretpassword" {
			t.Errorf("expected password 'mysecretpassword'; got %s", client.Options().Password)
		}
		if client.Options().DB != 1 {
			t.Errorf("expected DB 1; got %d", client.Options().DB)
		}
	})

	t.Run("parses URL with ACL username and password", func(t *testing.T) {
		client, err := NewClient("redis://customuser:mysecretpassword@localhost:6379/2")
		if err != nil {
			t.Fatalf("expected nil error; got %v", err)
		}
		if client.Options().Username != "customuser" {
			t.Errorf("expected username 'customuser'; got %s", client.Options().Username)
		}
		if client.Options().Password != "mysecretpassword" {
			t.Errorf("expected password 'mysecretpassword'; got %s", client.Options().Password)
		}
		if client.Options().DB != 2 {
			t.Errorf("expected DB 2; got %d", client.Options().DB)
		}
	})

	t.Run("overrides URL password with explicit password parameter", func(t *testing.T) {
		client, err := NewClient("redis://localhost:6379/0", "override_password")
		if err != nil {
			t.Fatalf("expected nil error; got %v", err)
		}
		if client.Options().Password != "override_password" {
			t.Errorf("expected password 'override_password'; got %s", client.Options().Password)
		}
	})

	t.Run("returns error on malformed URL", func(t *testing.T) {
		_, err := NewClient("invalid-url://bad")
		if err == nil {
			t.Error("expected error for invalid redis url; got nil")
		}
	})
}

