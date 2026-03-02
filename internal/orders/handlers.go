// Handlers: HTTP layer. Parse JSON body, get customerId from JWT context, call service.
//
// LEARNING: Handlers deal only with HTTP: read body, parse JSON, call service, write response.
// They map domain errors (ErrProductNotFound) to HTTP status codes (404). No business logic here.
package orders

import (
	"errors"
	"net/http"

	"github.com/sikozonpc/ecom/internal/auth"
	"github.com/sikozonpc/ecom/internal/json"
	"github.com/sikozonpc/ecom/internal/validate"
)

// NewHandler returns the HTTP handler for orders.
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
		json.WriteError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		Items []orderItem `json:"items"`
	}
	if err := json.Read(r, &body); err != nil {
		h.logger.Error("place order: bad body", "error", err)
		json.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	for _, item := range body.Items {
		if err := validate.ProductID(item.ProductID); err != nil {
			json.WriteError(w, r, http.StatusBadRequest, err.Error())
			return
		}
		if err := validate.OrderItemQuantity(item.Quantity); err != nil {
			json.WriteError(w, r, http.StatusBadRequest, err.Error())
			return
		}
	}

	tempOrder := createOrderParams{CustomerID: customerID, Items: body.Items}
	createdOrder, err := h.service.PlaceOrder(r.Context(), tempOrder)
	if err != nil {
		// LEARNING: errors.Is checks if err wraps ErrX. This works with fmt.Errorf("%w", ErrX).
		switch {
		case errors.Is(err, ErrInvalidInput):
			json.WriteError(w, r, http.StatusBadRequest, err.Error())
			return
		case errors.Is(err, ErrProductNotFound):
			json.WriteError(w, r, http.StatusNotFound, err.Error())
			return
		case errors.Is(err, ErrProductNoStock):
			json.WriteError(w, r, http.StatusConflict, err.Error())
			return
		default:
			h.logger.Error("place order failed", "error", err)
			json.WriteError(w, r, http.StatusInternalServerError, "internal server error")
			return
		}
	}
	json.Write(w, http.StatusCreated, createdOrder)
}
