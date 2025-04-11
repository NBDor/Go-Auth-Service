package test

import (
	"context"
	"testing"
	"time"

	"github.com/NBDor/Go-Auth-Service/internal/auth"
	"github.com/NBDor/Go-Auth-Service/internal/auth/providers/local"
	"github.com/NBDor/Go-Auth-Service/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// mockTokenStore is a simple implementation for testing
type mockTokenStore struct {
	revokedTokens map[string]time.Time
}

func newMockTokenStore() *mockTokenStore {
	return &mockTokenStore{
		revokedTokens: make(map[string]time.Time),
	}
}

func (m *mockTokenStore) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
	_, exists := m.revokedTokens[tokenID]
	return exists, nil
}

func (m *mockTokenStore) RevokeToken(ctx context.Context, tokenID string, expiresAt time.Time) error {
	m.revokedTokens[tokenID] = expiresAt
	return nil
}

func (m *mockTokenStore) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	return 0, nil
}

func TestProviderWithRevocation(t *testing.T) {
	// Create user store with a test user
	userStore := local.NewMemoryUserStore()
	
	// Create a test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := &local.StoredUser{
		ID:           "test-user-id",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		Roles:        []string{"user"},
	}
	
	ctx := context.Background()
	err := userStore.Create(ctx, testUser)
	assert.NoError(t, err)
	
	// Create token store
	tokenStore := newMockTokenStore()
	
	// Create provider
	config := local.Config{
		JWTSecret:       "test-secret",
		TokenExpiration: 1 * time.Hour,
	}
	provider := local.NewProviderWithRevocation(config, userStore, tokenStore)
	
	// Test authentication
	creds := auth.Credentials{
		Type:     "password",
		Username: "testuser",
		Password: "password123",
		Provider: "local",
	}
	
	user, err := provider.Authenticate(ctx, creds)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", user.Username)
	
	// Generate a token
	ctxWithUser := context.WithValue(ctx, "user", user)
	token, err := provider.RefreshToken(ctxWithUser, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	
	// Validate the token
	validatedUser, err := provider.ValidateToken(ctx, token)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", validatedUser.Username)
	
	// Revoke the token
	err = provider.RevokeToken(ctx, token)
	assert.NoError(t, err)
	
	// Validate the revoked token - should fail
	_, err = provider.ValidateToken(ctx, token)
	assert.Error(t, err)
	assert.Equal(t, jwt.ErrInvalidToken, err)
}
