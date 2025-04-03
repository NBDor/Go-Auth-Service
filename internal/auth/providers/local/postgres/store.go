package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/NBDor/Go-Auth-Service/internal/auth"
	"github.com/NBDor/Go-Auth-Service/internal/auth/providers/local"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// SQLUserStore implements the UserStore interface with PostgreSQL
type SQLUserStore struct {
	db *sqlx.DB
}

// userRow represents a row in the users table
type userRow struct {
	ID           string    `db:"id"`
	Username     string    `db:"username"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// NewSQLUserStore creates a new PostgreSQL-backed user store
func NewSQLUserStore(db *sqlx.DB) *SQLUserStore {
	return &SQLUserStore{
		db: db,
	}
}

// GetByID retrieves a user by ID
func (s *SQLUserStore) GetByID(ctx context.Context, id string) (*local.StoredUser, error) {
	var row userRow
	err := s.db.GetContext(ctx, &row, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, err
	}

	return s.assembleUser(ctx, &row)
}

// GetByUsername retrieves a user by username
func (s *SQLUserStore) GetByUsername(ctx context.Context, username string) (*local.StoredUser, error) {
	var row userRow
	err := s.db.GetContext(ctx, &row, "SELECT * FROM users WHERE username = $1", username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, err
	}

	return s.assembleUser(ctx, &row)
}

// GetByEmail retrieves a user by email
func (s *SQLUserStore) GetByEmail(ctx context.Context, email string) (*local.StoredUser, error) {
	var row userRow
	err := s.db.GetContext(ctx, &row, "SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, err
	}

	return s.assembleUser(ctx, &row)
}

// Create creates a new user
func (s *SQLUserStore) Create(ctx context.Context, user *local.StoredUser) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Generate ID if not provided
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	// Insert user
	_, err = tx.ExecContext(ctx, `
		INSERT INTO users (id, username, email, password_hash)
		VALUES ($1, $2, $3, $4)`,
		user.ID, user.Username, user.Email, user.PasswordHash)
	if err != nil {
		return err
	}

	// Insert roles
	if len(user.Roles) > 0 {
		for _, role := range user.Roles {
			_, err = tx.ExecContext(ctx, `
				INSERT INTO user_roles (user_id, role)
				VALUES ($1, $2)`,
				user.ID, role)
			if err != nil {
				return err
			}
		}
	}

	// Insert metadata
	if user.Metadata != nil {
		for key, value := range user.Metadata {
			jsonValue, err := json.Marshal(value)
			if err != nil {
				return err
			}

			_, err = tx.ExecContext(ctx, `
				INSERT INTO user_metadata (user_id, key, value)
				VALUES ($1, $2, $3)`,
				user.ID, key, jsonValue)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// Update updates an existing user
func (s *SQLUserStore) Update(ctx context.Context, user *local.StoredUser) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update user
	result, err := tx.ExecContext(ctx, `
		UPDATE users 
		SET username = $1, email = $2, password_hash = $3, updated_at = now()
		WHERE id = $4`,
		user.Username, user.Email, user.PasswordHash, user.ID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return auth.ErrUserNotFound
	}

	// Update roles (delete then insert)
	_, err = tx.ExecContext(ctx, "DELETE FROM user_roles WHERE user_id = $1", user.ID)
	if err != nil {
		return err
	}

	for _, role := range user.Roles {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO user_roles (user_id, role)
			VALUES ($1, $2)`,
			user.ID, role)
		if err != nil {
			return err
		}
	}

	// Update metadata (delete then insert)
	_, err = tx.ExecContext(ctx, "DELETE FROM user_metadata WHERE user_id = $1", user.ID)
	if err != nil {
		return err
	}

	for key, value := range user.Metadata {
		jsonValue, err := json.Marshal(value)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO user_metadata (user_id, key, value)
			VALUES ($1, $2, $3)`,
			user.ID, key, jsonValue)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Delete removes a user
func (s *SQLUserStore) Delete(ctx context.Context, id string) error {
	// The database is set up with ON DELETE CASCADE, so deleting from the users
	// table will automatically delete related roles and metadata
	result, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return auth.ErrUserNotFound
	}

	return nil
}

// assembleUser creates a complete StoredUser from a database row
func (s *SQLUserStore) assembleUser(ctx context.Context, row *userRow) (*local.StoredUser, error) {
	user := &local.StoredUser{
		ID:           row.ID,
		Username:     row.Username,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt.Unix(),
		UpdatedAt:    row.UpdatedAt.Unix(),
		Roles:        make([]string, 0),
		Metadata:     make(map[string]interface{}),
	}

	// Get roles
	var roles []string
	err := s.db.SelectContext(ctx, &roles, 
		"SELECT role FROM user_roles WHERE user_id = $1", user.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	user.Roles = roles

	// Get metadata
	type metadataRow struct {
		Key   string          `db:"key"`
		Value json.RawMessage `db:"value"`
	}
	var metadata []metadataRow
	err = s.db.SelectContext(ctx, &metadata, 
		"SELECT key, value FROM user_metadata WHERE user_id = $1", user.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	for _, meta := range metadata {
		var value interface{}
		if err := json.Unmarshal(meta.Value, &value); err != nil {
			return nil, err
		}
		user.Metadata[meta.Key] = value
	}

	return user, nil
}
