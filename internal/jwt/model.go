package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims payload structure used for cross-service authentication.
type Claims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// JWK represents a single JSON Web Key in JWKS specification (RFC 7517).
type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS represents a JSON Web Key Set container exposing public keys.
type JWKS struct {
	Keys []JWK `json:"keys"`
}
