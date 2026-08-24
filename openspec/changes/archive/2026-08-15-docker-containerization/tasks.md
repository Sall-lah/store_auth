## 1. Dockerfile & Build Context Setup

- [x] 1.1 Create `.dockerignore` excluding `.git`, `.env*`, `keys/*.pem`, `bin/`, and temporary build files
- [x] 1.2 Create multi-stage `Dockerfile` compiling Prisma Client Go and the Go application on Alpine

## 2. Docker Compose Orchestration

- [x] 2.1 Create `docker-compose.yml` defining `store_auth` and `redis` services on a shared network
- [x] 2.2 Configure Redis healthcheck, port mappings, environment variable forwarding, and `./keys` volume mount

## 3. Documentation & Verification

- [x] 3.1 Update `README.md` with Docker build and Docker Compose execution workflows
- [x] 3.2 Validate Dockerfile structure, Compose configuration, and `.dockerignore` completeness
