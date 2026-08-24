## 1. OpenAPI 3.1 Contract Definition

- [x] 1.1 Create canonical `api/openapi.yaml` specifying OpenAPI 3.1 metadata, server URLs, tags, and security schemes (Cookie auth and Bearer JWT).
- [x] 1.2 Define all endpoint paths and operations: `POST /api/auth/register`, `POST /api/auth/verify-otp`, `POST /api/auth/login`, `POST /api/auth/refresh`, `POST /api/auth/logout`, `POST /api/auth/forgot-password`, `POST /api/auth/reset-password`, `GET /api/auth/me`, and `GET /.well-known/jwks.json`.
- [x] 1.3 Define comprehensive component schemas for all request payloads (`RegisterRequest`, `LoginRequest`, `ForgotPasswordRequest`, `ResetPasswordRequest`, `VerifyOTPRequest`), response payloads (`UserResponse`, `JWKS`, `JWK`), and unified error format (`ErrorResponse`).

## 2. Embedded Swagger UI & Route Mounting

- [x] 2.1 Create `internal/docs` package embedding Swagger UI HTML, CSS, JS bundle, and `api/openapi.yaml` using Go `embed.FS`.
- [x] 2.2 Implement HTTP handlers in `internal/docs` to serve the Swagger UI interface and raw OpenAPI YAML/JSON endpoints.
- [x] 2.3 Mount `/docs`, `/docs/`, `/docs/openapi.yaml`, and `/docs/openapi.json` routes in `internal/router/router.go`.

## 3. Downstream Microservice Integration Guide

- [x] 3.1 Create `docs/MICROSERVICE_INTEGRATION.md` detailing the asymmetric RS256 token verification flow via `/.well-known/jwks.json`.
- [x] 3.2 Provide ready-to-use, copy-paste JWT authentication middleware code snippets in Go and Node.js/TypeScript for feature microservices.

## 4. Verification & Testing

- [x] 4.1 Add unit and integration tests in `internal/router/router_test.go` asserting HTTP 200 and proper content types on `/docs` and `/docs/openapi.yaml`.
- [x] 4.2 Validate full test suite execution (`go test ./...`) to ensure zero regressions across existing auth capabilities.
