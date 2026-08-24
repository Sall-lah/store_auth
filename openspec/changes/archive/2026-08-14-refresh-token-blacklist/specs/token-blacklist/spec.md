## ADDED Requirements

### Requirement: Blacklisted users are rejected instantly on API requests
The system SHALL check a Redis-based user blacklist in the authentication middleware after RSA signature validation. If the user's ID is found in the blacklist, the request SHALL be rejected immediately with HTTP 401 Unauthorized regardless of access token validity.

#### Scenario: Blacklisted user makes API request
- **WHEN** an API request is made with a cryptographically valid access token but the user's ID exists in the Redis blacklist (`revoked:user:<userID>`)
- **THEN** the system returns HTTP 401 Unauthorized with message "Account has been revoked"

#### Scenario: Non-blacklisted user makes API request
- **WHEN** an API request is made with a valid access token and the user's ID is NOT in the Redis blacklist
- **THEN** the request proceeds normally through the middleware chain

### Requirement: Account deletion or ban triggers instant blacklisting
The system SHALL add the user's ID to the Redis blacklist with a TTL equal to the access token maximum lifetime (15 minutes) when an account is deleted or banned. The system SHALL also revoke all refresh tokens for that user in PostgreSQL.

#### Scenario: Account is deleted
- **WHEN** a user account is deleted from the system
- **THEN** the system sets `SETEX revoked:user:<userID> 900 "account_deleted"` in Redis and revokes all refresh tokens for that user in PostgreSQL

#### Scenario: Account is banned (isActive set to false)
- **WHEN** an admin sets a user's `isActive` to `false`
- **THEN** the system sets `SETEX revoked:user:<userID> 900 "account_banned"` in Redis and revokes all refresh tokens for that user in PostgreSQL

### Requirement: Blacklist is resilient to Redis unavailability
The system SHALL treat Redis connection errors during blacklist checks as "not blacklisted" (fail-open) and log a warning. The refresh endpoint's PostgreSQL-based `isActive` check serves as the secondary revocation gate.

#### Scenario: Redis is unavailable during blacklist check
- **WHEN** the middleware attempts to check the Redis blacklist but Redis is unreachable
- **THEN** the system logs a warning and allows the request to proceed (fail-open behavior)

### Requirement: Blacklist keys auto-expire
The system SHALL set a TTL on all blacklist Redis keys equal to the access token maximum lifetime (900 seconds). After the TTL expires, the key is automatically removed by Redis, as all previously-issued access tokens will have expired naturally.

#### Scenario: Blacklist key expires after TTL
- **WHEN** 900 seconds pass after a user was blacklisted
- **THEN** the Redis key `revoked:user:<userID>` is automatically deleted and no longer checked
