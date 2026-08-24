## Why

The developer already has a local Redis instance running on the host machine. The current `docker-compose.yml` attempts to spin up an internal `redis` container binding to port `6379`, causing port collision errors or redundant resource consumption. Streamlining `docker-compose.yml` allows the `store_auth` container to connect cleanly to the host's existing Redis instance via `host.docker.internal`.

## What Changes

- Remove redundant `redis` container service and its associated named volume `redis_data` from `docker-compose.yml`.
- Configure `extra_hosts` on the `store_auth` service with `host.docker.internal:host-gateway` to allow reliable host machine communication across Docker environments.
- Update `REDIS_URL` environment variable in `docker-compose.yml` to default to `redis://host.docker.internal:6379` while supporting overrides via `.env`.
- Remove `depends_on: redis` healthcheck dependency from `store_auth` service so the container starts immediately.

## Capabilities

### New Capabilities

<!-- No new capability specs introduced -->

### Modified Capabilities
- `docker-deployment`: Update Docker Compose orchestration requirements to connect to external/host Redis instances rather than requiring an internal containerized Redis service.

## Impact

- **Affected Files**: `docker-compose.yml`, `README.md` (if Docker compose instructions are mentioned).
- **Environment**: Containerized `store_auth` service communicates with host/external Redis on `host.docker.internal:6379`.
- **Breaking Changes**: None. Port 8080 mapping and all security, JWT, and database environment variables remain identical.
