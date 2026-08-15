package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/redis/go-redis/v9"
	"store_auth/internal/jwt"
	platformRedis "store_auth/internal/platform/redis"
)

type contextKey string

const (
	UserContextKey contextKey = "user_claims"
)

type errorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

// Authenticate returns middleware that extracts and validates the RS256 JWT access token from the access_token HttpOnly cookie,
// and ensures the user account is not revoked in the Redis blacklist.
// Why: Enforces token validity statelessly while verifying real-time user revocation status without querying PostgreSQL.
func Authenticate(jwtService *jwt.Service, rdb *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("access_token")
			if err != nil || cookie.Value == "" {
				respondWithError(w, http.StatusUnauthorized, "Authentication cookie missing or empty")
				return
			}

			claims, err := jwtService.ValidateToken(cookie.Value)
			if err != nil {
				respondWithError(w, http.StatusUnauthorized, "Invalid or expired authentication token")
				return
			}

			if platformRedis.IsUserBlacklisted(r.Context(), rdb, claims.Subject) {
				respondWithError(w, http.StatusUnauthorized, "Account has been revoked")
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole returns middleware ensuring the authenticated user's role satisfies access requirements.
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserContextKey).(*jwt.Claims)
			if !ok || claims == nil {
				respondWithError(w, http.StatusUnauthorized, "Authentication required")
				return
			}

			allowed := false
			for _, r := range allowedRoles {
				if claims.Role == r {
					allowed = true
					break
				}
			}

			if !allowed {
				respondWithError(w, http.StatusForbidden, "Forbidden: insufficient role permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserClaimsFromContext extracts claims populated by the Authenticate middleware.
func GetUserClaimsFromContext(ctx context.Context) (*jwt.Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*jwt.Claims)
	return claims, ok
}

func respondWithError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{
		Error: message,
	})
}
