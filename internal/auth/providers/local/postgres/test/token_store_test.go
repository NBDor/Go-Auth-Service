//go:build database

package test

import (
	"context"
	"testing"
	"time"

	"github.com/NBDor/Go-Auth-Service/internal/auth/providers/local/postgres"
	"github.com/NBDor/Go-Auth-Service/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresTokenStore(t *testing.T) {
	// Set up the database
	dbConfig := database.DefaultConfig()
	
	// Override with environment variables if available
	if envConfig := database.NewConfigFromEnv(); envConfig.Host != "localhost" {
		dbConfig = envConfig
	}
	
	db, err := database.Connect(dbConfig)
	require.NoError(t, err)
	defer db.Close()
	
	err = database.Initialize(db)
	require.NoError(t, err)
	
	// Create a token store
	tokenStore := postgres.NewTokenStore(db)
	
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
	assert.GreaterOrEqual(t, count, int64(1)) // At least our expired token should be removed
	
	// Verify the expired token is gone
	isRevoked, err = tokenStore.IsRevoked(ctx, expiredTokenID)
	assert.NoError(t, err)
	assert.False(t, isRevoked)
}
