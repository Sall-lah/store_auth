## Context

The authentication handler in `internal/auth/handler.go` uses Go's standard `http.SetCookie` function to transmit the RS256 JWT `access_token` cookie to clients upon successful authentication. 

Currently, `SameSite` is set to `http.SameSiteStrictMode`.

## Goals / Non-Goals

**Goals:**
- Update `Login` and `Logout` handlers to set `SameSite: http.SameSiteLaxMode`.
- Retain `HttpOnly: true` to protect against client-side script token theft.
- Retain environment-aware `Secure: h.isProd` setting.

**Non-Goals:**
- Changing cookie name, storage backend, token claims, or expiration duration.
- Supporting legacy browsers without `SameSite` support.

## Decisions

### Decision 1: Use `http.SameSiteLaxMode` instead of `http.SameSiteStrictMode`
- **Rationale**: `SameSiteLaxMode` allows browsers to send the `access_token` cookie on top-level GET navigation requests (e.g. following a link to the store), preventing unexpected session drop-offs when users navigate to the app from external origins.
- **CSRF Trade-off**: `Lax` mode still prevents the browser from sending cookies on cross-site subrequests (such as `<img>`, `<iframe>`, or cross-site AJAX `POST` calls), preserving essential CSRF defense.

### Decision 2: Maintain `HttpOnly` and `Secure` Controls
- **Rationale**: `HttpOnly` prevents XSS token theft via `document.cookie`. `Secure` ensures cookies are transmitted strictly over HTTPS in production environments (`h.isProd`).

## Risks / Trade-offs

- [Risk] Cross-site top-level GET requests will send the cookie. → *Mitigation*: Ensure all state-mutating actions (e.g. updating profile, resetting password, initiating purchases) require non-GET HTTP methods (`POST`, `PUT`, `DELETE`), which are blocked by `SameSite=Lax`.
