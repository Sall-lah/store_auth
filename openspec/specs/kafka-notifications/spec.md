# kafka-notifications Specification

## Purpose
TBD - Direct publishing of authentication domain lifecycle events (registration OTP, password reset OTP) to Apache Kafka topic `auth.events`.

## Requirements

### Requirement: Service publishes authentication domain events to Kafka
The system SHALL publish structured JSON domain events to the configured Apache Kafka topic (`auth.events`) whenever authentication lifecycle events requiring customer notification occur. Each event SHALL be wrapped in a standard `EventEnvelope` containing `event_id` (UUID), `event_type`, `timestamp` (ISO8601), `producer` ("store_auth"), and `data` (JSON object).

#### Scenario: Successful Kafka event publication
- **WHEN** an OTP delivery event occurs and Kafka broker is reachable
- **THEN** the system publishes the message envelope to topic `auth.events` keyed by the user email without error

#### Scenario: Mock provider fallback
- **WHEN** `OTP_PROVIDER` is set to `mock`
- **THEN** the system logs the OTP code and event data to stdout without attempting Kafka network connections

### Requirement: Registration and Password Reset OTP event schemas
The system SHALL emit event payloads conforming to the `AuthOtpEventData` schema containing `email`, `code`, `name`, and `type`.
- For user registration, `event_type` SHALL be `auth.registration_otp` and `type` SHALL be `REGISTRATION`.
- For password reset, `event_type` SHALL be `auth.password_reset_otp` and `type` SHALL be `PASSWORD_RESET`.

#### Scenario: Emitting registration OTP event
- **WHEN** a user registers a new account
- **THEN** the system emits an envelope with `event_type: "auth.registration_otp"` and data payload containing the 6-digit code, destination email, user display name, and `type: "REGISTRATION"`

#### Scenario: Emitting password reset OTP event
- **WHEN** an existing user requests a password reset
- **THEN** the system emits an envelope with `event_type: "auth.password_reset_otp"` and data payload containing the 6-digit code, destination email, user display name, and `type: "PASSWORD_RESET"`
