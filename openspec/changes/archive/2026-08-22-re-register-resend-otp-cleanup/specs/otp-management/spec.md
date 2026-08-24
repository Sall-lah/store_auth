## ADDED Requirements

### Requirement: User can resend verification OTP
The system SHALL provide an endpoint `POST /api/auth/resend-otp` accepting an email and OTP type (`registration` or `password_reset`). For `registration`, if the account exists and is unverified (`is_active: false`), the system SHALL invalidate prior registration OTPs, generate a fresh 6-digit code, and publish an `auth.registration_otp` event. If the account is already active, the system SHALL return HTTP 400 Bad Request. For `password_reset`, if the account exists and is active, the system SHALL invalidate prior reset OTPs, generate a fresh 6-digit code, and publish an `auth.password_reset_otp` event. If the account does not exist or is inactive, the system SHALL return HTTP 200 without emitting events to prevent email enumeration.

#### Scenario: Successful registration OTP resend for unverified user
- **WHEN** a user posts `{ "email": "user@example.com", "type": "registration" }` and the user exists with `is_active: false`
- **THEN** the system invalidates old registration OTPs, issues a new 6-digit OTP, publishes `auth.registration_otp` to Kafka topic `auth.events`, and returns HTTP 200 with message "A new verification code has been sent."

#### Scenario: Registration OTP resend for already active user
- **WHEN** a user posts `{ "email": "user@example.com", "type": "registration" }` and the user exists with `is_active: true`
- **THEN** the system returns HTTP 400 Bad Request with message "Account is already verified. Please log in."

#### Scenario: Registration OTP resend for non-existent user
- **WHEN** a user posts `{ "email": "unknown@example.com", "type": "registration" }`
- **THEN** the system returns HTTP 400 Bad Request with message "User account not found"

#### Scenario: Password reset OTP resend for active user
- **WHEN** a user posts `{ "email": "user@example.com", "type": "password_reset" }` and the user exists with `is_active: true`
- **THEN** the system invalidates old reset OTPs, issues a new 6-digit OTP, publishes `auth.password_reset_otp` to Kafka topic `auth.events`, and returns HTTP 200 with a generic confirmation message

#### Scenario: Password reset OTP resend for non-existent or inactive user
- **WHEN** a user posts `{ "email": "unknown@example.com", "type": "password_reset" }`
- **THEN** the system returns HTTP 200 with the identical confirmation message without dispatching any notification event

#### Scenario: Invalid OTP resend type
- **WHEN** a user posts an unsupported or missing `type` parameter
- **THEN** the system returns HTTP 400 Bad Request with validation details

### Requirement: OTP secret code scrubbing
The system SHALL scrub and nullify the raw numeric OTP code stored in the database whenever an OTP is successfully marked as used, replaced by a resend, or invalidated upon account activation or password reset completion.

#### Scenario: Scrubbing code on OTP consumption
- **WHEN** an OTP is verified successfully during account activation or password reset
- **THEN** the system sets `used: true` and clears/nullifies the `code` field in the database record

#### Scenario: Scrubbing code on OTP invalidation
- **WHEN** an OTP is replaced by a newly issued OTP or invalidated after password change
- **THEN** the system marks pending OTPs as `used: true` and clears/nullifies the `code` field
