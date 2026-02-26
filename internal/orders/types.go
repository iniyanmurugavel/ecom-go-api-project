// Types and service interface for orders. Request body has items only; customerId comes from JWT.
package orders

import (
	"context"

	repo "github.com/sikozonpc/ecom/internal/adapters/postgresql/sqlc"
)

type orderItem struct {
	ProductID int64 `json:"productId"`
	Quantity  int32 `json:"quantity"`
}

// createOrderParams: customerID is from auth context (JWT), not from body.
type createOrderParams struct {
	CustomerID int64
	Items      []orderItem `json:"items"`
}

// Service: handlers call PlaceOrder; service runs in a DB transaction and returns repo.Order.
type Service interface {
	PlaceOrder(ctx context.Context, tempOrder createOrderParams) (repo.Order, error)
}