# user-registration Specification

## MODIFIED Requirements

### Requirement: User can register with email and password
The system SHALL allow new users to create an account by providing their name, email address, and password. The password SHALL be hashed using bcrypt (cost factor 12) before storage. The email address SHALL be unique across all users. New accounts SHALL be created with role `CUSTOMER` and `is_active` set to `false` until OTP verification is complete. The system SHALL generate a 6-digit OTP code and publish an `auth.registration_otp` notification event containing the user's name, email, code, and `type: "REGISTRATION"`.

#### Scenario: Successful registration
- **WHEN** a user submits a valid name, email, and password to `POST /api/auth/register`
- **THEN** the system creates a new user record with `is_active: false`, generates a 6-digit OTP code, dispatches the `auth.registration_otp` notification event, and returns HTTP 201 with a message indicating that an OTP has been sent

#### Scenario: Duplicate email registration
- **WHEN** a user submits an email that already exists in the system
- **THEN** the system returns HTTP 409 Conflict with an error message

#### Scenario: Invalid input
- **WHEN** a user submits missing or malformed fields (empty name, invalid email format, password shorter than 8 characters)
- **THEN** the system returns HTTP 400 Bad Request with validation error details
