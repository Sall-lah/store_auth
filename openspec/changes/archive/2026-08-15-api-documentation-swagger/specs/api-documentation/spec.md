## ADDED Requirements

### Requirement: OpenAPI 3.1 Specification Contract
The system SHALL provide an OpenAPI 3.1 specification describing all REST endpoints, request payloads, response structures, query parameters, error schemas, and security definitions for the `store_auth` service.

#### Scenario: OpenAPI Specification File Availability
- **WHEN** a client or developer requests `GET /docs/openapi.yaml` or `GET /docs/openapi.json`
- **THEN** the system SHALL return HTTP 200 with the valid OpenAPI 3.1 definition and matching `Content-Type` header (`application/yaml` or `application/json`).

#### Scenario: Comprehensive Endpoint Coverage
- **WHEN** the OpenAPI specification is inspected
- **THEN** it SHALL document `POST /api/auth/register`, `POST /api/auth/verify-otp`, `POST /api/auth/login`, `POST /api/auth/refresh`, `POST /api/auth/logout`, `POST /api/auth/forgot-password`, `POST /api/auth/reset-password`, `GET /api/auth/me`, and `GET /.well-known/jwks.json`.

#### Scenario: Security Schemes and Token Definitions
- **WHEN** inspecting the security schemes within the OpenAPI specification
- **THEN** it SHALL declare the Cookie authentication scheme (`access_token` and `refresh_token`) and Bearer JWT scheme, and document the standard error payload schema (`{"error": "string", "details": {}}`).

---

### Requirement: Interactive Swagger UI Route
The system SHALL serve an embedded Swagger UI web interface that renders the OpenAPI 3.1 specification for interactive endpoint testing and discovery.

#### Scenario: Accessing Swagger UI HTML
- **WHEN** a user navigates to `GET /docs` or `GET /docs/` via an HTTP client
- **THEN** the system SHALL return HTTP 200 with HTML content initializing Swagger UI pointed to the `store_auth` OpenAPI specification.

#### Scenario: Static Embed Integrity
- **WHEN** Swagger UI is accessed in an environment without internet access or CDN connectivity
- **THEN** the system SHALL serve all necessary assets or bundle them reliably so the documentation remains fully renderable.

---

### Requirement: Downstream Microservice Integration Reference
The system repository SHALL include a downstream microservice integration guide documenting the cryptographic verification of RS256 JWT tokens using the JWKS endpoint.

#### Scenario: Downstream Token Validation Reference
- **WHEN** a developer builds a downstream microservice
- **THEN** the documentation SHALL specify the JWT claims contract (`sub`, `email`, `role`, `exp`, `iss`), the `/.well-known/jwks.json` discovery flow, and provide implementation examples for consuming and validating tokens in downstream services.
