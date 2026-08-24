# rate-limiting Specification

## Purpose
TBD - Redis-backed sliding window counter rate limiting on public auth endpoints.

## Requirements

### Requirement: Public auth endpoints are rate limited via Redis
The system SHALL enforce distributed rate limiting on all public authentication endpoints (`/api/auth/register`, `/api/auth/login`, `/api/auth/verify-otp`, `/api/auth/forgot-password`, `/api/auth/reset-password`). Rate limiting SHALL use a sliding window counter algorithm backed by Redis, keyed by client IP address and endpoint path (format: `ratelimit:<ip>:<endpoint>`).

#### Scenario: Request within rate limit
- **WHEN** a client sends a request to a rate-limited endpoint and has not exceeded the allowed rate
- **THEN** the system processes the request normally

#### Scenario: Request exceeds rate limit
- **WHEN** a client sends a request to a rate-limited endpoint and has exceeded the allowed rate for their IP address and endpoint
- **THEN** the system returns HTTP 429 Too Many Requests with a `Retry-After` header indicating when the client can retry

### Requirement: Rate limit configuration
The system SHALL support configurable rate limit parameters: requests per window and window duration. These parameters SHALL be configurable via environment variables.

#### Scenario: Default rate limit values
- **WHEN** the service starts without explicit rate limit configuration
- **THEN** the system applies default rate limits of 10 requests per 1-second window per IP per endpoint

#### Scenario: Custom rate limit configuration
- **WHEN** the environment variables `RATE_LIMIT_MAX_REQUESTS` and `RATE_LIMIT_WINDOW_SECONDS` are set
- **THEN** the system uses the configured values for rate limiting

### Requirement: Redis connection for rate limiting
The system SHALL connect to a Redis instance using the `REDIS_URL` environment variable. The Redis client SHALL support unauthenticated connection URLs as well as password-authenticated and ACL-authenticated connection URLs (e.g., `redis://:password@host:port/db`, `redis://user:password@host:port/db`, and TLS `rediss://...`). The Redis client SHALL be initialized at service startup alongside the database client.

#### Scenario: Redis connection succeeds
- **WHEN** the service starts and `REDIS_URL` is set to a valid Redis connection string
- **THEN** the system connects to Redis and enables distributed rate limiting

#### Scenario: Password-authenticated Redis connection succeeds
- **WHEN** the service starts and `REDIS_URL` contains credentials (e.g., `redis://:password@localhost:6379/0`)
- **THEN** the system parses the credentials and successfully connects to the authenticated Redis instance

#### Scenario: Redis connection not configured
- **WHEN** the service starts without a `REDIS_URL` environment variable
- **THEN** the system fails to start and logs an error indicating that Redis is required

### Requirement: Rate limiting resilience on Redis failure
The system SHALL use a fail-open strategy when Redis is temporarily unreachable. If a rate limit check fails due to a Redis connection error, the request SHALL be allowed through and the failure SHALL be logged as a warning.

#### Scenario: Redis temporarily unavailable
- **WHEN** a client sends a request to a rate-limited endpoint and Redis is unreachable
- **THEN** the system allows the request through, logs a warning, and continues processing normally

#### Scenario: Redis recovers after outage
- **WHEN** Redis becomes reachable again after a temporary outage
- **THEN** the system resumes distributed rate limiting automatically on subsequent requests
