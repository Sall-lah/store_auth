package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"store_auth/internal/auth"
	"store_auth/internal/jwt"
	"store_auth/internal/middleware"
)

// SetupRouter mounts HTTP routes and attaches middleware chains for security, rate-limiting, and auth enforcement.
func SetupRouter(authHandler *auth.Handler, jwksHandler *jwt.Handler, rateLimiter *middleware.RateLimiter, jwtService *jwt.Service) http.Handler {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	// Public JWKS key set endpoint
	r.Get("/.well-known/jwks.json", jwksHandler.GetJWKS)

	// Public Auth endpoints protected by Redis rate limiting
	r.Route("/api/auth", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			if rateLimiter != nil {
				r.Use(rateLimiter.RateLimit)
			}
			r.Post("/register", authHandler.Register)
			r.Post("/verify-otp", authHandler.VerifyOTP)
			r.Post("/login", authHandler.Login)
			r.Post("/forgot-password", authHandler.ForgotPassword)
			r.Post("/reset-password", authHandler.ResetPassword)
		})

		r.Post("/logout", authHandler.Logout)

		// Protected endpoints requiring valid RS256 JWT cookie
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(jwtService))
			r.Get("/me", authHandler.GetMe)
		})
	})

	return r
}
