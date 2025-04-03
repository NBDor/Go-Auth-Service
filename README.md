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
│   └── server/            # Application entry points
│       └── main.go        # Main server code
├── internal/
│   ├── auth/              # Authentication logic
│   │   ├── provider.go    # Authentication provider interface
│   │   └── providers/     # Individual auth provider implementations
│   │       ├── local/     # Username/password authentication
│   │       │   └── postgres/  # PostgreSQL storage implementation
│   │       └── oauth2/    # OAuth2 authentication (planned)
│   ├── database/          # Database connectivity and migrations
│   ├── integration/       # Integration tests
│   ├── middleware/        # HTTP middleware components
│   └── server/            # HTTP server and router logic
├── pkg/
│   └── jwt/               # JWT utilities
├── .gitignore             # Git ignore file
├── Dockerfile             # Docker image configuration
├── docker-compose.yml     # Docker Compose configuration
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
- Automated integration tests for both in-memory and database modes
- Graceful fallback to in-memory storage if database is unavailable
- Modular architecture with clean separation of concerns
- Docker containerization
- CI/CD pipeline with GitHub Actions

### Planned
- OAuth2 authentication providers
- Role-based access control (RBAC)
- API endpoints for user management
- Token blacklisting and refresh
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

#### Authentication
To authenticate and get a JWT token:

```bash
curl -X POST http://localhost:8080/auth/login \
  -d "username=admin&password=admin123"
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

# Run in-memory tests (default, no database needed)
go test -v ./internal/integration/memory_auth_test.go

# Run database tests (requires PostgreSQL)
go test -tags=database -v ./internal/integration/db_auth_test.go
```

These tests verify both the health endpoint and authentication flows in their respective storage modes.

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
4. **Storage Layer** (`internal/auth/providers`): User and token storage implementations
5. **Database Layer** (`internal/database`): Database connection and schema management

This separation allows for easy testing, maintenance, and future expansion.

## Next Steps

1. Complete OAuth2 provider implementation
2. Add API endpoints for user management
3. Implement RBAC middleware
4. Add security features (rate limiting, token blacklisting)
5. Improve token management
6. Add observability (logging, metrics)
7. Create API documentation
8. Add multi-tenancy support
