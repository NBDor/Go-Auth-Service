package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

// handles revoked tokens storage and checking
type TokenStore struct {
	db *sqlx.DB
}

// represents a revoked JWT token
type RevokedToken struct {
	TokenID   string    `db:"token_id"`
	RevokedAt time.Time `db:"revoked_at"`
	ExpiresAt time.Time `db:"expires_at"`
}

// creates a new PostgreSQL-backed token store
func NewTokenStore(db *sqlx.DB) *TokenStore {
	return &TokenStore{
		db: db,
	}
}

// checks if a token has been revoked
func (s *TokenStore) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
	var token RevokedToken
	err := s.db.GetContext(ctx, &token, "SELECT * FROM revoked_tokens WHERE token_id = $1", tokenID)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Token not found in revoked list, so it's not revoked
			return false, nil
		}
		return false, err
	}
	
	// Token found in revoked list
	return true, nil
}

// adds a token to the revoked list
func (s *TokenStore) RevokeToken(ctx context.Context, tokenID string, expiresAt time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO revoked_tokens (token_id, expires_at)
		VALUES ($1, $2)
		ON CONFLICT (token_id) DO NOTHING`,
		tokenID, expiresAt)
	
	return err
}

// removes expired tokens from the revoked list
func (s *TokenStore) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM revoked_tokens
		WHERE expires_at < now()`)
	
	if err != nil {
		return 0, err
	}
	
	return result.RowsAffected()
}
