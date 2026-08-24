## ADDED Requirements

### Requirement: JWT claims map to downstream identity headers
The JWT access token claim structure SHALL support deterministic mapping to downstream HTTP headers by API Gateways and microservices, mapping `sub` to `X-User-Id`, `role` to `X-User-Role`, and `email` to `X-User-Email`.

#### Scenario: Gateway extracts claims to standard identity headers
- **WHEN** the API Gateway parses a valid access token
- **THEN** the gateway successfully populates `X-User-Id` from `sub`, `X-User-Role` from `role`, and `X-User-Email` from `email`

### Requirement: Support dual-mode JWT validation architecture
The system SHALL support both perimeter Gateway offloading (where downstream services consume trusted identity headers) and zero-trust verification (where downstream services independently fetch JWKS and verify RS256 token signatures).

#### Scenario: Downstream service in zero-trust mode validates token via JWKS
- **WHEN** a microservice receives a request with an `Authorization: Bearer` token or cookie without prior gateway offloading
- **THEN** the service fetches `/.well-known/jwks.json` from `store_auth` and validates the RSA signature locally
