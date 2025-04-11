package local

import (
	"context"
	"sync"
	"time"
)

// MemoryTokenStore implements token revocation with in-memory storage
type MemoryTokenStore struct {
	revokedTokens map[string]time.Time // Maps tokenID to expiry time
	mu            sync.RWMutex
}

// NewMemoryTokenStore creates a new in-memory token revocation store
func NewMemoryTokenStore() *MemoryTokenStore {
	return &MemoryTokenStore{
		revokedTokens: make(map[string]time.Time),
	}
}

// IsRevoked checks if a token has been revoked
func (s *MemoryTokenStore) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	expiry, exists := s.revokedTokens[tokenID]
	if !exists {
		return false, nil
	}
	
	// If token has expired, clean it up
	if time.Now().After(expiry) {
		s.mu.RUnlock()
		s.mu.Lock()
		delete(s.revokedTokens, tokenID)
		s.mu.Unlock()
		s.mu.RLock()
		return false, nil
	}
	
	return true, nil
}

// RevokeToken adds a token to the revoked list
func (s *MemoryTokenStore) RevokeToken(ctx context.Context, tokenID string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.revokedTokens[tokenID] = expiresAt
	return nil
}

// CleanupExpiredTokens removes expired tokens from the revoked list
func (s *MemoryTokenStore) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	var count int64
	
	for tokenID, expiry := range s.revokedTokens {
		if now.After(expiry) {
			delete(s.revokedTokens, tokenID)
			count++
		}
	}
	
	return count, nil
}
