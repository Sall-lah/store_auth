## Context

The `store_auth` service has evolved with multiple core capabilities including feature-based refactoring, refresh token rotation, Redis token blacklisting, sliding-window rate limiting, request sanitization, and embedded Swagger UI documentation. However, the root `README.md` and related documentation still reflect earlier iterations—omitting the `POST /api/auth/refresh` endpoint, missing several critical environment variables (such as token lifetimes, rate limit parameters, bcrypt cost, and SMTP settings), lacking a project structure overview, and omitting Swagger UI links.

This design establishes a structured, complete documentation baseline that fully mirrors the 12 OpenSpec capabilities and existing implementation.

## Goals / Non-Goals

**Goals:**
- Update `README.md` to accurately document all 9 endpoints (`GET /.well-known/jwks.json`, `POST /api/auth/register`, `POST /api/auth/verify-otp`, `POST /api/auth/login`, `POST /api/auth/refresh`, `POST /api/auth/logout`, `POST /api/auth/forgot-password`, `POST /api/auth/reset-password`, `GET /api/auth/me`) along with interactive documentation routes (`/docs`, `/swagger`, `/docs/openapi.yaml`).
- Document all environment variables defined in `internal/config/config.go` with their descriptions and default values.
- Document the feature-based package layout (`internal/auth`, `internal/jwt`, `internal/otp`, `internal/docs`, `internal/middleware`, `internal/platform`, `internal/router`, `internal/config`, `prisma/`).
- Document core security and authentication mechanics: RS256 asymmetric signing, HttpOnly + SameSite=Lax cookie management, Redis token blacklist, and sliding-window rate limiting.
- Verify consistency between `README.md`, `docs/MICROSERVICE_INTEGRATION.md`, and OpenAPI specifications (`api/openapi.yaml`, `internal/docs/openapi.yaml`).

**Non-Goals:**
- Modifying Go source code, routers, handlers, or middleware.
- Altering Prisma database schema or migrations.

## Decisions

### 1. Standardized Documentation Layout
- Organize `README.md` into logical sections:
  1. Service Overview & Architecture (RS256, JWKS, Gateway integration, Token rotation & blacklist)
  2. Tech Stack
  3. Project Directory Structure (Feature-based modular architecture)
  4. Environment Variables Configuration Table
  5. Getting Started & Local Setup (RSA key generation, Prisma client, Server launch)
  6. Docker Deployment (Standalone container with volume mount)
  7. API Endpoints Catalog & Auth Modes
  8. Interactive Swagger UI Access (`/docs`, `/swagger`)
  9. Downstream Microservice Integration Reference

### 2. Synchronization of OpenAPI and Embedded Documentation
- Ensure `api/openapi.yaml` and `internal/docs/openapi.yaml` remain identical and completely describe all schemas, response codes (200, 201, 400, 401, 403, 404, 409, 413, 429, 500), and security schemes.

### 3. Clear Differentiation of Auth Transport Modes
- Explicitly document both web browser transport (`HttpOnly`, `SameSite=Lax` cookies) and cross-service/mobile transport (`Authorization: Bearer <token>`).

## Risks / Trade-offs

- **[Documentation Drift Over Time]** → Mitigated by referencing OpenSpec specs and maintaining strict single-source schema definitions.
