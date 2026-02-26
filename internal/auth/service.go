// Auth service: register (create customer) and login (verify password, return JWT).
package auth

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	repo "github.com/sikozonpc/ecom/internal/adapters/postgresql/sqlc"
)

var (
	ErrEmailExists   = errors.New("email already registered")
	ErrInvalidCreds = errors.New("invalid email or password")
)

// RegisterInput is the request body for register.
type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// LoginInput is the request body for login.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Service defines auth operations.
type Service interface {
	Register(ctx context.Context, in RegisterInput) (repo.Customer, error)
	Login(ctx context.Context, in LoginInput) (customerID int64, email, token string, err error)
}

type svc struct {
	repo   *repo.Queries
	secret []byte
	logger interface {
		Error(msg string, args ...any)
	}
}

// NewService creates the auth service.
func NewService(repo *repo.Queries, secret []byte, logger interface {
	Error(msg string, args ...any)
}) Service {
	return &svc{repo: repo, secret: secret, logger: logger}
}

// Register hashes the password and creates a customer. Returns ErrEmailExists if email taken.
func (s *svc) Register(ctx context.Context, in RegisterInput) (repo.Customer, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return repo.Customer{}, err
	}
	customer, err := s.repo.CreateCustomer(ctx, repo.CreateCustomerParams{
		Email:        in.Email,
		PasswordHash: string(hash),
		Name:         in.Name,
	})
	if err != nil {
		// Check for unique violation (email exists)
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			return repo.Customer{}, ErrEmailExists
		}
		return repo.Customer{}, err
	}
	return customer, nil
}

// Login verifies email/password and returns a JWT.
func (s *svc) Login(ctx context.Context, in LoginInput) (customerID int64, email, token string, err error) {
	customer, err := s.repo.FindCustomerByEmail(ctx, in.Email)
	if err != nil {
		return 0, "", "", ErrInvalidCreds
	}
	if err := bcrypt.CompareHashAndPassword([]byte(customer.PasswordHash), []byte(in.Password)); err != nil {
		return 0, "", "", ErrInvalidCreds
	}
	tok, err := SignToken(s.secret, customer.ID, customer.Email)
	if err != nil {
		s.logger.Error("jwt sign failed", "error", err)
		return 0, "", "", err
	}
	return customer.ID, customer.Email, tok, nil
}
