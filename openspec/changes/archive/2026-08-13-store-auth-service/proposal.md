## Why

The web store application uses a microservices architecture and currently has no authentication service. Users cannot register, log in, or securely access protected resources across the system. A dedicated Auth microservice is needed to provide centralized identity management, secure credential handling, and stateless JWT-based authentication that other microservices can verify independently.

## What Changes

- Introduce a standalone Go-based Auth microservice (`store_auth`) with Prisma ORM connected to Supabase (PostgreSQL).
- Implement email/password registration and login with bcrypt password hashing.
- Add SMS/Email OTP-based 2FA verification for registration and password reset flows.
- Issue stateless RS256-signed JWTs delivered via `HttpOnly` cookies.
- Expose a `/.well-known/jwks.json` endpoint so other microservices can verify tokens using the public key without contacting the Auth service.
- Implement in-memory rate limiting on authentication endpoints to prevent brute-force attacks.
- Support two user roles: `CUSTOMER` and `ADMIN`, embedded as JWT claims.

## Capabilities

### New Capabilities
- `user-registration`: Email/password account creation with bcrypt hashing and SMS/Email OTP verification before activation.
- `user-login`: Email/password credential validation, JWT access token issuance via HttpOnly cookies.
- `password-reset`: OTP-verified password reset flow with bcrypt re-hashing.
- `jwt-management`: RS256 key pair signing, HttpOnly cookie transport, JWKS public key endpoint for cross-service verification.
- `rate-limiting`: Redis-backed sliding window counter rate limiting on public authentication endpoints (login, register, OTP, password reset), keyed by IP and endpoint.
- `role-authorization`: CUSTOMER and ADMIN role assignment, JWT claim embedding, and middleware-level role checking.

### Modified Capabilities
_(None — this is a greenfield service.)_

## Impact

- **New service**: Entirely new Go microservice added to the store ecosystem.
- **Database**: New tables in Supabase PostgreSQL — `users`, `otp_codes`, managed via Prisma schema.
- **Cross-service dependency**: All other microservices will depend on the JWKS endpoint or distributed public key to validate incoming JWTs.
- **Client integration**: Frontend must handle HttpOnly cookie-based auth flow (no manual token storage), OTP input screens for registration and password reset.
- **Infrastructure**: Requires RSA/ECDSA key pair generation and secure storage of the private key (environment variable or secrets manager). Requires an SMS/Email delivery provider (e.g., Twilio, SendGrid) for OTP codes. Requires a Redis instance for distributed rate limiting.
