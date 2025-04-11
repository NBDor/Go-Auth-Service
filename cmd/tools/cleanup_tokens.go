package main

import (
	"context"
	"log"
	"time"

	"github.com/NBDor/Go-Auth-Service/internal/auth/providers/local/postgres"
	"github.com/NBDor/Go-Auth-Service/internal/database"
)

func main() {
	log.Println("Starting token cleanup process...")

	// Initialize database connection
	dbConfig := database.NewConfigFromEnv()
	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create token store
	tokenStore := postgres.NewTokenStore(db)

	// Set context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Cleanup expired tokens
	count, err := tokenStore.CleanupExpiredTokens(ctx)
	if err != nil {
		log.Fatalf("Failed to cleanup tokens: %v", err)
	}

	log.Printf("Successfully removed %d expired tokens", count)
}
