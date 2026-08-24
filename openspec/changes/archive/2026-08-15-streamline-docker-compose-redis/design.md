## Context

The `store_auth` service utilizes Redis for rate limiting and user token revocation blacklisting. During local development, the developer runs an existing standalone Redis instance on the host machine (e.g. `localhost:6379`). The previous Docker Compose file bundled a containerized Redis instance, creating port binding conflicts on host port 6379 and ignoring the host Redis instance.

## Goals / Non-Goals

**Goals:**
- Streamline `docker-compose.yml` to run only the `store_auth` application container.
- Connect `store_auth` to the host's existing Redis server using `host.docker.internal:6379` by default.
- Enable overriding the Redis connection string via `.env` for remote/cloud Redis setups.
- Maintain cross-platform Docker networking support (Windows, macOS, Linux).

**Non-Goals:**
- Changing application Go code, Redis client logic, or fail-open middleware behavior.
- Managing the lifecycle or configuration of the host machine's Redis service.

## Decisions

### Decision 1: Use `extra_hosts: ["host.docker.internal:host-gateway"]`
- **Why**: While Docker Desktop on Windows/macOS maps `host.docker.internal` automatically, adding `host-gateway` ensures cross-platform compatibility across Linux Docker daemons as well without requiring host networking mode.
- **Alternatives Considered**: 
  - `network_mode: "host"`: Not reliably supported across all Docker Desktop installations on Windows and exposes unnecessary network surface.

### Decision 2: Remove `redis` service and `redis_data` volume
- **Why**: Having an unused service definition in Docker Compose leads to accidental container launches and port conflicts.
- **Alternatives Considered**:
  - Docker Compose profiles (e.g., `--profile with-redis`): Adds unnecessary CLI complexity when the developer's default workflow uses host Redis.

### Decision 3: Parameterize `REDIS_URL` with sensible host default
- **Why**: Setting `REDIS_URL=${REDIS_URL:-redis://host.docker.internal:6379}` in `docker-compose.yml` ensures that out-of-the-box compose runs hit host Redis, but allows custom `.env` configurations (e.g. remote Upstash URLs) to take precedence seamlessly.

## Risks / Trade-offs

- **[Risk] Host Redis only listening on 127.0.0.1 with protected-mode**: Docker container connects from the gateway subnet IP (e.g., `172.x.x.x` or `192.168.x.x`), which Redis might reject if bound strictly to localhost loopback.
  - **Mitigation**: Ensure host Redis configuration binds to `0.0.0.0` or allows gateway interface connections. Note that `store_auth` includes fail-open logging if Redis is unreachable.
- **[Risk] Container startup when host Redis is offline**: Container will start without a healthcheck block.
  - **Mitigation**: The Go service gracefully handles offline Redis with non-fatal logging and fail-open auth verification.
