package token

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// JWTMaker is responsible for creating and verifying JWT tokens.
type JWTMaker struct {
	signingKey string
	tokenTTL   time.Duration
}

// tokenClaims holds the custom claims for the JWT token, embedding the StandardClaims from the jwt package.
type tokenClaims struct {
	jwt.StandardClaims
	UserId int `json:"user_id"`
}

// NewJWTMaker creates a new JWTMaker with the given signing key and token time-to-live (TTL).
func New(signingKey string, tokenTTL time.Duration) TokenMaker {
	return &JWTMaker{
		signingKey: signingKey,
		tokenTTL:   tokenTTL,
	}
}

// CreateToken generates a new JWT token with a user ID, signed using the provided signing key.
func (j *JWTMaker) CreateToken(userId int) (string, error) {
	claims := &tokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(j.tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		UserId: userId,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.signingKey))
}

// VerifyToken parses and validates a given JWT token string, returning the user ID if valid.
func (j *JWTMaker) VerifyToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(j.signingKey), nil
	})

	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token")
	}

	return claims.UserId, nil
}
