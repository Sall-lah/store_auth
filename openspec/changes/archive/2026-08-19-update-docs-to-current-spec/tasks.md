## 1. Update Root README.md

- [x] 1.1 Update API Endpoints Overview table to include `POST /api/auth/refresh`, documentation routes (`/docs`, `/swagger`, `/docs/openapi.yaml`), and clear auth requirements.
- [x] 1.2 Add an Interactive API Documentation section detailing Swagger UI access at `/docs` and `/swagger`.
- [x] 1.3 Expand the Environment Variables reference table to cover all supported settings (`DATABASE_URL`, `REDIS_URL`, `SERVER_PORT`, `ENV`, `JWT_*`, `RATE_LIMIT_*`, `BCRYPT_COST`, `OTP_*`, `SMTP_*`).
- [x] 1.4 Add the Feature-Based Project Structure directory tree outlining `internal/auth`, `internal/jwt`, `internal/otp`, `internal/docs`, `internal/middleware`, `internal/platform`, `internal/router`, and `prisma/`.
- [x] 1.5 Add Core Security & Architecture section detailing RS256 token signing, refresh token rotation, Redis token blacklisting, sliding window rate limiting, and cookie security.

## 2. Review and Synchronize Supporting Documentation

- [x] 2.1 Review and align `docs/MICROSERVICE_INTEGRATION.md` with current claims, headers, JWKS discovery, and token rotation policies.
- [x] 2.2 Verify schema parity between `api/openapi.yaml` and embedded `internal/docs/openapi.yaml`.

## 3. Verification

- [x] 3.1 Verify all Markdown links, code blocks, environment variable tables, and endpoint tables for accuracy and consistency.
