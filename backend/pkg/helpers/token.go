package helpers

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenClaims holds the custom claims stored in JWT tokens.
type TokenClaims struct {
	// UserID is the authenticated user's unique identifier.
	UserID string `json:"userId"`
	// Email is the authenticated user's email address.
	Email string `json:"email"`
	// Role is the authenticated user's role.
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT token with the given user claims.
func GenerateToken(userID string, email string, role string, secret string, expiryHours int) (string, error) {
	claims := TokenClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken parses and validates a JWT token string.
// Returns the token claims if the token is valid.
func ValidateToken(tokenString string, secret string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
