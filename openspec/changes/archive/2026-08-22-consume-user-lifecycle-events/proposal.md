## Why

In a microservices architecture, user account lifecycle actions (such as an admin banning a user or a user deleting their account) originate in the downstream `user` service. To maintain zero-trust security and prevent banned/deleted users from continuing to use active 15-minute RS256 JWT tokens or rotating 7-day refresh tokens, `store_auth` must asynchronously consume user lifecycle domain events from Kafka and trigger immediate session revocation and credential synchronization.

## What Changes

- Add a background Kafka consumer group in `store_auth` listening to the user lifecycle event topic (`user.events`).
- Handle `user.banned` events:
  - Populate Redis token blacklist (`revoked:user:{userId}`) with a 15-minute TTL to block active access tokens immediately.
  - Revoke all active refresh tokens in PostgreSQL for the target user ID.
  - Set `isActive = false` on the user record in PostgreSQL to reject subsequent login attempts with `403 Forbidden`.
- Handle `user.deleted` events:
  - Populate Redis token blacklist (`revoked:user:{userId}`) with a 15-minute TTL.
  - Revoke all active refresh tokens in PostgreSQL for the target user ID.
  - Delete or purge user credentials from PostgreSQL so the email is released for future registration and PII is removed.
- Add configuration parameters for Kafka Consumer Group ID and user events topic.
- Implement graceful startup and shutdown hooks for the Kafka consumer worker.

## Capabilities

### New Capabilities
- `kafka-user-events`: Background Kafka consumer worker for processing asynchronous user lifecycle domain events (`user.banned`, `user.deleted`, `user.unbanned`) to trigger token revocation and credential synchronization.

### Modified Capabilities
<!-- No requirement changes to existing specs; token-blacklist and refresh-token specs already expect revocation triggers -->

## Impact

- **New Dependencies / Packages**: Kafka consumer group reader in pure Go (`github.com/segmentio/kafka-go`).
- **Runtime / Concurrency**: Background goroutine spawned during `cmd/server/main.go` startup with graceful termination on `SIGINT`/`SIGTERM`.
- **Database / Cache**: PostgreSQL queries to update/delete user records and revoke refresh tokens; Redis writes to `revoked:user:{userId}`.
- **Config**: New environment variables `KAFKA_TOPIC_USER_EVENTS` (default: `user.events`) and `KAFKA_CONSUMER_GROUP_AUTH` (default: `store-auth-user-events-group`).
