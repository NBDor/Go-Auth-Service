package local

import (
	"context"
	"time"

	"github.com/NBDor/Go-Auth-Service/internal/auth"
	"github.com/NBDor/Go-Auth-Service/pkg/jwt"
)

// TokenRevocationStore defines an interface for token revocation
type TokenRevocationStore interface {
	IsRevoked(ctx context.Context, tokenID string) (bool, error)
	RevokeToken(ctx context.Context, tokenID string, expiresAt time.Time) error
	CleanupExpiredTokens(ctx context.Context) (int64, error)
}

// ProviderWithRevocation extends the local provider with token revocation
type ProviderWithRevocation struct {
	*Provider
	tokenStore TokenRevocationStore
}

// NewProviderWithRevocation creates a new local provider with token revocation
func NewProviderWithRevocation(config Config, userStore UserStore, tokenStore TokenRevocationStore) *ProviderWithRevocation {
	provider := NewProvider(config, userStore)
	return &ProviderWithRevocation{
		Provider:   provider,
		tokenStore: tokenStore,
	}
}

// Name returns the provider identifier
func (p *ProviderWithRevocation) Name() string {
	return p.Provider.Name()
}

// Authenticate implements the Provider.Authenticate method
func (p *ProviderWithRevocation) Authenticate(ctx context.Context, creds auth.Credentials) (*auth.User, error) {
	return p.Provider.Authenticate(ctx, creds)
}

// ValidateToken overrides the base Provider.ValidateToken to check for revoked tokens
func (p *ProviderWithRevocation) ValidateToken(ctx context.Context, token string) (*auth.User, error) {
	// Parse and validate JWT token
	claims, err := p.jwtUtil.ValidateToken(token)
	if err != nil {
		return nil, err
	}
	
	// Get token ID
	tokenID, ok := claims["jti"].(string)
	if !ok {
		tokenID = token // Use the token as ID if jti not available
	}
	
	// Check if token is revoked
	isRevoked, err := p.tokenStore.IsRevoked(ctx, tokenID)
	if err != nil {
		return nil, err
	}
	if isRevoked {
		return nil, jwt.ErrInvalidToken
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

// RefreshToken generates a new token while invalidating the old one
func (p *ProviderWithRevocation) RefreshToken(ctx context.Context, token string) (string, error) {
	if token == "" {
		// Generate a new token for the user in the context
		return p.Provider.RefreshToken(ctx, token)
	}
	
	// Validate the old token
	claims, err := p.jwtUtil.ValidateToken(token)
	if err != nil && err != jwt.ErrExpiredToken {
		return "", err
	}
	
	// Revoke the old token
	tokenID, _ := claims["jti"].(string)
	if tokenID != "" {
		// Get expiry time
		expClaim, ok := claims["exp"].(float64)
		expiresAt := time.Now().Add(p.config.TokenExpiration)
		if ok {
			expiresAt = time.Unix(int64(expClaim), 0)
		}
		
		err = p.tokenStore.RevokeToken(ctx, tokenID, expiresAt)
		if err != nil {
			return "", err
		}
	}
	
	// Generate a new token
	return p.Provider.RefreshToken(ctx, token)
}

// RevokeToken overrides the base implementation to store revoked tokens
func (p *ProviderWithRevocation) RevokeToken(ctx context.Context, token string) error {
	// Parse the token to get its expiry
	claims, err := p.jwtUtil.ValidateToken(token)
	if err != nil && err != jwt.ErrExpiredToken {
		return err
	}
	
	// Get token ID
	tokenID, ok := claims["jti"].(string)
	if !ok {
		tokenID = token // Use the token as ID if jti not available
	}
	
	// Get expiry time
	expClaim, ok := claims["exp"].(float64)
	if !ok {
		// Default to the configured token expiration if exp claim not found
		return p.tokenStore.RevokeToken(ctx, tokenID, time.Now().Add(p.config.TokenExpiration))
	}
	
	// Store the token in the revocation list
	expiresAt := time.Unix(int64(expClaim), 0)
	return p.tokenStore.RevokeToken(ctx, tokenID, expiresAt)
}
