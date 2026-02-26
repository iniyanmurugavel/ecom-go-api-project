// Unit tests for the orders service using mock OrderRepo and TxBeginner (no real DB).
package orders

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	repo "github.com/sikozonpc/ecom/internal/adapters/postgresql/sqlc"
)

var noopLogger = slog.Default()

// fakeTx is a minimal pgx.Tx for tests. Only Commit and Rollback are used by the service.
// WithTx on the mock repo ignores the tx and returns mock OrderTxRepo, so we never call Query/Exec on this.
type fakeTx struct{}

func (fakeTx) Begin(context.Context) (pgx.Tx, error) { return fakeTx{}, nil }
func (fakeTx) Commit(context.Context) error         { return nil }
func (fakeTx) Rollback(context.Context) error       { return nil }
func (fakeTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	panic("not used in test")
}
func (fakeTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { panic("not used in test") }
func (fakeTx) LargeObjects() pgx.LargeObjects                        { panic("not used in test") }
func (fakeTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	panic("not used in test")
}
func (fakeTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	panic("not used in test")
}
func (fakeTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	panic("not used in test")
}
func (fakeTx) QueryRow(context.Context, string, ...any) pgx.Row {
	panic("not used in test")
}
func (fakeTx) Conn() *pgx.Conn { return nil }

// Ensure fakeTx implements pgx.Tx at compile time.
var _ pgx.Tx = fakeTx{}

// mockOrderTxRepo controls CreateOrder, FindProductByID, CreateOrderItem for tests.
type mockOrderTxRepo struct {
	createOrder     func(context.Context, int64) (repo.Order, error)
	findProductByID func(context.Context, int64) (repo.Product, error)
	createOrderItem func(context.Context, repo.CreateOrderItemParams) (repo.OrderItem, error)
}

func (m *mockOrderTxRepo) CreateOrder(ctx context.Context, customerID int64) (repo.Order, error) {
	if m.createOrder != nil {
		return m.createOrder(ctx, customerID)
	}
	return repo.Order{ID: 1, CustomerID: customerID}, nil
}

func (m *mockOrderTxRepo) FindProductByID(ctx context.Context, id int64) (repo.Product, error) {
	if m.findProductByID != nil {
		return m.findProductByID(ctx, id)
	}
	return repo.Product{ID: id, Quantity: 100}, nil
}

func (m *mockOrderTxRepo) CreateOrderItem(ctx context.Context, arg repo.CreateOrderItemParams) (repo.OrderItem, error) {
	if m.createOrderItem != nil {
		return m.createOrderItem(ctx, arg)
	}
	return repo.OrderItem{}, nil
}

// mockOrderRepo returns a fixed OrderTxRepo from WithTx (ignores tx).
type mockOrderRepo struct {
	txRepo OrderTxRepo
}

func (m *mockOrderRepo) WithTx(pgx.Tx) OrderTxRepo { return m.txRepo }

// mockTxBeginner returns (tx, nil) or (nil, err) for tests.
type mockTxBeginner struct {
	tx  pgx.Tx
	err error
}

func (m *mockTxBeginner) Begin(context.Context) (pgx.Tx, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.tx != nil {
		return m.tx, nil
	}
	return fakeTx{}, nil
}

func TestPlaceOrder_InvalidInput_MissingCustomerID(t *testing.T) {
	svc := NewService(&mockOrderRepo{txRepo: &mockOrderTxRepo{}}, &mockTxBeginner{}, noopLogger)
	_, err := svc.PlaceOrder(context.Background(), createOrderParams{Items: []orderItem{{ProductID: 1, Quantity: 1}}})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestPlaceOrder_InvalidInput_EmptyItems(t *testing.T) {
	svc := NewService(&mockOrderRepo{txRepo: &mockOrderTxRepo{}}, &mockTxBeginner{}, noopLogger)
	_, err := svc.PlaceOrder(context.Background(), createOrderParams{CustomerID: 1, Items: nil})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
	_, err = svc.PlaceOrder(context.Background(), createOrderParams{CustomerID: 1, Items: []orderItem{}})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestPlaceOrder_BeginError(t *testing.T) {
	beginErr := errors.New("db unavailable")
	svc := NewService(&mockOrderRepo{}, &mockTxBeginner{err: beginErr}, noopLogger)
	_, err := svc.PlaceOrder(context.Background(), createOrderParams{CustomerID: 1, Items: []orderItem{{ProductID: 1, Quantity: 1}}})
	if err != beginErr {
		t.Errorf("err = %v, want %v", err, beginErr)
	}
}

func TestPlaceOrder_ProductNotFound(t *testing.T) {
	txRepo := &mockOrderTxRepo{
		createOrder: func(ctx context.Context, customerID int64) (repo.Order, error) {
			return repo.Order{ID: 99, CustomerID: customerID}, nil
		},
		findProductByID: func(ctx context.Context, id int64) (repo.Product, error) {
			return repo.Product{}, errors.New("no rows")
		},
	}
	svc := NewService(&mockOrderRepo{txRepo: txRepo}, &mockTxBeginner{}, noopLogger)
	_, err := svc.PlaceOrder(context.Background(), createOrderParams{
		CustomerID: 1,
		Items:      []orderItem{{ProductID: 999, Quantity: 1}},
	})
	if !errors.Is(err, ErrProductNotFound) {
		t.Errorf("err = %v, want ErrProductNotFound", err)
	}
}

func TestPlaceOrder_NotEnoughStock(t *testing.T) {
	txRepo := &mockOrderTxRepo{
		createOrder: func(ctx context.Context, customerID int64) (repo.Order, error) {
			return repo.Order{ID: 99, CustomerID: customerID}, nil
		},
		findProductByID: func(ctx context.Context, id int64) (repo.Product, error) {
			return repo.Product{ID: id, Quantity: 2, PriceInCenters: 1000}, nil // only 2 in stock
		},
	}
	svc := NewService(&mockOrderRepo{txRepo: txRepo}, &mockTxBeginner{}, noopLogger)
	_, err := svc.PlaceOrder(context.Background(), createOrderParams{
		CustomerID: 1,
		Items:      []orderItem{{ProductID: 1, Quantity: 5}}, // ask for 5
	})
	if !errors.Is(err, ErrProductNoStock) {
		t.Errorf("err = %v, want ErrProductNoStock", err)
	}
}

func TestPlaceOrder_Success(t *testing.T) {
	orderID := int64(42)
	txRepo := &mockOrderTxRepo{
		createOrder: func(ctx context.Context, customerID int64) (repo.Order, error) {
			return repo.Order{ID: orderID, CustomerID: customerID}, nil
		},
		findProductByID: func(ctx context.Context, id int64) (repo.Product, error) {
			return repo.Product{ID: id, Quantity: 10, PriceInCenters: 1999}, nil
		},
	}
	svc := NewService(&mockOrderRepo{txRepo: txRepo}, &mockTxBeginner{}, noopLogger)
	order, err := svc.PlaceOrder(context.Background(), createOrderParams{
		CustomerID: 1,
		Items:      []orderItem{{ProductID: 1, Quantity: 2}},
	})
	if err != nil {
		t.Fatalf("PlaceOrder: %v", err)
	}
	if order.ID != orderID || order.CustomerID != 1 {
		t.Errorf("order = %+v", order)
	}
}
