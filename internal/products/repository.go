// Package products: repository interface for product data access.
// Defining this in the domain lets us test the service with a mock repo (no real DB).
package products

import (
	"context"

	repo "github.com/sikozonpc/ecom/internal/adapters/postgresql/sqlc"
)

// Repository is the interface the product service depends on.
// The sqlc-generated *repo.Queries implements this (ListProductsPaginated).
// In tests you can pass a fake that returns fixed data.
type Repository interface {
	ListProductsPaginated(ctx context.Context, arg repo.ListProductsPaginatedParams) ([]repo.Product, error)
}
