package database

import (
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// SimpleMigrate applies the hard-coded schema
func Migrate(db *sqlx.DB) error {
	// Apply our schema directly
	schema := `
	-- Create schema version table if it doesn't exist
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
	);

	-- Create users table
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(36) PRIMARY KEY,
		username VARCHAR(255) NOT NULL UNIQUE,
		email VARCHAR(255) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
	);

	-- Create roles table
	CREATE TABLE IF NOT EXISTS user_roles (
		user_id VARCHAR(36) REFERENCES users(id) ON DELETE CASCADE,
		role VARCHAR(50) NOT NULL,
		PRIMARY KEY (user_id, role)
	);

	-- Create metadata table
	CREATE TABLE IF NOT EXISTS user_metadata (
		user_id VARCHAR(36) REFERENCES users(id) ON DELETE CASCADE,
		key VARCHAR(255) NOT NULL,
		value JSONB NOT NULL,
		PRIMARY KEY (user_id, key)
	);

	-- Create token revocation table
	CREATE TABLE IF NOT EXISTS revoked_tokens (
		token_id VARCHAR(255) PRIMARY KEY,
		revoked_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL
	);

	-- Make sure we have a version record
	INSERT INTO schema_migrations (version)
	VALUES (1)
	ON CONFLICT (version) DO NOTHING;
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to apply schema: %w", err)
	}

	return nil
}

// RollbackLatest rolls back the latest migration
func RollbackLatest(db *sqlx.DB) error {
	return errors.New("rollback not implemented in simple migration")
}
