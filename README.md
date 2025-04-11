# Go-Auth-Service

A lightweight, secure authentication and authorization service built with Go. This microservice provides JWT-based authentication, role-based access control, and integrates with various identity providers. Designed for high performance, scalability, and easy integration with other polyglot microservices.

## Project Structure

```
Go-Auth-Service/
├── .github/
│   └── workflows/
│       ├── go.yml         # GitHub Actions CI workflow
│       └── test.yml       # Integration tests workflow
├── api/
│   └── v1/                # API endpoints version 1
├── cmd/
│   ├── server/            # Application entry points
│   │   └── main.go        # Main server code
│   └── tools/             # Utility tools and scripts
│       └── cleanup_tokens.go # Script to clean expired tokens
├── internal/
│   ├── auth/              # Authentication logic
│   │   ├── provider.go    # Authentication provider interface
│   │   └── providers/     # Individual auth provider implementations
│   │       ├── local/     # Username/password authentication
│   │       │   ├── postgres/  # PostgreSQL storage implementation
│   │       │   │   ├── store.go    # User storage
│   │       │   │   └── token_store.go  # Token revocation store
│   │       │   ├── memory_store.go    # In-memory user store
│   │       │   ├── memory_token_store.go  # In-memory token store
│   │       │   ├── provider.go       # Basic provider implementation
│   │       │   ├── provider_with_revocation.go # Enhanced provider with token revocation
│   │       │   └── user_store.go     # User store interface
│   │       └── oauth2/    # OAuth2 authentication (planned)
│   ├── database/          # Database connectivity and migrations
│   │   ├── database.go    # DB connection configuration
│   │   ├── migrations.go  # Migration system
│   │   └── migrations/    # SQL migration files
│   │       ├── 001_initial_schema.up.sql   # Initial schema setup
│   │       └── 001_initial_schema.down.sql # Schema rollback
│   ├── integration/       # Integration tests
│   │   ├── db_auth_test.go     # Database integration tests
│   │   ├── memory_auth_test.go # In-memory integration tests
│   │   ├── memory_token_test.go # In-memory token tests
│   │   └── token_revocation_test.go # Token revocation tests
│   └── server/            # HTTP server and router logic
│       └── router.go      # HTTP routing configuration
├── pkg/
│   └── jwt/               # JWT utilities
│       └── jwt.go         # JWT token generation and validation
├── .gitignore             # Git ignore file
├── Dockerfile             # Docker image configuration
├── docker-compose.yml     # Docker Compose configuration with PostgreSQL
├── go.mod                 # Go module definition
└── README.md              # This file
```

## Features

### Implemented
- Authentication provider interface (extensible design)
- Local username/password authentication
- JWT token generation and validation
- In-memory user store for development/testing
- PostgreSQL database integration for persistent storage
- Database migrations for schema versioning
- Token revocation and blacklisting
- Protected API endpoints with token validation
- Logout endpoint for token invalidation
- Comprehensive test suite for both in-memory and database modes
- Graceful fallback to in-memory storage if database is unavailable
- Modular architecture with clean separation of concerns
- Docker containerization
- CI/CD pipeline with GitHub Actions

### Planned
- OAuth2 authentication providers
- Role-based access control (RBAC)
- API endpoints for user management
- Token refresh
- Rate limiting and security features
- Observability (logging, metrics)

## Development Setup

This project uses Docker for development to avoid requiring local Go installation.

### Prerequisites

- Git
- Docker
- Docker Compose
- Go 1.22+ (optional, for local development)

### Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/NBDor/Go-Auth-Service.git
   cd Go-Auth-Service
   ```

2. Run the service using Docker Compose:
   ```bash
   docker-compose up
   ```

3. The service will be available at http://localhost:8080

### Using the API

#### Authentication and Token Management
To authenticate and manage JWT tokens:

```bash
# Login to get a token
curl -X POST http://localhost:8080/auth/login \
  -d "username=admin&password=admin123"

# Access protected endpoint
curl -X GET http://localhost:8080/auth/me \
  -H "Authorization: Bearer your-token-here"

# Logout/revoke a token
curl -X POST http://localhost:8080/auth/logout \
  -H "Authorization: Bearer your-token-here"
```

The default admin credentials are:
- Username: `admin`
- Password: `admin123`

### Testing

#### Testing Architecture

The project uses a dual testing approach:
- **In-memory tests**: Fast tests that run without external dependencies
- **Database tests**: Full integration tests that verify database functionality

The tests are separated using Go build tags, allowing them to run independently.

#### Running Tests Locally

```bash
# Install dependencies
go mod tidy

# Run unit tests for storage implementations
go test -v ./internal/auth/providers/local/test/...

# Run in-memory integration tests (no database needed)
go test -v ./internal/integration/...

# Run database integration tests (requires PostgreSQL)
go test -v -tags=database ./internal/integration/...
```

These tests verify both the authentication and token revocation functionality in both storage modes (in-memory and PostgreSQL).

### Common Development Commands

Building the service:
```bash
# Local build
go build -o main ./cmd/server

# Docker build
docker run --rm -v $(pwd):/app -w /app golang:1.22 go build -o main ./cmd/server
```

Adding dependencies:
```bash
go get github.com/some/dependency
go mod tidy
```

## Configuration

### Database Configuration

The service uses PostgreSQL for persistent storage with automatic fallback to in-memory storage if the database is unavailable. Database configuration can be customized through environment variables:

- `DB_HOST`: Database hostname (default: localhost)
- `DB_PORT`: Database port (default: 5432)
- `DB_USER`: Database username (default: postgres)
- `DB_PASSWORD`: Database password (default: postgres)
- `DB_NAME`: Database name (default: auth_service)
- `DB_SSLMODE`: SSL mode (default: disable)
- `DB_MAX_CONNS`: Maximum number of open connections (default: 25)
- `DB_MAX_IDLE`: Maximum number of idle connections (default: 5)
- `DB_TIMEOUT`: Connection timeout in seconds (default: 5s)

### Token Management

The service provides comprehensive token management capabilities:

- **Token Generation**: Creates JWT tokens with secure random IDs
- **Token Validation**: Validates tokens for protected API endpoints
- **Token Revocation**: Allows users to invalidate tokens before expiration
- **Revocation Storage**: Persists revoked tokens in PostgreSQL or memory
- **Token Cleanup**: Includes a utility to purge expired tokens from storage

The token revocation system ensures that logged-out sessions cannot be reused, even if the token hasn't expired yet.

### JWT Configuration

JWT settings can be customized through environment variables:

- `JWT_SECRET`: Secret key for signing JWTs (default: change-me-in-production)
- `TOKEN_EXPIRY`: Token expiration time (default: 24h)

## CI/CD Pipeline

This project uses GitHub Actions for continuous integration with separate workflows for different testing scenarios:

### Build Workflow (`go.yml`)
1. Builds the Go application
2. Runs in-memory tests (no database required)
3. Builds a Docker image

### Database Test Workflow (`test.yml`)
1. Runs two parallel job streams:
   - In-memory tests: Fast verification of core functionality
   - Database tests: Full integration tests with PostgreSQL

This approach ensures fast feedback on basic functionality while still thoroughly testing database integration.

## Architecture

The service follows a clean architecture pattern with clearly separated concerns:

1. **Command Layer** (`cmd/server`): Entry point and server lifecycle management
2. **Router Layer** (`internal/server`): HTTP routing and request handling
3. **Auth Layer** (`internal/auth`): Authentication and provider interfaces
4. **Storage Layer** (`internal/auth/providers`):
   - User store interfaces and implementations (memory and PostgreSQL)
   - Token store for revocation management (memory and PostgreSQL)
5. **Database Layer** (`internal/database`): Database connection, migrations, and schema management
6. **Token Management** (`pkg/jwt`): JWT token generation, validation, and revocation

This separation allows for easy testing, maintenance, and future expansion. The service can operate in two modes:

1. **Database Mode**: Uses PostgreSQL for persistent storage of users and revoked tokens
2. **Fallback Mode**: Uses in-memory storage when the database is unavailable

## Next Steps

1. Complete OAuth2 provider implementation
2. Add API endpoints for user management
3. Implement RBAC middleware
4. Add security features (rate limiting)
5. Add token refresh functionality
6. Add observability (logging, metrics)
7. Create API documentation
8. Add multi-tenancy support
