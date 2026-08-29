## Why

Currently, Swagger UI documentation in `internal/docs/index.html` hardcodes an absolute spec URL (`/docs/openapi.yaml`). When `store_auth` is deployed behind a reverse proxy or API gateway under a path prefix (e.g. `/auth/docs` or `/api/v1/auth/docs`), the browser requests `/docs/openapi.yaml` from the root domain, stripping the gateway path prefix and resulting in HTTP 404 errors. Serving the documentation with relative path resolution ensures Swagger UI functions seamlessly in standalone, containerized, and subpath-proxied environments.

## What Changes

- Update Swagger UI initialization in `internal/docs/index.html` to dynamically resolve the OpenAPI YAML specification path relative to the current URL path.
- Support both trailing slash (`/docs/`) and non-trailing slash (`/docs`) access patterns, as well as prefixed proxy subpaths (e.g. `/auth/docs`), resolving correctly to `openapi.yaml`.
- Update unit tests in `internal/docs/handler_test.go` to assert that served HTML contains dynamic relative spec resolution logic.

## Capabilities

### New Capabilities
<!-- None -->

### Modified Capabilities
- `api-documentation`: Refine Swagger UI HTML serving scenario to require relative specification resolution to support reverse proxy path prefixes and trailing slash variations.

## Impact

- **Affected code**: `internal/docs/index.html`, `internal/docs/handler_test.go`.
- **APIs**: No breaking changes to existing endpoints. `/docs`, `/docs/`, and `/docs/openapi.yaml` retain their current routing and response formats.
- **Dependencies**: No external library additions.
