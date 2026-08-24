# password-reset Specification

## MODIFIED Requirements

### Requirement: User can request a password reset
The system SHALL allow users to request a password reset by providing their registered email address. The system SHALL generate a 6-digit OTP code and publish an `auth.password_reset_otp` notification event containing the user's display name, email, code, and `type: "PASSWORD_RESET"`. The system SHALL NOT reveal whether the email exists in the system to prevent user enumeration.

#### Scenario: Password reset request for existing user
- **WHEN** a user submits a registered email to `POST /api/auth/forgot-password`
- **THEN** the system generates a 6-digit OTP with a 5-minute expiry, dispatches the `auth.password_reset_otp` notification event, and returns HTTP 200 with a generic success message

#### Scenario: Password reset request for non-existent email
- **WHEN** a user submits an email that does not exist in the system
- **THEN** the system returns HTTP 200 with the same generic success message (no information leakage)
