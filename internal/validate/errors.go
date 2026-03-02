package validate

import "errors"

var (
	ErrEmpty            = errors.New("field is required")
	ErrTooLong          = errors.New("field exceeds maximum length")
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong  = errors.New("password exceeds maximum length")
	ErrQuantityInvalid  = errors.New("quantity must be greater than 0")
	ErrProductIDInvalid = errors.New("product ID must be greater than 0")
)
