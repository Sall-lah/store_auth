## Context

`store_auth` is a Go-based authentication service utilizing RS256 asymmetric JWTs and an RFC 7517 JWKS endpoint (`/.well-known/jwks.json`). Downstream services (e.g., product catalog, orders, payments) and frontend clients require clear documentation of endpoints, payloads, error schemas, and cryptographic token verification routines.

Currently, the service only has a Postman collection. To support automated tooling, interactive documentation, and seamless microservice integration, we need an OpenAPI 3.1 specification, an embedded Swagger UI endpoint, and an integration guide.

## Goals / Non-Goals

**Goals:**
- Provide a clean, comprehensive OpenAPI 3.1 specification (`api/openapi.yaml`).
- Embed and serve Swagger UI directly from the Go binary at `GET /docs` and `GET /docs/*` using Go's `embed.FS`.
- Provide a dedicated downstream microservice integration guide (`docs/MICROSERVICE_INTEGRATION.md`) detailing JWKS caching, RS256 token verification, and copy-paste middleware recipes.
- Maintain zero external runtime dependencies and full offline capability.

**Non-Goals:**
- Rewriting existing authentication handlers or breaking existing API schemas.
- Adding heavyweight third-party documentation servers or node-based build steps.

## Decisions

### Decision 1: Canonical OpenAPI 3.1 YAML File (`api/openapi.yaml`)
- **Rationale**: OpenAPI 3.1 is the modern standard for API descriptions, supporting full JSON Schema compatibility and rich security declarations. Handcrafting/maintaining `api/openapi.yaml` allows high-precision type definitions without polluting Go handler code with repetitive annotations.
- **Alternatives Considered**:
  - *Code-comment generation (`swaggo/swag`)*: Restricted to OpenAPI 2.0/3.0, clutters Go codebase with verbose annotations.

### Decision 2: Embedded Static Swagger UI via `embed.FS` in `internal/docs`
- **Rationale**: Using Go 1.16+ `embed.FS` allows bundling Swagger UI HTML/assets directly into the compiled Go binary. The server can deliver Swagger UI and raw OpenAPI specs via standard `http.Handler` without requiring external CDN access or separate static file deployment.
- **Alternatives Considered**:
  - *External CDN iframe*: Requires constant internet access, fails in air-gapped or offline local development environments.

### Decision 3: Microservice Integration Guide in `docs/MICROSERVICE_INTEGRATION.md`
- **Rationale**: OpenAPI documents HTTP endpoints well, but does not explain decentralized token verification (JWKS public key fetching, in-memory caching, RSA signature validation). A dedicated markdown guide provides ready-to-use middleware code examples in Go and Node.js for downstream developers.

## Risks / Trade-offs

- **[Risk] Specification Drift**: Changes to Go handler models might not be reflected in `api/openapi.yaml`.
  - *Mitigation*: Router tests and schema validation unit tests to verify endpoint route parity.
- **[Risk] Binary Size Increase**: Embedding Swagger UI assets might increase Go binary size.
  - *Mitigation*: Use a lightweight, minified Swagger UI bundle (less than 1.5MB total).
