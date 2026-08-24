## ADDED Requirements

### Requirement: Downstream services enforce role authorization from gateway-injected role header
In an API Gateway offloaded architecture, downstream microservices SHALL read the trusted `X-User-Role` header to enforce role-based access control (RBAC) without parsing JWT signatures locally.

#### Scenario: Downstream endpoint permits authorized role from header
- **WHEN** a request reaches a role-restricted endpoint with header `X-User-Role: ADMIN`
- **THEN** the service verifies the role satisfies the permission requirement and executes the handler

#### Scenario: Downstream endpoint rejects unauthorized role from header
- **WHEN** a request reaches an admin-restricted endpoint with header `X-User-Role: CUSTOMER`
- **THEN** the service rejects the request with HTTP 403 Forbidden
