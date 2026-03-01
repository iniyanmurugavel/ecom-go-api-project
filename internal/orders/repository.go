// Package orders: repository interfaces for order data access.
// The service uses these instead of concrete *repo.Queries so we can test with mocks.
//
// LEARNING: Interfaces let us "inject" different implementations. In production we use
// *repo.Queries (sqlc-generated). In tests we use mockOrderTxRepo that returns fixed data.
// The service doesn't know or care which one it gets — that's dependency injection.
package orders

import (
	"context"

	"github.com/jackc/pgx/v5"
	repo "github.com/sikozonpc/ecom/internal/adapters/postgresql/sqlc"
)

// OrderTxRepo runs inside a transaction (CreateOrder, FindProductByID, CreateOrderItem, DecrementProductStock).
// *repo.Queries returned by WithTx(tx) implements this.
type OrderTxRepo interface {
	CreateOrder(ctx context.Context, customerID int64) (repo.Order, error)
	FindProductByID(ctx context.Context, id int64) (repo.Product, error)
	CreateOrderItem(ctx context.Context, arg repo.CreateOrderItemParams) (repo.OrderItem, error)
	DecrementProductStock(ctx context.Context, arg repo.DecrementProductStockParams) (int64, error)
}

// OrderRepo can run queries in a transaction via WithTx.
// *repo.Queries implements this: WithTx(tx) returns a *Queries that implements OrderTxRepo.
type OrderRepo interface {
	WithTx(tx pgx.Tx) OrderTxRepo
}

// TxBeginner starts a DB transaction. *pgxpool.Pool implements this.
// The service calls Begin(ctx) then repo.WithTx(tx) to run order logic in one transaction.
type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// NewOrderRepo wraps the sqlc *Queries so it implements OrderRepo.
// WithTx returns *Queries, which implements OrderTxRepo; we expose it as OrderTxRepo so the interface is satisfied.
func NewOrderRepo(q *repo.Queries) OrderRepo {
	return &orderRepoAdapter{q: q}
}

type orderRepoAdapter struct {
	q *repo.Queries
}

func (a *orderRepoAdapter) WithTx(tx pgx.Tx) OrderTxRepo {
	return a.q.WithTx(tx)
}
