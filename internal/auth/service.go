// Package auth: business logic for registration and login.
// Uses bcrypt for password hashing, JWT for tokens. Handlers sign the token; service returns user info.
package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	repo "github.com/sikozonpc/ecom/internal/adapters/postgresql/sqlc"
)

var (
	ErrEmailExists  = errors.New("email already registered")
	ErrInvalidCreds = errors.New("invalid email or password")
)

type Service interface {
	Register(ctx context.Context, email, password, name string) (id int64, emailOut, nameOut string, err error)
	Login(ctx context.Context, email, password string) (token string, customerID int64, emailOut string, err error)
}

type svc struct {
	repo AuthRepository
}

func NewService(repo AuthRepository) Service {
	return &svc{repo: repo}
}

func (s *svc) Register(ctx context.Context, email, password, name string) (int64, string, string, error) {
	// LEARNING: Never store plain passwords. bcrypt hashes with a salt; CompareHashAndPassword
	// verifies later. DefaultCost is 10 (2^10 iterations — balance of security and speed).
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, "", "", err
	}
	row, err := s.repo.CreateCustomer(ctx, repo.CreateCustomerParams{
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return 0, "", "", ErrEmailExists
		}
		return 0, "", "", err
	}
	return row.ID, row.Email, row.Name, nil
}

func (s *svc) Login(ctx context.Context, email, password string) (string, int64, string, error) {
	cust, err := s.repo.FindCustomerByEmail(ctx, email)
	if err != nil {
		return "", 0, "", ErrInvalidCreds
	}
	if err := bcrypt.CompareHashAndPassword([]byte(cust.PasswordHash), []byte(password)); err != nil {
		return "", 0, "", ErrInvalidCreds
	}
	return "", cust.ID, cust.Email, nil
}

// isUniqueViolation checks if the error is a PostgreSQL unique constraint violation.
// LEARNING: PostgreSQL error code 23505 = unique_violation. We use errors.As to unwrap
// the error chain and check if it's a *pgconn.PgError with that code.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
