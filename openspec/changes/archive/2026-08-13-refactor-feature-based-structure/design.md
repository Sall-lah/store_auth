## Context

The `store_auth` service currently uses a horizontal layer structure (`internal/handler`, `internal/service`, `internal/repository`, `internal/model`). As features grow, this causes file scattering across directories.

Refactoring to a feature-based structure reorganizes code around business capabilities while adhering to Go's package boundary rules and preventing circular import dependencies.

## Goals / Non-Goals

**Goals:**
- Reorganize domain code into self-contained feature packages: `internal/auth`, `internal/otp`, `internal/jwt`.
- Isolate shared infrastructure setup into `internal/platform/redis`, `internal/middleware`, and `internal/config`.
- Keep external API endpoints, request/response formats, database models, and JWT claims completely unchanged.

**Non-Goals:**
- Changing database schema (`prisma/schema.prisma`).
- Adding new endpoints or modifying existing auth business logic.

## Decisions

### Decision 1: Feature-Based Directory Structure

We split `internal/` into feature packages:

```
internal/
├── auth/
│   ├── handler.go         # Register, Login, Logout, Me endpoints
│   ├── service.go         # Register, Login, Reset password business logic
│   ├── repository.go      # User database queries
│   └── model.go           # User DTOs (RegisterRequest, LoginRequest, UserResponse)
├── otp/
│   ├── handler.go         # Verify OTP, Forgot password HTTP endpoints
│   ├── service.go         # OTP code generation and verification rules
│   ├── repository.go      # OTP database queries
│   ├── sender.go          # OTPSender interface & LogOTPSender implementation
│   └── model.go           # OTPCode, VerifyOTPRequest, ResetPasswordRequest
├── jwt/
│   ├── service.go         # RS256 token issuance & verification
│   ├── handler.go         # GET /.well-known/jwks.json handler
│   └── model.go           # JWKS, JWK, and Claims models
├── middleware/
│   ├── auth.go            # JWT cookie validation & context extraction
│   └── ratelimit.go       # Redis sliding window counter rate limiter
├── config/
│   └── config.go          # Environment configuration
└── platform/
    └── redis/             # Redis client initialization
```

### Decision 2: Avoid Circular Imports via Interface Injection

`internal/auth` needs to trigger OTP generation and JWT signing. To prevent circular imports:
- `auth.Service` accepts `*otp.Service` and `*jwt.Service` during initialization in `cmd/server/main.go`.
- Each feature package defines its own request/response DTOs and errors locally.

## Risks / Trade-offs

- **Risk:** Import path changes across all files may break compilation during migration.  
  *Mitigation:* Perform refactoring in steps (create feature modules -> move functions -> update imports) and run `go build ./...` to verify clean compilation.
