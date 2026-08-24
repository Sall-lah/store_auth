## Why

The `store_auth` microservice currently relies on external services for persistence (Supabase PostgreSQL) and rate limiting / caching (Redis on host / external instance). The `docker-compose.yml` file only defined a single service (`store_auth`) without spinning up auxiliary containers, acting solely as a redundant runner wrapper. Removing it simplifies project maintenance, avoids confusion around multi-container orchestration, and aligns local execution with native Go tooling (`go run cmd/server/main.go`) or direct Docker container runs (`docker run`).

## What Changes

- **Remove Docker Compose orchestration file**: Delete `docker-compose.yml`.
- **Update Documentation**: Update `README.md` to remove Docker Compose instructions and provide direct `docker build` / `docker run` instructions for containerized execution.
- **Update Specification**: Modify `docker-deployment` spec to retire the requirement for Docker Compose while preserving the multi-stage `Dockerfile`, security hardening, non-root user, and secret volume isolation requirements.

## Capabilities

### New Capabilities
<!-- None -->

### Modified Capabilities
- `docker-deployment`: Retire Requirement `Docker Compose Local Orchestration` while retaining multi-stage build, secret/key volume mounting, and container runtime specifications.

## Impact

- **Affected Files**:
  - `docker-compose.yml` (deleted)
  - `README.md` (updated setup / running guide)
  - `openspec/specs/docker-deployment/spec.md` (spec updated via delta spec)
- **APIs & Dependencies**: No impact on Go source code, APIs, or database models.
- **Dev Workflow**: Developers running with Docker will use standard `docker build` / `docker run` commands or native Go commands.
