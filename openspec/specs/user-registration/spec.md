# user-registration Specification

## Purpose
TBD - Account registration with bcrypt hashing and 2FA OTP verification.

## Requirements

### Requirement: User can register with email and password
The system SHALL allow new users to create an account by providing their name, email address, and password. The password SHALL be hashed using bcrypt (cost factor 12) before storage. If an account already exists with `is_active: false`, the system SHALL update the existing unverified account's name and password hash, invalidate any prior pending OTP codes, generate a fresh 6-digit OTP code, and dispatch the `auth.registration_otp` event. If an account already exists with `is_active: true`, the system SHALL reject the request. New accounts SHALL be created with role `CUSTOMER` and `is_active` set to `false` until OTP verification is complete. The system SHALL generate a 6-digit OTP code and publish an `auth.registration_otp` notification event containing the user's name, email, code, and `type: "REGISTRATION"`.

#### Scenario: Successful registration for new user
- **WHEN** a user submits a valid name, email, and password to `POST /api/auth/register` and no account exists
- **THEN** the system creates a new user record with `is_active: false`, generates a 6-digit OTP code, dispatches the `auth.registration_otp` notification event, and returns HTTP 201 with a message indicating that an OTP has been sent

#### Scenario: Re-registration for unverified existing user
- **WHEN** a user submits valid registration credentials for an email that already exists in the system with `is_active: false`
- **THEN** the system updates the user's name and bcrypt password hash, invalidates and nullifies previous pending OTP codes, generates a fresh 6-digit OTP, publishes the `auth.registration_otp` event, and returns HTTP 200/201 indicating a new verification code has been dispatched

#### Scenario: Duplicate active email registration
- **WHEN** a user submits an email that already exists in the system with `is_active: true`
- **THEN** the system returns HTTP 409 Conflict with an error message indicating the email is already registered

#### Scenario: Invalid input
- **WHEN** a user submits missing or malformed fields (empty name, invalid email format, password shorter than 8 characters)
- **THEN** the system returns HTTP 400 Bad Request with validation error details

### Requirement: User must verify OTP to activate account
The system SHALL require users to verify a 6-digit OTP code sent during registration before their account is activated. The OTP code SHALL expire after 5 minutes. A maximum of 5 verification attempts SHALL be allowed per OTP code. Upon successful verification, the system SHALL mark the OTP code as used, nullify/clear the raw OTP code from persistence, and activate the user account (`is_active: true`).

#### Scenario: Successful OTP verification
- **WHEN** a user submits the correct OTP code within 5 minutes and under 5 attempts
- **THEN** the system sets `is_active: true` on the user record, marks the OTP as used, scrubs/nullifies the code in database storage, and returns HTTP 200

#### Scenario: Expired OTP
- **WHEN** a user submits an OTP code after the 5-minute expiry window
- **THEN** the system returns HTTP 410 Gone with a message indicating the code has expired

#### Scenario: Exceeded maximum attempts
- **WHEN** a user has failed OTP verification 5 times for the same code
- **THEN** the system returns HTTP 429 Too Many Requests and invalidates the OTP code

#### Scenario: Invalid OTP code
- **WHEN** a user submits an incorrect OTP code
- **THEN** the system increments the attempt counter and returns HTTP 400 with an error message
