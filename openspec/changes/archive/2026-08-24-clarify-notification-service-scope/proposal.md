## Why

In the distributed microservice architecture, there is potential ambiguity between two notification concepts:
1. Transactional external OTP delivery (SMS/Email) handled by store_notification.
2. In-app user notification feeds and read-states managed internally by store_user.

Clarifying this boundary in documentation and specs ensures maintainers and downstream consumers understand that store_notification is a stateless delivery mechanism dedicated exclusively to OTP delivery and is not involved in user notification persistence or account deletion cleanup.

## What Changes

- Update docs/MICROSERVICE_INTEGRATION.md Section 8.2 to explicitly document that store_notification is used exclusively for transactional OTP code delivery (via Kafka auth.events) and maintains no persistent user notification state.
- Update openspec/specs/kafka-notifications/spec.md purpose and requirement descriptions to clearly define the transactional OTP delivery scope and distinction from in-app notifications.

## Capabilities

### New Capabilities
<!-- None -->

### Modified Capabilities
- kafka-notifications: Clarify transactional OTP delivery scope and boundaries relative to in-app notification management.

## Impact

- **Documentation**: docs/MICROSERVICE_INTEGRATION.md updated with explicit architecture callouts.
- **Specifications**: openspec/specs/kafka-notifications/spec.md updated with refined purpose and non-goals.
- **Code**: No runtime backend Go code changes required (existing producer and consumer logic are already fully aligned).
