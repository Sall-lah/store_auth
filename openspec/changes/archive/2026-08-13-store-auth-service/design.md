## Context

The web store platform uses a microservices architecture. Currently, no authentication service exists. This service (`store_auth`) will be the centralized identity provider for the entire ecosystem.

**Tech Stack:**
- **Language:** Go
- **ORM:** Prisma (`prisma-client-go`)
- **Database:** Supabase (PostgreSQL)
- **Password Hashing:** bcrypt
- **Token Format:** JWT (RS256)
- **Token Transport:** HttpOnly, Secure, SameSite cookies
- **Rate Limiting:** Redis-backed sliding window counter
- **2FA:** SMS/Email OTP (6-digit numeric code)

**Current State:** Greenfield — no existing auth code, database tables, or user accounts.

## Goals / Non-Goals

**Goals:**
- Provide secure email/password registration and login.
- Verify user identity via SMS/Email OTP during registration and password reset.
- Issue short-lived RS256-signed JWT access tokens via HttpOnly cookies.
- Expose a JWKS endpoint (`/.well-known/jwks.json`) so other microservices can verify tokens without contacting the Auth service.
- Rate-limit all public authentication endpoints to prevent brute-force attacks.
- Support two roles (`CUSTOMER`, `ADMIN`) embedded as JWT claims.

**Non-Goals:**
- OAuth/social login providers (Google, GitHub, etc.) — out of scope for this change.
- TOTP/Authenticator app support — only SMS/Email OTP.
- User profile management beyond core auth fields (name, email, role).
- Refresh token rotation or database-backed session management — pure stateless JWT.
- Admin user management UI — admin accounts are seeded or created via internal tooling.
- Email/SMS provider implementation — the service will define an interface; the concrete provider is injected via configuration.

## Decisions

### Decision 1: Asymmetric JWT Signing (RS256)

**Choice:** RS256 (RSA 2048-bit key pair) for JWT signing.

**Rationale:** In a microservices architecture, only the Auth service needs the private key to sign tokens. All other services verify tokens using the public key fetched from the JWKS endpoint. This eliminates shared secret distribution (HS256) and prevents any non-Auth service from forging tokens.

**Alternatives Considered:**
- **HS256 (Shared Secret):** Simpler but requires distributing the same secret to every service. Any compromised service can forge tokens for the entire system.
- **ES256 (ECDSA):** Smaller key/signature size and faster verification. Viable alternative, but RS256 has broader library support in the Go ecosystem and is the industry default for JWKS.

### Decision 2: Redis-Backed Distributed Rate Limiting

**Choice:** Sliding window counter algorithm backed by Redis (`github.com/redis/go-redis/v9`), keyed by client IP address and endpoint path (`ratelimit:<ip>:<endpoint>`).

**Rationale:** The Auth service needs to support horizontal scaling from the start. Redis-backed rate limiting ensures consistent throttling across multiple service replicas. The sliding window counter avoids the boundary-burst problem of fixed windows while being simpler and more memory-efficient than sliding window logs.

**Failure Strategy:** Fail-open — if Redis is temporarily unreachable, requests are allowed through and the failure is logged as a warning. This prioritizes user availability (ability to log in) over brute-force protection during brief Redis outages.

**Alternatives Considered:**
- **In-memory token bucket (`golang.org/x/time/rate`):** Zero-dependency and low-latency, but rate limits are not shared across replicas and are lost on restart.
- **Sliding window log (Redis ZADD):** 100% accurate but stores a timestamp per request, leading to higher memory usage under load.
- **Lua token bucket script:** Precise burst control, but adds complexity of maintaining a Lua script in Redis.

### Decision 3: SMS/Email OTP for 2FA

**Choice:** 6-digit numeric OTP code sent via SMS or Email, with a 5-minute expiry and a maximum of 5 verification attempts.

**Rationale:** OTP via SMS/Email is the most accessible 2FA method for e-commerce customers. It requires no app installation and works across all devices. Used specifically during registration (account activation) and password reset flows.

**Alternatives Considered:**
- **TOTP (Authenticator App):** More secure against SIM-swap attacks, but higher friction for casual e-commerce customers who may not have an authenticator app installed.

### Decision 4: bcrypt Password Hashing

**Choice:** bcrypt with a cost factor of 12.

**Rationale:** bcrypt is well-tested, widely supported in Go (`golang.org/x/crypto/bcrypt`), and the cost factor makes brute-force attacks computationally expensive. Cost 12 provides a good balance between security and login latency (~250ms hash time).

**Alternatives Considered:**
- **Argon2id:** Newer and memory-hard (resistant to GPU attacks). However, bcrypt is more battle-tested and the Go standard library support is more mature.

### Decision 5: Prisma Go Client with Supabase

**Choice:** `github.com/steebchen/prisma-client-go` connected to Supabase PostgreSQL via connection pooler (Transaction mode, port 6543).

**Rationale:** Prisma provides a type-safe, schema-first approach to database access. The Go client generates strongly-typed query builders from the Prisma schema. Supabase connection pooler (PgBouncer) handles connection management efficiently for serverless-style workloads.

### Decision 6: Database Schema

**Tables:**

```
┌─────────────────────────────────────────────────┐
│ users                                           │
├─────────────────────────────────────────────────┤
│ id          UUID     PK, default(uuid())        │
│ email       String   UNIQUE, NOT NULL            │
│ password    String   NOT NULL (bcrypt hash)      │
│ name        String   NOT NULL                    │
│ role        Enum     CUSTOMER | ADMIN            │
│ is_active   Boolean  default(false)              │
│ created_at  DateTime default(now())              │
│ updated_at  DateTime @updatedAt                  │
└─────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────┐
│ otp_codes                                       │
├─────────────────────────────────────────────────┤
│ id          UUID     PK, default(uuid())        │
│ user_id     UUID     FK → users.id              │
│ code        String   NOT NULL (6-digit)         │
│ type        Enum     REGISTRATION | PASSWORD_RESET│
│ attempts    Int      default(0), max 5          │
│ expires_at  DateTime NOT NULL                    │
│ used        Boolean  default(false)              │
│ created_at  DateTime default(now())              │
└─────────────────────────────────────────────────┘
```

### Decision 7: API Endpoint Design

```
POST   /api/auth/register          Register new user (sends OTP)
POST   /api/auth/verify-otp        Verify OTP code (activates account or authorizes reset)
POST   /api/auth/login             Login with email/password (returns JWT cookie)
POST   /api/auth/logout            Clear auth cookie
POST   /api/auth/forgot-password   Request password reset (sends OTP)
POST   /api/auth/reset-password    Submit new password (after OTP verification)
GET    /api/auth/me                Get current user from JWT claims
GET    /.well-known/jwks.json      Public key endpoint for cross-service JWT verification
```

### Decision 8: JWT Payload Structure

```json
{
  "sub": "usr_uuid_here",
  "email": "user@example.com",
  "role": "CUSTOMER",
  "iat": 1691234567,
  "exp": 1691238167
}
```

- `exp` set to 1 hour from issuance.
- No refresh token mechanism — user re-authenticates after expiry.

### Decision 9: Project Structure

```
store_auth/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Environment configuration (DB_URL, REDIS_ADDR, etc.)
│   ├── handler/
│   │   └── auth_handler.go      # HTTP handlers
│   ├── middleware/
│   │   ├── auth.go              # JWT auth middleware
│   │   └── ratelimit.go         # Redis rate limiting middleware
│   ├── redis/
│   │   └── client.go            # Redis client initialization
│   ├── model/
│   │   └── models.go            # Domain models
│   ├── repository/
│   │   └── user_repo.go         # Prisma-backed data access
│   ├── service/
│   │   ├── auth_service.go      # Business logic
│   │   ├── jwt_service.go       # JWT signing & verification
│   │   └── otp_service.go       # OTP generation & validation
│   └── router/
│       └── router.go            # Route registration
├── prisma/
│   └── schema.prisma            # Database schema
├── .env.example
├── go.mod
└── go.sum
```

## Risks / Trade-offs

### Risk 1: Redis Dependency for Rate Limiting
**Risk:** Rate limiting now depends on Redis availability. If Redis goes down, rate limiting is disabled (fail-open strategy).
**Mitigation:** Redis is a mature, operationally simple dependency. The fail-open strategy ensures users can still authenticate during brief Redis outages. Monitor Redis health and set up alerting. Redis Sentinel or managed Redis (e.g., Upstash, ElastiCache) can provide high availability if needed.

### Risk 2: Pure Stateless JWT Cannot Be Instantly Revoked
**Risk:** A compromised JWT remains valid until its `exp` (1 hour). There is no server-side session to delete.
**Mitigation:** Keep token lifetime short (1 hour). For critical security events (password change, admin ban), implement a lightweight in-memory token blacklist that expires entries after the token's remaining TTL. This is a future enhancement if needed.

### Risk 3: SMS OTP Delivery Reliability
**Risk:** SMS delivery can be delayed or fail due to carrier issues, especially internationally.
**Mitigation:** Support both SMS and Email OTP delivery. Allow the user to request a resend. Implement retry logic in the OTP delivery interface.

### Risk 4: bcrypt CPU Cost Under Load
**Risk:** bcrypt with cost 12 takes ~250ms per hash. Under high concurrent login/registration load, this can saturate CPU.
**Mitigation:** Rate limiting on login/register endpoints caps the maximum concurrent bcrypt operations. Monitor CPU usage and adjust cost factor if needed.

### Risk 5: Prisma Go Client Maturity
**Risk:** `prisma-client-go` is community-maintained and less mature than the official TypeScript client.
**Mitigation:** Pin to a stable version. Write repository-layer abstractions so the ORM can be swapped if issues arise.
