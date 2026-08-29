## 1. Swagger UI Relative Path Implementation

- [x] 1.1 Update `internal/docs/index.html` to dynamically calculate the relative OpenAPI spec URL using runtime `window.location.pathname`.
- [x] 1.2 Verify that both trailing-slash (`/docs/`) and non-trailing-slash (`/docs`) URLs as well as subpath proxy routes (e.g. `/auth/docs`) correctly resolve the specification URL to `openapi.yaml`.

## 2. Verification and Tests

- [x] 2.1 Update or add unit tests in `internal/docs/handler_test.go` asserting that Swagger UI HTML output contains relative spec path resolution logic.
- [x] 2.2 Run full test suite (`go test ./...`) to ensure all router and documentation tests pass without regressions.
