## 1. Project Initialization

- [x] 1.1 Initialize Go module (`go mod init`) and create the project directory structure (`cmd/server/`, `internal/config/`, `internal/handler/`, `internal/middleware/`, `internal/model/`, `internal/repository/`, `internal/service/`, `internal/router/`, `prisma/`)
- [x] 1.2 Install Go dependencies: `github.com/golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt`, `github.com/redis/go-redis/v9`, HTTP router (e.g., `chi` or `gorilla/mux`), and `github.com/steebchen/prisma-client-go`
- [x] 1.3 Create `.env.example` with required environment variables: `DATABASE_URL`, `REDIS_URL`, `JWT_PRIVATE_KEY_PATH`, `JWT_PUBLIC_KEY_PATH`, `JWT_EXPIRY`, `BCRYPT_COST`, `OTP_EXPIRY_MINUTES`, `OTP_MAX_ATTEMPTS`, `RATE_LIMIT_MAX_REQUESTS`, `RATE_LIMIT_WINDOW_SECONDS`, `OTP_PROVIDER`, `SERVER_PORT`
- [x] 1.4 Create `internal/config/config.go` to load and validate environment variables into a typed config struct (including `REDIS_URL`, `RATE_LIMIT_MAX_REQUESTS`, `RATE_LIMIT_WINDOW_SECONDS`)

## 2. Database Schema & Prisma Setup

- [x] 2.1 Create `prisma/schema.prisma` with the `users` table (id UUID PK, email unique, password, name, role enum CUSTOMER/ADMIN, is_active boolean default false, created_at, updated_at) and `otp_codes` table (id UUID PK, user_id FK, code string, type enum REGISTRATION/PASSWORD_RESET, attempts int default 0, expires_at datetime, used boolean default false, created_at)
- [x] 2.2 Configure Prisma datasource to use Supabase PostgreSQL connection string from `DATABASE_URL`
- [x] 2.3 Generate Prisma Go client and run initial migration against Supabase

## 3. Domain Models

- [x] 3.1 Create `internal/model/models.go` with domain types: `User`, `OTPCode`, `Role` enum (`CUSTOMER`, `ADMIN`), `OTPType` enum (`REGISTRATION`, `PASSWORD_RESET`), and request/response DTOs (`RegisterRequest`, `LoginRequest`, `VerifyOTPRequest`, `ForgotPasswordRequest`, `ResetPasswordRequest`, `UserResponse`, `ErrorResponse`)

## 4. Repository Layer

- [x] 4.1 Create `internal/repository/user_repo.go` with Prisma-backed functions: `CreateUser`, `FindUserByEmail`, `FindUserByID`, `UpdateUserPassword`, `ActivateUser`
- [x] 4.2 Create `internal/repository/otp_repo.go` with Prisma-backed functions: `CreateOTP`, `FindLatestOTPByUserAndType`, `IncrementOTPAttempts`, `MarkOTPUsed`, `InvalidateOTPsByUserAndType`

## 5. Service Layer — JWT

- [x] 5.1 Create `internal/service/jwt_service.go` with RSA key pair loading from PEM files, `GenerateToken(user) -> JWT string` using RS256 signing with claims (`sub`, `email`, `role`, `iat`, `exp`), and `ValidateToken(tokenString) -> Claims`
- [x] 5.2 Create JWKS endpoint handler: parse the RSA public key into JWK format and serve it at `GET /.well-known/jwks.json` with fields `kty`, `n`, `e`, `alg`, `kid`

## 6. Service Layer — OTP

- [x] 6.1 Create `internal/service/otp_service.go` with `GenerateOTP(userID, otpType) -> OTPCode` (generates a cryptographically random 6-digit code, stores in DB with 5-minute expiry)
- [x] 6.2 Implement `VerifyOTP(userID, code, otpType) -> error` with checks: code exists, not expired, not used, attempts < 5. Increment attempts on failure, mark as used on success
- [x] 6.3 Define OTP delivery interface `OTPSender` with method `Send(phone/email, code) error` and create a placeholder/logging implementation for development

## 7. Service Layer — Auth Business Logic

- [x] 7.1 Create `internal/service/auth_service.go` with `Register(req RegisterRequest) -> error`: validate input, check email uniqueness, hash password with bcrypt (cost 12), create user with `is_active: false`, generate OTP, send via `OTPSender`
- [x] 7.2 Implement `VerifyRegistrationOTP(req VerifyOTPRequest) -> error`: verify OTP code, activate user account (`is_active: true`)
- [x] 7.3 Implement `Login(req LoginRequest) -> (User, JWT string, error)`: find user by email, verify bcrypt hash, check `is_active`, generate JWT
- [x] 7.4 Implement `ForgotPassword(email) -> error`: find user (return success even if not found to prevent enumeration), generate PASSWORD_RESET OTP, send via `OTPSender`
- [x] 7.5 Implement `ResetPassword(req ResetPasswordRequest) -> error`: verify OTP, validate new password, hash with bcrypt, update user password, invalidate all existing OTP codes for user

## 8. Middleware

- [x] 8.1 Create `internal/redis/client.go` with Redis client initialization using `github.com/redis/go-redis/v9` and the `REDIS_URL` environment variable. Include a health check ping on startup
- [x] 8.2 Create `internal/middleware/ratelimit.go` implementing Redis-backed sliding window counter rate limiting keyed by client IP and endpoint path (key format: `ratelimit:<ip>:<endpoint>`). Use Redis `INCR` + `EXPIRE` commands. Return HTTP 429 with `Retry-After` header when exceeded. Use configurable max requests and window duration from config. Implement fail-open: if Redis is unreachable, allow the request through and log a warning
- [x] 8.3 Create `internal/middleware/auth.go` implementing JWT authentication middleware: extract `access_token` cookie, validate JWT signature and expiration, extract claims (`sub`, `email`, `role`), attach to request context. Return HTTP 401 on failure
- [x] 8.4 Create role-checking middleware/helper `RequireRole(roles ...Role)` that reads the role from request context and returns HTTP 403 if the user's role is not in the allowed list

## 9. HTTP Handlers

- [x] 9.1 Create `internal/handler/auth_handler.go` with handler for `POST /api/auth/register`: parse and validate request body, call auth service Register, return HTTP 201 or appropriate error
- [x] 9.2 Handler for `POST /api/auth/verify-otp`: parse request, call VerifyRegistrationOTP or authorize password reset depending on OTP type, return HTTP 200 or error
- [x] 9.3 Handler for `POST /api/auth/login`: parse request, call auth service Login, set HttpOnly cookie (`access_token`, Secure, SameSite=Strict, Path=/, Max-Age=3600), return HTTP 200 with user profile
- [x] 9.4 Handler for `POST /api/auth/logout`: clear auth cookie (empty value, Max-Age=0), return HTTP 200
- [x] 9.5 Handler for `POST /api/auth/forgot-password`: parse email, call ForgotPassword, always return HTTP 200 with generic message
- [x] 9.6 Handler for `POST /api/auth/reset-password`: parse request (OTP code + new password), call ResetPassword, return HTTP 200 or error
- [x] 9.7 Handler for `GET /api/auth/me`: read user claims from context (set by auth middleware), return HTTP 200 with user profile

## 10. Router & Server Entry Point

- [x] 10.1 Create `internal/router/router.go` to register all routes with appropriate middleware: rate limiter on public auth endpoints, auth middleware on protected endpoints (`/api/auth/me`), JWKS endpoint without middleware
- [x] 10.2 Create `cmd/server/main.go` entry point: load config, initialize Prisma client, initialize Redis client, wire up services/repositories/handlers, start HTTP server with graceful shutdown

## 11. Validation & Error Handling

- [x] 11.1 Create input validation helpers: email format validation, password minimum length (8 chars), required field checks. Return structured `ErrorResponse` with field-level errors
- [x] 11.2 Implement consistent error response format across all handlers: `{ "error": "message", "details": {...} }` with appropriate HTTP status codes

## 12. RSA Key Pair Generation & Documentation

- [x] 12.1 Add a script or Makefile target to generate RSA 2048-bit key pair (`private.pem` and `public.pem`) for local development
- [x] 12.2 Update `README.md` with project setup instructions: environment variables, key generation, database migration, and running the server
