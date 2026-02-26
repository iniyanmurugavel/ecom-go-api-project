// Middleware: extract Bearer token, verify JWT, put customer ID in context.
package auth

import (
	"net/http"
	"strings"
)

// RequireAuth returns middleware that validates the JWT and adds customer ID to context.
// Expects "Authorization: Bearer <token>". Returns 401 JSON if missing or invalid.
func RequireAuth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"missing Authorization header"}`, http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"invalid Authorization format, use Bearer <token>"}`, http.StatusUnauthorized)
				return
			}
			tokenString := strings.TrimSpace(parts[1])
			if tokenString == "" {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"missing token"}`, http.StatusUnauthorized)
				return
			}
			claims, err := VerifyToken(secret, tokenString)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}
			ctx := WithCustomerID(r.Context(), claims.CustomerID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
