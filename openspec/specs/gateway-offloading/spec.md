# gateway-offloading Specification

## Purpose
Perimeter API Gateway offloading, anti-spoofing header sanitization, subrequest authentication verification via store_auth, and downstream identity header propagation.

## Requirements

### Requirement: Gateway sanitizes incoming identity headers to prevent spoofing
The API Gateway SHALL strip any client-supplied `X-User-Id`, `X-User-Role`, and `X-User-Email` headers from incoming requests before forwarding traffic to upstream internal microservices.

#### Scenario: Client attempts header spoofing
- **WHEN** an external client sends a request containing `X-User-Role: ADMIN` or `X-User-Id: arbitrary-id`
- **THEN** the API Gateway strips these headers so they cannot reach downstream services unverified

### Requirement: Gateway offloads authentication verification to store_auth and injects verified identity headers
The API Gateway SHALL verify caller authentication tokens on protected routes by executing an internal subrequest to `${AUTH_SERVICE_URL}/api/auth/me` before forwarding the client request downstream. The subrequest SHALL forward the client's `Authorization` header, `Cookie`, `X-Original-URI`, `X-Original-Method`, and `X-Request-ID`. `store_auth` SHALL validate the RS256 token and Redis blacklist status, returning `X-User-Id`, `X-User-Role`, and `X-User-Email` in response headers. Upon HTTP 200 from the subrequest, the Gateway SHALL extract these headers and inject them into the downstream proxy request. If authentication fails, the Gateway SHALL immediately reject the request with HTTP 401 Unauthorized.

#### Scenario: Valid bearer token or cookie allows request and injects user context
- **WHEN** an external client sends a request to a protected endpoint with a valid `access_token` cookie or `Authorization: Bearer <token>`
- **THEN** the API Gateway executes an authentication subrequest to `store_auth`, receives HTTP 200 along with verified user metadata (`X-User-Id`, `X-User-Role`, `X-User-Email`), injects these headers into the downstream request, and proxies to the target feature service

#### Scenario: Expired, invalid, or missing token rejected at Gateway perimeter
- **WHEN** an external client sends a request to a protected endpoint without authentication or with an invalid/expired token
- **THEN** the `store_auth` subrequest returns HTTP 401 Unauthorized and the API Gateway immediately terminates the client request with HTTP 401 Unauthorized without invoking downstream services

#### Scenario: Revoked user rejected by store_auth subrequest via Redis blacklist
- **WHEN** a request with a valid JWT is received for a user ID present in the Redis blacklist
- **THEN** `store_auth`'s `/api/auth/me` returns HTTP 401 Unauthorized ("Account has been revoked") during the subrequest and the API Gateway terminates the request with HTTP 401 Unauthorized

### Requirement: Gateway routes store_auth endpoints and JWKS discovery
The API Gateway SHALL route `/api/auth/*`, `/.well-known/*`, and `/docs` paths directly to the `store_auth` upstream service. `store_auth` SHALL publish its RSA public key via `/.well-known/jwks.json` for consumption by external clients and zero-trust microservices.

#### Scenario: Routing auth traffic and JWKS discovery
- **WHEN** a client makes a request to `/api/auth/login` or `/.well-known/jwks.json`
- **THEN** the API Gateway proxies the request directly to the `store_auth` backend instance
