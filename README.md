# Store Auth Microservice (`store_auth`)

`store_auth` is a centralized identity provider and authentication microservice built in Go for the web store platform. It provides stateless RS256 JWT issuance, email/password credentials handling, SMS/Email OTP 2FA verification, Redis sliding-window rate limiting, and a JWKS public key endpoint (`/.well-known/jwks.json`) for cross-service token verification.

---

## 🛠 Tech Stack

- **Language:** Go (1.22+)
- **ORM:** Prisma Client Go (`github.com/steebchen/prisma-client-go`)
- **Database:** Supabase (PostgreSQL)
- **Token Signing:** RS256 (RSA 2048-bit key pair)
- **Rate Limiting:** Redis sliding window (`github.com/redis/go-redis/v9`)
- **Password Hashing:** bcrypt (Cost 12)
- **Router:** Chi HTTP Router (`github.com/go-chi/chi/v5`)

---

## 🚀 Getting Started

### 1. Environment Setup

Copy `.env.example` to `.env` and fill in your Supabase PostgreSQL database URL and Redis connection string:

```bash
cp .env.example .env
```

Key environment variables:

| Variable | Description | Default |
| :--- | :--- | :--- |
| `DATABASE_URL` | PostgreSQL connection string | *Required* |
| `REDIS_URL` | Redis instance URL | `redis://localhost:6379` |
| `SERVER_PORT` | HTTP Server port | `8080` |
| `JWT_PRIVATE_KEY_PATH` | Path to RSA private key PEM | `./keys/private.pem` |
| `JWT_PUBLIC_KEY_PATH` | Path to RSA public key PEM | `./keys/public.pem` |

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

To push schema changes to Supabase database:

```bash
go run github.com/steebchen/prisma-client-go db push
```

---
### 4. Running the Service

Start the Auth HTTP server:

```bash
go run cmd/server/main.go
```

The server starts on `http://localhost:8080`.

---

## 🔐 API Endpoints Overview

| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :--- |
| `GET` | `/.well-known/jwks.json` | JWKS RSA public keys for token verification | ❌ No |
| `POST` | `/api/auth/register` | Register new pending account (triggers OTP) | ❌ No (Rate limited) |
| `POST` | `/api/auth/verify-otp` | Verify 6-digit OTP code (activates user) | ❌ No (Rate limited) |
| `POST` | `/api/auth/login` | Authenticate credentials & receive `access_token` cookie | ❌ No (Rate limited) |
| `POST` | `/api/auth/logout` | Clear HttpOnly authentication cookie | ❌ No |
| `POST` | `/api/auth/forgot-password` | Request password reset OTP | ❌ No (Rate limited) |
| `POST` | `/api/auth/reset-password` | Submit new password after OTP verification | ❌ No (Rate limited) |
| `GET` | `/api/auth/me` | Retrieve authenticated user profile | ✅ Yes (`access_token` cookie) |
