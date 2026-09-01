## Why

The `store_auth` microservice documentation currently uses a basic format that lacks the detailed architectural clarity, visual flow diagrams, comprehensive configuration reference, and standardized presentation established across the platform's microservices (such as `store_order`). 

Updating `README.md` to the standardized style improves developer onboarding, clarifies cryptographic token verification & JWKS discovery for downstream service maintainers, documents Kafka event schemas and Redis rate limiting policies, and provides clear local development and Docker deployment instructions.

## What Changes

- **Badges & Header**: Add standard shields for Go 1.25+, Chi v5, PostgreSQL, Prisma Client Go, RS256 JWT / JWKS, Apache Kafka, and Redis.
- **Table of Contents**: Add structured navigation with standard emoji anchors.
- **Architecture & Sequence Diagrams**: Include Mermaid architecture diagram illustrating API Gateway proxying, token verification, Redis rate-limiting/blacklisting, Kafka streaming, and database persistence, as well as an Authentication & Token Lifecycle State Machine diagram.
- **Key Features**: Highlight RS256 asymmetric signing, dual transport (Cookie/Bearer), single-use refresh token rotation, Redis sliding-window rate limiting, Kafka user event handling (`user.banned`, `user.deleted`), and interactive OpenAPI/Swagger docs.
- **Repository Structure**: Update ASCII directory tree to match the exact file layout of `store_auth`.
- **Environment Variable Catalog**: Expand `.env` configuration table to document all settings, defaults, types, and descriptions.
- **Database & Prisma Guide**: Detail Prisma schema migration and Go client generation steps.
- **API Endpoint Catalog**: Add comprehensive endpoint reference table with HTTP methods, paths, auth headers, rate limiting flags, and descriptions.
- **Kafka Pipeline & Redis Rate Limiting**: Document inbound (`user.events`) and outbound (`auth.events`) event contracts, as well as sliding-window rate limits and fail-open behavior.
- **Testing & Docker Deployment**: Document test execution commands and Docker multi-stage container build and execution with RSA key mounts.

## Capabilities

### New Capabilities

### Modified Capabilities
- `api-documentation`: Update repository documentation requirements to mandate the standardized microservice README style (badges, Mermaid architecture diagrams, detailed env var tables, Kafka event matrix, and rate limiting rules).

## Impact

- `README.md`: Overhauled to follow the standardized microservice structure.
- Developer experience and downstream microservice integration documentation improved with clear architectural diagrams and endpoint references.
- No application code or runtime behavior is altered.
