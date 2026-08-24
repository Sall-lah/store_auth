# refresh-token Specification

## Purpose
TBD - Stateful opaque refresh token lifecycle management including database persistence, token rotation on refresh, and reuse detection.

## Requirements

### Requirement: System issues refresh token on login
The system SHALL generate a cryptographically random opaque refresh token (32 bytes, base64url-encoded) on successful login, store its SHA-256 hash in PostgreSQL with an expiration of 7 days, and deliver it via an HttpOnly cookie named `refresh_token` scoped to `Path=/api/auth/refresh`.

#### Scenario: Refresh token issued alongside access token
- **WHEN** a user successfully authenticates via `POST /api/auth/login`
- **THEN** the system sets two HttpOnly cookies: `access_token` (15-minute expiry, `Path=/`) and `refresh_token` (7-day expiry, `Path=/api/auth/refresh`), and stores the SHA-256 hash of the refresh token in the `refresh_tokens` table linked to the user

### Requirement: User can refresh access token
The system SHALL provide a `POST /api/auth/refresh` endpoint that validates the refresh token cookie against the database, verifies the user account is active and not blacklisted, fetches the latest user role, and issues a new short-lived access token with updated claims.

#### Scenario: Successful token refresh
- **WHEN** a user sends `POST /api/auth/refresh` with a valid, unexpired, unrevoked `refresh_token` cookie and the user account is active
- **THEN** the system returns HTTP 200, sets a new `access_token` cookie (15-minute expiry) with the user's current `role` from the database, rotates the refresh token (invalidates old, issues new), and sets a new `refresh_token` cookie

#### Scenario: Expired refresh token
- **WHEN** a user sends `POST /api/auth/refresh` with an expired refresh token
- **THEN** the system returns HTTP 401 Unauthorized and clears both cookie values

#### Scenario: Revoked refresh token (not reuse)
- **WHEN** a user sends `POST /api/auth/refresh` with a refresh token that has been explicitly revoked (e.g., by logout)
- **THEN** the system returns HTTP 401 Unauthorized and clears both cookie values

#### Scenario: Inactive user account during refresh
- **WHEN** a user sends `POST /api/auth/refresh` with a valid refresh token but the user's `isActive` is `false`
- **THEN** the system returns HTTP 403 Forbidden, revokes all refresh tokens for the user, and clears both cookie values

### Requirement: Refresh token rotation with reuse detection
The system SHALL rotate refresh tokens on every successful refresh. If a previously-rotated (already-used) refresh token is presented, the system SHALL revoke ALL refresh tokens for that user as a security breach response.

#### Scenario: Token rotation on refresh
- **WHEN** a refresh is successful
- **THEN** the old refresh token record is marked as `revoked=true` in the database and a new refresh token is issued

#### Scenario: Reuse of rotated token detected
- **WHEN** a request presents a refresh token whose hash exists in the database with `revoked=true`
- **THEN** the system revokes ALL refresh tokens for that user, returns HTTP 401 Unauthorized, and clears both cookie values
