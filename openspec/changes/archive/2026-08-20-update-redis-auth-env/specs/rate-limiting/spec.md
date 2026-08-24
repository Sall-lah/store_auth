## MODIFIED Requirements

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
