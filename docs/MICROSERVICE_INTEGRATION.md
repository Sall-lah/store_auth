# Downstream Microservice & API Gateway Integration Guide

This guide details how downstream feature services (e.g. `order-service`, `product-service`, `payment-service`) and API Gateways (e.g. NGINX, Heroku Gateway, Traefik, Kong, KrakenD) integrate with and authenticate requests issued by `store_auth`.

---

## 1. Architectural Overview

`store_auth` uses **RS256 Asymmetric Cryptography** (RSA Signature with SHA-256) and standard RFC 7517 JWKS discovery. In production architectures, services are fronted by an **API Gateway**.

### Pattern A: Production API Gateway Offloading via Subrequest (Recommended)

In this pattern, the API Gateway serves as the single public entrypoint, terminates TLS, handles CORS, strips client `X-User-*` headers, and offloads authentication verification to `store_auth` via internal subrequests (`auth_request /_auth_verify` -> `/api/auth/me`). `store_auth` validates the RS256 token and Redis blacklist status, returning `X-User-*` headers for the Gateway to forward downstream.

```
                              PUBLIC INTERNET (HTTPS)
                                         │
                         [ Client (Web / Mobile / SPA) ]
                                         │  (Bearer JWT or HttpOnly Cookie)
                                         ▼
                     ┌────────────────────────────────────────┐
                     │     API GATEWAY (e.g. NGINX/Heroku)    │
                     ├────────────────────────────────────────┤
                     │ 1. Terminates TLS / Handles CORS       │
                     │ 2. Strips external 'X-User-*' headers  │  <-- Anti-Spoofing
                     │ 3. Subrequest: auth_request to         │
                     │    store_auth /api/auth/me             │
                     │ 4. Injects verified identity headers:  │
                     │    • X-User-Id: 9b1deb4d-...           │
                     │    • X-User-Role: CUSTOMER             │
                     │    • X-User-Email: user@example.com    │
                     └───────┬────────────────────────┬───────┘
                             │ (Internal Subrequest)  │ (Proxies enriched request)
                             ▼                        ▼
              ┌─────────────────────────────┐  ┌─────────────────────────────┐
              │         store_auth          │  │        order-service        │
              │         (Port 8080)         │  │         (Port 8081)         │
              ├─────────────────────────────┤  ├─────────────────────────────┤
              │ • Signs tokens (PrivKey)    │  │ Read headers directly:      │
              │ • Exposes JWKS (PubKey)     │  │   r.Header.Get("X-User-Id") │
              │ • Validates token & Redis   │  │ Zero JWT crypto overhead!   │
              │   blacklist on /api/auth/me │  │                             │
              └─────────────────────────────┘  └─────────────────────────────┘
```

---

### Pattern B: Perimeter Reverse Proxy with Downstream JWKS Validation (Zero-Trust)

If your internal microservice network requires zero-trust end-to-end token verification, the Gateway simply proxies the `Authorization: Bearer <token>` header, and feature services fetch `store_auth`'s public keys via `/.well-known/jwks.json` to verify the RS256 signature locally in memory.

```
                     ┌────────────────────────────────────────┐
                     │     API GATEWAY (Reverse Proxy)        │
                     │  • Passes Bearer token downstream      │
                     └───────────────────┬────────────────────┘
                                         │
                     ┌───────────────────┴────────────────────┐
                     │ (Fetches JWKS once on boot)            │
                     ▼                                        ▼
      ┌─────────────────────────────┐          ┌─────────────────────────────┐
      │         store_auth          │          │  Feature Service (Orders)   │
      │  Exposes /.well-known/jwks  │          │  • Validates JWT via JWKS   │
      └─────────────────────────────┘          │  • Zero runtime DB latency  │
                                               └─────────────────────────────┘
```

---

## 2. JWT Claims & Identity Header Specification

### JWT Claims Payload (Issued by `store_auth`)

```json
{
  "sub": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
  "email": "customer@example.com",
  "role": "CUSTOMER",
  "exp": 1740000000,
  "iat": 1739999100,
  "iss": "store_auth"
}
```

| Claim | Header Injected by Gateway | Type | Description |
| :--- | :--- | :--- | :--- |
| `sub` | `X-User-Id` | `string (UUID)` | Unique User ID in database. Use as foreign key for orders/carts. |
| `email` | `X-User-Email` | `string` | User email address. |
| `role` | `X-User-Role` | `string` | User permission role: `"CUSTOMER"` or `"ADMIN"`. |
| `exp` | *(Perimeter validated)* | `int64` | Expiration timestamp (Unix epoch seconds). Lifespan is 15 minutes. |
| `iss` | *(Perimeter validated)* | `string` | Issuer identifier: `"store_auth"`. |

---

## 3. Public Key Discovery (JWKS)

- **Endpoint**: `GET http://<auth-service-host>:8080/.well-known/jwks.json`
- **Specification**: [RFC 7517 (JSON Web Key Set)](https://datatracker.ietf.org/doc/html/rfc7517)

### Sample Response:
```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "alg": "RS256",
      "kid": "store-auth-primary-key",
      "n": "u1R5Zn...",
      "e": "AQAB"
    }
  ]
}
```

---

## 4. API Gateway Configuration (NGINX on Heroku / Docker)

The API Gateway lives in its own dedicated repository and manages perimeter routing to `store_auth` and feature microservices.

### Security: Anti-Spoofing Header Sanitization
The Gateway **MUST** strip client-supplied `X-User-*` headers before proxying requests to prevent malicious users from injecting arbitrary user identities.

### Production `nginx.conf.template` (Heroku Dyno Compatible)

```nginx
# nginx.conf.template
events {
    worker_connections 1024;
}

http {
    include       mime.types;
    default_type  application/octet-stream;
    sendfile      on;

    # Heroku dynamic DNS resolver
    resolver 8.8.8.8 1.1.1.1 valid=30s ipv6=off;

    # Upstream timeouts
    proxy_connect_timeout 5s;
    proxy_read_timeout    25s;

    server {
        listen ${PORT};

        # Global anti-spoofing: Strip client identity headers
        proxy_set_header X-User-Id "";
        proxy_set_header X-User-Role "";
        proxy_set_header X-User-Email "";

        # Standard Proxy Headers
        proxy_set_header Host $http_host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $http_x_forwarded_proto;

        # ----------------------------------------------------------------------
        # Internal Auth Subrequest Verification (Offloading to store_auth)
        # ----------------------------------------------------------------------
        location = /_auth_verify {
            internal;
            if ($request_method = OPTIONS) {
                return 200;
            }

            set $auth_backend "https://store-auth.herokuapp.com";
            proxy_pass $auth_backend/api/auth/me;
            proxy_pass_request_body off;
            proxy_set_header Content-Length "";
            proxy_set_header X-Original-URI $request_uri;
            proxy_set_header X-Original-Method $request_method;
            proxy_set_header Authorization $http_authorization;
            proxy_set_header Cookie $http_cookie;
        }

        # ----------------------------------------------------------------------
        # 1. Auth Service Routes & JWKS
        # ----------------------------------------------------------------------
        location /api/auth/ {
            set $auth_backend "https://store-auth.herokuapp.com";
            proxy_pass $auth_backend;
        }

        location = /.well-known/jwks.json {
            set $auth_backend "https://store-auth.herokuapp.com";
            proxy_pass $auth_backend/.well-known/jwks.json;
        }

        location /docs {
            set $auth_backend "https://store-auth.herokuapp.com";
            proxy_pass $auth_backend;
        }

        # ----------------------------------------------------------------------
        # 2. Feature Services Routes (e.g. Orders - Protected via Auth Subrequest)
        # ----------------------------------------------------------------------
        location /api/orders/ {
            # 1. Offload verification to store_auth /api/auth/me
            auth_request /_auth_verify;

            # 2. Extract verified claims from store_auth response headers
            auth_request_set $auth_user_id $upstream_http_x_user_id;
            auth_request_set $auth_user_role $upstream_http_x_user_role;
            auth_request_set $auth_user_email $upstream_http_x_user_email;

            # 3. Inject verified identity headers into downstream proxy request
            proxy_set_header X-User-Id $auth_user_id;
            proxy_set_header X-User-Role $auth_user_role;
            proxy_set_header X-User-Email $auth_user_email;

            set $orders_backend "https://store-orders.herokuapp.com";
            proxy_pass $orders_backend;
        }

        location /api/products/ {
            set $products_backend "https://store-products.herokuapp.com";
            proxy_pass $products_backend;
        }
    }
}
```

### Dockerfile for Heroku Deployment

```dockerfile
FROM nginx:alpine

# Copy configuration template and dynamic port startup script
COPY nginx.conf.template /etc/nginx/templates/default.conf.template

CMD ["sh", "-c", "envsubst '${PORT}' < /etc/nginx/templates/default.conf.template > /etc/nginx/nginx.conf && exec nginx -g 'daemon off;'"]
```

---

## 5. Downstream Microservice Code Recipes

### Recipe 1: Consuming Injected Gateway Headers (Polyglot / 0 Dependencies)

When using **Pattern A (Gateway Offloading)**, downstream microservices do not need any JWT verification dependencies. Simply read the trusted headers:

#### Go:
```go
package middleware

import (
	"context"
	"net/http"
)

type UserContext struct {
	ID    string
	Role  string
	Email string
}

func GatewayIdentityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-Id")
		if userID == "" {
			http.Error(w, `{"error":"Unauthorized: Missing user context"}`, http.StatusUnauthorized)
			return
		}

		user := &UserContext{
			ID:    userID,
			Role:  r.Header.Get("X-User-Role"),
			Email: r.Header.Get("X-User-Email"),
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
```

#### Node.js / TypeScript (Express):
```typescript
import { Request, Response, NextFunction } from "express";

export interface GatewayRequest extends Request {
  user?: {
    id: string;
    role: string;
    email: string;
  };
}

export function gatewayIdentity(req: GatewayRequest, res: Response, next: NextFunction) {
  const userId = req.headers["x-user-id"] as string;
  if (!userId) {
    return res.status(401).json({ error: "Unauthorized: Missing user context" });
  }

  req.user = {
    id: userId,
    role: (req.headers["x-user-role"] as string) || "CUSTOMER",
    email: (req.headers["x-user-email"] as string) || "",
  };

  next();
}
```

---

### Recipe 2: Standalone RS256 JWKS Verification (Zero-Trust / Direct Call)

When using **Pattern B (Zero-Trust)** or local development without an API gateway:

#### Go (Golang):
```bash
go get github.com/golang-jwt/jwt/v5
```

```go
package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

type JWK struct {
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type AuthValidator struct {
	jwksURL   string
	publicKey *rsa.PublicKey
	mu        sync.RWMutex
	lastFetch time.Time
}

func NewAuthValidator(jwksURL string) *AuthValidator {
	v := &AuthValidator{jwksURL: jwksURL}
	_ = v.refreshKey()
	return v
}

func (v *AuthValidator) refreshKey() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	resp, err := http.Get(v.jwksURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}

	if len(jwks.Keys) == 0 {
		return errors.New("no keys in JWKS")
	}

	key := jwks.Keys[0]
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return err
	}

	var eInt int
	for _, b := range eBytes {
		eInt = (eInt << 8) | int(b)
	}

	v.publicKey = &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: eInt,
	}
	v.lastFetch = time.Now()
	return nil
}

func (v *AuthValidator) AuthenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := ""

		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		} else if cookie, err := r.Cookie("access_token"); err == nil {
			tokenStr = cookie.Value
		}

		if tokenStr == "" {
			http.Error(w, `{"error":"Missing authentication token"}`, http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, errors.New("unexpected signing method")
			}
			v.mu.RLock()
			pk := v.publicKey
			v.mu.RUnlock()
			return pk, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, `{"error":"Invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*UserClaims)
		if !ok {
			http.Error(w, `{"error":"Invalid token claims"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
```

---

## 6. Instant Revocation (Redis Blacklist)

When an admin suspends a user or when a logout/password-reset occurs, `store_auth` publishes a blacklist entry into Redis:
- **Redis Key**: `revoked:user:{userId}`
- **TTL**: Matching maximum remaining JWT lifespan (15 minutes).

In **Pattern A (Gateway Subrequest Offloading)**, `store_auth` automatically evaluates the Redis blacklist during the internal `/api/auth/me` subrequest. The API Gateway and downstream services do not need direct Redis access.

If using **Pattern B (Zero-Trust)** and your service requires real-time sub-second revocation enforcement:
```go
if rdb.Exists(ctx, "revoked:user:"+claims.Subject).Val() > 0 {
    http.Error(w, `{"error":"Account revoked"}`, http.StatusUnauthorized)
    return
}
```

---

## 7. Testing & Troubleshooting Checklist

1. **Anti-Spoofing Verification**: Test sending raw `X-User-Id: attacker-uuid` from Postman to the API Gateway. Ensure NGINX strips the header and the downstream service rejects or replaces it with the authenticated identity.
2. **Clock Skew**: Ensure all servers synchronize time using NTP. Tokens expired by more than a few seconds will be rejected.
3. **CORS & Cookies**: If using cookies across multiple subdomains (e.g. `auth.store.com` and `orders.store.com`), configure your DNS and cookie domain to `.store.com` with `SameSite=Lax` and `Secure=true`.
4. **Interactive Testing via Swagger UI**: Test your endpoints interactively anytime by visiting `http://localhost:8080/docs` (standalone) or `https://store-gateway.herokuapp.com/docs` (via Gateway).
