## Context

The `store_auth` project is a Go-based identity and authentication service. Previously, `docker-compose.yml` was used to spin up `store_auth` in a container while linking to an external/host Redis and Supabase Postgres database. Because no auxiliary containers (database or Redis) were managed by Docker Compose, the compose configuration added unnecessary complexity and maintenance overhead without multi-service orchestration benefits.

## Goals / Non-Goals

**Goals:**
- Eliminate `docker-compose.yml` to simplify project root and reduce redundant tooling.
- Preserve container image build support via `Dockerfile` and `.dockerignore` for production containerization and CI/CD workflows.
- Update `README.md` to guide developers on running via native Go or standard `docker run`.

**Non-Goals:**
- Modifying the Go application code, database schema, or environment variables.
- Modifying the production `Dockerfile` or build stages.

## Decisions

### Decision 1: Direct Docker Run / Native Go in place of Docker Compose
- **Context**: Developers need to run the service locally either natively or within a container.
- **Choice**: Remove `docker-compose.yml`. For local development, native `go run cmd/server/main.go` is the primary and fastest workflow. For container testing, provide the equivalent `docker run` command with `--env-file .env` and volume mount `-v ./keys:/app/keys:ro`.
- **Alternatives Considered**:
  - *Keep docker-compose.yml with embedded Redis*: Adds extra service definitions that clash with existing host Redis setups and external Supabase databases.

## Risks / Trade-offs

- **[Risk] Container testing requires longer docker run CLI command** → **Mitigation**: Document clear `docker build` and `docker run` snippet with `--env-file` in `README.md`.
