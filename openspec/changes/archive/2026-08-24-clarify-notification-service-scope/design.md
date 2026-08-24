## Context

In our microservices architecture, there are two distinct concepts related to notifications:
1. **`store_notification`**: A standalone, stateless service that listens to `auth.events` on Apache Kafka to deliver one-time passwords (OTPs) via email or SMS.
2. **In-App Notifications**: A user profile feature hosted inside `store_user` providing an in-app notification inbox (`/api/users/notifications`) with read states and persistent database storage.

This design clarifies these architectural boundaries in `store_auth`'s integration documentation and OpenSpec specifications.

## Goals / Non-Goals

**Goals:**
- Provide unambiguous architecture guidelines in `docs/MICROSERVICE_INTEGRATION.md` detailing that `store_notification` is strictly used for transactional OTP code delivery.
- Update `openspec/specs/kafka-notifications/spec.md` to formally document the scope and non-goals.

**Non-Goals:**
- Modify backend Go code or event structs (the existing producer implementation in `internal/otp/sender.go` is already functionally complete and correctly scoped).
- Introduce in-app notification features to `store_auth` (in-app notifications remain solely in `store_user`).

## Decisions

### Decision: Explicit Architecture Callouts in Integration Guide
- **Rationale**: Developers integrating or maintaining microservices should immediately know that `store_notification` does not maintain user state or require deletion hooks upon account removal.
- **Alternatives Considered**: Creating a new shared document vs updating `docs/MICROSERVICE_INTEGRATION.md`. Selected `docs/MICROSERVICE_INTEGRATION.md` as it is the canonical integration guide.

## Risks / Trade-offs

- **[Risk] None (Documentation & Spec refinement only)** → Zero runtime impact or migration downtime.
