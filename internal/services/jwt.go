package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWTService holds the Secret
type JWTService struct {
	Secret []byte
}

// New creates a service with your Secret
func NewJwtService(Secret string) *JWTService {
	return &JWTService{Secret: []byte(Secret)}
}

// GenerateJWT - create a token that expires on the 1st day of the next month
func (s *JWTService) GenerateJWT() (string, error) {

	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   "tradingview",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.Secret)
}

// ValidateJWT ensures the token is valid and not expired
func (s *JWTService) ValidateJWT(tokenString string) (*jwt.RegisteredClaims, error) {
	claims := &jwt.RegisteredClaims{}
	parsed, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return s.Secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token expired")
	}
	return claims, nil
}
