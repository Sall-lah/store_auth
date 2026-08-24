## Why

The current API documentation (`api/openapi.yaml` and `docs/MICROSERVICE_INTEGRATION.md`) portrays a direct client-to-service architecture where downstream services and external clients interact directly with `store_auth` without an API Gateway. In production environments (such as cloud deployments or Heroku), traffic is routed through a dedicated API Gateway (e.g., NGINX, Traefik, KrakenD, or Kong). Introducing API Gateway integration documentation and updating the OpenAPI server and architecture contracts ensures downstream service teams, frontend developers, and DevOps engineers understand the unified gateway routing model, perimeter security, and identity header propagation.

## What Changes

- Update `api/openapi.yaml` server configurations to document the API Gateway entrypoint alongside local direct server URLs.
- Update `docs/MICROSERVICE_INTEGRATION.md` to document the API Gateway architecture patterns (Gateway Offloaded Authentication & Header Injection vs. Perimeter Reverse Proxy with downstream RS256 JWKS verification).
- Document NGINX / Heroku API Gateway configuration examples, header propagation contracts (`X-User-Id`, `X-User-Role`, `X-User-Email`), and anti-spoofing sanitization rules.
- Clarify repository boundaries (API Gateway residing in its own repository/configuration while `store_auth` serves as the identity provider).

## Capabilities

### New Capabilities
<!-- None: No completely new standalone capabilities introduced. -->

### Modified Capabilities
- `api-documentation`: Update the OpenAPI specification and downstream microservice integration guide to reflect API Gateway routing, gateway server URLs, and gateway authentication/header injection models.

## Impact

- `api/openapi.yaml`: Server definitions and routing context updated.
- `docs/MICROSERVICE_INTEGRATION.md`: Expanded with Gateway architecture diagrams, NGINX configuration recipes, and identity header propagation specifications.
- Downstream microservices can consume both direct JWKS validation and Gateway-injected headers.
