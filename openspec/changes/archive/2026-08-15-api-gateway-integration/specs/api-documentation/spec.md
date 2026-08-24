## MODIFIED Requirements

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

### Requirement: Downstream Microservice Integration Reference
The system repository SHALL include an architectural and microservice integration guide documenting cryptographic verification of RS256 JWT tokens using the JWKS endpoint as well as API Gateway perimeter routing and header propagation.

#### Scenario: Downstream Token Validation Reference
- **WHEN** a developer builds a downstream microservice
- **THEN** the documentation SHALL specify the JWT claims contract (`sub`, `email`, `role`, `exp`, `iss`), the `/.well-known/jwks.json` discovery flow, and provide implementation examples for consuming and validating tokens in downstream services.

#### Scenario: API Gateway Architecture and Header Propagation Reference
- **WHEN** an engineer configures an API Gateway (such as NGINX or Heroku-deployed gateway) in front of the platform microservices
- **THEN** the integration guide SHALL document the architectural data flow, the anti-spoofing requirement to sanitize incoming `X-User-*` headers, the forwarded header contract (`X-User-Id`, `X-User-Role`, `X-User-Email`), and sample gateway proxy configurations.
