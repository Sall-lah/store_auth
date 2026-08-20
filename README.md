# Store Auth Microservice (`store_auth`)

`store_auth` is a centralized identity provider and authentication microservice built in Go for the web store platform. It provides stateless RS256 JWT issuance, email/password credentials handling, SMS/Email OTP 2FA verification, refresh token rotation, Redis token blacklisting, Redis sliding-window rate limiting, interactive Swagger UI documentation, and a JWKS public key endpoint (`/.well-known/jwks.json`) for cross-service token verification.

---

## 🛠 Tech Stack

- **Language:** Go (1.22+)
- **ORM:** Prisma Client Go (`github.com/steebchen/prisma-client-go`)
- **Database:** PostgreSQL (Supabase)
- **Token Signing:** RS256 (RSA 2048-bit key pair)
- **Rate Limiting & Blacklist:** Redis sliding window & token revocation (`github.com/redis/go-redis/v9`)
- **Password Hashing:** bcrypt (Cost 12)
- **Router:** Chi HTTP Router (`github.com/go-chi/chi/v5`)
- **API Documentation:** OpenAPI 3.1 & Swagger UI (`internal/docs`)

---

## 🏛 Architecture & Security Features

- **Asymmetric RS256 JWT Signing:** Tokens are signed using a private RSA key. Downstream services verify signatures locally using public keys from the JWKS endpoint (`/.well-known/jwks.json`) with zero runtime database latency.
- **Dual Authentication Transport:** Supports both browser-based `HttpOnly`, `SameSite=Lax` cookies (`access_token`, `refresh_token`) and mobile/service `Authorization: Bearer <token>` headers.
- **Refresh Token Rotation & Invalidation:** Secure token refresh flow (`POST /api/auth/refresh`) with single-use token rotation and database-backed tracking in PostgreSQL.
- **Redis Token Blacklisting:** Instant token revocation on logout (`POST /api/auth/logout`) with TTL matching remaining token lifespan.
- **Sliding-Window Rate Limiting:** Redis-backed rate limiting across sensitive public authentication routes to prevent brute-force attacks.
- **Request Sanitization & Body Limits:** Enforces a 1MB payload size limit (`middleware.MaxBody`) and sanitizes incoming user input.

---

## 📁 Project Directory Structure

The project adheres to a modular, feature-based architecture where distinct capabilities reside in their own package under `internal/`:

```
store_auth/
├── api/
│   └── openapi.yaml                  # Canonical OpenAPI 3.1 specification
├── cmd/
│   └── server/
│       └── main.go                   # Application entrypoint & dependency wiring
├── docs/
│   └── MICROSERVICE_INTEGRATION.md   # Downstream integration & API Gateway guide
├── internal/
│   ├── auth/                         # Authentication feature (login, register, refresh, logout, user repo)
│   ├── config/                       # Environment configuration loader & validation
│   ├── docs/                         # Embedded Swagger UI handler & OpenAPI spec asset
│   ├── jwt/                          # RS256 token service, claims, and JWKS handler
│   ├── middleware/                   # Rate limiting, JWT authentication, body limit middleware
│   ├── otp/                          # One-Time Password service, repository, and email sender
│   ├── platform/
│   │   └── redis/                    # Redis client initialization & helper functions
│   └── router/                       # Chi router topology & route definitions
├── prisma/
│   └── schema.prisma                 # Prisma database schema definition
├── scripts/
│   └── gen_keys.go                   # RSA 2048-bit keypair generator script
├── Dockerfile                        # Multi-stage production container build
├── go.mod                            # Go module dependencies
└── README.md                         # Project documentation
```

---

## 🚀 Getting Started

### 1. Environment Setup

Copy `.env.example` to `.env` and configure your database, Redis, and secret keys:

```bash
cp .env.example .env
```

#### Environment Variables Reference

| Variable | Description | Default | Required |
| :--- | :--- | :--- | :--- |
| `DATABASE_URL` | PostgreSQL connection string | — | **Yes** |
| `REDIS_URL` | Redis connection URL (`redis://[:password@]host:port[/db]` or `rediss://...`) | `redis://localhost:6379` | No |
| `REDIS_PASSWORD` | Optional Redis authentication password (overrides URL password if set) | — | No |
| `SERVER_PORT` | HTTP server port | `8080` | No |
| `ENV` | Application environment (`development` / `production`) | `development` | No |
| `JWT_PRIVATE_KEY_PATH` | Path to RSA private key PEM file | `./keys/private.pem` | No |
| `JWT_PUBLIC_KEY_PATH` | Path to RSA public key PEM file | `./keys/public.pem` | No |
| `JWT_ACCESS_EXPIRY_MINUTES` | Access token lifespan in minutes | `15` | No |
| `JWT_REFRESH_EXPIRY_DAYS` | Refresh token lifespan in days | `7` | No |
| `RATE_LIMIT_MAX_REQUESTS` | Maximum requests per sliding window | `10` | No |
| `RATE_LIMIT_WINDOW_SECONDS`| Sliding rate limit window duration in seconds | `1` | No |
| `BCRYPT_COST` | Hashing cost factor for bcrypt | `12` | No |
| `OTP_PROVIDER` | OTP dispatch provider (`mock` or `smtp`) | `mock` | No |
| `OTP_EXPIRY_MINUTES` | OTP code expiration time in minutes | `5` | No |
| `OTP_MAX_ATTEMPTS` | Maximum allowed invalid OTP submission attempts | `5` | No |
| `SMTP_HOST` | SMTP server hostname for email delivery | `smtp.gmail.com` | No |
| `SMTP_PORT` | SMTP server port | `587` | No |
| `SMTP_USERNAME` | SMTP authentication username / sender address | — | No |
| `SMTP_PASSWORD` | SMTP authentication app password | — | No |
| `SMTP_FROM` | Outgoing sender email address | `${SMTP_USERNAME}` | No |

---

### 2. Generate RSA Key Pair

Generate local development 2048-bit RSA keys (`private.pem` & `public.pem`) in `./keys/`:

```bash
go run scripts/gen_keys.go
```

---

### 3. Database Migration & Prisma Client Generation

Generate the Prisma Go client code from `prisma/schema.prisma`:

```bash
go run github.com/steebchen/prisma-client-go generate
```

To push schema changes to your database:

```bash
go run github.com/steebchen/prisma-client-go db push
```

---

### 4. Running the Service Locally

Start the Auth HTTP server:

```bash
go run cmd/server/main.go
```

The server starts on `http://localhost:8080`.

---

### 5. Running with Docker

Ensure RSA keys are generated in `./keys/` first:

```bash
go run scripts/gen_keys.go
```

Build the production container image:

```bash
docker build -t store_auth .
```

Run the containerized service:

```bash
docker run -d \
  --name store_auth_app \
  -p 8080:8080 \
  --env-file .env \
  -v ./keys:/app/keys:ro \
  store_auth
```

- The API will be available on `http://localhost:8080`
- RSA keys are mounted read-only from `./keys` into the container at `/app/keys`
- To stop the service: `docker stop store_auth_app && docker rm store_auth_app`

---

## 📖 Interactive API Documentation (Swagger UI)

Interactive Swagger UI documentation is embedded directly into the microservice:

- **Swagger UI Dashboard:** `http://localhost:8080/docs` or `http://localhost:8080/swagger`
- **Raw OpenAPI 3.1 YAML:** `http://localhost:8080/docs/openapi.yaml`

---

## 🔐 API Endpoints Overview

| Method | Endpoint | Description | Auth Required | Rate Limited |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/.well-known/jwks.json` | JWKS RSA public keys for token verification | ❌ No | ❌ No |
| `GET` | `/docs` | Interactive Swagger UI API documentation | ❌ No | ❌ No |
| `GET` | `/docs/openapi.yaml` | Raw OpenAPI 3.1 YAML specification | ❌ No | ❌ No |
| `POST` | `/api/auth/register` | Register new pending account (triggers OTP) | ❌ No | ✅ Yes |
| `POST` | `/api/auth/verify-otp` | Verify 6-digit OTP code (activates user & issues tokens) | ❌ No | ✅ Yes |
| `POST` | `/api/auth/login` | Authenticate credentials & receive access/refresh tokens | ❌ No | ✅ Yes |
| `POST` | `/api/auth/refresh` | Rotate and issue new access token using refresh token | ❌ No | ✅ Yes |
| `POST` | `/api/auth/logout` | Invalidate refresh token and blacklist access token | ❌ No | ❌ No |
| `POST` | `/api/auth/forgot-password` | Request password reset OTP code via email | ❌ No | ✅ Yes |
| `POST` | `/api/auth/reset-password` | Reset password using verified OTP code | ❌ No | ✅ Yes |
| `GET` | `/api/auth/me` | Retrieve authenticated user profile | ✅ Yes (Cookie / Bearer) | ❌ No |

---

## 🌐 Downstream Microservice & Gateway Integration

For detailed instructions on verifying tokens in downstream services, NGINX API Gateway routing, anti-spoofing header stripping, and polyglot middleware recipes, see:

📖 **[Downstream Microservice & API Gateway Integration Guide](docs/MICROSERVICE_INTEGRATION.md)**
