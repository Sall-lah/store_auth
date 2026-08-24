## 1. OpenAPI Specification Updates

- [x] 1.1 Update `servers` block in `api/openapi.yaml` to include API Gateway entrypoint (e.g. `https://api.store.example.com` and `https://store-gateway.herokuapp.com`) alongside direct local server
- [x] 1.2 Validate `api/openapi.yaml` syntax and OpenAPI 3.1.0 schema compliance

## 2. Downstream & Gateway Integration Guide

- [x] 2.1 Update `docs/MICROSERVICE_INTEGRATION.md` architectural overview with Gateway-centric ASCII diagrams
- [x] 2.2 Add Section on API Gateway Header Propagation (`X-User-Id`, `X-User-Role`, `X-User-Email`) and anti-spoofing header sanitization
- [x] 2.3 Add NGINX / Heroku Gateway deployment guide with dynamic DNS resolver and proxy configuration recipes

## 3. Verification & Validation

- [x] 3.1 Verify Swagger UI and documentation rendering
- [x] 3.2 Verify consistency across specs, OpenAPI file, and integration guides
