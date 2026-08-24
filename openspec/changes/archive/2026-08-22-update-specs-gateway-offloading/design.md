## Context

`store_auth` provides identity, registration, login, JWT token issuance, and OTP flows in an e-commerce microservices platform. As the system scales, downstream services (`order-service`, `product-service`, `payment-service`) require authenticated caller context.

Previously, microservices were expected to either accept direct tokens or handle authentication redundantly. With API Gateway Offloading (e.g., NGINX, Heroku Gateway, Kong), the gateway acts as the single public perimeter entrypoint that terminates TLS, handles CORS, sanitizes headers to prevent spoofing, validates JWT access tokens against the RS256 JWKS exposed by `store_auth`, and injects verified identity headers (`X-User-Id`, `X-User-Role`, `X-User-Email`) into downstream HTTP requests.

This design aligns `store_auth` specifications with this production gateway offloading pattern while preserving zero-trust direct verification as a secondary fallback.

## Goals / Non-Goals

**Goals:**
- Document and formalize the API Gateway Offloading pattern across specifications and contracts.
- Define identity header injection standards (`X-User-Id`, `X-User-Role`, `X-User-Email`) and anti-spoofing header stripping rules at the gateway perimeter.
- Specify the JWKS discovery contract (`GET /.well-known/jwks.json`) and RS256 token validation mechanics.
- Support dual-mode authentication: Pattern A (Gateway Offloading with trusted headers) for production microservices and Pattern B (Zero-Trust standalone JWKS verification) for direct calls or perimeter pass-through.

**Non-Goals:**
- Removing JWT signing or JWKS capabilities from `store_auth` (the gateway relies on `store_auth`'s JWKS).
- Modifying `store_auth` internal database schema or auth endpoint paths (`/api/auth/*`).
- Mandating a specific gateway vendor (architecture is compatible with NGINX, Kong, Envoy, Traefik, AWS API Gateway, etc.).

## Decisions

### Decision 1: Gateway Ingress Header Mapping Contract
- **Decision**: In the gateway offloading model, the perimeter gateway maps verified JWT claims to HTTP headers:
  - `claims.sub` -> `X-User-Id` (UUID string)
  - `claims.role` -> `X-User-Role` (`"CUSTOMER"` or `"ADMIN"`)
  - `claims.email` -> `X-User-Email` (string)
- **Rationale**: Downstream services can authenticate requests in $O(1)$ without JWT parsing libraries, public key fetches, or database lookups.
- **Alternatives Considered**:
  - *Internal token exchange / mTLS with client certs*: High operational complexity; excessive overhead for internal service-to-service communication within private VPCs.
  - *Downstream services query store_auth on every request*: Introduces severe latency bottlenecks and single points of failure.

### Decision 2: Anti-Spoofing Perimeter Header Stripping
- **Decision**: The API Gateway MUST strip incoming `X-User-Id`, `X-User-Role`, and `X-User-Email` headers from client requests before forwarding them to any upstream service.
- **Rationale**: Prevents external clients from forging user identities or escalating privileges to `ADMIN` by sending custom headers directly.

### Decision 3: Public Key Discovery via Standard RFC 7517 JWKS
- **Decision**: `store_auth` exposes its RSA public key at `GET /.well-known/jwks.json` with standard fields (`kty`, `alg`, `use`, `kid`, `n`, `e`).
- **Rationale**: Gateways and zero-trust downstream microservices can dynamically fetch and cache public keys, enabling seamless key rotation without redeploying downstream services.

### Decision 4: Instant Revocation via Shared Redis Blacklist
- **Decision**: On user suspension, logout, or password reset, `store_auth` writes to Redis key `blacklist:user:{userId}` with a 15-minute TTL. The perimeter gateway checks this Redis key during token validation.
- **Rationale**: Provides real-time revocation without converting JWT into a stateful database token.

## Risks / Trade-offs

- **[Risk] Downstream services exposed directly to public internet without gateway**
  - *Mitigation*: Internal services MUST be deployed in a private VPC network accessible only via the API Gateway, or use zero-trust JWT validation middleware (Pattern B) if exposed.
- **[Risk] Header injection / spoofing if gateway misconfigured**
  - *Mitigation*: Gateway configuration template explicitly includes global header stripping (`proxy_set_header X-User-* ""` in NGINX) before proxying.
- **[Risk] Gateway cache out-of-sync during RSA key rotation**
  - *Mitigation*: JWKS keys include unique `kid` identifiers. Gateways invalidate or refresh cached JWKS when encountering an unrecognized `kid`.
