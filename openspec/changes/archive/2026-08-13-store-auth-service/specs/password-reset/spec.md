## ADDED Requirements

### Requirement: User can request a password reset
The system SHALL allow users to request a password reset by providing their registered email address. The system SHALL generate a 6-digit OTP code and send it via SMS/Email. The system SHALL NOT reveal whether the email exists in the system to prevent user enumeration.

#### Scenario: Password reset request for existing user
- **WHEN** a user submits a registered email to `POST /api/auth/forgot-password`
- **THEN** the system generates a 6-digit OTP with a 5-minute expiry, sends it to the user, and returns HTTP 200 with a generic success message

#### Scenario: Password reset request for non-existent email
- **WHEN** a user submits an email that does not exist in the system
- **THEN** the system returns HTTP 200 with the same generic success message (no information leakage)

### Requirement: User can reset password after OTP verification
The system SHALL allow users to set a new password after successfully verifying the OTP code from the password reset flow. The new password SHALL be hashed using bcrypt (cost factor 12) before storage.

#### Scenario: Successful password reset
- **WHEN** a user submits a valid OTP code and a new password to `POST /api/auth/reset-password`
- **THEN** the system verifies the OTP, hashes the new password with bcrypt, updates the user record, marks the OTP as used, and returns HTTP 200

#### Scenario: Expired or invalid OTP during reset
- **WHEN** a user submits an expired or incorrect OTP code with a new password
- **THEN** the system returns the appropriate error (HTTP 410 for expired, HTTP 400 for invalid) and does NOT update the password

#### Scenario: Weak new password
- **WHEN** a user submits a new password shorter than 8 characters
- **THEN** the system returns HTTP 400 Bad Request with validation error details
