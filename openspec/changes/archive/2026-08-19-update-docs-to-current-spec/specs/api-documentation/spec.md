## ADDED Requirements

### Requirement: Repository Documentation and Quickstart Reference
The system repository SHALL provide comprehensive, up-to-date documentation in `README.md` covering the complete API endpoint catalog, interactive Swagger UI navigation, exhaustive environment variable configuration reference, feature-based project architecture, and core security mechanisms.

#### Scenario: Comprehensive API Endpoints Overview Table
- **WHEN** a developer inspects `README.md`
- **THEN** the API Endpoints Overview table SHALL list all public and protected HTTP routes (`GET /.well-known/jwks.json`, `POST /api/auth/register`, `POST /api/auth/verify-otp`, `POST /api/auth/login`, `POST /api/auth/refresh`, `POST /api/auth/logout`, `POST /api/auth/forgot-password`, `POST /api/auth/reset-password`, `GET /api/auth/me`, and `/docs`, `/swagger`).

#### Scenario: Full Environment Configuration Table
- **WHEN** configuring the microservice environment
- **THEN** `README.md` SHALL document all supported environment variables including database connection, Redis URL, server port, environment mode, JWT key paths and expiry durations, rate limiting thresholds, bcrypt cost, and OTP/SMTP configuration variables.

#### Scenario: Modular Project Structure Reference
- **WHEN** a developer reviews project architecture
- **THEN** `README.md` SHALL illustrate the feature-based package organization (`internal/auth`, `internal/jwt`, `internal/otp`, `internal/docs`, `internal/middleware`, `internal/platform`, `internal/router`, `internal/config`, `prisma/`).
