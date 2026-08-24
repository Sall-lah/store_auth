## Why

The current API endpoints accept raw input payloads without performing string trimming, case normalization, HTML/XSS sanitization, or request body payload size capping. Implementing comprehensive request sanitization and input normalization will protect the application from memory exhaustion (DoS), Stored XSS, non-printable control characters, and case-sensitivity authentication bugs.

## What Changes

- Add request payload body size limiting middleware (`http.MaxBytesReader`) to restrict incoming JSON payload size (e.g., max 1 MB).
- Add string input normalization (whitespace trimming and lowercase conversion for emails) across all authentication requests (Register, Login, Verify OTP, Forgot Password, Reset Password).
- Add HTML/XSS sanitization and control-character stripping for string fields such as user `Name`.
- Update input validation routines to reject or sanitize invalid/unsafe unicode characters.

## Capabilities

### New Capabilities
- `request-sanitization`: Enforces HTTP request body payload limits, string input trimming, email normalization, and XSS/HTML control character sanitization across incoming HTTP requests.

### Modified Capabilities
- `user-registration`: Input parameters (`email`, `name`) must be normalized and sanitized prior to validation and database persistence.
- `user-login`: Input `email` must be trimmed and lowercased prior to authentication lookups.
- `password-reset`: Input `email` must be trimmed and lowercased prior to OTP generation and password reset verification.

## Impact

- `internal/auth/handler.go`: Updated request decoders and input validation to apply string trimming, lowercase email normalization, and XSS sanitization.
- `internal/middleware/`: Add request body size limiting middleware or helper.
- `internal/router/router.go`: Wire body size limiting middleware.
- Dependencies: Potential addition of standard string utilities or lightweight HTML sanitization helper if needed.
