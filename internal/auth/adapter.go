// Package auth: adapter that wraps sqlc Queries to implement AuthRepository.
package auth

import (
	"context"

	repo "github.com/sikozonpc/ecom/internal/adapters/postgresql/sqlc"
)

// NewAuthRepo returns an AuthRepository backed by *repo.Queries.
func NewAuthRepo(q *repo.Queries) AuthRepository {
	return &authRepoAdapter{q: q}
}

type authRepoAdapter struct {
	q *repo.Queries
}

func (a *authRepoAdapter) CreateCustomer(ctx context.Context, arg repo.CreateCustomerParams) (repo.CreateCustomerRow, error) {
	return a.q.CreateCustomer(ctx, arg)
}

func (a *authRepoAdapter) FindCustomerByEmail(ctx context.Context, email string) (repo.Customer, error) {
	return a.q.FindCustomerByEmail(ctx, email)
}
