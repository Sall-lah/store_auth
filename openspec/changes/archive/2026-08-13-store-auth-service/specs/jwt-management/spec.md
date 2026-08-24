## ADDED Requirements

### Requirement: JWT tokens are signed with RS256
The system SHALL sign all JWT access tokens using an RSA 2048-bit private key with the RS256 algorithm. The JWT payload SHALL contain `sub` (user UUID), `email`, `role` (CUSTOMER or ADMIN), `iat` (issued at), and `exp` (expiration) claims. Tokens SHALL expire 1 hour after issuance.

#### Scenario: Token issuance on login
- **WHEN** a user successfully authenticates
- **THEN** the system generates a JWT signed with the RS256 private key containing the user's `sub`, `email`, `role`, `iat`, and `exp` claims, with `exp` set to 1 hour from `iat`

#### Scenario: Token contains correct claims
- **WHEN** a JWT is decoded
- **THEN** it SHALL contain exactly the fields: `sub` (UUID string), `email` (string), `role` ("CUSTOMER" or "ADMIN"), `iat` (Unix timestamp), `exp` (Unix timestamp)

### Requirement: JWT tokens are delivered via HttpOnly cookies
The system SHALL deliver JWT tokens exclusively via HTTP cookies with the following attributes: `HttpOnly`, `Secure`, `SameSite=Strict`, and `Path=/`.

#### Scenario: Cookie attributes on login
- **WHEN** a JWT is issued after successful login
- **THEN** the response sets a cookie named `access_token` with attributes `HttpOnly=true`, `Secure=true`, `SameSite=Strict`, `Path=/`, and `Max-Age` equal to the token's remaining lifetime

#### Scenario: Cookie not accessible via JavaScript
- **WHEN** client-side JavaScript attempts to read `document.cookie`
- **THEN** the `access_token` cookie SHALL NOT be visible due to the `HttpOnly` attribute

### Requirement: JWKS endpoint exposes public key
The system SHALL expose a `GET /.well-known/jwks.json` endpoint that returns the RSA public key in JSON Web Key Set format. This endpoint SHALL be publicly accessible without authentication.

#### Scenario: Other microservice fetches public key
- **WHEN** a microservice sends a GET request to `/.well-known/jwks.json`
- **THEN** the system returns HTTP 200 with a JSON body containing the RSA public key in JWK format with fields `kty`, `n`, `e`, `alg`, and `kid`

#### Scenario: JWKS endpoint requires no authentication
- **WHEN** an unauthenticated request is sent to `/.well-known/jwks.json`
- **THEN** the system returns HTTP 200 (no JWT cookie or auth header required)
