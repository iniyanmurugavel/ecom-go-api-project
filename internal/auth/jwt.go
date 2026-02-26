// JWT: sign and verify tokens. Claims include customer ID and email.
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims holds the JWT payload. jwt.RegisteredClaims adds exp, iat, etc.
type Claims struct {
	CustomerID int64  `json:"customer_id"`
	Email      string `json:"email"`
	jwt.RegisteredClaims
}

// DefaultExpiry is how long a token is valid (24 hours).
const DefaultExpiry = 24 * time.Hour

var (
	ErrInvalidToken = errors.New("invalid token")
)

// SignToken creates a JWT with the given customer ID and email. Uses secret for HMAC signing.
func SignToken(secret []byte, customerID int64, email string) (string, error) {
	claims := Claims{
		CustomerID: customerID,
		Email:      email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(DefaultExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// VerifyToken parses and validates the token. Returns claims or error.
func VerifyToken(secret []byte, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
