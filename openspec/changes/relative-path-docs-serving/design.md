## Context

store_auth embeds Swagger UI in internal/docs/index.html to provide interactive API documentation. Previously, the Swagger UI bundle was configured with url: /docs/openapi.yaml. When accessed directly on http://localhost:8080/docs, this succeeds. However, when deployed in containerized environments, Kubernetes ingress, or API Gateway setups where requests are mapped under a subpath prefix (e.g. https://gateway/auth/docs), the browser requests /docs/openapi.yaml from the root domain, stripping /auth and failing with HTTP 404.

## Goals / Non-Goals

**Goals:**
- Dynamically resolve the OpenAPI specification YAML path relative to the active URL location in internal/docs/index.html.
- Support both trailing slash (/docs/, /swagger/) and non-trailing slash (/docs, /swagger) URL paths.
- Support arbitrary reverse proxy path prefixes (e.g., /auth/docs, /api/v1/auth/docs) with zero configuration.
- Preserve existing Go router behavior, HTTP status codes (200 OK on GET /docs), and existing unit tests.

**Non-Goals:**
- Introducing server-side redirects (e.g. 301 from /docs to /docs/) which can conflict with reverse proxy path rewriting.
- Modifying the embedded OpenAPI 3.1 YAML content or route definitions.

## Decisions

### Decision 1: Dynamic Client-Side Path Derivation in index.html
- **Rationale**: Browser URL resolution differs depending on whether a URL contains a trailing slash. If a static relative URL like ./openapi.yaml is used, visiting /docs resolves to /openapi.yaml (404), while visiting /docs/ resolves to /docs/openapi.yaml. By inspecting window.location.pathname at runtime:
  `javascript
  const currentPath = window.location.pathname.replace(/\/+$/, '');
  const basePath = currentPath.replace(/\/(docs|swagger)$/, '');
  const specUrl = (basePath || '') + '/docs/openapi.yaml';
  `
  or:
  `javascript
  const specUrl = window.location.pathname.endsWith('/') ? './openapi.yaml' : './docs/openapi.yaml';
  `
  Swagger UI will always resolve to the correct path regardless of proxy prefix or trailing slash.
- **Alternatives Considered**:
  - *Static relative string (./openapi.yaml)*: Fails when accessed without trailing slash (/docs).
  - *Backend HTTP Redirect*: Adding an HTTP 301 redirect in Go from /docs to /docs/ would alter the contract and break existing router unit tests.

## Risks / Trade-offs

- **[Risk] Reliance on Client JavaScript**: Spec URL resolution happens inside window.onload.
  - *Mitigation*: Swagger UI is an entirely client-side JavaScript application; it already requires JavaScript to fetch the spec and mount the DOM.
