.PHONY: build run test clean docker-build

# Build the application
build:
	go build -o bin/app ./cmd/main.go

# Build static binary
build-static:
	CGO_ENABLED=1 go build -ldflags="-w -s -extldflags '-static'" -tags sqlite_omit_load_extension -o bin/app-static ./cmd/main.go

# Run the application
run:
	go run ./cmd/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f app app-static data.db

# Build Docker image locally
docker-build:
	docker buildx bake local

# Build multi-architecture images
docker-build-multi:
	docker buildx bake

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Download dependencies
deps:
	go mod download
	go mod tidy

# Development setup
dev-setup: deps
	mkdir -p bin/