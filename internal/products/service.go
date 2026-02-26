// Package products: business logic for product listing. Service talks to DB via repo (sqlc).
package products

import (
	"context"

	repo "github.com/sikozonpc/ecom/internal/adapters/postgresql/sqlc"
)

// Service is the interface handlers depend on; keeps handlers testable without real DB.
type Service interface {
	ListProducts(ctx context.Context) ([]repo.Product, error)
}

type svc struct {
	repo repo.Querier // Querier = interface implemented by sqlc Queries (ListProducts, etc.)
}

func NewService(repo repo.Querier) Service {
	return &svc{repo: repo}
}

func (s *svc) ListProducts(ctx context.Context) ([]repo.Product, error) {
	return s.repo.ListProducts(ctx)
}
