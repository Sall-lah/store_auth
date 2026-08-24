## 1. Auth Handler Cookie Configuration

- [x] 1.1 Update `Login` handler in `internal/auth/handler.go` to set `SameSite: http.SameSiteLaxMode`
- [x] 1.2 Update `Logout` handler in `internal/auth/handler.go` to set `SameSite: http.SameSiteLaxMode`

## 2. Testing & Verification

- [x] 2.1 Verify auth handler unit tests pass (`go test ./internal/auth/...`)
- [x] 2.2 Verify full test suite passes (`go test ./...`)
