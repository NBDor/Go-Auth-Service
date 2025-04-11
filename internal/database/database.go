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
	
	if maxConns := getEnvAsInt("DB_MAX_CONNS", 0); maxConns != 0 {
		config.MaxConns = maxConns
	}
	
	if maxIdle := getEnvAsInt("DB_MAX_IDLE", 0); maxIdle != 0 {
		config.MaxIdle = maxIdle
	}
	
	if timeout := getEnvAsInt("DB_TIMEOUT", 0); timeout != 0 {
		config.Timeout = time.Duration(timeout) * time.Second
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

// establishes a connection to the database and runs migrations
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

// Initializes the database by running migrations
func Initialize(db *sqlx.DB) error {
	return Migrate(db)
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
