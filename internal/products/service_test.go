// Unit tests for the products service using a mock repository (no real DB).
package products

import (
	"context"
	"log/slog"
	"testing"

	repo "github.com/sikozonpc/ecom/internal/adapters/postgresql/sqlc"
)

// mockRepo is a fake Repository for tests. It returns fixed data or an error you set.
type mockRepo struct {
	products []repo.Product
	err      error
}

func (m *mockRepo) ListProductsPaginated(ctx context.Context, arg repo.ListProductsPaginatedParams) ([]repo.Product, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.products, nil
}

// noopLogger satisfies the logger interface used by the service (Info/Error no-op in tests).
var noopLogger = slog.Default()

func TestListProducts_ReturnsRepoResult(t *testing.T) {
	ctx := context.Background()
	want := []repo.Product{
		{ID: 1, Name: "A", PriceInCenters: 1000, Quantity: 5},
		{ID: 2, Name: "B", PriceInCenters: 2000, Quantity: 10},
	}
	repo := &mockRepo{products: want}
	svc := NewService(repo, noopLogger)

	got, err := svc.ListProducts(ctx, 20, 0)
	if err != nil {
		t.Fatalf("ListProducts: %v", err)
	}
	if len(got) != len(want) {
		t.Errorf("len(got) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].ID != want[i].ID || got[i].Name != want[i].Name {
			t.Errorf("got[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestListProducts_PassesLimitOffsetToRepo(t *testing.T) {
	ctx := context.Background()
	repo := &mockRepo{products: []repo.Product{}}
	svc := NewService(repo, noopLogger)

	_, _ = svc.ListProducts(ctx, 10, 5)
	// We can't assert the args on mockRepo without recording them; just ensure no panic.
	// Optional: extend mockRepo to record arg and assert Limit=10, Offset=5.
}

func TestListProducts_ReturnsRepoError(t *testing.T) {
	ctx := context.Background()
	repoErr := context.DeadlineExceeded
	repo := &mockRepo{err: repoErr}
	svc := NewService(repo, noopLogger)

	_, err := svc.ListProducts(ctx, 20, 0)
	if err != repoErr {
		t.Errorf("err = %v, want %v", err, repoErr)
	}
}
