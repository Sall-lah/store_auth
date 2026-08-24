## Why

The current authentication system issues a single long-lived RS256 access token (configurable via `JWT_EXPIRY_HOURS`) delivered in an HttpOnly cookie. Because token validation is purely stateless (RSA signature + expiration check in middleware), the server cannot instantly revoke a token when an account is deleted, banned, or has its role changed. The user remains authorized until the access token naturally expires. Adding refresh token rotation with a Redis-backed user blacklist solves both problems: short-lived access tokens reduce the window of exposure, while the Redis blacklist enables instant (0-second) revocation when needed.

## What Changes

- Introduce a `RefreshToken` model in PostgreSQL (via Prisma) to store hashed refresh tokens per user with expiration and revocation tracking.
- Add `POST /api/auth/refresh` endpoint that validates the refresh token against the database, re-checks `user.isActive` and fetches the latest `user.role`, then issues a new short-lived access token and rotates the refresh token.
- Reduce access token lifetime from hours to **15 minutes**.
- Set refresh token lifetime to **7 days**, delivered in a separate HttpOnly cookie scoped to `Path=/api/auth/refresh`.
- Update `Login` handler to issue both `access_token` and `refresh_token` cookies.
- Update `Logout` handler to revoke the refresh token in the database and clear both cookies.
- Add Redis-based user blacklist (`revoked:user:<userID>`) for **instant token invalidation** when an account is deleted or banned, checked in the auth middleware on every request.
- Implement refresh token rotation with reuse detection: if a previously-used refresh token is presented, revoke all refresh tokens for that user (security breach response).

## Capabilities

### New Capabilities
- `refresh-token`: Stateful refresh token lifecycle including issuance, rotation, reuse detection, and revocation via PostgreSQL storage.
- `token-blacklist`: Redis-backed instant user-level token revocation checked in the auth middleware, with automatic TTL-based cleanup.

### Modified Capabilities
- `jwt-management`: Reduce access token expiry from configurable hours to 15 minutes. Update login to issue dual cookies (access + refresh). Update logout to clear both cookies and revoke refresh token.
- `user-login`: Login response now sets two HttpOnly cookies instead of one.

## Impact

- `prisma/schema.prisma`: Add `RefreshToken` model with `tokenHash`, `expiresAt`, `revoked`, and `User` relation.
- `internal/auth/handler.go`: Update `Login` and `Logout` handlers for dual-cookie flow. Add new `Refresh` handler.
- `internal/auth/service.go`: Add `RefreshToken`, `RevokeRefreshToken`, and `RevokeAllUserTokens` service methods.
- `internal/auth/repository.go`: Add refresh token CRUD operations.
- `internal/middleware/auth.go`: Add Redis blacklist check after RSA signature validation.
- `internal/router/router.go`: Add `POST /api/auth/refresh` route.
- `internal/config/config.go`: Add `JWTAccessExpiryMinutes`, `JWTRefreshExpiryDays` config fields.
- `internal/platform/redis/client.go`: Add blacklist helper methods (`BlacklistUser`, `IsUserBlacklisted`).
