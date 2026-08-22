package user

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	kafkaGo "github.com/segmentio/kafka-go"
	"store_auth/internal/platform/kafka"
	platformRedis "store_auth/internal/platform/redis"
)

// UserRepository defines the persistence contract required for user account deactivation and deletion.
// Why: Decouples the Kafka consumer domain from specific database drivers and facilitates isolated unit testing.
type UserRepository interface {
	DeactivateUser(ctx context.Context, id string) error
	DeleteUser(ctx context.Context, id string) error
}

// RefreshTokenRepository defines the persistence contract required for bulk refresh token revocation.
// Why: Enables invalidating active user sessions without tight coupling to the auth repository implementation.
type RefreshTokenRepository interface {
	RevokeAllUserRefreshTokens(ctx context.Context, userID string) error
}

// Consumer processes asynchronous user lifecycle events to synchronize account state and revoke active sessions.
type Consumer struct {
	kafkaConsumer kafka.Consumer
	userRepo      UserRepository
	refreshRepo   RefreshTokenRepository
	rdb           *redis.Client
	blacklistTTL  time.Duration
}

// NewConsumer initializes a Consumer instance with persistence, caching, and messaging dependencies.
// Why: Injects decoupled repository and message reader dependencies for robust event-driven user lifecycle management.
func NewConsumer(
	kafkaConsumer kafka.Consumer,
	userRepo UserRepository,
	refreshRepo RefreshTokenRepository,
	rdb *redis.Client,
) *Consumer {
	return &Consumer{
		kafkaConsumer: kafkaConsumer,
		userRepo:      userRepo,
		refreshRepo:   refreshRepo,
		rdb:           rdb,
		blacklistTTL:  15 * time.Minute,
	}
}

// Start begins the event consumption loop and blocks until the provided context is cancelled or a fatal error occurs.
// Why: Continuously listens for account lifecycle events in the background while respecting graceful server shutdown signals.
func (c *Consumer) Start(ctx context.Context) {
	log.Println("[UserEventConsumer] Background worker started listening for user lifecycle events...")

	for {
		select {
		case <-ctx.Done():
			log.Println("[UserEventConsumer] Context cancelled; stopping event loop.")
			return
		default:
		}

		msg, err := c.kafkaConsumer.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				return
			}
			log.Printf("[UserEventConsumer] Error fetching message from Kafka: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		c.processMessage(ctx, msg)

		if err := c.kafkaConsumer.CommitMessages(ctx, msg); err != nil {
			log.Printf("[UserEventConsumer] Failed to commit message offset for partition %d offset %d: %v",
				msg.Partition, msg.Offset, err)
		}
	}
}

// processMessage unmarshals and routes domain events to specialized handlers.
// Why: Isolates malformed payloads from crashing the worker and guarantees offsets are advanced.
func (c *Consumer) processMessage(ctx context.Context, msg kafkaGo.Message) {
	log.Printf("[UserEventConsumer] Received message on topic '%s' partition %d offset %d: %s", msg.Topic, msg.Partition, msg.Offset, string(msg.Value))

	var event LifecycleEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Printf("[UserEventConsumer] Warning: Malformed event payload (offset %d): %v", msg.Offset, err)
		return
	}

	if event.UserID == "" {
		log.Printf("[UserEventConsumer] Warning: Missing userId in event payload (offset %d)", msg.Offset)
		return
	}

	switch event.Event {
	case EventUserBanned:
		c.handleUserBanned(ctx, event)
	case EventUserDeleted:
		c.handleUserDeleted(ctx, event)
	default:
		log.Printf("[UserEventConsumer] Unhandled lifecycle event type '%s' for user %s; skipping.", event.Event, event.UserID)
	}
}

// handleUserBanned deactivates the user record, revokes all refresh tokens, and blacklists active JWTs in Redis.
// Why: Immediately invalidates all existing and future sessions when an administrator bans an account.
func (c *Consumer) handleUserBanned(ctx context.Context, event LifecycleEvent) {
	reason := event.Reason
	if reason == "" {
		reason = "account_banned"
	}

	log.Printf("[UserEventConsumer] Processing '%s' for user %s (reason: %s)...", EventUserBanned, event.UserID, reason)

	if err := platformRedis.BlacklistUser(ctx, c.rdb, event.UserID, c.blacklistTTL, reason); err != nil {
		log.Printf("[UserEventConsumer] Error writing Redis blacklist for banned user %s: %v", event.UserID, err)
	}

	if err := c.refreshRepo.RevokeAllUserRefreshTokens(ctx, event.UserID); err != nil {
		log.Printf("[UserEventConsumer] Error revoking refresh tokens for banned user %s: %v", event.UserID, err)
	}

	if err := c.userRepo.DeactivateUser(ctx, event.UserID); err != nil {
		log.Printf("[UserEventConsumer] Error deactivating banned user %s: %v", event.UserID, err)
	}

	log.Printf("[UserEventConsumer] Successfully processed ban and session revocation for user %s.", event.UserID)
}

// handleUserDeleted purges user credentials from the database, revokes refresh tokens, and blacklists active JWTs in Redis.
// Why: Fulfills privacy and account removal requirements while blocking any in-flight access tokens.
func (c *Consumer) handleUserDeleted(ctx context.Context, event LifecycleEvent) {
	reason := event.Reason
	if reason == "" {
		reason = "account_deleted"
	}

	log.Printf("[UserEventConsumer] Processing '%s' for user %s...", EventUserDeleted, event.UserID)

	if err := platformRedis.BlacklistUser(ctx, c.rdb, event.UserID, c.blacklistTTL, reason); err != nil {
		log.Printf("[UserEventConsumer] Error writing Redis blacklist for deleted user %s: %v", event.UserID, err)
	}

	if err := c.refreshRepo.RevokeAllUserRefreshTokens(ctx, event.UserID); err != nil {
		log.Printf("[UserEventConsumer] Error revoking refresh tokens for deleted user %s: %v", event.UserID, err)
	}

	if err := c.userRepo.DeleteUser(ctx, event.UserID); err != nil {
		log.Printf("[UserEventConsumer] Error deleting user record %s: %v", event.UserID, err)
	}

	log.Printf("[UserEventConsumer] Successfully processed account deletion for user %s.", event.UserID)
}

// Close terminates the underlying Kafka message reader connection.
// Why: Cleans up partition locks and network resources upon worker shutdown.
func (c *Consumer) Close() error {
	if c.kafkaConsumer != nil {
		return c.kafkaConsumer.Close()
	}
	return nil
}
