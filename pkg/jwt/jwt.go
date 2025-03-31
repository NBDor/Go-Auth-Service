package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// JWT token generation and validation
type Util struct {
	secret    []byte
	expiresIn time.Duration
}

func NewUtil(secret string, expiresIn time.Duration) *Util {
	return &Util{
		secret:    []byte(secret),
		expiresIn: expiresIn,
	}
}

// creates a new JWT token with the provided claims
func (u *Util) GenerateToken(claims map[string]interface{}) (string, error) {
	now := time.Now()
	
	// Create the token with standard claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat": now.Unix(),                      // Issued at
		"exp": now.Add(u.expiresIn).Unix(),     // Expiration time
		"nbf": now.Unix(),                      // Not valid before
	})
	
	// Add custom claims
	for key, value := range claims {
		token.Claims.(jwt.MapClaims)[key] = value
	}
	
	// Sign and return the token
	return token.SignedString(u.secret)
}

// checks if a token is valid and returns its claims
func (u *Util) ValidateToken(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return u.secret, nil
	})
	
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	
	if !token.Valid {
		return nil, ErrInvalidToken
	}
	
	// Extract and return the claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}
	
	return nil, ErrInvalidToken
}

// validates an existing token and issues a new one
func (u *Util) RefreshToken(tokenString string) (string, error) {
	claims, err := u.ValidateToken(tokenString)
	if err != nil && !errors.Is(err, ErrExpiredToken) {
		// Allow refresh of expired tokens, but not invalid ones
		return "", err
	}
	
	// Generate a new token with the same claims
	return u.GenerateToken(claims)
}