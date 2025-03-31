package local

import (
	"context"
)

type StoredUser struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
	Roles        []string
	Metadata     map[string]interface{}
	CreatedAt    int64
	UpdatedAt    int64
}

type UserStore interface {
	GetByID(ctx context.Context, id string) (*StoredUser, error)
	
	GetByUsername(ctx context.Context, username string) (*StoredUser, error)
	
	GetByEmail(ctx context.Context, email string) (*StoredUser, error)
	
	Create(ctx context.Context, user *StoredUser) error
	
	Update(ctx context.Context, user *StoredUser) error
	
	Delete(ctx context.Context, id string) error
}