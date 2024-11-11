package utils

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type JWTService interface {
	GenerateAccessToken(expirationTime time.Duration, userId string) (string, error)
	GenerateRefreshToken(expirationTime time.Duration, userId string, jti string) (string, error)
	ValidateToken(token string) (*jwt.Token, error)
}

type JwtCustomClaim struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type jwtService struct {
	secret string
}

func NewJWTService(secret string) JWTService {
	return &jwtService{
		secret: secret,
	}
}

func (s *jwtService) GenerateAccessToken(expirationTime time.Duration, userId string) (string, error) {
	expiration := time.Now().Add(expirationTime)
	claims := &JwtCustomClaim{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiration),
		},
		UserID: userId,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *jwtService) GenerateRefreshToken(expirationTime time.Duration, userId string, jti string) (string, error) {
	expiration := time.Now().Add(expirationTime)
	claims := &JwtCustomClaim{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiration),
			ID:        jti,
		},
		UserID: userId,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *jwtService) ValidateToken(tokenStr string) (*jwt.Token, error) {
	claims := &JwtCustomClaim{}
	return jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		return []byte(s.secret), nil
	})
}
