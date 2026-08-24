## Why

Currently, the authentication service sets `SameSite=Strict` on the `access_token` HTTP cookie. While `Strict` offers strong CSRF protection, it prevents the authentication cookie from being sent on top-level GET navigations coming from external sites or links (e.g. email links, external redirects, or cross-domain client applications). Updating the cookie attribute to `SameSite=Lax` preserves robust CSRF defense against cross-site state-changing requests (POST, PUT, DELETE) while allowing seamless top-level user navigation.

## What Changes

- Update `Login` handler cookie generation to set `SameSite: http.SameSiteLaxMode`.
- Update `Logout` handler cookie invalidation to set `SameSite: http.SameSiteLaxMode`.
- Preserve existing `HttpOnly: true`, `Path: "/"`, `MaxAge`, and dynamic `Secure: h.isProd` settings.
- Update `jwt-management` capability specification to reflect `SameSite=Lax` requirement.

## Capabilities

### Modified Capabilities
- `jwt-management`: Update cookie transport specification requirement from `SameSite=Strict` to `SameSite=Lax`.

## Impact

- `internal/auth/handler.go`: Update `Login` and `Logout` handlers to set `SameSite: http.SameSiteLaxMode`.
- Endpoints: `/api/v1/auth/login`, `/api/v1/auth/logout`.
- Client impact: Clients navigated from external links will now present the authentication cookie on top-level GET requests without losing session context.
