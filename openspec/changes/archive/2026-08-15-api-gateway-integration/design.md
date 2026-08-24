## Context

`store_auth` issues RS256 asymmetric JWTs and exposes public RSA keys at `GET /.well-known/jwks.json`. In production environments, client requests are routed through a reverse proxy or API Gateway (such as NGINX on Heroku, Traefik, or KrakenD) rather than hitting `store_auth` directly. 

Currently, `api/openapi.yaml` and `docs/MICROSERVICE_INTEGRATION.md` describe only direct client-to-service connections. This design establishes the technical specifications and documentation for API Gateway integration, server endpoint definitions, and identity header propagation.

## Goals / Non-Goals

**Goals:**
- Update `api/openapi.yaml` to include the API Gateway server endpoint definitions and document path mapping.
- Update `docs/MICROSERVICE_INTEGRATION.md` with end-to-end API Gateway architecture diagrams, data flows, and NGINX on Heroku deployment and configuration templates.
- Define standard downstream identity propagation headers (`X-User-Id`, `X-User-Role`, `X-User-Email`).
- Detail perimeter security and anti-spoofing requirements (stripping external client `X-User-*` headers).
- Maintain clear architectural separation: `store_auth` provides identity, token issuance, and JWKS; the API Gateway operates in an independent repository/infrastructure.

**Non-Goals:**
- Implementing the Gateway application code inside this repository (Gateway lives in its own repository).
- Modifying core Go authentication logic, database schema, or RS256 token generation routines.

## Decisions

### Decision 1: Multi-Server OpenAPI 3.1 Definition
- **Choice**: Add API Gateway server entrypoints to `api/openapi.yaml` alongside the local development server.
- **Rationale**: Downstream frontend clients and API tools can toggle between direct local development (`http://localhost:8080`) and the API Gateway endpoint (`https://api.yourdomain.com` or `https://store-gateway.herokuapp.com`).
- **Alternatives Considered**: Documenting only the Gateway URL (would break local standalone developer testing).

### Decision 2: Document Gateway Offloaded Auth & Perimeter Header Propagation
- **Choice**: Document Gateway Offloaded Authentication as the primary production pattern, with Downstream JWKS verification as an alternative for zero-trust architectures.
- **Contract**:
  - External client passes `Authorization: Bearer <token>` or HTTP-only cookies to Gateway.
  - Gateway validates JWT signature using `GET /.well-known/jwks.json`.
  - Gateway forwards sanitized headers: `X-User-Id`, `X-User-Role`, `X-User-Email` to downstream services.
  - Downstream services read standard HTTP headers without requiring JWT cryptography libraries.

### Decision 3: NGINX on Heroku Reference Architecture
- **Choice**: Include complete NGINX configuration templates and Heroku container/buildpack instructions in `docs/MICROSERVICE_INTEGRATION.md`.
- **Key Configuration Elements**:
  - Dynamic DNS resolution (`resolver 8.8.8.8 valid=30s;`) to prevent stale IP routing on Heroku dyno restarts.
  - Strict header sanitization (`proxy_set_header X-User-Id ""` on unauthenticated routes) to prevent header injection attacks.
  - Upstream route proxying for `/api/auth/*` -> `store_auth`.

## Risks / Trade-offs

- **[Risk: Header Spoofing]** An external attacker sends crafted `X-User-Id: admin-uuid` headers directly to the API Gateway.
  - **Mitigation**: Document mandatory gateway header sanitization rules in the integration guide where all external `X-User-*` headers are stripped or overwritten before forwarding to downstream services.
- **[Risk: Stale JWKS Keys]** Gateway caches public keys indefinitely during key rotation.
  - **Mitigation**: Document JWKS cache TTL (e.g., 24 hours with lazy refetch on unknown `kid`).
- **[Risk: Heroku Request Timeout]** Heroku imposes a 30-second request router limit.
  - **Mitigation**: Document fast upstream timeouts in NGINX proxy settings (`proxy_connect_timeout 5s; proxy_read_timeout 25s;`).
