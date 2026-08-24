# Design: Redis Authentication Environment Configuration

## Context

`store_auth` uses Redis for token revocation blacklisting and sliding window rate limiting. The Redis connection is managed through `internal/platform/redis/client.go` using `redis.ParseURL(redisURL)`.

While `redis.ParseURL` supports connection URLs containing credentials (such as passwords and ACL usernames), existing environment templates only provide an unauthenticated local example (`redis://localhost:6379`). Users configuring password-protected Redis instances for local development or staging require clear examples and instructions, including special character URL encoding.

## Goals / Non-Goals

**Goals:**
- Provide clear syntax examples in `.env.example` for local Redis instances with password authentication (`redis://:password@host:port/db`) and ACL users (`redis://username:password@host:port/db`).
- Document URL percent-encoding rules in `.env.example` and `.env.production.example` for passwords containing special characters (e.g., `@`, `:`, `#`, `%`).
- Update `README.md` to reflect authenticated Redis connection options.

**Non-Goals:**
- Refactoring `internal/platform/redis` to use discrete environment variables (`REDIS_HOST`, `REDIS_PASSWORD`), as `REDIS_URL` standard is sufficient and 12-factor compliant.
- Changing Redis driver libraries or connection lifecycle.

## Decisions

### Decision 1: Maintain Standard `REDIS_URL` Scheme
- **Rationale**: `redis.ParseURL` from `go-redis/v9` supports RFC-compliant Redis URI schemes (`redis://` and `rediss://`), supporting auth, db selection, and TLS out of the box.
- **Alternatives Considered**:
  - *Discrete Env Vars (`REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`)*: Adds unnecessary configuration complexity and parsing logic when standard URI connection strings are universally supported.

## Risks / Trade-offs

- **[Special Characters in Passwords]** → Complex passwords with `@` or `:` can break URI parsing if not percent-encoded.
  - *Mitigation*: Explicitly document URL-encoding guidance directly in `.env.example`, `.env.production.example`, and `README.md`.
