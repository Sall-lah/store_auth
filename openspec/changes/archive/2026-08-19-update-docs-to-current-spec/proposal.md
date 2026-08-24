## Why

The repository documentation has drifted out of sync with recent system capabilities and OpenSpec specifications. Specifically, `README.md` is missing the `/api/auth/refresh` endpoint in its API table, lacks documentation for the embedded Swagger UI (`/docs`, `/swagger`), omits several critical environment variables (e.g., token expiration durations, Redis rate limiting thresholds, bcrypt cost, and SMTP configurations), and does not detail the modular feature-based architecture. Aligning all documentation with the current specifications ensures developer clarity, accurate onboarding, and seamless integration for downstream microservices.

## What Changes

- Update `README.md` with:
  - Complete API Endpoints Overview table including `POST /api/auth/refresh` and documentation routes (`/docs`, `/swagger`, `/docs/openapi.yaml`).
  - Dedicated Interactive API Documentation section detailing Swagger UI usage.
  - Comprehensive Environment Variables table covering all configuration parameters (`DATABASE_URL`, `REDIS_URL`, `SERVER_PORT`, `ENV`, `JWT_*`, `RATE_LIMIT_*`, `BCRYPT_COST`, `OTP_*`, `SMTP_*`).
  - Feature-based project structure directory tree reflecting `internal/auth`, `internal/jwt`, `internal/otp`, `internal/docs`, `internal/middleware`, `internal/platform`, `internal/router`, and `prisma/`.
  - Architecture overview detailing RS256 token issuance, Refresh Token Rotation, Redis Token Blacklisting, and Sliding Window Rate Limiting.
- Verify and synchronize `docs/MICROSERVICE_INTEGRATION.md`, `api/openapi.yaml`, and `internal/docs/openapi.yaml` to ensure complete consistency with current endpoints, security schemas, and claims contracts.

## Capabilities

### New Capabilities
<!-- None: Documentation updates map to the existing api-documentation capability -->

### Modified Capabilities
- `api-documentation`: Expand documentation requirements to specify comprehensive repository README coverage (endpoints, Swagger UI routes, complete environment configuration table, project layout, and security architecture overview).

## Impact

- **Documentation Files**: `README.md`, `docs/MICROSERVICE_INTEGRATION.md`.
- **OpenAPI Specification**: `api/openapi.yaml`, `internal/docs/openapi.yaml`.
- **APIs & Code**: No breaking code changes or runtime behavior modifications; purely documentation and specification alignment.
