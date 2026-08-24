## MODIFIED Requirements

### Requirement: JWT tokens are signed with RS256
The system SHALL sign all JWT access tokens using an RSA 2048-bit private key with the RS256 algorithm. The JWT payload SHALL contain `sub` (user UUID), `email`, `role` (CUSTOMER or ADMIN), `iat` (issued at), and `exp` (expiration) claims. Tokens SHALL expire **15 minutes** after issuance.

#### Scenario: Token issuance on login
- **WHEN** a user successfully authenticates
- **THEN** the system generates a JWT signed with the RS256 private key containing the user's `sub`, `email`, `role`, `iat`, and `exp` claims, with `exp` set to **15 minutes** from `iat`

#### Scenario: Token contains correct claims
- **WHEN** a JWT is decoded
- **THEN** it SHALL contain exactly the fields: `sub` (UUID string), `email` (string), `role` ("CUSTOMER" or "ADMIN"), `iat` (Unix timestamp), `exp` (Unix timestamp)

### Requirement: JWT tokens are delivered via HttpOnly cookies
The system SHALL deliver JWT access tokens exclusively via HTTP cookies with the following attributes: `HttpOnly`, `Secure`, `SameSite=Lax`, and `Path=/`. The system SHALL also deliver a separate `refresh_token` cookie with attributes: `HttpOnly`, `Secure`, `SameSite=Lax`, and `Path=/api/auth/refresh`.

#### Scenario: Cookie attributes on login
- **WHEN** a JWT is issued after successful login
- **THEN** the response sets a cookie named `access_token` with attributes `HttpOnly=true`, `Secure=true`, `SameSite=Lax`, `Path=/`, and `Max-Age=900` (15 minutes), and a cookie named `refresh_token` with attributes `HttpOnly=true`, `Secure=true`, `SameSite=Lax`, `Path=/api/auth/refresh`, and `Max-Age=604800` (7 days)

#### Scenario: Cookie not accessible via JavaScript
- **WHEN** client-side JavaScript attempts to read `document.cookie`
- **THEN** neither the `access_token` nor the `refresh_token` cookie SHALL be visible due to the `HttpOnly` attribute

#### Scenario: Cookie attributes on logout
- **WHEN** a user logs out via `POST /api/auth/logout`
- **THEN** the system clears both `access_token` and `refresh_token` cookies by setting them to empty values with `Max-Age=-1`
