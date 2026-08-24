# Technical Design: Unverified Re-Registration, Resend OTP Endpoint, and OTP Code Nullification

## Context

In `store_auth`, user activation and password resets depend on 6-digit numeric OTPs. Currently, if an unverified user's OTP expires or is lost:
1. Re-registering with the same email returns `409 Conflict` because the email already exists in `users` (even with `is_active: false`).
2. Logging in returns `403 Forbidden` because `is_active` is `false`.
3. There is no resend endpoint, permanently locking the user out.
4. Plain numeric OTP strings remain permanently in the database table even after being verified or invalidated.

This design resolves the trapped registration state, introduces a secure rate-limited resend endpoint, and establishes an OTP secret scrubbing lifecycle.

## Goals / Non-Goals

**Goals:**
- Enable unverified accounts (`is_active = false`) to re-submit `POST /api/auth/register` to update their profile/credentials and receive a fresh registration OTP.
- Provide a dedicated `POST /api/auth/resend-otp` endpoint supporting both `registration` and `password_reset` flows.
- Dispatch resend events to Kafka (`auth.events`) with identical schema (`auth.registration_otp` / `auth.password_reset_otp`) and fresh `event_id` for seamless `store_notification` processing.
- Nullify / scrub raw OTP codes in `otp_codes` persistence once used, invalidated, or superseded.
- Protect the resend endpoint with Redis-backed rate limiting and prevent account enumeration attacks.

**Non-Goals:**
- Modifying `store_notification` event consumer or email templates (existing templates and topics are completely reused).
- Adding multi-channel SMS / WhatsApp delivery (remains email-based via notification service).

## Decisions

### 1. Inactive Account Re-Registration Strategy
- **Decision**: When `POST /api/auth/register` is called with an email matching an existing account where `is_active == false`:
  - Update user's name and re-hash new password with bcrypt (cost factor 12).
  - Invalidate/nullify any existing pending registration OTPs for that user.
  - Generate a new 6-digit OTP code and publish `auth.registration_otp` to Kafka topic `auth.events`.
  - Respond with `200 OK` or `201 Created` with `"Registration pending activation. A new OTP has been sent."`
- **Alternative Considered**: Returning `409 Conflict` with a specific error code directing them to `/resend-otp`. 
  - *Rationale*: Allowing re-registration provides superior UX if a user made a typo in their name or forgot the password they typed moments earlier.

### 2. Dedicated `POST /api/auth/resend-otp` Route & Schema
- **Decision**: Mount `POST /api/auth/resend-otp` in the rate-limited `/api/auth` group.
  - Payload: `{ "email": string, "type": "registration" | "password_reset" }`
  - Flow for `registration`:
    - If user not found -> Return `400 Bad Request` ("User not found").
    - If user is active -> Return `400 Bad Request` ("Account is already verified. Please log in.").
    - If user is inactive -> Invalidate existing registration OTPs, issue fresh OTP, emit `auth.registration_otp`.
  - Flow for `password_reset`:
    - If user not found or inactive -> Return `200 OK` (anti-enumeration shield).
    - If user is active -> Invalidate existing reset OTPs, issue fresh OTP, emit `auth.password_reset_otp`.

### 3. Event Contract & Idempotency
- **Decision**: Resend operations construct standard `EventEnvelope` with a newly generated UUID `event_id` and publish to Kafka topic `auth.events`.
  - *Rationale*: A new `event_id` ensures `store_notification`'s Redis idempotency check processes the delivery without duplicate collision.

### 4. OTP Secret Nullification in Database
- **Decision**: Update Prisma schema to make `code` in `OTPCode` nullable (`code String?` or string clearing). Upon successful verification, expiration cleanup, or invalidation:
  - Set `used = true` and `code = null` (or scrubbed).
  - *Rationale*: Minimizes exposure of short numeric passcodes in database dumps or database access logs while preserving row history for attempt throttling and audit timestamps.

## Architecture Diagram

```
                        REGISTRATION & RESEND TOPOLOGY
                        ══════════════════════════════

  POST /api/auth/register (existing unverified email)
          │
          ▼
  ┌────────────────────────────────────────────────────────┐
  │ 1. Update User password & name in DB                   │
  │ 2. Nullify prior pending OTP codes                     │
  │ 3. Create fresh OTP & publish auth.registration_otp    │
  └────────────────────────────────────────────────────────┘

  POST /api/auth/resend-otp { email, type }
          │
          ├────────► type: "registration"
          │            └─► If inactive: nullify old OTP, issue new OTP & event
          │            └─► If active: 400 Bad Request
          │
          └────────► type: "password_reset"
                       └─► If active: nullify old OTP, issue new OTP & event
                       └─► If not found/inactive: 200 OK (shielded)
```

## Risks / Trade-offs

- **[Risk] OTP Email Spam / Flooding** → *Mitigation*: Rate limit `POST /api/auth/resend-otp` via Redis sliding window (10 requests / 60s per IP) and invalidate previous active codes upon each generation.
- **[Risk] Account Enumeration on Password Reset** → *Mitigation*: Return identical HTTP 200 responses regardless of whether the email exists for `type: "password_reset"`.
- **[Risk] In-Flight OTP Collision** → *Mitigation*: Invalidate prior OTPs before creating a new one, ensuring only the most recently issued code can be validated.
