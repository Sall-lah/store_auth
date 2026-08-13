package jwt

import (
	"encoding/json"
	"net/http"
)

// Handler serves HTTP requests for the JWKS public key endpoint.
type Handler struct {
	jwtService *Service
}

// NewHandler constructs a JWKS Handler instance.
func NewHandler(jwtService *Service) *Handler {
	return &Handler{
		jwtService: jwtService,
	}
}

// GetJWKS handles HTTP GET requests for /.well-known/jwks.json to expose public keys for token verification.
func (h *Handler) GetJWKS(w http.ResponseWriter, r *http.Request) {
	jwks := h.jwtService.GetJWKS()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(jwks)
}
