## MODIFIED Requirements

### Requirement: Repository Documentation and Quickstart Reference
The system repository SHALL provide comprehensive, up-to-date documentation in `README.md` structured according to the standardized microservice README specification, containing tech stack badges, table of contents, Mermaid architectural and authentication lifecycle state diagrams, key feature breakdown, technology stack catalog, repository structure, prerequisites and exhaustive environment configuration tables, database migration and Prisma ORM setup, local quickstart instructions, API endpoint and interactive Swagger documentation guides, Kafka event pipeline matrices, Redis rate limiting and token blacklist rules, automated test execution commands, and production Docker deployment workflows.

#### Scenario: Tech Stack Badges and Table of Contents
- **WHEN** a developer views the header of `README.md`
- **THEN** it SHALL render badges for Go, Chi v5, PostgreSQL, Prisma Go Client, RS256 JWT / JWKS, Apache Kafka, and Redis, accompanied by a complete Table of Contents linked to document sections.

#### Scenario: Mermaid Architectural and Authentication State Diagrams
- **WHEN** inspecting the architecture section of `README.md`
- **THEN** it SHALL present a Mermaid flowchart mapping client and API Gateway traffic through Chi router, middleware, handlers, service layers, Kafka producers/consumers, Redis rate-limiting/blacklisting, and PostgreSQL database, as well as a Mermaid state machine or sequence diagram illustrating the authentication and token lifecycle.

#### Scenario: Comprehensive API Endpoints Overview Table
- **WHEN** a developer inspects `README.md`
- **THEN** the API Endpoints Overview table SHALL list all public and protected HTTP routes (`GET /.well-known/jwks.json`, `POST /api/auth/register`, `POST /api/auth/verify-otp`, `POST /api/auth/resend-otp`, `POST /api/auth/login`, `POST /api/auth/refresh`, `POST /api/auth/logout`, `POST /api/auth/forgot-password`, `POST /api/auth/reset-password`, `GET /api/auth/me`, and `/docs`, `/swagger`) with HTTP methods, paths, required auth/headers, rate limiting status, and descriptions.

#### Scenario: Full Environment Configuration Table
- **WHEN** configuring the microservice environment
- **THEN** `README.md` SHALL document all supported environment variables including variable name, type, default value, and description for database connection, Redis URL, server port, environment mode, JWT key paths and expiry durations, rate limiting thresholds, bcrypt cost, OTP configuration, and Kafka broker/topic settings.

#### Scenario: Modular Project Structure Reference
- **WHEN** a developer reviews project architecture
- **THEN** `README.md` SHALL illustrate the feature-based package organization (`cmd/`, `api/`, `docs/`, `internal/auth`, `internal/config`, `internal/docs`, `internal/jwt`, `internal/middleware`, `internal/otp`, `internal/platform`, `internal/router`, `internal/sanitizer`, `internal/user`, `prisma/`, `scripts/`).

#### Scenario: Kafka Event Pipeline and Redis Policy Documentation
- **WHEN** reviewing asynchronous event integration and caching rules
- **THEN** `README.md` SHALL specify the outbound domain events (`auth.events`), inbound consumer events (`user.events` for `user.banned` and `user.deleted`), and the Redis sliding-window rate limiting policies and fail-open resilience rules.
