# Frontend Service Dockerfile
# Multi-stage build for optimal image size

# Stage 1: Builder
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install git (sometimes needed for go modules)
RUN apk add --no-cache git

# Copy go mod files first for better caching
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 for static linking
# GOOS=linux for Linux containers
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o frontend-server .

# Stage 2: Runtime
FROM alpine:latest

# Install ca-certificates for HTTPS requests to backend
RUN apk --no-cache add ca-certificates

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/frontend-server .

# Copy web assets (templates and static files)
COPY --from=builder /app/web ./web

# Copy config directory
COPY --from=builder /app/config ./config

# Copy environment file for defaults
COPY --from=builder /app/.env .

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port 3000 (default frontend port)
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/ || exit 1


ENV API_BASE_URL=http://api-container:8080/api


# Run the application
CMD ["./frontend-server"]