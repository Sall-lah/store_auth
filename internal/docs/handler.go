package docs

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.yaml
var openAPISpecYAML []byte

//go:embed index.html
var swaggerUIHTML []byte

// Handler serves embedded API documentation and interactive Swagger UI assets.
type Handler struct{}

// NewHandler constructs a docs Handler instance.
// Why: Provides a self-contained HTTP handler for Swagger UI and raw OpenAPI specs without external filesystem dependencies.
func NewHandler() *Handler {
	return &Handler{}
}

// ServeUI serves the Swagger UI HTML page.
// Why: Enables interactive API exploration, endpoint testing, and schema visualization in the browser.
func (h *Handler) ServeUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(swaggerUIHTML)
}

// ServeSpecYAML serves the raw OpenAPI 3.1 YAML definition.
// Why: Allows automated client generators (e.g. oapi-codegen, openapi-generator) and linters to download the API contract.
func (h *Handler) ServeSpecYAML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openAPISpecYAML)
}
