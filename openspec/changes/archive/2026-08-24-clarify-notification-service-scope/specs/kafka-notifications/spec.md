## MODIFIED Requirements

### Requirement: Service publishes authentication domain events to Kafka
The system SHALL publish structured JSON domain events to the configured Apache Kafka topic (`auth.events`) exclusively for transactional authentication lifecycle events requiring customer OTP notification (registration and password reset). The system SHALL NOT expect or require the notification service to persist in-app notification state or participate in user lifecycle deactivation. Each event SHALL be wrapped in a standard `EventEnvelope` containing `event_id` (UUID), `event_type`, `timestamp` (ISO8601), `producer` ("store_auth"), and `data` (JSON object).

#### Scenario: Successful Kafka event publication
- **WHEN** an OTP delivery event occurs and Kafka broker is reachable
- **THEN** the system publishes the message envelope to topic `auth.events` keyed by the user email without error

#### Scenario: Mock provider fallback
- **WHEN** `OTP_PROVIDER` is set to `mock`
- **THEN** the system logs the OTP code and event data to stdout without attempting Kafka network connections
