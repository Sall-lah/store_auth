## ADDED Requirements

### Requirement: User can log in with email and password
The system SHALL authenticate users by validating their email and password against the stored bcrypt hash. Upon successful authentication, the system SHALL issue a JWT access token signed with RS256 and deliver it via an HttpOnly, Secure, SameSite=Strict cookie.

#### Scenario: Successful login
- **WHEN** a user submits a valid email and correct password to `POST /api/auth/login` and the account is active
- **THEN** the system returns HTTP 200, sets an HttpOnly cookie containing a signed JWT with claims (`sub`, `email`, `role`, `iat`, `exp`), and returns the user profile in the response body

#### Scenario: Invalid credentials
- **WHEN** a user submits an incorrect email or password
- **THEN** the system returns HTTP 401 Unauthorized with a generic error message (SHALL NOT reveal whether the email or password was incorrect)

#### Scenario: Inactive account
- **WHEN** a user attempts to log in with correct credentials but `is_active` is `false`
- **THEN** the system returns HTTP 403 Forbidden indicating the account has not been verified

### Requirement: User can log out
The system SHALL provide a logout endpoint that clears the authentication cookie.

#### Scenario: Successful logout
- **WHEN** a user sends a request to `POST /api/auth/logout`
- **THEN** the system clears the HttpOnly auth cookie by setting it to an empty value with `Max-Age=0` and returns HTTP 200

### Requirement: User can view their own profile
The system SHALL provide an endpoint to retrieve the current authenticated user's profile information from their JWT claims.

#### Scenario: Authenticated user requests profile
- **WHEN** an authenticated user sends a request to `GET /api/auth/me` with a valid JWT cookie
- **THEN** the system returns HTTP 200 with the user's `id`, `email`, `name`, and `role`

#### Scenario: Unauthenticated request
- **WHEN** a request is sent to `GET /api/auth/me` without a valid JWT cookie
- **THEN** the system returns HTTP 401 Unauthorized
