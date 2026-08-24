# Tasks: Integrate Notification Service for OTP Delivery via Kafka

## 1. Dependencies & Configuration Cleanup

- [x] 1.1 Add `github.com/segmentio/kafka-go` dependency to `go.mod`
- [x] 1.2 Remove `SMTPHost`, `SMTPPort`, `SMTPUsername`, `SMTPPassword`, and `SMTPFrom` from `internal/config/config.go` and add `KafkaBrokers`, `KafkaTopicAuthEvents`, and update `OTPProvider`
- [x] 1.3 Update `.env.example`, `.env.production.example`, and `README.md` to remove `SMTP_*` variables and document `KAFKA_BROKERS`, `KAFKA_TOPIC_AUTH_EVENTS`, and `OTP_PROVIDER`

## 2. Kafka Platform Producer Implementation

- [x] 2.1 Implement `KafkaProducer` wrapper in `internal/platform/kafka/producer.go` with pure Go segmentio writer, smart dialer, and clean connection shutdown
- [x] 2.2 Add unit tests for Kafka producer interface and fallback dialer configuration

## 3. OTP Sender Refactoring

- [x] 3.1 Upgrade `OTPSender` interface in `internal/otp/sender.go` to `SendOTP(ctx context.Context, email, code, name string, otpType Type) error`
- [x] 3.2 Delete `SMTPOTPSender` and remove `net/smtp` imports from `internal/otp/sender.go`
- [x] 3.3 Implement `KafkaOTPSender` constructing `EventEnvelope` (`auth.registration_otp`, `auth.password_reset_otp`) matching `store_notification` schemas
- [x] 3.4 Update `LogOTPSender` and provider factory `NewOTPSender` for mock and Kafka delivery modes
- [x] 3.5 Update `SendOTP` in `internal/otp/service.go` to propagate context, display name, and OTP type

## 4. Auth Service & Server Wiring

- [x] 4.1 Update `Register` in `internal/auth/service.go` to dispatch registration OTP event with user display name
- [x] 4.2 Update `ForgotPassword` in `internal/auth/service.go` to dispatch password reset OTP event with user display name
- [x] 4.3 Wire Kafka producer initialization, `KafkaOTPSender` injection, and graceful shutdown into `cmd/server/main.go`

## 5. Verification & Testing

- [x] 5.1 Update unit tests in `internal/otp/otp_test.go`, `internal/config/config_test.go`, and `internal/auth/handler_test.go`
- [x] 5.2 Run `go test ./...` and `go build ./...` to verify clean compilation, zero linter errors, and passing tests
