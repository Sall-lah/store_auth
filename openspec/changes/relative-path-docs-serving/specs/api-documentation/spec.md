## MODIFIED Requirements

### Requirement: Interactive Swagger UI Route
The system SHALL serve an embedded Swagger UI web interface that renders the OpenAPI 3.1 specification for interactive endpoint testing and discovery.

#### Scenario: Accessing Swagger UI HTML
- **WHEN** a user navigates to GET /docs or GET /docs/ via an HTTP client
- **THEN** the system SHALL return HTTP 200 with HTML content initializing Swagger UI configured to resolve the store_auth OpenAPI specification relatively based on the current URL path.

#### Scenario: Reverse Proxy and Gateway Subpath Compatibility
- **WHEN** a user accesses Swagger UI under a prefixed reverse proxy subpath or with/without a trailing slash (e.g. /docs, /docs/, /auth/docs, /auth/docs/)
- **THEN** the browser SHALL dynamically resolve the specification URL to the corresponding openapi.yaml without stripping the URL subpath prefix.

#### Scenario: Static Embed Integrity
- **WHEN** Swagger UI is accessed in an environment without internet access or CDN connectivity
- **THEN** the system SHALL serve all necessary assets or bundle them reliably so the documentation remains fully renderable.
