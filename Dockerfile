# Build stage
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk update && apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments for cross-compilation
ARG TARGETOS
ARG TARGETARCH

# Build the application with static linking to avoid SQLite dependencies
RUN CGO_ENABLED=1 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags="-w -s -extldflags '-static'" -tags sqlite_omit_load_extension \
    -o app ./cmd/main.go

# Final stage - use distroless for minimal attack surface
FROM gcr.io/distroless/static-debian12:latest

# Copy binary and migrations from builder
COPY --from=builder /app/app /app
COPY --from=builder /app/migrations /migrations

# Expose port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/app"]