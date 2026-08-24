# Proposal: Simplify OTP Verification and Lifecycle State Handling

## Why

During integration and end-to-end testing of `store_auth` on Podman, verifying valid registration and password reset OTPs (`POST /api/auth/verify-otp`) failed with `HTTP 500 Internal Server Error`.

Investigation revealed that when an OTP is verified, `MarkOTPUsedAndScrubCode` in `internal/otp/repository.go` attempts to clear the raw code string using `db.OTPCode.Code.SetOptional(nil)`. In Prisma Client Go, passing `nil` to `SetOptional` during an `Update` call triggers an internal query engine error on PostgreSQL.

Furthermore, attempting to nullify the code string is unnecessary for lifecycle safety because:
1. `FindLatestOTPByUserAndType` strictly queries only records where `used = false`.
2. `VerifyOTP` explicitly blocks replay with `if otpRecord.Used { return ErrOTPAlreadyUsed }`.
3. The OTP code has a strict 5-minute time-to-live (`expires_at`) and 5-attempt rate-limit lockout.

This proposal simplifies the OTP consumption and invalidation lifecycle to strictly transition the state to `used = true` without modifying or clearing the `code` column, preserving the audit trail and resolving the 500 error across all OTP verification workflows.

## What Changes

- **Simplify OTP Consumption**: Update `MarkOTPUsed` in `internal/otp/repository.go` to update only `used = true`, removing `db.OTPCode.Code.SetOptional(nil)`.
- **Simplify Batch Invalidation**: Update `InvalidateOTPsByUserAndType` in `internal/otp/repository.go` to update only `used = true`, removing `db.OTPCode.Code.SetOptional(nil)`.
- **Preserve Verification Auditing**: Retain the generated 6-digit numeric string alongside `created_at`, `expires_at`, `attempts`, and `used = true` to maintain a reliable audit history of activations and password recovery requests.
- **Update OpenSpec Specifications**: Update `specs/otp-management/spec.md` to reflect the `used = true` state transition requirement instead of nullification.

## Capabilities

### Modified Capabilities
- `otp-management`: Replaces the code nullification requirement with a durable `used = true` consumption lifecycle state machine.

## Impact

- **Database / Schema**: No schema migrations required; existing `otp_codes` table and `schema.prisma` are preserved.
- **Internal Modules**:
  - `internal/otp/repository.go`: Simplify `MarkOTPUsed` and `InvalidateOTPsByUserAndType` update statements.
  - `internal/otp/otp_test.go`: Update unit tests to assert `used == true` state transition without expecting `code` to be null.
- **APIs**: `POST /api/auth/verify-otp`, `POST /api/auth/resend-otp`, `POST /api/auth/reset-password`, and `POST /api/auth/register` return `HTTP 200 OK` on valid OTP verification instead of `HTTP 500`.
