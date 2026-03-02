// Handlers: HTTP layer only. Parse request, call service, map errors to status codes, write JSON.
// No business logic here — that lives in the service.
//
// LEARNING: Handler reads query params (limit, offset), applies defaults, calls service,
// and writes JSON. Pagination params are validated here; the service just passes them to the repo.
package products

import (
	"net/http"
	"strconv"

	"github.com/sikozonpc/ecom/internal/json"
)

// Default pagination when client does not send limit/offset.
const (
	DefaultLimit  = 20
	DefaultOffset = 0
	MaxLimit      = 100
)

// NewHandler returns the HTTP handler for products. It needs the service and a logger (e.g. *slog.Logger).
func NewHandler(service Service, logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}) *handler {
	return &handler{service: service, logger: logger}
}

type handler struct {
	service Service
	logger  interface {
		Info(msg string, args ...any)
		Error(msg string, args ...any)
	}
}

// ListProducts handles GET /v1/products. Query params: limit (default 20, max 100), offset (default 0).
// Returns 200 + JSON array of products, or 500 on DB error.
func (h *handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	limit := int32(DefaultLimit)
	offset := int32(DefaultOffset)
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 32); err == nil && n > 0 {
			limit = int32(n)
			if limit > MaxLimit {
				limit = MaxLimit
			}
		}
	}
	if s := r.URL.Query().Get("offset"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 32); err == nil && n >= 0 {
			offset = int32(n)
		}
	}

	products, err := h.service.ListProducts(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("list products failed", "error", err)
		json.WriteError(w, r, http.StatusInternalServerError, "internal server error")
		return
	}
	json.Write(w, http.StatusOK, products)
}
