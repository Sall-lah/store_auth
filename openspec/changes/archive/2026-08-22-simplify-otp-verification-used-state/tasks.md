# Tasks: Simplify OTP Verification and State Management

- [x] **1. Refactor OTP Repository Update Queries** <!-- id: 1-refactor-otp-repo -->
  - [x] Update `MarkOTPUsedAndScrubCode` / `MarkOTPUsed` in `internal/otp/repository.go` to set only `used = true`, removing `db.OTPCode.Code.SetOptional(nil)`.
  - [x] Update `InvalidateOTPsByUserAndType` in `internal/otp/repository.go` to set only `used = true`, removing `db.OTPCode.Code.SetOptional(nil)`.

- [x] **2. Update Unit Tests** <!-- id: 2-update-unit-tests -->
  - [x] Verify and update `internal/otp/otp_test.go` and `internal/auth/auth_test.go` to validate `used = true` transitions.
  - [x] Run `go test ./...` to ensure all package unit tests pass.

- [x] **3. Container Build & End-to-End Verification on Podman** <!-- id: 3-container-verification -->
  - [x] Rebuild and restart `store_auth_app` on Podman.
  - [x] Run the complete automated end-to-end PowerShell test suite (`run_tests.ps1` / `test_e2e.go`) covering registration, OTP verification, login, `/me` profile inspection, token rotation, forgot password, reset password, logout, rate limiting, and gateway routing.
  - [x] Verify 100% test success rate across all microservice endpoints.
