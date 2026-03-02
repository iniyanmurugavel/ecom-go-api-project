// JWT: sign and verify tokens. Claims include customer ID and email.
//
// LEARNING: JWT = JSON Web Token. It's a signed payload (claims) that the client sends in
// Authorization: Bearer <token>. We verify the signature with our secret — if valid, we trust
// the customer ID and email inside. No server-side session needed (stateless).
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims holds the JWT payload. jwt.RegisteredClaims adds exp, iat, iss, aud, etc.
type Claims struct {
	CustomerID int64  `json:"customer_id"`
	Email      string `json:"email"`
	jwt.RegisteredClaims
}

// JWTConfig holds JWT settings from env. Pass to SignToken and VerifyToken.
type JWTConfig struct {
	Secret   []byte
	Expiry   time.Duration
	Issuer   string
	Audience string
}

var (
	ErrInvalidToken = errors.New("invalid token")
)

// SignToken creates a JWT with the given customer ID and email. Uses config for secret, expiry, iss, aud.
func SignToken(cfg *JWTConfig, customerID int64, email string) (string, error) {
	now := time.Now()
	claims := Claims{
		CustomerID: customerID,
		Email:      email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.Expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    cfg.Issuer,
			Audience:  jwt.ClaimStrings{cfg.Audience},
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(cfg.Secret)
}

// VerifyToken parses and validates the token. Validates iss and aud when set in config.
func VerifyToken(cfg *JWTConfig, tokenString string) (*Claims, error) {
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return cfg.Secret, nil
	}
	opts := []jwt.ParserOption{}
	if cfg.Issuer != "" {
		opts = append(opts, jwt.WithIssuer(cfg.Issuer))
	}
	if cfg.Audience != "" {
		opts = append(opts, jwt.WithAudience(cfg.Audience))
	}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, keyFunc, opts...)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
