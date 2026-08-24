## Context

The `store_auth` microservice is written in Go 1.22+ using Prisma Client Go (`github.com/steebchen/prisma-client-go`), Chi router, RS256 token signing, and Redis for rate limiting and token blacklisting. 

To enable isolated local execution alongside other microservices (API Gateway, store services, frontend) and prepare for cloud container deployments, `store_auth` requires a robust containerization architecture. A key technical requirement is properly handling Prisma Client Go's pre-compiled query engine binary during container builds, as well as securely mounting RSA key pairs without baking secrets into the Docker image.

## Goals / Non-Goals

**Goals:**
- Provide a production-grade multi-stage `Dockerfile` producing a minimal, secure container image.
- Handle Prisma Client Go query engine binary generation seamlessly within the Docker build pipeline.
- Supply a complete `docker-compose.yml` defining `store_auth` and a containerized `redis` service on a shared network with healthchecks.
- Provide strict `.dockerignore` rules to avoid leaking sensitive `.env` files, build caches, or private keys into image layers.
- Document container build, volume mount, and Compose orchestration workflows in `README.md`.

**Non-Goals:**
- Creating Kubernetes / Helm manifests (kept out of scope for initial containerization).
- Bundling a local PostgreSQL database into Compose (external Supabase PostgreSQL is used as the primary database).

## Decisions

### Decision 1: Multi-Stage Build with Alpine Runtime

- **Choice**: Use `golang:1.22-alpine` for the builder stage and `alpine:3.20` for the final runtime stage, installing `ca-certificates`, `openssl`, and `tzdata`.
- **Rationale**: 
  - Prisma Client Go generates a Go client that connects via a pre-compiled Prisma query engine (`query-engine-linux-musl`). Alpine provides a lightweight (< 30MB base) footprint and native compatibility when `openssl` / `ca-certificates` are installed.
  - Multi-stage build discards the Go compiler, package caches, and build tools, leaving only the compiled executable and query engine binary.
- **Alternatives Considered**: 
  - *Distroless*: Distroless lacks basic shell tools and requires manual glibc/musl dependency troubleshooting for Prisma engine binaries.
  - *Ubuntu / Debian Slim*: Larger image footprint (~150MB+ vs ~35MB Alpine).

### Decision 2: Volume-Mounted RSA Key Pairs and Secret Isolation

- **Choice**: Mount `./keys` directory from host into container at `/app/keys:ro` (read-only) and load paths via `JWT_PRIVATE_KEY_PATH` and `JWT_PUBLIC_KEY_PATH`.
- **Rationale**: 
  - Prevents sensitive private RSA keys from ever being baked into Docker image layers.
  - Allows different environments (local development vs staging vs production) to supply their own keys via volume mounts or secret managers (e.g. Docker Secrets / K8s secrets).
- **Alternatives Considered**: 
  - *Generating keys automatically inside container on boot*: Causes key mismatches across container restarts and breaks existing client tokens.

### Decision 3: Docker Compose Environment & Redis Service Wiring

- **Choice**: Define a `redis` container (`redis:7-alpine`) alongside `store_auth` on a custom bridge network (`store-network`), with a Redis healthcheck.
- **Service Configuration**:
  ```yaml
  services:
    redis:
      image: redis:7-alpine
      ports:
        - "6379:6379"
      healthcheck:
        test: ["CMD", "redis-cli", "ping"]
        interval: 5s
        timeout: 3s
        retries: 5
    store_auth:
      build: .
      ports:
        - "8080:8080"
      environment:
        - REDIS_URL=redis://redis:6379
        - DATABASE_URL=${DATABASE_URL}
      volumes:
        - ./keys:/app/keys:ro
      depends_on:
        redis:
          condition: service_healthy
  ```
- **Rationale**: Downstream microservice developers can boot `store_auth` + `redis` with a single command `docker compose up -d` without needing local Redis installed.

## Risks / Trade-offs

- **[Risk: Prisma Query Engine Mismatch on Alpine]** Prisma Client Go might fail to run if `openssl` or musl compatibility libraries are missing.
  - **Mitigation**: Explicitly install `openssl` and `ca-certificates` in the Alpine runtime stage and verify engine generation during the builder stage (`go run github.com/steebchen/prisma-client-go generate`).
- **[Risk: Missing Host Keys Directory]** If a developer runs `docker compose up` without generating RSA keys first (`go run scripts/gen_keys.go`), `store_auth` fails to start.
  - **Mitigation**: Document step-by-step instructions in `README.md` to generate keys prior to first `docker compose up`, and verify clean error messaging in application logs.
- **[Risk: Accidental Secret Leakage in Image Context]** Developers might copy `.env` into image layers.
  - **Mitigation**: Add explicit `.dockerignore` patterns matching `.env`, `.env.*`, `keys/*.pem`, `.git`, `bin`, and test caches.
