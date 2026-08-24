# Proposal: Integrate Notification Service for OTP Delivery via Kafka

## Why

`store_auth` currently relies on local console logs or direct SMTP email transport embedded inside the auth microservice to deliver OTP codes. In our decoupled microservices architecture, a dedicated `store_notification` service is now online to handle event ingestion, Redis deduplication, template rendering, and email dispatch. Delegating OTP delivery to `store_notification` via Apache Kafka removes email infrastructure concerns from `store_auth`, eliminates direct SMTP configuration, and aligns auth with the platform's asynchronous event-driven standard.

## What Changes

- **Kafka Event Publication**: Introduce pure Go Kafka message publishing via `segmentio/kafka-go` in `store_auth` to dispatch domain events to the `auth.events` topic.
- **Asynchronous OTP Dispatch**: Emit `auth.registration_otp` and `auth.password_reset_otp` events conforming to the standard `EventEnvelope` containing email, 6-digit OTP code, user name, and OTP type.
- **Enhanced `OTPSender` Interface**: Upgrade `OTPSender` to accept `context.Context`, destination email, OTP code, user display name, and OTP type for personalized email generation.
- **Direct SMTP Removal (BREAKING)**: Completely remove direct SMTP email sending (`SMTPOTPSender`), eliminating `net/smtp` dependencies from `store_auth`.
- **Configuration & Environment Cleanup (BREAKING)**: Remove `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, and `SMTP_FROM` from application config and `.env` templates. Introduce `KAFKA_BROKERS` and `KAFKA_TOPIC_AUTH_EVENTS`.
- **Testing & Mocking**: Retain `LogOTPSender` under `OTP_PROVIDER=mock` for zero-broker offline development and unit tests.

## Capabilities

### New Capabilities
- `kafka-notifications`: Direct publication of authenticated domain events (`auth.registration_otp`, `auth.password_reset_otp`) onto Kafka topic `auth.events` using structured event envelopes.

### Modified Capabilities
- `user-registration`: Emits asynchronous `auth.registration_otp` event with user name and code upon new account creation instead of invoking local SMTP.
- `password-reset`: Emits asynchronous `auth.password_reset_otp` event with user name and code upon password reset request instead of invoking local SMTP.

## Impact

- **Dependencies**: Added `github.com/segmentio/kafka-go`.
- **Infrastructure / Environment**: Removed SMTP environment variables. Added `KAFKA_BROKERS` (default `localhost:9092`) and `KAFKA_TOPIC_AUTH_EVENTS` (default `auth.events`).
- **Internal Modules**:
  - `internal/config`: Removed SMTP properties, added Kafka settings.
  - `internal/platform/kafka`: New Kafka producer client.
  - `internal/otp`: Removed `SMTPOTPSender`, implemented `KafkaOTPSender`, updated `OTPSender` interface.
  - `internal/auth`: Updated `Register` and `ForgotPassword` service calls to pass user name and OTP type.
  - `cmd/server/main.go`: Wired Kafka producer lifecycle and graceful shutdown.
- **External Integration**: Produces events consumed by `store_notification` consumer group `store-notification-group-auth-events`.
