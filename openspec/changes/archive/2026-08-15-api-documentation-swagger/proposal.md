## Why

As the microservice ecosystem expands, downstream services (such as product, order, and billing services) need a clear, standardized API contract and integration specification to interact with `store_auth` and independently validate user tokens. Introducing an OpenAPI 3.1 specification with an embedded Swagger UI and a comprehensive downstream service integration guide will eliminate integration guesswork, provide interactive endpoint testing, and enable automated SDK/DTO generation for other services.

## What Changes

- **OpenAPI 3.1 Specification**: Define a complete, standardized `openapi.yaml` contract covering all authentication endpoints, request/response models, cookie/header authentication schemes, error schemas, and the `/.well-known/jwks.json` key endpoint.
- **Embedded Swagger UI**: Mount an interactive Swagger UI dashboard directly within the `store_auth` Go server at `GET /docs` and `GET /swagger`, serving the OpenAPI document from embedded static assets with zero external dependencies.
- **Downstream Integration Documentation**: Provide architecture guidance and copy-paste code recipes for downstream microservices to consume `/.well-known/jwks.json` and validate RS256 JWTs locally in-memory.

## Capabilities

### New Capabilities
- `api-documentation`: OpenAPI 3.1 specification definition, embedded Swagger UI HTTP route serving, and downstream microservice integration guide for cryptographic token validation.

### Modified Capabilities
<!-- None: Existing authentication, registration, rate limiting, and token behavior remain backward-compatible without spec modifications. -->

## Impact

- **Affected Code**: `internal/router/router.go` (new documentation routes), new `docs/` package or static embed files for Swagger UI & OpenAPI spec.
- **APIs**: New endpoints exposed:
  - `GET /docs` & `GET /docs/` (Swagger UI HTML)
  - `GET /docs/openapi.yaml` & `GET /docs/openapi.json` (OpenAPI 3.1 spec)
- **Dependencies**: Static Swagger UI bundle or lightweight Go swagger middleware (`http.FS` / `embed`). Zero breaking changes to existing `/api/auth/*` or `/.well-known/jwks.json` routes.
