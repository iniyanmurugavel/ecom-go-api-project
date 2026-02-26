// Handlers: HTTP layer. Parse JSON body, get customerId from JWT context, call service.
package orders

import (
	"errors"
	"net/http"

	"github.com/sikozonpc/ecom/internal/auth"
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

// PlaceOrder handles POST /v1/orders. Requires Authorization: Bearer <token>.
// Body: { "items": [ { "productId": 1, "quantity": 2 } ] }. Customer ID comes from JWT.
func (h *handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	customerID := auth.CustomerIDFromContext(r.Context())
	if customerID == 0 {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var body struct {
		Items []orderItem `json:"items"`
	}
	if err := json.Read(r, &body); err != nil {
		h.logger.Error("place order: bad body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tempOrder := createOrderParams{CustomerID: customerID, Items: body.Items}
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
