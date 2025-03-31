package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NBDor/Go-Auth-Service/internal/auth"
	"github.com/NBDor/Go-Auth-Service/internal/auth/providers/local"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Create a registry for authentication providers
	providerRegistry := auth.NewProviderRegistry()

	// Set up a local authentication provider with in-memory storage
	userStore := local.NewMemoryUserStore()
	localProviderConfig := local.DefaultConfig()
	localProvider := local.NewProvider(localProviderConfig, userStore)
	providerRegistry.Register(localProvider)

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

	// Set up HTTP server
	mux := http.NewServeMux()

	// Add a basic health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","providers":["%s"]}`, localProvider.Name())
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
		token, err := localProvider.RefreshToken(r.Context(), "")
		if err != nil {
			http.Error(w, "Error generating token", http.StatusInternalServerError)
			return
		}

		// Return token in response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"token":"%s","user":{"id":"%s","username":"%s","email":"%s"}}`, 
			token, user.ID, user.Username, user.Email)
	})

	// Configure the server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}