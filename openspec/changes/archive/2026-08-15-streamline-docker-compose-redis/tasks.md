## 1. Docker Compose Configuration

- [x] 1.1 Remove `redis` service and `redis_data` volume from `docker-compose.yml`
- [x] 1.2 Add `extra_hosts` mapping `host.docker.internal:host-gateway` to `store_auth` in `docker-compose.yml`
- [x] 1.3 Update `REDIS_URL` environment variable to default to `redis://host.docker.internal:6379`
- [x] 1.4 Remove `depends_on: redis` section from `store_auth` service in `docker-compose.yml`

## 2. Documentation and Verification

- [x] 2.1 Verify `docker-compose.yml` syntax and validity
- [x] 2.2 Update any relevant Docker orchestration notes in project documentation
