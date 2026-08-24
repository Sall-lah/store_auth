## 1. Directory Structure Setup

- [x] 1.1 Create feature module directories (`internal/auth`, `internal/otp`, `internal/jwt`) and platform directory (`internal/platform/redis`)

## 2. JWT Feature Relocation

- [x] 2.1 Create `internal/jwt/model.go` containing `Claims`, `JWK`, and `JWKS` types
- [x] 2.2 Relocate `internal/service/jwt_service.go` into `internal/jwt/service.go` with updated package declaration and imports
- [x] 2.3 Create `internal/jwt/handler.go` with `GetJWKS` HTTP handler

## 3. OTP Feature Relocation

- [x] 3.1 Create `internal/otp/model.go` containing `OTPCode`, `VerifyOTPRequest`, and `ResetPasswordRequest` types
- [x] 3.2 Create `internal/otp/sender.go` with `OTPSender` interface and `LogOTPSender` implementation
- [x] 3.3 Relocate `internal/repository/otp_repo.go` into `internal/otp/repository.go` with updated package declaration
- [x] 3.4 Relocate `internal/service/otp_service.go` into `internal/otp/service.go` using `internal/otp/repository.go` and `internal/otp/sender.go`

## 4. Auth Feature Relocation

- [x] 4.1 Create `internal/auth/model.go` containing `User`, `Role` enum, `RegisterRequest`, `LoginRequest`, `ForgotPasswordRequest`, `UserResponse`, and `ErrorResponse` types
- [x] 4.2 Relocate `internal/repository/user_repo.go` into `internal/auth/repository.go`
- [x] 4.3 Relocate `internal/service/auth_service.go` into `internal/auth/service.go` using `internal/auth/repository.go`, `*otp.Service`, and `*jwt.Service`
- [x] 4.4 Create `internal/auth/handler.go` consolidating HTTP handlers for `/register`, `/verify-otp`, `/login`, `/logout`, `/forgot-password`, `/reset-password`, and `/me`

## 5. Shared Infrastructure & Middlewares

- [x] 5.1 Relocate `internal/redis/client.go` to `internal/platform/redis/client.go`
- [x] 5.2 Update `internal/middleware/auth.go` and `internal/middleware/ratelimit.go` package imports to reference new feature types (`jwt.Service`, `auth.Role`)

## 6. Wiring & Verification

- [x] 6.1 Update `internal/router/router.go` to register routes using feature handlers (`auth.Handler`, `jwt.Handler`)
- [x] 6.2 Update `cmd/server/main.go` to wire feature services and repositories
- [x] 6.3 Remove old layered directories (`internal/handler`, `internal/service`, `internal/repository`, `internal/model`, `internal/redis`)
- [x] 6.4 Run `go build ./...` to verify clean compilation of the feature-based codebase
