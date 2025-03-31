# Go-Auth-Service

A lightweight, secure authentication and authorization service built with Go. This microservice provides JWT-based authentication, role-based access control, and integrates with various identity providers. Designed for high performance, scalability, and easy integration with other polyglot microservices.

## Project Structure

```
Go-Auth-Service/
├── .github/
│   └── workflows/
│       └── go.yml         # GitHub Actions CI workflow
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
│   │       └── oauth2/    # OAuth2 authentication
│   └── middleware/        # HTTP middleware components
├── pkg/
│   └── jwt/               # JWT utilities
├── .gitignore             # Git ignore file
├── Dockerfile             # Docker image configuration
├── docker-compose.yml     # Docker Compose configuration
├── go.mod                 # Go module definition
└── README.md              # This file
```

## Features

- JWT-based authentication
- Role-based access control (RBAC)
- Multiple authentication providers:
  - Local username/password
  - OAuth2 (Google, GitHub, etc.)
- API rate limiting
- Secure password hashing

## Development Setup

This project uses Docker for development to avoid requiring local Go installation.

### Prerequisites

- Git
- Docker
- Docker Compose

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

### Common Development Commands

Building the service:
```bash
docker run --rm -v $(pwd):/app -w /app golang:1.22 go build -o main ./cmd/server
```

Running tests:
```bash
docker run --rm -v $(pwd):/app -w /app golang:1.22 go test ./...
```

Adding dependencies:
```bash
docker run --rm -v $(pwd):/app -w /app golang:1.22 go get github.com/some/dependency
```

## CI/CD Pipeline

This project uses GitHub Actions for continuous integration. The workflow:

1. Builds the Go application
2. Runs tests
3. Builds a Docker image
4. Verifies the Docker container works correctly
