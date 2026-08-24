## ADDED Requirements

### Requirement: Gateway sanitizes incoming identity headers to prevent spoofing
The API Gateway SHALL strip any client-supplied `X-User-Id`, `X-User-Role`, and `X-User-Email` headers from incoming requests before forwarding traffic to upstream internal microservices.

#### Scenario: Client attempts header spoofing
- **WHEN** an external client sends a request containing `X-User-Role: ADMIN` or `X-User-Id: arbitrary-id`
- **THEN** the API Gateway strips these headers so they cannot reach downstream services unverified

### Requirement: Gateway validates JWT signatures and injects verified identity headers
The API Gateway SHALL validate incoming RS256 JWT access tokens (from the `access_token` cookie or `Authorization: Bearer <token>` header) using `store_auth`'s JWKS. Upon successful validation, the Gateway SHALL inject verified `X-User-Id`, `X-User-Role`, and `X-User-Email` headers into the downstream request.

#### Scenario: Valid JWT token validated by Gateway
- **WHEN** an external client sends a valid, unexpired JWT access token
- **THEN** the API Gateway extracts `sub`, `role`, and `email` claims and injects them as `X-User-Id`, `X-User-Role`, and `X-User-Email` headers before forwarding the request to downstream services

#### Scenario: Expired or invalid JWT rejected by Gateway
- **WHEN** an external client sends an invalid or expired JWT token
- **THEN** the API Gateway returns HTTP 401 Unauthorized without forwarding the request downstream

### Requirement: Gateway enforces token revocation via Redis blacklist
The API Gateway SHALL verify that the token's subject (`sub`) is not marked as revoked in Redis under key `blacklist:user:{userId}`.

#### Scenario: Revoked user rejected at Gateway perimeter
- **WHEN** a request with a valid JWT is received for a user ID present in the Redis blacklist
- **THEN** the API Gateway returns HTTP 401 Unauthorized

### Requirement: Gateway routes store_auth endpoints and JWKS discovery
The API Gateway SHALL route `/api/auth/*`, `/.well-known/*`, and `/docs` paths directly to the `store_auth` upstream service.

#### Scenario: Routing auth traffic to store_auth
- **WHEN** a client makes a request to `/api/auth/login` or `/.well-known/jwks.json`
- **THEN** the API Gateway proxies the request directly to the `store_auth` backend instance
