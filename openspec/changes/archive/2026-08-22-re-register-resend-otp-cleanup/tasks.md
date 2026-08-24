## 1. Database & Schema Updates

- [x] 1.1 Update `prisma/schema.prisma` to make `OTPCode.code` nullable (`String?` or support clearing) and run `prisma db push` / generate client
- [x] 1.2 Update `internal/otp/repository.go` and `internal/otp/model.go` to support scrubbing/nullifying OTP codes upon verification, replacement, and invalidation

## 2. Unverified Re-Registration Support

- [x] 2.1 Update `internal/auth/repository.go` to add unverified user update method (`UpdateUnverifiedUserCredentials`)
- [x] 2.2 Update `internal/auth/service.go` `Register` logic to handle inactive accounts (`is_active: false`) by updating credentials, scrubbing prior OTPs, and issuing a new registration OTP
- [x] 2.3 Update `internal/auth/service.go` `VerifyRegistrationOTP` to scrub the OTP code upon successful activation

## 3. Resend OTP Service & Endpoint

- [x] 3.1 Define `ResendOTPRequest` in `internal/auth/model.go` with request validation
- [x] 3.2 Implement `ResendOTP` in `internal/auth/service.go` handling both `registration` (unverified accounts) and `password_reset` (active accounts with anti-enumeration shielding)
- [x] 3.3 Implement `ResendOTP` handler in `internal/auth/handler.go` with sanitization and status code responses
- [x] 3.4 Mount `POST /api/auth/resend-otp` in `internal/router/router.go` inside the rate-limited group

## 4. Password Reset OTP Scrubbing

- [x] 4.1 Update `internal/auth/service.go` `ResetPassword` to scrub the verified OTP code and invalidate/scrub all remaining reset OTPs for that user

## 5. API Documentation & OpenAPI Specification

- [x] 5.1 Update `api/openapi.yaml` and `internal/docs/openapi.yaml` to document `POST /api/auth/resend-otp` endpoint, schemas, and updated registration status codes

## 6. Testing & Verification

- [x] 6.1 Add unit tests covering unverified re-registration, resend OTP for registration and password reset, and OTP code scrubbing in `internal/auth` and `internal/otp`
- [x] 6.2 Execute full test suite (`go test ./...`) and verify all tests pass
