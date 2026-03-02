// Package auth: JWT-based authentication. Customer ID is stored in request context.
package auth

import "context"

// contextKey is the type for context keys (avoids collisions).
// LEARNING: Using a custom type (not string) for context keys prevents other packages from
// accidentally using the same key. context.WithValue(ctx, key, value) stores; ctx.Value(key) retrieves.
type contextKey string

const customerIDKey contextKey = "customer_id"

// CustomerIDFromContext returns the authenticated customer's ID from the request context.
// Returns 0 if not set (e.g. no token or invalid token).
func CustomerIDFromContext(ctx context.Context) int64 {
	if id, ok := ctx.Value(customerIDKey).(int64); ok {
		return id
	}
	return 0
}

// WithCustomerID puts the customer ID into the context for downstream handlers.
func WithCustomerID(ctx context.Context, customerID int64) context.Context {
	return context.WithValue(ctx, customerIDKey, customerID)
}
