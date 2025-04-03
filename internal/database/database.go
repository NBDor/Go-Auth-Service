package database

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// Config holds database connection configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int
	MaxIdle  int
	Timeout  time.Duration
}

// DefaultConfig returns a default database configuration
func DefaultConfig() Config {
	return Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "auth_service",
		SSLMode:  "disable",
		MaxConns: 25,
		MaxIdle:  5,
		Timeout:  5 * time.Second,
	}
}

// NewConfigFromEnv creates a config from environment variables
func NewConfigFromEnv() Config {
	config := DefaultConfig()
	
	if host := getEnv("DB_HOST", ""); host != "" {
		config.Host = host
	}
	
	if port := getEnvAsInt("DB_PORT", 0); port != 0 {
		config.Port = port
	}
	
	if user := getEnv("DB_USER", ""); user != "" {
		config.User = user
	}
	
	if password := getEnv("DB_PASSWORD", ""); password != "" {
		config.Password = password
	}
	
	if dbName := getEnv("DB_NAME", ""); dbName != "" {
		config.DBName = dbName
	}
	
	if sslMode := getEnv("DB_SSLMODE", ""); sslMode != "" {
		config.SSLMode = sslMode
	}
	
	return config
}

// DSN returns a PostgreSQL connection string
func (c Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// establishes a connection to the database
func Connect(config Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", config.DSN())
	if err != nil {
		return nil, fmt.Errorf("database connection error: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxConns)
	db.SetMaxIdleConns(config.MaxIdle)
	db.SetConnMaxLifetime(config.Timeout)

	return db, nil
}

// creates necessary tables if they don't exist
func Initialize(db *sqlx.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(36) PRIMARY KEY,
		username VARCHAR(255) NOT NULL UNIQUE,
		email VARCHAR(255) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
	);

	CREATE TABLE IF NOT EXISTS user_roles (
		user_id VARCHAR(36) REFERENCES users(id) ON DELETE CASCADE,
		role VARCHAR(50) NOT NULL,
		PRIMARY KEY (user_id, role)
	);

	CREATE TABLE IF NOT EXISTS user_metadata (
		user_id VARCHAR(36) REFERENCES users(id) ON DELETE CASCADE,
		key VARCHAR(255) NOT NULL,
		value JSONB NOT NULL,
		PRIMARY KEY (user_id, key)
	);

	CREATE TABLE IF NOT EXISTS revoked_tokens (
		token_id VARCHAR(255) PRIMARY KEY,
		revoked_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL
	);
	`

	_, err := db.Exec(schema)
	return err
}

// Helper functions for environment variables

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

// getEnvAsInt retrieves an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
