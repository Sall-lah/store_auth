# kafka-user-events Specification

## ADDED Requirements

### Requirement: Consumer group processes user.banned events
The system SHALL run a background Kafka consumer group subscribing to the user lifecycle event topic (`user.events`). Upon receiving a `user.banned` event with a payload containing `userId`, the system SHALL:
1. Set a Redis token blacklist key `revoked:user:<userID>` with a TTL of 900 seconds (15 minutes).
2. Revoke all active refresh tokens for that `userID` in PostgreSQL.
3. Update the user record in PostgreSQL to set `isActive = false`.

#### Scenario: User banned event received
- **WHEN** a valid JSON event `{"event": "user.banned", "userId": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d", "reason": "Terms violation"}` is received on the topic
- **THEN** the system blacklists the user ID in Redis with 900s TTL, marks all user refresh tokens revoked in PostgreSQL, and sets `isActive = false` on the user record

### Requirement: Consumer group processes user.deleted events
Upon receiving a `user.deleted` event with a payload containing `userId`, the system SHALL:
1. Set a Redis token blacklist key `revoked:user:<userID>` with a TTL of 900 seconds (15 minutes).
2. Revoke and cascade-delete all refresh tokens and OTP records for that `userID` in PostgreSQL.
3. Delete the user record from PostgreSQL to remove credentials and free the email address for future registration.

#### Scenario: User deleted event received
- **WHEN** a valid JSON event `{"event": "user.deleted", "userId": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"}` is received on the topic
- **THEN** the system blacklists the user ID in Redis with 900s TTL, purges all refresh tokens, and deletes the user record from the Auth database

### Requirement: Consumer handles malformed messages and unknown events safely
The system SHALL acknowledge and commit Kafka message offsets even if the payload is malformed or represents an unknown event type, logging an error or warning without crashing the consumer worker.

#### Scenario: Malformed event received
- **WHEN** a non-JSON or invalid payload is read from the topic
- **THEN** the system logs an error, commits the offset, and continues processing subsequent messages without terminating the service

### Requirement: Graceful consumer lifecycle management
The system SHALL start the Kafka consumer worker during server initialization in a background goroutine and shut it down cleanly upon receiving OS termination signals (`SIGINT`, `SIGTERM`), waiting for active message processing to complete within a configurable timeout.

#### Scenario: Server shutdown signal received
- **WHEN** the server receives a `SIGTERM` or `SIGINT` signal
- **THEN** the consumer stops fetching new messages, completes inflight event processing, commits offsets, and closes the Kafka reader connection cleanly
