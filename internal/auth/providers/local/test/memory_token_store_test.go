package test

import (
	"context"
	"testing"
	"time"

	"github.com/NBDor/Go-Auth-Service/internal/auth/providers/local"
	"github.com/stretchr/testify/assert"
)

func TestMemoryTokenStore(t *testing.T) {
	// Create a token store
	tokenStore := local.NewMemoryTokenStore()
	
	// Test token revocation
	ctx := context.Background()
	tokenID := "test-token-" + time.Now().Format(time.RFC3339)
	expiresAt := time.Now().Add(24 * time.Hour)
	
	// 1. Check token is not revoked initially
	isRevoked, err := tokenStore.IsRevoked(ctx, tokenID)
	assert.NoError(t, err)
	assert.False(t, isRevoked)
	
	// 2. Revoke the token
	err = tokenStore.RevokeToken(ctx, tokenID, expiresAt)
	assert.NoError(t, err)
	
	// 3. Check token is now revoked
	isRevoked, err = tokenStore.IsRevoked(ctx, tokenID)
	assert.NoError(t, err)
	assert.True(t, isRevoked)
	
	// 4. Test cleanup of expired tokens
	expiredTokenID := "expired-token-" + time.Now().Format(time.RFC3339)
	expiredAt := time.Now().Add(-1 * time.Hour) // Already expired
	
	err = tokenStore.RevokeToken(ctx, expiredTokenID, expiredAt)
	assert.NoError(t, err)
	
	count, err := tokenStore.CleanupExpiredTokens(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count) // Should be exactly 1 (our expired token)
	
	// Verify the expired token is gone
	isRevoked, err = tokenStore.IsRevoked(ctx, expiredTokenID)
	assert.NoError(t, err)
	assert.False(t, isRevoked)
}
