// Auth handlers: register and login. Return JWT on success.
package auth

import (
	"errors"
	"net/http"

	"github.com/sikozonpc/ecom/internal/json"
)

// NewHandler returns the auth HTTP handler.
func NewHandler(service Service, logger interface {
	Error(msg string, args ...any)
}) *handler {
	return &handler{service: service, logger: logger}
}

type handler struct {
	service Service
	logger  interface {
		Error(msg string, args ...any)
	}
}

// Register handles POST /v1/auth/register. Body: { email, password, name }.
func (h *handler) Register(w http.ResponseWriter, r *http.Request) {
	var in RegisterInput
	if err := json.Read(r, &in); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if in.Email == "" || in.Password == "" || in.Name == "" {
		http.Error(w, `{"error":"email, password, and name are required"}`, http.StatusBadRequest)
		return
	}
	customer, err := h.service.Register(r.Context(), in)
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			http.Error(w, `{"error":"email already registered"}`, http.StatusConflict)
			return
		}
		h.logger.Error("register failed", "error", err)
		http.Error(w, `{"error":"registration failed"}`, http.StatusInternalServerError)
		return
	}
	// Don't return password_hash; return customer without sensitive fields
	resp := map[string]any{
		"id":    customer.ID,
		"email": customer.Email,
		"name":  customer.Name,
	}
	json.Write(w, http.StatusCreated, resp)
}

// Login handles POST /v1/auth/login. Body: { email, password }. Returns JWT.
func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	var in LoginInput
	if err := json.Read(r, &in); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if in.Email == "" || in.Password == "" {
		http.Error(w, `{"error":"email and password are required"}`, http.StatusBadRequest)
		return
	}
	customerID, email, token, err := h.service.Login(r.Context(), in)
	if err != nil {
		if errors.Is(err, ErrInvalidCreds) {
			http.Error(w, `{"error":"invalid email or password"}`, http.StatusUnauthorized)
			return
		}
		h.logger.Error("login failed", "error", err)
		http.Error(w, `{"error":"login failed"}`, http.StatusInternalServerError)
		return
	}
	json.Write(w, http.StatusOK, map[string]any{
		"token":       token,
		"customer_id": customerID,
		"email":       email,
	})
}
