# Design: Integrate Notification Service for OTP Delivery via Kafka

## Context

`store_auth` generates 6-digit one-time passcodes (OTPs) for user registration and password recovery. Previously, OTPs were delivered either by logging to standard output (`mock`) or sending direct HTML emails via standard library `net/smtp` (`smtp`).

A dedicated `store_notification` microservice is now deployed and operational, consuming events from Kafka topics `auth.events` and `order.events`. It encapsulates email template rendering, Redis deduplication locks, and external SMTP gateway communication.

This design transitions `store_auth` to publish structured domain events to Kafka, eliminating embedded SMTP transport from the auth codebase.

## Goals / Non-Goals

**Goals:**
- Implement a thread-safe, pure Go Kafka producer client using `github.com/segmentio/kafka-go` in `internal/platform/kafka`.
- Publish `auth.registration_otp` and `auth.password_reset_otp` event envelopes to topic `auth.events` keyed by user email.
- Upgrade the `OTPSender` interface to pass `context.Context`, user display name, and OTP type for template personalization.
- Implement `KafkaOTPSender` for production and retain `LogOTPSender` for mock/local testing.
- Remove all `SMTPOTPSender` code, `net/smtp` imports, and `SMTP_*` environment variables from configuration and documentation.
- Wire Kafka producer lifecycle and graceful shutdown into `cmd/server/main.go`.

**Non-Goals:**
- Implementing Kafka consumer workers inside `store_auth` (`store_auth` is exclusively a publisher).
- Creating a database-backed transactional outbox table for OTPs (direct resilient publishing is preferred for fast user authentication flows).
- Implementing HTML email templates inside `store_auth` (templates are owned and rendered by `store_notification`).

## Decisions

### Decision 1: Use `segmentio/kafka-go` for pure Go event publishing
- **Rationale**: `segmentio/kafka-go` provides a clean, idiomatic Go API without requiring CGo or external `librdkafka` C libraries, ensuring fast cross-platform compilation and lightweight container builds.
- **Alternatives Considered**:
  - `confluent-kafka-go`: Requires CGo and `librdkafka`, complicating Windows/Linux development and CI/CD pipelines.
  - `Shopify/sarama`: Heavyweight and complex API compared to `segmentio/kafka-go`.

### Decision 2: Direct Kafka Publishing instead of Transactional Outbox
- **Rationale**: Auth OTP delivery is an interactive user request where immediate feedback is essential. Direct publication with connection timeouts provides immediate status back to the client without the overhead of outbox polling routines or extra database tables.
- **Alternatives Considered**:
  - *Transactional Outbox Table*: Increases schema complexity and adds background polling loop latency which is unnecessary for ephemeral 5-minute OTP codes.

### Decision 3: Upgrade `OTPSender` Interface Signature
- **Rationale**: Rich email notifications require the user's name (for greetings) and specific OTP event types (`REGISTRATION` vs `PASSWORD_RESET`). Updating `OTPSender` to `SendOTP(ctx context.Context, email, code, name string, otpType Type) error` maintains abstraction while supplying all payload requirements.
- **Alternatives Considered**:
  - Separate interfaces for each event: Unnecessary fragmentation for a single notification domain concept.

### Decision 4: Complete Removal of SMTP Transport and Configuration
- **Rationale**: Removing `SMTPOTPSender` and all `SMTP_*` configuration variables enforces single-responsibility architecture across the microservices ecosystem. It prevents configuration drift and removes unused credentials from environment files.
- **Alternatives Considered**:
  - Keeping SMTP as a fallback provider: Increases codebase maintenance burden and creates competing email delivery paths.

## Event Contract & Data Schema

### Topic: `auth.events`

```json
{
  "event_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
  "event_type": "auth.registration_otp",
  "timestamp": "2026-08-21T14:40:00Z",
  "producer": "store_auth",
  "data": {
    "email": "customer@example.com",
    "code": "123456",
    "name": "Jane Doe",
    "type": "REGISTRATION"
  }
}
```

## Risks / Trade-offs

- **[Kafka Broker Unavailability]** → Handled via `WriteTimeout` (5s). If the broker is unreachable in production, `Publish` returns an error, alerting the user to retry. In local dev/testing, `OTP_PROVIDER=mock` bypasses Kafka entirely.
- **[Host / Container DNS Resolution]** → The smart dialer in `internal/platform/kafka/producer.go` detects if `kafka` or `broker` hostnames fail lookup and resolves to `127.0.0.1`, allowing seamless local execution on host machines against containerized Kafka brokers.
