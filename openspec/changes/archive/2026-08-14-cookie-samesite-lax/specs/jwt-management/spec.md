# jwt-management Specification Delta

## Requirements

### Requirement: JWT tokens are delivered via HttpOnly cookies
The system SHALL deliver JWT tokens exclusively via HTTP cookies with the following attributes: `HttpOnly`, `Secure`, `SameSite=Lax`, and `Path=/`.

#### Scenario: Cookie attributes on login
- **WHEN** a JWT is issued after successful login
- **THEN** the response sets a cookie named `access_token` with attributes `HttpOnly=true`, `Secure=true` (in production), `SameSite=Lax`, `Path=/`, and `Max-Age` equal to the token's remaining lifetime
