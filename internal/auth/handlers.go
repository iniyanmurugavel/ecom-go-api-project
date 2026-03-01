package auth

import (
	"errors"
	"net/http"

	"github.com/sikozonpc/ecom/internal/json"
	"github.com/sikozonpc/ecom/internal/validate"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token      string `json:"token"`
	CustomerID int64  `json:"customer_id"`
	Email      string `json:"email"`
}

type Handler struct {
	svc       Service
	jwtConfig *JWTConfig
	logger    interface {
		Error(msg string, args ...any)
	}
}

func NewHandler(svc Service, jwtConfig *JWTConfig, logger interface {
	Error(msg string, args ...any)
}) *Handler {
	return &Handler{svc: svc, jwtConfig: jwtConfig, logger: logger}
}

// Register handles POST /v1/auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.Read(r, &req); err != nil {
		json.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := validate.Email(req.Email); err != nil {
		json.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := validate.Password(req.Password); err != nil {
		json.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := validate.Name(req.Name); err != nil {
		json.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	id, email, name, err := h.svc.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			json.WriteError(w, r, http.StatusConflict, "email already registered")
			return
		}
		h.logger.Error("auth register failed", "error", err)
		json.WriteError(w, r, http.StatusInternalServerError, "registration failed")
		return
	}
	json.Write(w, http.StatusCreated, map[string]any{
		"id":    id,
		"email": email,
		"name":  name,
	})
}

// Login handles POST /v1/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.Read(r, &req); err != nil {
		json.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := validate.Email(req.Email); err != nil {
		json.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if req.Password == "" {
		json.WriteError(w, r, http.StatusBadRequest, "password is required")
		return
	}
	_, customerID, email, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCreds) {
			json.WriteError(w, r, http.StatusUnauthorized, "invalid email or password")
			return
		}
		h.logger.Error("auth login failed", "error", err)
		json.WriteError(w, r, http.StatusInternalServerError, "login failed")
		return
	}
	token, err := SignToken(h.jwtConfig, customerID, email)
	if err != nil {
		h.logger.Error("auth sign token failed", "error", err)
		json.WriteError(w, r, http.StatusInternalServerError, "login failed")
		return
	}
	json.Write(w, http.StatusOK, loginResponse{
		Token:      token,
		CustomerID: customerID,
		Email:      email,
	})
}
