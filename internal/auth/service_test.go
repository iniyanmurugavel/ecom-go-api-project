// Unit tests for the auth service using mock AuthRepository.
package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	repo "github.com/sikozonpc/ecom/internal/adapters/postgresql/sqlc"
	"golang.org/x/crypto/bcrypt"
)

func bcryptHash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

type mockAuthRepo struct {
	createCustomer      func(context.Context, repo.CreateCustomerParams) (repo.CreateCustomerRow, error)
	findCustomerByEmail func(context.Context, string) (repo.Customer, error)
}

func (m *mockAuthRepo) CreateCustomer(ctx context.Context, arg repo.CreateCustomerParams) (repo.CreateCustomerRow, error) {
	if m.createCustomer != nil {
		return m.createCustomer(ctx, arg)
	}
	return repo.CreateCustomerRow{ID: 1, Email: arg.Email, Name: arg.Name}, nil
}

func (m *mockAuthRepo) FindCustomerByEmail(ctx context.Context, email string) (repo.Customer, error) {
	if m.findCustomerByEmail != nil {
		return m.findCustomerByEmail(ctx, email)
	}
	return repo.Customer{}, errors.New("not found")
}

func TestRegister_EmailExists(t *testing.T) {
	mockRepo := &mockAuthRepo{
		createCustomer: func(ctx context.Context, arg repo.CreateCustomerParams) (repo.CreateCustomerRow, error) {
			return repo.CreateCustomerRow{}, &pgconn.PgError{Code: "23505"}
		},
	}
	svc := NewService(mockRepo)
	_, _, _, err := svc.Register(context.Background(), "a@b.com", "password123", "Alice")
	if !errors.Is(err, ErrEmailExists) {
		t.Errorf("err = %v, want ErrEmailExists", err)
	}
}

func TestRegister_Success(t *testing.T) {
	mockRepo := &mockAuthRepo{
		createCustomer: func(ctx context.Context, arg repo.CreateCustomerParams) (repo.CreateCustomerRow, error) {
			return repo.CreateCustomerRow{ID: 42, Email: arg.Email, Name: arg.Name}, nil
		},
	}
	svc := NewService(mockRepo)
	id, email, name, err := svc.Register(context.Background(), "a@b.com", "password123", "Alice")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if id != 42 || email != "a@b.com" || name != "Alice" {
		t.Errorf("got id=%d email=%s name=%s", id, email, name)
	}
}

func TestLogin_InvalidCreds_NotFound(t *testing.T) {
	mockRepo := &mockAuthRepo{
		findCustomerByEmail: func(ctx context.Context, email string) (repo.Customer, error) {
			return repo.Customer{}, errors.New("no rows")
		},
	}
	svc := NewService(mockRepo)
	_, _, _, err := svc.Login(context.Background(), "a@b.com", "wrong")
	if !errors.Is(err, ErrInvalidCreds) {
		t.Errorf("err = %v, want ErrInvalidCreds", err)
	}
}

func TestLogin_InvalidCreds_BadPassword(t *testing.T) {
	hash, _ := bcryptHash("correct")
	mockRepo := &mockAuthRepo{
		findCustomerByEmail: func(ctx context.Context, email string) (repo.Customer, error) {
			return repo.Customer{ID: 1, Email: email, PasswordHash: hash, Name: "Alice"}, nil
		},
	}
	svc := NewService(mockRepo)
	_, _, _, err := svc.Login(context.Background(), "a@b.com", "wrongpassword")
	if !errors.Is(err, ErrInvalidCreds) {
		t.Errorf("err = %v, want ErrInvalidCreds", err)
	}
}

func TestLogin_Success(t *testing.T) {
	hash, _ := bcryptHash("secret123")
	mockRepo := &mockAuthRepo{
		findCustomerByEmail: func(ctx context.Context, email string) (repo.Customer, error) {
			return repo.Customer{ID: 99, Email: "a@b.com", PasswordHash: hash, Name: "Alice"}, nil
		},
	}
	svc := NewService(mockRepo)
	_, customerID, email, err := svc.Login(context.Background(), "a@b.com", "secret123")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if customerID != 99 || email != "a@b.com" {
		t.Errorf("got customerID=%d email=%s", customerID, email)
	}
}
