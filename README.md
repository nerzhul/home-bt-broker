# home-bt-broker

A Bluetooth broker web application built with Go, Echo framework, and SQLite. This application provides RESTful APIs for managing Bluetooth devices via BlueZ and storing authentication tokens.

## Features

- **Multi-architecture support**: Built for both ARM64 and AMD64 architectures using Docker Bake
- **SQLite database**: Lightweight database with migrations managed by golang-migrate
- **BlueZ integration**: Full Bluetooth device management via D-Bus
- **Health checks**: Includes `/readyz` and `/livez` endpoints for Kubernetes/container orchestration
- **RESTful API**: CRUD operations for managing username/token pairs and Bluetooth devices
- **Docker support**: Multi-stage builds with distroless final image for security
- **Comprehensive testing**: Unit tests with 76.9% coverage and mocked dependencies
- **CI/CD automation**: GitHub Actions for testing and multi-architecture Docker builds

## API Endpoints

### Health Checks
- `GET /readyz` - Readiness check (includes database connectivity test)
- `GET /livez` - Liveness check

### Token Management
- `POST /api/v1/tokens` - Create a new username/token pair
- `GET /api/v1/tokens` - Get all tokens
- `GET /api/v1/tokens/{username}` - Get token for specific username
- `DELETE /api/v1/tokens/{username}` - Delete token for specific username

### Bluetooth Management
- `GET /api/v1/bluetooth/adapters` - List all Bluetooth adapters
- `GET /api/v1/bluetooth/adapters/{adapter_mac}/devices` - List all devices for an adapter by MAC address
- `GET /api/v1/bluetooth/adapters/{adapter_mac}/devices/trusted` - List trusted devices for an adapter by MAC address
- `GET /api/v1/bluetooth/adapters/{adapter_mac}/devices/connected` - List connected devices for an adapter by MAC address
- `POST /api/v1/bluetooth/adapters/{adapter_mac}/devices/{device_mac}/pair` - Pair with a device by MAC address (auto-accepts PIN)
- `POST /api/v1/bluetooth/adapters/{adapter_mac}/devices/{device_mac}/connect` - Connect to a device by MAC address
- `POST /api/v1/bluetooth/adapters/{adapter_mac}/devices/{device_mac}/trust` - Trust a device by MAC address
- `DELETE /api/v1/bluetooth/adapters/{adapter_mac}/devices/{device_mac}` - Remove a device by MAC address

## Quick Start

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

## Requirements

- BlueZ installed and running (for Bluetooth functionality)
- D-Bus system bus access
- Appropriate permissions for Bluetooth operations

## Example Usage

### Token Management
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

### Bluetooth Management
```bash
# List all Bluetooth adapters
curl http://localhost:8080/api/v1/bluetooth/adapters

# List devices for adapter with MAC address AA:BB:CC:DD:EE:00
curl http://localhost:8080/api/v1/bluetooth/adapters/AA:BB:CC:DD:EE:00/devices

# List trusted devices
curl http://localhost:8080/api/v1/bluetooth/adapters/AA:BB:CC:DD:EE:00/devices/trusted

# List connected devices
curl http://localhost:8080/api/v1/bluetooth/adapters/AA:BB:CC:DD:EE:00/devices/connected

# Pair with device (auto-accepts PIN)
curl -X POST http://localhost:8080/api/v1/bluetooth/adapters/AA:BB:CC:DD:EE:00/devices/11:22:33:44:55:66/pair

# Connect to device
curl -X POST http://localhost:8080/api/v1/bluetooth/adapters/AA:BB:CC:DD:EE:00/devices/11:22:33:44:55:66/connect

# Trust device
curl -X POST http://localhost:8080/api/v1/bluetooth/adapters/AA:BB:CC:DD:EE:00/devices/11:22:33:44:55:66/trust

# Remove device
curl -X DELETE http://localhost:8080/api/v1/bluetooth/adapters/AA:BB:CC:DD:EE:00/devices/11:22:33:44:55:66
```

## CI/CD

The project includes automated GitHub Actions workflows:

### Workflows

- **Unit Tests** (`.github/workflows/test.yml`): Runs on every push and pull request
  - Executes all unit tests with coverage reporting
  - Uploads coverage artifacts
  - Supports Go 1.22

- **Docker Build** (`.github/workflows/docker.yml`): Builds multi-architecture Docker images
  - Builds for both AMD64 and ARM64 architectures
  - Pushes to GitHub Container Registry (ghcr.io)
  - Runs on pushes to main/develop and on releases

- **Release Pipeline** (`.github/workflows/release.yml`): Complete CI/CD for releases
  - Runs tests with coverage threshold check (70% minimum)
  - Builds and pushes Docker images on successful tests
  - Triggers on GitHub releases and manual dispatch

### Container Registry

Images are published to GitHub Container Registry:
```
ghcr.io/nerzhul/home-bt-broker:latest
ghcr.io/nerzhul/home-bt-broker:v1.0.0  # version tags
```

### Usage in CI

```bash
# Run tests locally (same as CI)
make test

# Build Docker images locally
make docker-build        # Single architecture
make docker-build-multi  # Multi-architecture
```
