# Proposal: Unverified Re-Registration, Resend OTP Endpoint, and OTP Code Nullification

## Why

Currently, when a user registers but fails to verify their 6-digit registration OTP within the 5-minute expiry window, any subsequent registration attempt fails with `409 Conflict` (`"email address is already registered"`). Because there is no endpoint to resend registration OTPs and the unverified account cannot log in (`403 Forbidden`), the user becomes permanently trapped in an unactivated state. Furthermore, OTP numeric codes are retained in plain text inside the database table after being consumed.

This change resolves the trapped account issue by allowing inactive accounts to re-register (refreshing their credentials and issuing a new OTP), introduces a dedicated self-service `POST /api/auth/resend-otp` endpoint, and nullifies OTP code secrets in persistence once consumed or invalidated.

## What Changes

- **Unverified Re-Registration Support**: If `POST /api/auth/register` is invoked with an email that exists in `users` with `is_active = false`, update the account's password hash and display name, invalidate any prior pending OTPs, issue a fresh OTP code, and respond with `200 OK` / `201 Created`.
- **Self-Service Resend OTP Endpoint**: Add `POST /api/auth/resend-otp` accepting `{ "email": "...", "type": "registration" | "password_reset" }`.
  - For `registration`: Resends activation OTP for inactive accounts, or returns `400 Bad Request` if already active.
  - For `password_reset`: Dispatches reset OTP to active accounts while shielding account existence to prevent email enumeration.
- **Identical Event Routing via Kafka**: Resend OTP events publish the exact same `auth.registration_otp` and `auth.password_reset_otp` event envelopes to the `auth.events` Kafka topic with a newly generated `event_id` (UUID), preserving full compatibility with `store_notification`.
- **OTP Secret Nullification**: Update Prisma schema to allow nullable `code String?` on `OTPCode` (or blank/delete on consumed state), clearing the raw OTP digits whenever an OTP is marked as used, replaced, or invalidated upon account activation or password reset.
- **OpenAPI Documentation**: Update OpenAPI spec and Swagger UI to document the new `POST /api/auth/resend-otp` endpoint and updated `POST /api/auth/register` behaviors.

## Capabilities

### New Capabilities
- `otp-management`: Exposes self-service OTP resend capabilities via `POST /api/auth/resend-otp` and enforces OTP code scrubbing/nullification lifecycle.

### Modified Capabilities
- `user-registration`: Updates registration behavior to support unverified account re-registration and auto-refresh of OTPs, and nullifies OTP codes upon activation.
- `password-reset`: Nullifies OTP codes upon successful password reset completion and integrates with the resend workflow.

## Impact

- **Database / Schema**: Update `prisma/schema.prisma` to make `OTPCode.code` nullable (`String?` or updated helper methods).
- **APIs**: New route `POST /api/auth/resend-otp` mounted with rate limiting.
- **Internal Modules**:
  - `internal/auth`: Update `Register`, `VerifyRegistrationOTP`, and `ResetPassword` logic.
  - `internal/otp`: Add resend orchestration and `MarkOTPUsedAndScrubCode` / `NullifyOTPCode` in repository.
  - `internal/router`: Register `POST /api/auth/resend-otp`.
  - `api/openapi.yaml` & `internal/docs`: Document new endpoint and schemas.
