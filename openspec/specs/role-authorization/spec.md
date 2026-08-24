# role-authorization Specification

## Purpose
TBD - User role assignment and role-based authorization checking.

## Requirements

### Requirement: Users are assigned a role at registration
The system SHALL assign every new user the `CUSTOMER` role by default upon registration. The system SHALL support two roles: `CUSTOMER` and `ADMIN`.

#### Scenario: New user receives CUSTOMER role
- **WHEN** a new user completes registration
- **THEN** the user record is created with `role` set to `CUSTOMER`

#### Scenario: Admin role assignment
- **WHEN** an admin account needs to be created
- **THEN** it SHALL be created via database seeding or internal tooling (not via the public registration endpoint)

### Requirement: JWT contains user role
The system SHALL embed the user's role in the JWT payload as the `role` claim. Other microservices SHALL use this claim to enforce authorization decisions.

#### Scenario: Customer JWT claims
- **WHEN** a user with role `CUSTOMER` logs in
- **THEN** the issued JWT contains `"role": "CUSTOMER"`

#### Scenario: Admin JWT claims
- **WHEN** a user with role `ADMIN` logs in
- **THEN** the issued JWT contains `"role": "ADMIN"`

### Requirement: Auth middleware validates JWT and extracts role
The system SHALL provide a middleware that validates the JWT from the HttpOnly cookie, extracts the user claims (including role), and attaches them to the request context. Protected endpoints SHALL use this middleware.

#### Scenario: Valid JWT passes middleware
- **WHEN** a request with a valid, non-expired JWT cookie reaches a protected endpoint
- **THEN** the middleware extracts `sub`, `email`, and `role` from the JWT and attaches them to the request context for downstream handlers

#### Scenario: Missing or invalid JWT fails middleware
- **WHEN** a request without a JWT cookie or with an invalid/expired JWT reaches a protected endpoint
- **THEN** the middleware returns HTTP 401 Unauthorized and does not invoke the downstream handler

### Requirement: Role-based endpoint protection
The system SHALL support restricting endpoints to specific roles. An admin-only endpoint SHALL reject requests from users with the `CUSTOMER` role.

#### Scenario: Admin accesses admin-only endpoint
- **WHEN** an authenticated user with role `ADMIN` accesses an admin-restricted endpoint
- **THEN** the system allows the request to proceed

#### Scenario: Customer accesses admin-only endpoint
- **WHEN** an authenticated user with role `CUSTOMER` accesses an admin-restricted endpoint
- **THEN** the system returns HTTP 403 Forbidden

### Requirement: Downstream services enforce role authorization from gateway-injected role header
In an API Gateway offloaded architecture, downstream microservices SHALL read the trusted `X-User-Role` header to enforce role-based access control (RBAC) without parsing JWT signatures locally.

#### Scenario: Downstream endpoint permits authorized role from header
- **WHEN** a request reaches a role-restricted endpoint with header `X-User-Role: ADMIN`
- **THEN** the service verifies the role satisfies the permission requirement and executes the handler

#### Scenario: Downstream endpoint rejects unauthorized role from header
- **WHEN** a request reaches an admin-restricted endpoint with header `X-User-Role: CUSTOMER`
- **THEN** the service rejects the request with HTTP 403 Forbidden
