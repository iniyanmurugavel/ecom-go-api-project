package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	internaljson "github.com/sikozonpc/ecom/internal/json"
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
	svc    Service
	secret []byte
}

func NewHandler(svc Service, secret []byte) *Handler {
	return &Handler{svc: svc, secret: secret}
}

// Register handles POST /v1/auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := internaljson.Read(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Email == "" || req.Password == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "email, password, and name are required")
		return
	}
	id, email, name, err := h.svc.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			writeError(w, http.StatusConflict, "email already registered")
			return
		}
		log.Println("auth register:", err)
		writeError(w, http.StatusInternalServerError, "registration failed")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":    id,
		"email": email,
		"name":  name,
	})
}

// Login handles POST /v1/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := internaljson.Read(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}
	_, customerID, email, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCreds) {
			writeError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		log.Println("auth login:", err)
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}
	token, err := SignToken(h.secret, customerID, email)
	if err != nil {
		log.Println("auth sign token:", err)
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}
	internaljson.Write(w, http.StatusOK, loginResponse{
		Token:      token,
		CustomerID: customerID,
		Email:      email,
	})
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
