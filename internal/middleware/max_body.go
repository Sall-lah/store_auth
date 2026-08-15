package middleware

import (
	"net/http"
)

// DefaultMaxBodyBytes defines the maximum allowed size (1 MB) for HTTP JSON request payloads.
const DefaultMaxBodyBytes int64 = 1 << 20 // 1,048,576 bytes (1 MB)

// MaxBody returns an HTTP middleware that limits the request payload read stream.
// Why: Shields the server against Denial-of-Service (DoS) and memory exhaustion attacks caused by arbitrarily large incoming request payloads.
func MaxBody(maxBytes int64) func(http.Handler) http.Handler {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBodyBytes
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// LimitRequestBody wraps an HTTP request's body with http.MaxBytesReader.
// Why: Provides direct, programmatic payload restriction when handler-level enforcement is required independently of middleware routing chains.
func LimitRequestBody(w http.ResponseWriter, r *http.Request, maxBytes int64) {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBodyBytes
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
}
