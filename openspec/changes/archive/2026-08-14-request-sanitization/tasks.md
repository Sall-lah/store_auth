## 1. Input Sanitizer Package

- [x] 1.1 Create `internal/sanitizer/sanitizer.go` with functions for `NormalizeEmail`, `SanitizeName`, and `SanitizeCode`
- [x] 1.2 Write unit tests in `internal/sanitizer/sanitizer_test.go` verifying whitespace trimming, lowercase conversion, HTML tag stripping, and control character removal

## 2. Request Body Limit Middleware

- [x] 2.1 Create request body max size reader helper/middleware in `internal/middleware/max_body.go` (1 MB limit)
- [x] 2.2 Wire payload size limit into incoming HTTP handler requests in `internal/auth/handler.go`

## 3. Auth Handlers Integration & Verification

- [x] 3.1 Update `Register` handler to sanitize `Email`, `Name`, and restrict request body size
- [x] 3.2 Update `VerifyOTP` handler to sanitize `Email`, `Code`, and restrict request body size
- [x] 3.3 Update `Login` handler to sanitize `Email`, and restrict request body size
- [x] 3.4 Update `ForgotPassword` and `ResetPassword` handlers to sanitize `Email`, `Code`, and restrict request body size
- [x] 3.5 Run `go test ./...` to verify all unit tests pass cleanly
