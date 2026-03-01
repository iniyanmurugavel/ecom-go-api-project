// Package orders: business logic for placing an order.
// Validates input, runs everything in one DB transaction, rolls back on any error.
//
// LEARNING: This is the "service" layer. It contains business rules (validation, stock checks)
// and coordinates the repository. It does NOT know about HTTP — that's the handler's job.
package orders

import (
	"context"
	"errors"
	"fmt"

	repo "github.com/sikozonpc/ecom/internal/adapters/postgresql/sqlc"
)

// Domain errors: handlers map these to HTTP status (400, 404, 409, 500).
// LEARNING: We use sentinel errors (errors.New) so handlers can check with errors.Is(err, ErrX).
// This keeps domain errors separate from HTTP — the handler decides the status code.
var (
	ErrInvalidInput    = errors.New("invalid input")
	ErrProductNotFound = errors.New("product not found")
	ErrProductNoStock  = errors.New("product has not enough stock")
)

type svc struct {
	repo     OrderRepo     // WithTx(tx) gives an OrderTxRepo for the transaction
	beginner TxBeginner    // Begin(ctx) starts a transaction (e.g. *pgxpool.Pool)
	logger   interface {
		Info(msg string, args ...any)
		Error(msg string, args ...any)
	}
}

// NewService builds the order service. repo is usually *sqlc.Queries; beginner is usually *pgxpool.Pool.
// In tests you pass mocks that implement OrderRepo and TxBeginner.
func NewService(repo OrderRepo, beginner TxBeginner, logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}) Service {
	return &svc{repo: repo, beginner: beginner, logger: logger}
}

// PlaceOrder validates customerId and items, then runs in one transaction:
// create order, for each item check product and stock, create order_item. Rollback on any error.
func (s *svc) PlaceOrder(ctx context.Context, tempOrder createOrderParams) (repo.Order, error) {
	if tempOrder.CustomerID == 0 {
		return repo.Order{}, fmt.Errorf("%w: customer ID is required", ErrInvalidInput)
	}
	if len(tempOrder.Items) == 0 {
		return repo.Order{}, fmt.Errorf("%w: at least one item is required", ErrInvalidInput)
	}

	tx, err := s.beginner.Begin(ctx)
	if err != nil {
		return repo.Order{}, err
	}
	// LEARNING: defer tx.Rollback runs when the function returns. If we Commit first, Rollback
	// is a no-op. If we return early (error), Rollback undoes the partial work. This pattern
	// ensures we never leave a transaction open.
	defer tx.Rollback(ctx) // if we return before Commit, Rollback runs (no-op after Commit)

	qtx := s.repo.WithTx(tx) // all following queries use this transaction

	order, err := qtx.CreateOrder(ctx, tempOrder.CustomerID)
	if err != nil {
		return repo.Order{}, err
	}

	for _, item := range tempOrder.Items {
		product, err := qtx.FindProductByID(ctx, item.ProductID)
		if err != nil {
			return repo.Order{}, ErrProductNotFound
		}
		if product.Quantity < item.Quantity {
			return repo.Order{}, ErrProductNoStock
		}
		_, err = qtx.CreateOrderItem(ctx, repo.CreateOrderItemParams{
			OrderID:    order.ID,
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			PriceCents: product.PriceInCenters,
		})
		if err != nil {
			return repo.Order{}, err
		}
		// Decrement stock atomically. If another request took the last units, UPDATE affects 0 rows
		// and returns error (e.g. pgx.ErrNoRows). We return ErrProductNoStock; transaction rolls back.
		_, err = qtx.DecrementProductStock(ctx, repo.DecrementProductStockParams{
			Quantity: item.Quantity,
			ID:       item.ProductID,
		})
		if err != nil {
			return repo.Order{}, ErrProductNoStock
		}
	}

	tx.Commit(ctx)
	return order, nil
}
