// Package products: business logic for product listing.
// The service talks to the DB via the Repository interface (no raw SQL here).
package products

import (
	"context"

	repo "github.com/sikozonpc/ecom/internal/adapters/postgresql/sqlc"
)

// Service is the interface handlers depend on. Easy to mock in tests (no real DB).
type Service interface {
	// ListProducts returns products with pagination. Use limit/offset from query params (e.g. ?limit=20&offset=0).
	ListProducts(ctx context.Context, limit, offset int32) ([]repo.Product, error)
}

type svc struct {
	repo   Repository
	logger interface {
		Info(msg string, args ...any)
		Error(msg string, args ...any)
	}
}

// NewService builds the product service. repo is usually *sqlc.Queries; in tests pass a mock Repository.
func NewService(repo Repository, logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}) Service {
	return &svc{repo: repo, logger: logger}
}

// ListProducts calls the repo with limit/offset. Defaults are applied in the handler (e.g. limit 20, offset 0).
func (s *svc) ListProducts(ctx context.Context, limit, offset int32) ([]repo.Product, error) {
	return s.repo.ListProductsPaginated(ctx, repo.ListProductsPaginatedParams{Limit: limit, Offset: offset})
}
