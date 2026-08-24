## Why

As `store_auth` transitions into a multi-service microservice ecosystem, downstream services (such as orders, products, and payments) and API Gateways (NGINX, Kong, Heroku Gateway) require a standardized integration architecture. Relying solely on downstream services to parse cookies and perform local JWT verification introduces duplicate cryptographic overhead, coupling, and security risks. Transitioning to an API Gateway Offloading architecture centralizes TLS termination, CORS, rate limiting, and JWT signature verification at the perimeter gateway while forwarding verified identity headers downstream. Updating the specifications ensures the architecture, JWKS contracts, anti-spoofing rules, and dual-mode authorization (gateway offloaded vs zero-trust) are formally documented and verified.

## What Changes

- **Add Gateway Offloading Architecture**: Introduce formal specification for perimeter API Gateway offloading where the gateway validates RS256 JWT tokens using JWKS and injects verified identity headers (`X-User-Id`, `X-User-Role`, `X-User-Email`) to internal services.
- **Enforce Anti-Spoofing Header Sanitization**: Define perimeter requirements ensuring client-supplied `X-User-*` headers are stripped before upstream proxying.
- **Update JWT Management Specification**: Formalize RFC 7517 JWKS discovery endpoint contract (`/.well-known/jwks.json`) and claim-to-header mapping (`sub` -> `X-User-Id`, `role` -> `X-User-Role`, `email` -> `X-User-Email`) for gateway and service integration.
- **Update Role Authorization Specification**: Clarify role verification mechanics to support both perimeter header extraction (`X-User-Role`) and direct token context inspection.

## Capabilities

### New Capabilities
- `gateway-offloading`: Defines API Gateway perimeter authentication offloading, anti-spoofing header sanitization, routing rules for `store_auth`, and downstream identity header propagation.

### Modified Capabilities
- `jwt-management`: Extends token specification to define JWKS consumption contracts, claim mapping rules, and support for perimeter validation alongside HttpOnly cookie transport.
- `role-authorization`: Updates authorization requirements to allow role enforcement from gateway-injected identity headers in downstream services alongside internal middleware token checks.

## Impact

- **Downstream Services**: Can now authenticate incoming requests with zero cryptographic/DB overhead by reading trusted `X-User-Id`, `X-User-Role`, and `X-User-Email` headers.
- **API Gateway**: Becomes responsible for perimeter JWT validation, CORS, TLS termination, anti-spoofing header stripping, and Redis blacklist checking.
- **store_auth**: Remains the authoritative identity provider, token issuer, and JWKS publisher (`/.well-known/jwks.json`).
- **Documentation & Specs**: `docs/MICROSERVICE_INTEGRATION.md` and OpenSpec specifications are aligned with current gateway offloading architecture.
