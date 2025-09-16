# home-bt-broker

A simple web application to manage Bluetooth broker tokens, built with Go, Echo framework, and SQLite.

## Features

- **Multi-architecture support**: Built for both ARM64 and AMD64 architectures using Docker Bake
- **SQLite database**: Lightweight database with migrations managed by golang-migrate
- **Health checks**: Includes `/readyz` and `/livez` endpoints for Kubernetes/container orchestration
- **RESTful API**: CRUD operations for managing username/token pairs
- **Docker support**: Multi-stage builds with distroless final image for security

## API Endpoints

### Health Checks
- `GET /readyz` - Readiness check (includes database connectivity test)
- `GET /livez` - Liveness check

### Token Management
- `POST /api/v1/tokens` - Create a new username/token pair
- `GET /api/v1/tokens` - Get all tokens
- `GET /api/v1/tokens/{username}` - Get token for specific username
- `DELETE /api/v1/tokens/{username}` - Delete token for specific username

## Quick Start

### Using Docker Compose
```bash
docker-compose up --build
```

### Using Docker Bake (Multi-architecture)
```bash
# Build for local testing
docker buildx bake local

# Build for both AMD64 and ARM64
docker buildx bake
```

### Local Development
```bash
go mod download
go run ./cmd/main.go
```

## Configuration

Environment variables:
- `PORT`: Server port (default: 8080)
- `DATABASE_PATH`: SQLite database file path (default: ./data.db)

## Example Usage

```bash
# Create a token
curl -X POST -H "Content-Type: application/json" \
  -d '{"username":"user1","token":"secret123"}' \
  http://localhost:8080/api/v1/tokens

# Get all tokens
curl http://localhost:8080/api/v1/tokens

# Get specific token
curl http://localhost:8080/api/v1/tokens/user1

# Delete token
curl -X DELETE http://localhost:8080/api/v1/tokens/user1
```
