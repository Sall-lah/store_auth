## 1. Configuration & Consumer Infrastructure

- [x] 1.1 Add `KafkaTopicUserEvents` and `KafkaConsumerGroupAuth` configuration fields and env loaders in `internal/config/config.go`.
- [x] 1.2 Implement pure-Go Kafka consumer reader wrapper with group rebalancing in `internal/platform/kafka/consumer.go`.
- [x] 1.3 Add unit tests for Kafka consumer configuration in `internal/config/config_test.go`.

## 2. Event Handler & Business Logic

- [x] 2.1 Define user lifecycle event payload schemas and types in `internal/user/event.go`.
- [x] 2.2 Implement `UserEventConsumer` service to process `user.banned` and `user.deleted` events with Redis blacklisting and DB revocation in `internal/user/consumer.go`.
- [x] 2.3 Add comprehensive unit tests mocking DB and Redis for `user.banned`, `user.deleted`, and malformed payloads in `internal/user/consumer_test.go`.

## 3. Server Lifecycle Integration & Documentation

- [x] 3.1 Wire background Kafka consumer worker lifecycle and graceful shutdown in `cmd/server/main.go`.
- [x] 3.2 Update `.env.example`, `.env.production.example`, and `README.md` with Kafka user lifecycle consumer settings.

## 4. Container Build & Verification

- [x] 4.1 Rebuild the Podman container image (`podman build -t store_auth .`) to verify multi-stage compilation with new Kafka consumer dependencies.
- [x] 4.2 Start and verify the updated Podman container before running end-to-end integration and smoke tests.
