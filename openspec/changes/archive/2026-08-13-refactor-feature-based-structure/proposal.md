## Why

The current `store_auth` service uses a horizontal, layer-based directory structure (`internal/handler`, `internal/service`, `internal/repository`, `internal/model`). As the application evolves, related code for a single domain capability (such as authentication or OTP logic) is scattered across multiple packages, increasing cognitive load when maintaining or adding features. Refactoring to a feature-based (vertical slice) directory structure groups handler, service, repository, and model logic by domain feature (`auth`, `otp`, `jwt`), improving code locality and maintainability.

## What Changes

- Reorganize the `internal/` package layout from layer-based directories to feature-focused modules: `internal/auth`, `internal/otp`, `internal/jwt`.
- Move shared cross-cutting concerns to dedicated utility/infrastructure packages: `internal/middleware`, `internal/config`, `internal/platform/redis`.
- Update all package imports and method invocations across the codebase to match the new feature-based structure.
- Update `cmd/server/main.go` and `internal/router/router.go` to wire feature modules cleanly.

## Capabilities

### New Capabilities
_(None — this is an internal architectural refactoring.)_

### Modified Capabilities
_(None — external API contracts, JWT claims, and endpoint behaviors remain completely unchanged.)_

## Impact

- **Internal Packages**: All files in `internal/handler`, `internal/service`, `internal/repository`, and `internal/model` will be relocated into feature modules under `internal/auth`, `internal/otp`, and `internal/jwt`.
- **Server Entrypoint**: `cmd/server/main.go` and `internal/router/router.go` will be updated to import and wire feature modules.
- **External Dependencies & Database**: Unchanged (Prisma Client Go schema and Redis setup remain identical).
- **Public API & Specs**: No breaking changes or spec modifications.
