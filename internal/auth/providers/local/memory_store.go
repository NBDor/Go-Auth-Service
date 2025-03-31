package local

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/NBDor/Go-Auth-Service/internal/auth"
	"github.com/google/uuid"
)

// implements UserStore with an in-memory map
// This is suitable for development and testing, but not for production
type MemoryUserStore struct {
	users     map[string]*StoredUser // Indexed by ID
	usernames map[string]string      // Maps username to ID
	emails    map[string]string      // Maps email to ID
	mu        sync.RWMutex
}

func NewMemoryUserStore() *MemoryUserStore {
	return &MemoryUserStore{
		users:     make(map[string]*StoredUser),
		usernames: make(map[string]string),
		emails:    make(map[string]string),
	}
}

func (s *MemoryUserStore) GetByID(ctx context.Context, id string) (*StoredUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	user, exists := s.users[id]
	if !exists {
		return nil, auth.ErrUserNotFound
	}
	
	return cloneUser(user), nil
}

func (s *MemoryUserStore) GetByUsername(ctx context.Context, username string) (*StoredUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	id, exists := s.usernames[username]
	if !exists {
		return nil, auth.ErrUserNotFound
	}
	
	return cloneUser(s.users[id]), nil
}

func (s *MemoryUserStore) GetByEmail(ctx context.Context, email string) (*StoredUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	id, exists := s.emails[email]
	if !exists {
		return nil, auth.ErrUserNotFound
	}
	
	return cloneUser(s.users[id]), nil
}

func (s *MemoryUserStore) Create(ctx context.Context, user *StoredUser) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Check if username or email already exists
	if _, exists := s.usernames[user.Username]; exists {
		return errors.New("username already exists")
	}
	
	if _, exists := s.emails[user.Email]; exists {
		return errors.New("email already exists")
	}
	
	// Generate ID if not provided
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	
	now := time.Now().Unix()
	user.CreatedAt = now
	user.UpdatedAt = now
	
	// Store the user
	userCopy := cloneUser(user)
	s.users[user.ID] = userCopy
	s.usernames[user.Username] = user.ID
	s.emails[user.Email] = user.ID
	
	// Update the original with the generated ID
	user.ID = userCopy.ID
	
	return nil
}

// Update modifies an existing user
func (s *MemoryUserStore) Update(ctx context.Context, user *StoredUser) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Check if user exists
	existingUser, exists := s.users[user.ID]
	if !exists {
		return auth.ErrUserNotFound
	}
	
	// Check for username change
	if user.Username != existingUser.Username {
		if _, exists := s.usernames[user.Username]; exists {
			return errors.New("username already exists")
		}
		
		// Update username mapping
		delete(s.usernames, existingUser.Username)
		s.usernames[user.Username] = user.ID
	}
	
	// Check for email change
	if user.Email != existingUser.Email {
		if _, exists := s.emails[user.Email]; exists {
			return errors.New("email already exists")
		}
		
		// Update email mapping
		delete(s.emails, existingUser.Email)
		s.emails[user.Email] = user.ID
	}
	
	// Update timestamp
	user.UpdatedAt = time.Now().Unix()
	user.CreatedAt = existingUser.CreatedAt
	
	// Store updated user
	s.users[user.ID] = cloneUser(user)
	
	return nil
}

func (s *MemoryUserStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	user, exists := s.users[id]
	if !exists {
		return auth.ErrUserNotFound
	}
	
	// Remove from all maps
	delete(s.usernames, user.Username)
	delete(s.emails, user.Email)
	delete(s.users, id)
	
	return nil
}

func cloneUser(user *StoredUser) *StoredUser {
	if user == nil {
		return nil
	}
	
	roles := make([]string, len(user.Roles))
	copy(roles, user.Roles)
	
	metadata := make(map[string]interface{})
	for k, v := range user.Metadata {
		metadata[k] = v
	}
	
	return &StoredUser{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Roles:        roles,
		Metadata:     metadata,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}