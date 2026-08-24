## MODIFIED Requirements

### Requirement: Docker Compose Local Orchestration
The system SHALL provide a `docker-compose.yml` file to run the `store_auth` containerized microservice and connect to a host or external Redis instance using host gateway resolution and environment variable configuration.

#### Scenario: Spin up store_auth with Docker Compose
- **WHEN** `docker compose up -d` is executed
- **THEN** the `store_auth` container starts up, binds to port 8080, resolves `host.docker.internal` via host gateway, and connects to the existing Redis instance on the host machine or external URL specified by `REDIS_URL`.
