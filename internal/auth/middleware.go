// Middleware: extract Bearer token, verify JWT, put customer ID in context.
//
// LEARNING: Middleware is a function that returns func(http.Handler) http.Handler. It wraps
// the next handler. Here we: 1) read Authorization header, 2) verify JWT, 3) put customer ID
// in context via WithCustomerID, 4) call next.ServeHTTP with the updated context.
package auth

import (
	"net/http"
	"strings"

	"github.com/sikozonpc/ecom/internal/json"
)

// RequireAuth returns middleware that validates the JWT and adds customer ID to context.
// Expects "Authorization: Bearer <token>". Returns 401 JSON if missing or invalid.
func RequireAuth(cfg *JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				json.WriteError(w, r, http.StatusUnauthorized, "missing Authorization header")
				return
			}
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				json.WriteError(w, r, http.StatusUnauthorized, "invalid Authorization format, use Bearer <token>")
				return
			}
			tokenString := strings.TrimSpace(parts[1])
			if tokenString == "" {
				json.WriteError(w, r, http.StatusUnauthorized, "missing token")
				return
			}
			claims, err := VerifyToken(cfg, tokenString)
			if err != nil {
				json.WriteError(w, r, http.StatusUnauthorized, "invalid or expired token")
				return
			}
			// LEARNING: Context carries request-scoped values. We add customer ID so handlers
			// can get it via CustomerIDFromContext(r.Context()) without parsing the token again.
			ctx := WithCustomerID(r.Context(), claims.CustomerID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
