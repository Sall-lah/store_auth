# api-documentation Specification

## Purpose
Standardized OpenAPI 3.1 specification, embedded interactive Swagger UI documentation dashboard, and downstream microservice cryptographic token verification reference.

## Requirements

### Requirement: OpenAPI 3.1 Specification Contract
The system SHALL provide an OpenAPI 3.1 specification describing all REST endpoints, request payloads, response structures, query parameters, error schemas, security definitions, and API Gateway server endpoints for the `store_auth` service.

#### Scenario: OpenAPI Specification File Availability
- **WHEN** a client or developer requests `GET /docs/openapi.yaml` or `GET /docs/openapi.json`
- **THEN** the system SHALL return HTTP 200 with the valid OpenAPI 3.1 definition and matching `Content-Type` header (`application/yaml` or `application/json`).

#### Scenario: Comprehensive Endpoint Coverage
- **WHEN** the OpenAPI specification is inspected
- **THEN** it SHALL document `POST /api/auth/register`, `POST /api/auth/verify-otp`, `POST /api/auth/login`, `POST /api/auth/refresh`, `POST /api/auth/logout`, `POST /api/auth/forgot-password`, `POST /api/auth/reset-password`, `GET /api/auth/me`, and `GET /.well-known/jwks.json`.

#### Scenario: Security Schemes and Token Definitions
- **WHEN** inspecting the security schemes within the OpenAPI specification
- **THEN** it SHALL declare the Cookie authentication scheme (`access_token` and `refresh_token`) and Bearer JWT scheme, and document the standard error payload schema (`{"error": "string", "details": {}}`).

#### Scenario: Multi-Server and Gateway Environment Support
- **WHEN** developers or automated tooling inspect the `servers` block in the OpenAPI specification
- **THEN** it SHALL document both the direct local development server URL and the unified API Gateway server entrypoint (e.g. production/staging gateway URL) with appropriate descriptions.

### Requirement: Interactive Swagger UI Route
The system SHALL serve an embedded Swagger UI web interface that renders the OpenAPI 3.1 specification for interactive endpoint testing and discovery.

#### Scenario: Accessing Swagger UI HTML
- **WHEN** a user navigates to `GET /docs` or `GET /docs/` via an HTTP client
- **THEN** the system SHALL return HTTP 200 with HTML content initializing Swagger UI pointed to the `store_auth` OpenAPI specification.

#### Scenario: Static Embed Integrity
- **WHEN** Swagger UI is accessed in an environment without internet access or CDN connectivity
- **THEN** the system SHALL serve all necessary assets or bundle them reliably so the documentation remains fully renderable.

### Requirement: Downstream Microservice Integration Reference
The system repository SHALL include an architectural and microservice integration guide documenting cryptographic verification of RS256 JWT tokens using the JWKS endpoint as well as API Gateway perimeter routing and header propagation.

#### Scenario: Downstream Token Validation Reference
- **WHEN** a developer builds a downstream microservice
- **THEN** the documentation SHALL specify the JWT claims contract (`sub`, `email`, `role`, `exp`, `iss`), the `/.well-known/jwks.json` discovery flow, and provide implementation examples for consuming and validating tokens in downstream services.

#### Scenario: API Gateway Architecture and Header Propagation Reference
- **WHEN** an engineer configures an API Gateway (such as NGINX or Heroku-deployed gateway) in front of the platform microservices
- **THEN** the integration guide SHALL document the architectural data flow, the anti-spoofing requirement to sanitize incoming `X-User-*` headers, the forwarded header contract (`X-User-Id`, `X-User-Role`, `X-User-Email`), and sample gateway proxy configurations.

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

