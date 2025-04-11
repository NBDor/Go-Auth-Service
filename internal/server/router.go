package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/NBDor/Go-Auth-Service/internal/auth"
	"github.com/NBDor/Go-Auth-Service/internal/auth/providers/local"
	"github.com/NBDor/Go-Auth-Service/internal/auth/providers/local/postgres"
	"github.com/NBDor/Go-Auth-Service/internal/database"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// SetupRouter creates and configures the HTTP router with all routes
func SetupRouter() (http.Handler, *auth.ProviderRegistry) {
	// Create a registry for authentication providers
	providerRegistry := auth.NewProviderRegistry()

	// Initialize database connection
	log.Println("Initializing database connection...")
	dbConfig := database.NewConfigFromEnv()
	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		log.Println("Falling back to in-memory storage")
		useInMemoryStorage(providerRegistry)
	} else {
		log.Println("Successfully connected to database")
		// Initialize database schema
		if err := database.Initialize(db); err != nil {
			log.Printf("Failed to initialize database schema: %v", err)
			log.Println("Falling back to in-memory storage")
			useInMemoryStorage(providerRegistry)
		} else {
			// Set up PostgreSQL user store
			usePostgresStorage(providerRegistry, db)
		}
	}

	// Set up HTTP server
	mux := http.NewServeMux()

	// Add a basic health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		// Get the list of registered providers
		providers := providerRegistry.ListProviders()
		providerNames := make([]string, len(providers))
		for i, p := range providers {
			providerNames[i] = p.Name()
		}
		
		fmt.Fprintf(w, `{"status":"ok","providers":["%s"]}`, strings.Join(providerNames, `","`))
	})

	// Add a basic authentication endpoint
	mux.HandleFunc("POST /auth/login", func(w http.ResponseWriter, r *http.Request) {
		// Parse username and password from request
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Authenticate the user
		creds := auth.Credentials{
			Type:     "password",
			Username: username,
			Password: password,
			Provider: "local",
		}

		provider, exists := providerRegistry.Get("local")
		if !exists {
			http.Error(w, "Authentication provider not available", http.StatusInternalServerError)
			return
		}

		user, err := provider.Authenticate(r.Context(), creds)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Generate a JWT token
		ctx := context.WithValue(r.Context(), "user", user)
		token, err := provider.RefreshToken(ctx, "")
		if err != nil {
			log.Printf("Token generation error: %v", err)
			http.Error(w, "Error generating token", http.StatusInternalServerError)
			return
		}

		// Return token in response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"token":"%s","user":{"id":"%s","username":"%s","email":"%s"}}`, 
			token, user.ID, user.Username, user.Email)
	})

	// Add a protected endpoint that requires authentication
	mux.HandleFunc("GET /auth/me", func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}
		
		// Remove "Bearer " prefix if present
		token := auth
		if len(auth) > 7 && strings.ToUpper(auth[0:7]) == "BEARER " {
			token = auth[7:]
		}
		
		// Get the provider
		provider, exists := providerRegistry.Get("local")
		if !exists {
			http.Error(w, "Authentication provider not available", http.StatusInternalServerError)
			return
		}
		
		// Validate token
		user, err := provider.ValidateToken(r.Context(), token)
		if err != nil {
			log.Printf("Token validation error: %v", err)
			http.Error(w, "Invalid or revoked token", http.StatusUnauthorized)
			return
		}
		
		// Return user info
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"user":{"id":"%s","username":"%s","email":"%s","roles":["%s"]}}`,
			user.ID, user.Username, user.Email, strings.Join(user.Roles, "\",\""))
	})

	// Add a token revocation endpoint
	mux.HandleFunc("POST /auth/logout", func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "Missing Authorization header", http.StatusBadRequest)
			return
		}
		
		// Remove "Bearer " prefix if present
		token := auth
		if len(auth) > 7 && strings.ToUpper(auth[0:7]) == "BEARER " {
			token = auth[7:]
		}
		
		// Get the provider
		provider, exists := providerRegistry.Get("local")
		if !exists {
			http.Error(w, "Authentication provider not available", http.StatusInternalServerError)
			return
		}
		
		// Revoke the token
		err := provider.RevokeToken(r.Context(), token)
		if err != nil {
			log.Printf("Token revocation error: %v", err)
			http.Error(w, "Error revoking token", http.StatusInternalServerError)
			return
		}
		
		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message":"Successfully logged out"}`)
	})

	return mux, providerRegistry
}

// useInMemoryStorage sets up the in-memory user store for authentication
func useInMemoryStorage(registry *auth.ProviderRegistry) {
	userStore := local.NewMemoryUserStore()
	tokenStore := local.NewMemoryTokenStore()
	localProviderConfig := getJWTConfig()
	localProvider := local.NewProviderWithRevocation(localProviderConfig, userStore, tokenStore)
	registry.Register(localProvider)

	// Add a sample user for testing
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	sampleUser := &local.StoredUser{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		Roles:        []string{"user"},
		Metadata:     map[string]interface{}{"created_by": "system"},
	}
	ctx := context.Background()
	_ = userStore.Create(ctx, sampleUser)
	
	log.Println("Initialized in-memory user store with test user: testuser")
}

// usePostgresStorage sets up the PostgreSQL user store for authentication
func usePostgresStorage(registry *auth.ProviderRegistry, db *sqlx.DB) {
	userStore := postgres.NewSQLUserStore(db)
	tokenStore := postgres.NewTokenStore(db)
	localProviderConfig := getJWTConfig()
	log.Printf("Using JWT config: secret=%s, expiry=%s", 
		localProviderConfig.JWTSecret[:3]+"...", localProviderConfig.TokenExpiration)
	
	localProvider := local.NewProviderWithRevocation(localProviderConfig, userStore, tokenStore)
	registry.Register(localProvider)

	// Check if we need to create an admin user
	ctx := context.Background()
	_, err := userStore.GetByUsername(ctx, "admin")
	if err != nil && err.Error() == auth.ErrUserNotFound.Error() {
		// Create default admin user
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		adminUser := &local.StoredUser{
			Username:     "admin",
			Email:        "admin@example.com",
			PasswordHash: string(hashedPassword),
			Roles:        []string{"admin", "user"},
			Metadata:     map[string]interface{}{"created_by": "system"},
		}
		err = userStore.Create(ctx, adminUser)
		if err != nil {
			log.Printf("Failed to create admin user: %v", err)
		} else {
			log.Println("Created default admin user: admin")
		}
	}
}

// Get the JWT configuration from environment variables
func getJWTConfig() local.Config {
	config := local.DefaultConfig()
	
	// Get JWT secret from env
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		config.JWTSecret = secret
	}
	
	// Get token expiry from env
	if expiryStr := os.Getenv("TOKEN_EXPIRY"); expiryStr != "" {
		if expiry, err := time.ParseDuration(expiryStr); err == nil {
			config.TokenExpiration = expiry
		}
	}
	
	return config
}
