// Package auth: repository interface for customer data access.
// Allows testing the auth service with a mock (no real DB).
//
// LEARNING: AuthRepository defines what the auth service needs from the DB. The adapter
// (auth/adapter.go) wraps *repo.Queries to implement this. Tests use mockAuthRepo.
package auth

import (
	"context"

	repo "github.com/sikozonpc/ecom/internal/adapters/postgresql/sqlc"
)

// AuthRepository is the interface the auth service depends on.
// *repo.Queries implements this via the authRepoAdapter.
type AuthRepository interface {
	CreateCustomer(ctx context.Context, arg repo.CreateCustomerParams) (repo.CreateCustomerRow, error)
	FindCustomerByEmail(ctx context.Context, email string) (repo.Customer, error)
}
