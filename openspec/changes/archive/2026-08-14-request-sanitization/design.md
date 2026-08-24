## Context

The current authentication endpoints in `internal/auth/handler.go` decode JSON requests directly from `r.Body` into DTOs and perform basic validation (`mail.ParseAddress`, minimum password length) without normalizing input fields, stripping HTML/XSS content, or limiting payload body size. 

To harden the service against DoS via memory exhaustion, Stored XSS attacks, non-printable control characters, and case-sensitivity authentication errors, a structured request sanitization layer is needed.

## Goals / Non-Goals

**Goals:**
- Implement payload size limits via `http.MaxBytesReader` across all JSON-accepting auth endpoints (1 MB maximum).
- Implement input string trimming and email normalization (`strings.ToLower`) before validation, authentication lookups, and account creation.
- Implement HTML tag stripping and control-character removal for display name inputs (`req.Name`).
- Maintain existing API response formats and error schemas.

**Non-Goals:**
- Modifying password characters or hashing algorithms (passwords must retain exact character sequences).
- Implementing Web Application Firewall (WAF) or IP-level payload inspection (handled upstream by reverse proxies/gateways).

## Decisions

### Decision 1: Create a dedicated `internal/sanitizer` package for modular input cleansing
- **Rationale**: Keeps sanitization logic separate from HTTP transport handlers and business service logic, adhering to human-readable, modular Go architecture.
- **Alternatives Considered**:
  - Inline sanitization inside `handler.go`: Rejected to prevent code duplication across registration, login, and reset handlers.
  - Using an external HTML sanitizer library: Rejected to avoid unnecessary third-party dependencies for simple text string cleansing. Standard library regex/string processing is lightweight and reliable for basic name sanitization.

### Decision 2: Enforce Request Body Size Limits via `http.MaxBytesReader`
- **Rationale**: Go's `http.MaxBytesReader(w, r.Body, maxBytes)` protects against memory exhaustion attacks by returning an error if a client sends a payload exceeding 1 MB (1,048,576 bytes).
- **Alternatives Considered**:
  - Reading `r.Body` into `io.LimitReader`: Doesn't automatically notify the underlying HTTP server to close oversized request connections. `http.MaxBytesReader` properly updates the response writer state.

### Decision 3: Order of Operations (Decode -> Sanitize -> Validate)
- **Rationale**: Request payload must be decoded into struct fields first, then sanitized (trimming whitespace, normalizing case, stripping tags), and finally validated against structural rules (`validateEmail`, `validatePassword`).

## Risks / Trade-offs

- [Risk] Aggressive HTML stripping on names might remove valid characters (e.g. `<` or `>`). → *Mitigation*: Replace HTML tags while allowing standard names with apostrophes, hyphens, and standard Unicode characters.
- [Risk] `http.MaxBytesReader` returns error during `json.Decode`. → *Mitigation*: Provide clear `"Request payload exceeds maximum allowed size"` error response when body limit is breached.
