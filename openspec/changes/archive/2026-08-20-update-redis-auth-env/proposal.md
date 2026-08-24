# Proposal: Update Redis Authentication Environment Examples and Documentation

## Why

Local development and production environments frequently require Redis instances configured with passwords or ACL credentials. While the application's Redis client natively supports authenticated URLs via `redis.ParseURL`, the existing `.env.example`, `.env.production.example`, and `README.md` lack explicit syntax examples and URL encoding guidance for password-protected Redis setups.

## What Changes

- Update `.env.example` to include commented examples for password-protected Redis (`redis://:password@localhost:6379/0`) and ACL credentials (`redis://user:password@localhost:6379/0`).
- Add documentation notes in `.env.example` and `.env.production.example` explaining percent-encoding requirements for special characters in Redis passwords.
- Update `README.md` environment variable table with the expanded `REDIS_URL` format description.

## Capabilities

### New Capabilities
<!-- None -->

### Modified Capabilities
- `rate-limiting`: Document and verify support for authenticated Redis connection URLs.

## Impact

- Affected files: `.env.example`, `.env.production.example`, `README.md`, `openspec/specs/rate-limiting/spec.md`.
- No breaking changes or Go code modifications required.
