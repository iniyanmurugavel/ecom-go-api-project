// Package json helps handlers send JSON responses and parse JSON request bodies.
package json

import (
	"encoding/json"
	"net/http"
)

// Write sets Content-Type, status code, and encodes data as JSON. Use for all JSON responses.
func Write(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Read decodes request body into data; DisallowUnknownFields rejects extra JSON keys.
func Read(r *http.Request, data any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(data)
}