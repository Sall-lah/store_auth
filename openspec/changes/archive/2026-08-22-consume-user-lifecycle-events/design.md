## Context

`store_auth` acts as the Identity & Access Provider (IdP) for the platform, issuing RS256 JWT access tokens (15m expiration) and managing refresh tokens (7d rotation in PostgreSQL) and instantaneous token revocation via Redis (`revoked:user:{userId}`).
The business-facing user profile and admin account operations (such as user account deletion and admin user banning) reside in a separate `user` microservice. When an account is banned or deleted in `user`, `store_auth` must react asynchronously via Kafka to invalidate sessions and synchronize credential states.

## Goals / Non-Goals

**Goals:**
- Provide a robust background Kafka consumer package in `internal/platform/kafka/consumer.go` and `internal/user/consumer.go` using `segmentio/kafka-go`.
- Listen to topic configured by `KAFKA_TOPIC_USER_EVENTS` (default: `user.events`) with consumer group `KAFKA_CONSUMER_GROUP_AUTH` (default: `store-auth-user-events-group`).
- On `user.banned`: Blacklist user ID in Redis with 900s TTL, revoke all active refresh tokens in PostgreSQL, and update `isActive = false` in `users` table.
- On `user.deleted`: Blacklist user ID in Redis with 900s TTL, revoke refresh tokens, and delete credentials from `users` table.
- Implement graceful worker shutdown without message loss or dangling locks.

**Non-Goals:**
- Handling profile attributes (e.g., avatar, bio, address) in `store_auth` (owned by `user` microservice).
- Direct synchronous HTTP endpoints for banning/deleting (decoupled via event-driven Kafka architecture).
- Handling notifications to the banned/deleted user (handled by `notification-service`).

## Decisions

### 1. Pure Go Kafka Consumer with `segmentio/kafka-go`
* **Decision**: Use `kafkaGo.NewReader` with consumer group support instead of cgo-based libraries (like `confluent-kafka-go`).
* **Rationale**: Consistent with existing `internal/platform/kafka/producer.go`, zero external C library dependencies, easy cross-platform compilation, and native Go context cancellation support.
* **Alternatives Considered**:
  - `confluent-kafka-go` / `librdkafka`: Faster for extremely high throughput, but complicates Docker builds and cross-compilation on Windows/Alpine.

### 2. Event Payload Schema
* **Decision**: Standard JSON payload with envelope:
  ```json
  {
    "event": "user.banned",
    "userId": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
    "timestamp": "2026-08-22T21:00:00Z",
    "reason": "Terms violation"
  }
  ```
* **Rationale**: Simple, self-describing, and easily extensible for additional lifecycle events (e.g., `user.unbanned`, `user.password_reset`).

### 3. Graceful Background Lifecycle in `cmd/server/main.go`
* **Decision**: Run the consumer loop in a goroutine with a context cancelled upon OS signal (`SIGINT`/`SIGTERM`), waiting with `sync.WaitGroup` up to 5 seconds during server shutdown.
* **Rationale**: Prevents message processing interruptions and offset commit race conditions.

## Risks / Trade-offs

- **[Eventual Consistency Latency]**: There is a ~10-100ms gap between an admin clicking "Ban" and the Redis blacklist being populated.
  - *Mitigation*: 100ms is well within acceptable security thresholds for web store applications. If sub-millisecond guarantees are ever needed, direct Redis writing can be layered.
- **[Redis Outage During Event Processing]**: If Redis is temporarily unreachable when processing a `user.banned` event.
  - *Mitigation*: The consumer retries with backoff and marks refresh tokens in PostgreSQL, ensuring refresh flows are blocked even if Redis fails open.
- **[Poison Pill / Malformed Messages]**: A malformed message blocking the partition.
  - *Mitigation*: Wrap JSON unmarshaling in error logging and commit offset immediately to unblock consumer partition.

## Migration Plan

1. Rebuild Podman container (`podman build -t store_auth .`) and verify clean container startup.
2. Deploy new `store_auth` version with `KAFKA_TOPIC_USER_EVENTS` and consumer enabled.
3. Verify consumer group joins and offsets start from latest.
4. Deploy `user` microservice with producer publishing to `user.events`.
5. Rollback: If Kafka is unavailable, disable consumer via `ENABLE_USER_EVENTS_CONSUMER=false` (defaults to enabled when Kafka brokers are configured).
