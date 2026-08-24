## 1. Environment Template Updates

- [x] 1.1 Update `.env.example` to document Redis password syntax (`redis://:password@host:port/db`), ACL format, and special character percent-encoding rules
- [x] 1.2 Update `.env.production.example` with explicit URL-encoding notes for complex production passwords

## 2. Documentation Updates

- [x] 2.1 Update `README.md` environment variable table with authenticated `REDIS_URL` formats and syntax notes

## 3. Verification

- [x] 3.1 Run tests (`go test ./...`) to ensure existing configuration and Redis utilities maintain full integrity
