## Why

To support seamless local integration with other microservices (such as API Gateway, Product Catalog, Order Service, and Frontend) and simplify staging/production deployment, `store_auth` needs a reproducible containerization setup. Containerizing the service packages its Go runtime, Prisma Client query engine binary, Redis dependencies, and RSA key mounts into standard Docker and Docker Compose definitions, eliminating local environment mismatches.

## What Changes

- Add a multi-stage `Dockerfile` tailored for Go 1.22+ and Prisma Client Go (handling Prisma query-engine Linux binaries and creating a lightweight, secure production runtime container).
- Add `.dockerignore` to exclude build artifacts, git history, local node_modules, keys (unless mapped), and sensitive environment files from the build context.
- Add `docker-compose.yml` defining the `store_auth` service along with a companion `redis` service, shared Docker network, healthchecks, and volume mounts for RSA key pairs and configuration.
- Update `README.md` with Docker container build and run instructions.

## Capabilities

### New Capabilities
- `docker-deployment`: Defines container packaging, multi-stage build pipelines, volume mount requirements for RSA keys, Redis dependency wiring, and Docker Compose orchestration for the `store_auth` microservice.

### Modified Capabilities
<!-- None: No core authentication requirement or spec-level behavior changes. -->

## Impact

- **New files**: `Dockerfile`, `.dockerignore`, `docker-compose.yml`, `docker-compose.override.yml.example`.
- **Documentation**: Updated `README.md` with Docker build & Compose execution guides.
- **Dependencies & Build**: Uses standard Docker multi-stage builds with Debian/Alpine base images compatible with Prisma Client Go query engine binaries.
- **Security**: Ensures RSA private/public keys and `.env` credentials are appropriately mounted via volumes or injected via environment variables without baking secrets into Docker image layers.
