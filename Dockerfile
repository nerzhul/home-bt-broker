# Build stage
FROM --platform=$BUILDPLATFORM rust:alpine AS builder

# Install build dependencies (needed for SQLite and linking)
RUN apk update && apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Cache dependencies by building a dummy project first
COPY Cargo.toml Cargo.lock ./
RUN mkdir -p src && echo "fn main() {}" > src/main.rs \
    && cargo build --release \
    && rm -rf src

# Copy source code, static assets and migrations
COPY src ./src
COPY static ./static
COPY migrations ./migrations

# Touch main.rs to force rebuild after dummy build
RUN touch src/main.rs && cargo build --release

# Final stage - use distroless for minimal attack surface
FROM gcr.io/distroless/static-debian12:latest

# Copy binary, static assets and migrations from builder
COPY --from=builder /app/target/release/home-bt-broker /app
COPY --from=builder /app/static /static
COPY --from=builder /app/migrations /migrations

# Expose port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/app"]
