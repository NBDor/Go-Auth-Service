package local

import (
	"context"
	"time"

	"github.com/NBDor/Go-Auth-Service/internal/auth"
	"github.com/NBDor/Go-Auth-Service/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	JWTSecret string
	
	TokenExpiration time.Duration
	
	PasswordValidator func(password string) error // Optional function to validate password requirements
}

func DefaultConfig() Config {
	return Config{
		JWTSecret:       "change-me-in-production", // Should be overridden in production
		TokenExpiration: 24 * time.Hour,
		PasswordValidator: func(password string) error {
			if len(password) < 8 {
				return auth.ErrInvalidCredentials
			}
			return nil
		},
	}
}

// implements username/password authentication with JWT tokens
type Provider struct {
	config    Config
	userStore UserStore
	jwtUtil   *jwt.Util
}

// creates a new local authentication provider
func NewProvider(config Config, userStore UserStore) *Provider {
	jwtUtil := jwt.NewUtil(config.JWTSecret, config.TokenExpiration)
	return &Provider{
		config:    config,
		userStore: userStore,
		jwtUtil:   jwtUtil,
	}
}

// returns the provider identifier
func (p *Provider) Name() string {
	return "local"
}

func (p *Provider) Authenticate(ctx context.Context, creds auth.Credentials) (*auth.User, error) {
	if creds.Type != "password" {
		return nil, auth.ErrInvalidCredentials
	}
	
	user, err := p.userStore.GetByUsername(ctx, creds.Username)
	if err != nil {
		return nil, auth.ErrInvalidCredentials
	}
	
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password))
	if err != nil {
		return nil, auth.ErrInvalidCredentials
	}
	
	authUser := &auth.User{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Roles:    user.Roles,
	}
	
	return authUser, nil
}

func (p *Provider) ValidateToken(ctx context.Context, token string) (*auth.User, error) {
	// Parse and validate JWT token
	claims, err := p.jwtUtil.ValidateToken(token)
	if err != nil {
		return nil, err
	}
	
	// Get user ID from claims
	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, auth.ErrInvalidCredentials
	}
	
	user, err := p.userStore.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	authUser := &auth.User{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Roles:    user.Roles,
	}
	
	return authUser, nil
}

func (p *Provider) RefreshToken(ctx context.Context, token string) (string, error) {
	user, err := p.ValidateToken(ctx, token)
	if err != nil {
		return "", err
	}
	
	claims := map[string]interface{}{
		"sub":    user.ID,
		"roles":  user.Roles,
		"email":  user.Email,
		"name":   user.Username,
		"provider": "local",
	}
	
	return p.jwtUtil.GenerateToken(claims)
}

// invalidates a token
// Note: Simple implementation. Production would use a token blacklist or shorter expiration times
func (p *Provider) RevokeToken(ctx context.Context, token string) error {
	// This simple implementation doesn't maintain a blacklist
	// In a production environment, you'd add the token to a blacklist stored in Redis or a database
	return nil
}