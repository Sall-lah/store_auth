## Context

The `store_auth` microservice currently uses a single stateless RS256 access token (configurable expiry in hours) delivered via an HttpOnly cookie. Token validation in `middleware.Authenticate` performs only in-memory RSA signature verification and expiration checks — no database or Redis queries. Redis is present in the stack but used exclusively for API rate limiting (`internal/middleware/ratelimit.go`). The Prisma schema defines `User` and `OTPCode` models in PostgreSQL.

This design introduces a two-tier token system (short-lived access + long-lived refresh) and a Redis-backed user blacklist for instant revocation, while preserving the existing stateless validation path for normal API requests.

## Goals / Non-Goals

**Goals:**
- Reduce access token lifetime to 15 minutes to minimize exposure window.
- Enable seamless session continuity via refresh tokens (7-day lifetime) without requiring re-authentication.
- Provide instant (0-second) token revocation when a user account is deleted or banned via Redis blacklist.
- Detect refresh token reuse (stolen token replay) and revoke all user sessions as a security response.
- Keep normal API request validation fast and stateless (only adding a single Redis `EXISTS` check for blacklist).

**Non-Goals:**
- Multi-device session management UI (listing/revoking individual sessions by device).
- Sliding refresh token expiration (refresh token TTL is fixed, not extended on use).
- Access token blacklisting by individual `jti` claim (blacklist operates at user-level granularity).
- WebSocket or SSE token refresh mechanisms.

## Decisions

### Decision 1: Refresh tokens stored in PostgreSQL (not Redis)

**Choice**: Store refresh tokens as hashed records in a new `RefreshToken` Prisma model.

**Rationale**: Refresh token operations (`/auth/refresh`) happen infrequently (at most once per 15 minutes per user). PostgreSQL provides durable storage that survives Redis flushes/restarts, enables easy querying for "revoke all tokens for user", and keeps the data model co-located with the `User` table for cascading deletes.

**Alternatives considered**:
- *Redis storage*: Faster lookups but refresh operations are infrequent enough that PostgreSQL latency is negligible. Redis key expiration semantics don't easily support reuse detection or audit trails.

### Decision 2: Opaque refresh tokens (not JWT)

**Choice**: Generate refresh tokens as cryptographically random opaque strings (`crypto/rand`, 32 bytes, base64url-encoded). Store only the SHA-256 hash in the database.

**Rationale**: Refresh tokens are always validated against the database, so there is no benefit to making them self-contained JWTs. Opaque tokens are shorter, cannot be decoded by clients, and the hash-only storage pattern means a database breach does not expose usable tokens.

### Decision 3: Redis blacklist keyed by `userId` with TTL

**Choice**: On account deletion/ban, set `SETEX revoked:user:<userID> 900 "reason"` in Redis (TTL = 900 seconds = 15 minutes = access token max lifetime). The `middleware.Authenticate` function checks `EXISTS revoked:user:<userID>` after RSA signature validation.

**Rationale**: The blacklist key only needs to live as long as the longest possible unexpired access token. After 15 minutes, all previously-issued access tokens have expired naturally, and the refresh token path already checks `user.isActive` in PostgreSQL. This keeps Redis memory usage near zero.

**Alternatives considered**:
- *Go `sync.Map` in-memory cache*: Lowest latency but requires Redis Pub/Sub for multi-instance coordination. Added complexity not justified for current single-instance deployment.
- *Token version counter*: Requires reading a version number on every request from Redis anyway, similar overhead with more complexity.

### Decision 4: Refresh token rotation with reuse detection

**Choice**: Every successful `/auth/refresh` call invalidates the old refresh token and issues a new one. If a request presents a refresh token that is already marked as `revoked=true` (indicating it was already rotated), the server revokes ALL refresh tokens for that user.

**Rationale**: This pattern detects stolen-token replay attacks. If an attacker uses a stolen refresh token after the legitimate user already rotated it, the reuse triggers a full session wipe, forcing re-authentication on all devices.

### Decision 5: Separate cookie paths for access and refresh tokens

**Choice**: `access_token` cookie uses `Path=/`, `refresh_token` cookie uses `Path=/api/auth/refresh`.

**Rationale**: The browser only sends the long-lived refresh token to the refresh endpoint, never to regular API routes. This reduces the attack surface — even if a vulnerability exposes request cookies on an API endpoint, the refresh token is not included.

## Risks / Trade-offs

- **[Redis unavailability]** → If Redis is down, the blacklist check in middleware will fail. Mitigation: Treat Redis errors as "not blacklisted" (fail-open for availability) with a warning log. The refresh endpoint still checks `user.isActive` in PostgreSQL as a secondary gate.
- **[15-minute revocation gap without blacklist]** → Without the Redis blacklist, a banned user could operate for up to 15 minutes. Mitigation: The blacklist is the primary instant revocation mechanism; this gap only exists if Redis is unavailable.
- **[Database load on refresh]** → Each refresh call queries PostgreSQL for the refresh token record and user status. Mitigation: Refreshes happen at most once per 15 minutes per user, well within PostgreSQL capacity.
- **[Refresh token reuse false positives]** → Network retries could cause a legitimate client to present a rotated token. Mitigation: Implement a brief grace window (e.g., 30 seconds) where the old token is still accepted if the same IP presents it.
