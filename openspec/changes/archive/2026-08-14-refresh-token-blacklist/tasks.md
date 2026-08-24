## 1. Database Schema & Migration

- [x] 1.1 Add `RefreshToken` model to `prisma/schema.prisma` with fields: `id` (UUID), `userId` (UUID FK to User), `tokenHash` (unique string), `expiresAt` (DateTime), `revoked` (Boolean default false), `createdAt` (DateTime). Add `refreshTokens RefreshToken[]` relation to `User` model.
- [x] 1.2 Run `go run github.com/steebchen/prisma-client-go generate` to regenerate the Prisma Go client.
- [x] 1.3 Run `go run github.com/steebchen/prisma-client-go db push` to apply the schema to the database.

## 2. Configuration Updates

- [x] 2.1 Add `JWTAccessExpiryMinutes` (default 15) and `JWTRefreshExpiryDays` (default 7) fields to `internal/config/config.go`, loaded from `JWT_ACCESS_EXPIRY_MINUTES` and `JWT_REFRESH_EXPIRY_DAYS` environment variables.
- [x] 2.2 Update `.env.example` and `.env.production.example` with the new environment variables.
- [x] 2.3 Update `cmd/server/main.go` to pass `JWTAccessExpiryMinutes` (instead of `JWTExpiryHours`) to `jwt.NewService`.

## 3. JWT Service Changes

- [x] 3.1 Update `internal/jwt/service.go` `NewService` to accept expiry in minutes instead of hours. Update the `expiry` field calculation accordingly.

## 4. Redis Blacklist Implementation

- [x] 4.1 Add `BlacklistUser(ctx, userID, ttl, reason)` method to `internal/platform/redis/client.go` that executes `SETEX revoked:user:<userID> <ttl> <reason>`.
- [x] 4.2 Add `IsUserBlacklisted(ctx, userID)` method to `internal/platform/redis/client.go` that executes `EXISTS revoked:user:<userID>`. Return `false` on Redis connection errors (fail-open) with a warning log.
- [x] 4.3 Update `internal/middleware/auth.go` `Authenticate` to accept a Redis client parameter and check `IsUserBlacklisted` after RSA signature validation. If blacklisted, return HTTP 401 with `"Account has been revoked"`.

## 5. Refresh Token Repository

- [x] 5.1 Create `internal/auth/refresh_repository.go` with methods: `CreateRefreshToken(ctx, userID, tokenHash, expiresAt)`, `FindRefreshTokenByHash(ctx, tokenHash)`, `RevokeRefreshToken(ctx, tokenID)`, `RevokeAllUserRefreshTokens(ctx, userID)`.

## 6. Auth Service Updates

- [x] 6.1 Add refresh token generation logic to `internal/auth/service.go`: generate 32 random bytes, base64url-encode as opaque token, compute SHA-256 hash, store hash via `refresh_repository.go`.
- [x] 6.2 Add `RefreshToken(ctx, rawToken)` method to `internal/auth/service.go`: hash the incoming token, look up in DB, check if revoked (trigger reuse detection if already revoked), verify user is active, fetch latest user role, rotate token (revoke old + issue new), generate new access token with updated claims.
- [x] 6.3 Add `RevokeRefreshToken(ctx, rawToken)` method for logout flow: hash the incoming token and mark as revoked in DB.
- [x] 6.4 Add `RevokeAllUserTokens(ctx, userID)` method: revoke all refresh tokens for the user in DB and blacklist the user in Redis with 900-second TTL.

## 7. Auth Handler Updates

- [x] 7.1 Update `Login` handler in `internal/auth/handler.go` to set both `access_token` cookie (`Path=/`, `MaxAge=900`) and `refresh_token` cookie (`Path=/api/auth/refresh`, `MaxAge=604800`).
- [x] 7.2 Add `Refresh` handler in `internal/auth/handler.go`: extract `refresh_token` from cookie, call `service.RefreshToken`, set new dual cookies on success, clear both cookies and return 401/403 on failure.
- [x] 7.3 Update `Logout` handler in `internal/auth/handler.go`: extract `refresh_token` from cookie, call `service.RevokeRefreshToken`, clear both `access_token` and `refresh_token` cookies.

## 8. Router Updates

- [x] 8.1 Add `POST /api/auth/refresh` route in `internal/router/router.go` (public, rate-limited).
- [x] 8.2 Update `middleware.Authenticate` call in router to pass Redis client for blacklist checking.

## 9. Verification

- [x] 9.1 Verify login returns both `access_token` and `refresh_token` cookies with correct attributes.
- [x] 9.2 Verify `POST /api/auth/refresh` issues new access token and rotates refresh token.
- [x] 9.3 Verify refresh with expired or revoked token returns 401.
- [x] 9.4 Verify refresh with inactive user returns 403 and revokes all tokens.
- [x] 9.5 Verify reuse of a rotated refresh token revokes all user tokens.
- [x] 9.6 Verify blacklisted user is rejected instantly on API requests.
- [x] 9.7 Verify logout clears both cookies and revokes refresh token.
