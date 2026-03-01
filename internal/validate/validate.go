// Package validate provides input validation helpers.
//
// LEARNING: We validate at the handler layer before calling the service. This keeps invalid
// data out of the business logic. Each function returns an error from errors.go (e.g. ErrEmpty).
package validate

import (
	"net/mail"
	"strings"
)

const (
	MaxEmailLen   = 254
	MaxPasswordLen = 72 // bcrypt limit
	MinPasswordLen = 8
	MaxNameLen    = 100
)

// Email checks format and length.
func Email(s string) error {
	if s == "" {
		return ErrEmpty
	}
	if len(s) > MaxEmailLen {
		return ErrTooLong
	}
	if _, err := mail.ParseAddress(s); err != nil {
		return ErrInvalidEmail
	}
	return nil
}

// Password checks length (bcrypt has 72 byte limit).
func Password(s string) error {
	if s == "" {
		return ErrEmpty
	}
	if len(s) < MinPasswordLen {
		return ErrPasswordTooShort
	}
	if len(s) > MaxPasswordLen {
		return ErrPasswordTooLong
	}
	return nil
}

// Name checks non-empty and max length.
func Name(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return ErrEmpty
	}
	if len(s) > MaxNameLen {
		return ErrTooLong
	}
	return nil
}

// OrderItemQuantity checks quantity > 0.
func OrderItemQuantity(q int32) error {
	if q <= 0 {
		return ErrQuantityInvalid
	}
	return nil
}

// ProductID checks productId > 0.
func ProductID(id int64) error {
	if id <= 0 {
		return ErrProductIDInvalid
	}
	return nil
}
