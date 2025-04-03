package auth

import (
	"context"
	"errors"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrProviderNotEnabled = errors.New("authentication provider not enabled")
)

type User struct {
	ID       string
	Username string
	Email    string
	Roles    []string
	Metadata map[string]interface{} // Flexible field for provider-specific data
}

type Credentials struct {
	Type     string                 // "password", "token", "oauth", etc.
	Username string                 // username/password auth
	Password string                 // username/password auth
	Token    string                 // token-based auth
	Provider string                 // "google", "github", "local", etc.
	Params   map[string]interface{} // Additional provider-specific parameters
}

// interface that all authentication providers must implement
type Provider interface {
	Name() string // unique identifier for this provider
	
	Authenticate(ctx context.Context, creds Credentials) (*User, error)
	
	ValidateToken(ctx context.Context, token string) (*User, error)
	
	RefreshToken(ctx context.Context, token string) (string, error)
	
	RevokeToken(ctx context.Context, token string) error
}

// manages the available authentication providers
type ProviderRegistry struct {
	providers map[string]Provider
}

// creates a new provider registry
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]Provider),
	}
}

// adds a provider to the registry
func (r *ProviderRegistry) Register(provider Provider) {
	r.providers[provider.Name()] = provider
}

func (r *ProviderRegistry) Get(name string) (Provider, bool) {
	provider, exists := r.providers[name]
	return provider, exists
}

func (r *ProviderRegistry) GetAll() map[string]Provider {
	return r.providers
}

func (r *ProviderRegistry) ListProviders() []Provider {
	providers := make([]Provider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}
	return providers
}