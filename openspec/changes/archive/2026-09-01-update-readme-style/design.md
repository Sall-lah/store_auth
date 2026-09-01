## Context

`store_auth` is a centralized identity provider and cryptographic authentication microservice written in Go. The repository's documentation needs to be upgraded to match the standardized microservice README style exemplified by `store_order`. The standardized layout provides immediate architectural clarity through badges, table of contents, Mermaid diagrams, exhaustive configuration guides, endpoint matrices, Kafka streaming contracts, and container execution instructions.

## Goals / Non-Goals

**Goals:**
- Upgrade `README.md` to the standardized microservice layout with shield badges, complete table of contents, Mermaid diagrams, feature highlights, environment configuration, database setup, endpoint tables, Kafka pipelines, Redis security rules, testing, and Docker deployment.
- Accurately document all `store_auth` capabilities: RS256 JWT signing, public JWKS endpoint (`/.well-known/jwks.json`), OTP verification, refresh token rotation, Redis blacklisting and sliding-window rate limiting, and Kafka event streaming (`auth.events` and `user.events`).
- Ensure all environment variables, routes, and architectural components accurately reflect the current codebase (`internal/config/config.go`, `internal/router/router.go`, `api/openapi.yaml`).

**Non-Goals:**
- Modifying Go application logic, routing, or database models.
- Changing OpenAPI specifications or documentation assets.

## Decisions

### Decision 1: Architecture Overview with Dual Mermaid Diagrams
- **Choice**: Include both an overall architectural flowchart (Client -> API Gateway -> Router -> Middleware -> Handlers -> DB / Redis / Kafka) and an Authentication & Token Lifecycle State Machine diagram.
- **Rationale**: Demonstrates how `store_auth` fits into the multi-service platform and clearly maps the lifecycle transitions from unverified registration through active sessions, refresh rotation, and revocation.
- **Alternatives Considered**: Static text descriptions (less intuitive for visual scanning and downstream service developers).

### Decision 2: Exhaustive Environment Configuration Table
- **Choice**: Present a structured Markdown table including Variable, Type, Default, and Description for all config keys defined in `internal/config/config.go`.
- **Rationale**: Eliminates guesswork during local onboarding and deployment.

### Decision 3: Explicit Kafka Event Schemas & Inbound/Outbound Topics
- **Choice**: Document outbound domain events (`auth.events` including `auth.otp_created`, `auth.user_registered`, `auth.user_logged_in`, `auth.password_reset`) and inbound consumer events (`user.events` for `user.banned` and `user.deleted`).
- **Rationale**: Downstream microservices (such as `store_notification` and `store_user`) require exact topic and event trigger references for integration.

## Risks / Trade-offs

- **[Risk: Drift between README and code]** → **Mitigation**: Cross-verify every documented route with `internal/router/router.go` and every env var with `internal/config/config.go`.
- **[Risk: Syntax errors in Mermaid diagrams]** → **Mitigation**: Validate Mermaid nodes and state machine transitions against standard Mermaid parser rules.
