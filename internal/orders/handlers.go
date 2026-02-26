// Handlers: HTTP layer. Parse JSON body, call service, map domain errors to status codes (400, 404, 409, 500), write JSON.
package orders

import (
	"errors"
	"net/http"

	"github.com/sikozonpc/ecom/internal/json"
)

// NewHandler returns the HTTP handler for orders. Needs the service and a logger.
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

// PlaceOrder handles POST /v1/orders. Body: { "customerId": 1, "items": [ { "productId": 1, "quantity": 2 } ] }.
// Returns 201 + order JSON, or 400 (bad request), 404 (product not found), 409 (not enough stock), 500 (server error).
func (h *handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	var tempOrder createOrderParams
	if err := json.Read(r, &tempOrder); err != nil {
		h.logger.Error("place order: bad body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdOrder, err := h.service.PlaceOrder(r.Context(), tempOrder)
	if err != nil {
		// Map domain errors to HTTP status so clients get stable status codes
		switch {
		case errors.Is(err, ErrInvalidInput):
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		case errors.Is(err, ErrProductNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		case errors.Is(err, ErrProductNoStock):
			http.Error(w, err.Error(), http.StatusConflict) // 409 Conflict
			return
		default:
			h.logger.Error("place order failed", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}
	json.Write(w, http.StatusCreated, createdOrder)
}
