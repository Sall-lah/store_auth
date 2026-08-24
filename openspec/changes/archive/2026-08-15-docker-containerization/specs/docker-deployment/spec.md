## ADDED Requirements

### Requirement: Multi-Stage Docker Container Build
The system SHALL provide a multi-stage `Dockerfile` that compiles the Go application with Prisma Client Go and produces a lightweight, secure production container image.

#### Scenario: Build image successfully
- **WHEN** `docker build -t store_auth:latest .` is executed
- **THEN** the builder stage downloads Go dependencies, generates the Prisma Client Go binary/bindings, compiles the `store_auth` server binary, and produces a minimal runtime image containing only the executable and necessary CA/SSL certificates.

#### Scenario: Production runtime does not leak build artifacts or sources
- **WHEN** the container image is inspected
- **THEN** the image MUST NOT contain Go source code, build cache, or unneeded build tooling.

### Requirement: RSA Key and Secret Configuration Isolation
The containerized service SHALL accept configuration and cryptographic credentials via environment variables and mounted volumes without embedding secrets in the container image.

#### Scenario: Mount RSA keys at runtime
- **WHEN** the container is started with a volume mapping host RSA keys to `/keys` and environment variable `JWT_PRIVATE_KEY_PATH=/keys/private.pem` and `JWT_PUBLIC_KEY_PATH=/keys/public.pem`
- **THEN** the service successfully reads the RSA key pair, starts the HTTP server, and serves the JWKS endpoint.

#### Scenario: Missing required environment variables triggers clear startup failure
- **WHEN** the container is started without required variables such as `DATABASE_URL`
- **THEN** the container logs a descriptive startup error and exits cleanly without hanging.

### Requirement: Docker Compose Local Orchestration
The system SHALL provide a `docker-compose.yml` file to orchestrate `store_auth` and its companion `redis` service on a shared bridge network for local development and microservice testing.

#### Scenario: Spin up services with Docker Compose
- **WHEN** `docker compose up -d` is executed
- **THEN** both `store_auth` and `redis` containers start up, `redis` becomes healthy, and `store_auth` binds to port 8080 and successfully communicates with the Redis container via service DNS.

#### Scenario: Build context exclusion via .dockerignore
- **WHEN** Docker builds the image context
- **THEN** `.dockerignore` MUST prevent `.git`, `.env`, `.env.*`, local binaries, and sensitive key files from being copied into the Docker build context.
