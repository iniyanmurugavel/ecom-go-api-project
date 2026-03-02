// Package json helps handlers send JSON responses and parse JSON request bodies.
//
// LEARNING: Centralizing JSON I/O here ensures consistent Content-Type, error format, and
// request_id in errors. DisallowUnknownFields rejects extra keys (helps catch client bugs).
package json

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// Write sets Content-Type, status code, and encodes data as JSON. Use for all JSON responses.
func Write(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// ErrorResponse is the standard JSON error shape. Use WriteError for consistent error responses.
type ErrorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

// WriteError writes a JSON error response with status code. Includes request_id from chi middleware when available.
func WriteError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := ErrorResponse{Error: msg}
	if id := middleware.GetReqID(r.Context()); id != "" {
		resp.RequestID = id
	}
	json.NewEncoder(w).Encode(resp)
}

// Read decodes request body into data; DisallowUnknownFields rejects extra JSON keys.
// LEARNING: "data any" means any type — we pass a pointer to a struct. json.Decoder fills it.
func Read(r *http.Request, data any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(data)
}