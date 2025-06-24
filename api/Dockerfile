# API Backend Dockerfile
# Multi-stage build for optimal image size

# Stage 1: Builder
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies (git for go modules, gcc for SQLite CGO)
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=1 is required for SQLite driver
# Static linking for security
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o server .

# Stage 2: Runtime
FROM alpine:latest

# Install ca-certificates for HTTPS requests and sqlite for database operations
RUN apk --no-cache add ca-certificates sqlite

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/server .

# Copy database directory structure (will be empty initially)
COPY --from=builder /app/database ./database

# Copy configuration files
COPY --from=builder /app/config ./config

# Copy environment file for defaults
COPY --from=builder /app/.env .

# Copy existing database if it exists
COPY --from=builder /app/DBPath ./DBPath

# Ensure DBPath directory exists and set proper permissions
RUN mkdir -p ./DBPath && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port 8080 (default API port)
EXPOSE 8080

# Create volume for database persistence
VOLUME ["/app/DBPath"]

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Run the application
CMD ["./server"]