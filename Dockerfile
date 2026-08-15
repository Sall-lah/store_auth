# ==========================================
# Stage 1: Build & Prisma Generation
# ==========================================
FROM golang:alpine AS builder

# Install build dependencies for Prisma Client Go and CGO/SSL toolchain
RUN apk add --no-cache git ca-certificates openssl build-base tzdata

WORKDIR /app

# Cache Go dependency layers
COPY go.mod go.sum ./
RUN go mod download

# Copy Prisma schema and generate Linux-compatible client & query engine
COPY prisma ./prisma
RUN go run github.com/steebchen/prisma-client-go generate

# Copy source code and build production binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/store_auth ./cmd/server/main.go

# ==========================================
# Stage 2: Minimal Production Runtime
# ==========================================
FROM alpine:3.20

# Install CA certificates, timezone data, and OpenSSL required by Prisma runtime
RUN apk add --no-cache ca-certificates tzdata openssl libc6-compat

# Create non-root system user and group for hardened security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Prepare keys directory and permissions
RUN mkdir -p /app/keys && chown -R appuser:appgroup /app

# Copy compiled binary from builder
COPY --from=builder --chown=appuser:appgroup /app/bin/store_auth /app/store_auth

# Default environment configuration
ENV SERVER_PORT=8080 \
    ENV=production \
    JWT_PRIVATE_KEY_PATH=/app/keys/private.pem \
    JWT_PUBLIC_KEY_PATH=/app/keys/public.pem

# Expose HTTP service port
EXPOSE 8080

# Switch to non-root user
USER appuser

# Declare volume mount point for runtime RSA cryptographic keys
VOLUME ["/app/keys"]

ENTRYPOINT ["/app/store_auth"]
