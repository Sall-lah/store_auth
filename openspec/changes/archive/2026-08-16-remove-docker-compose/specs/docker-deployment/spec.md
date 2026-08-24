## REMOVED Requirements

### Requirement: Docker Compose Local Orchestration
**Reason**: Docker Compose provided redundant orchestration for a standalone microservice whose dependencies (Supabase Postgres and Redis) reside externally or directly on the host machine.
**Migration**: Run the service using native Go commands (`go run cmd/server/main.go`) or standalone Docker container commands (`docker build` and `docker run` with volume and environment variable flags).
