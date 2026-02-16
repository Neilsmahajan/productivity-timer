# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Install swag for generating Swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger docs
RUN swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o productivity-timer ./cmd/api

# Production stage
FROM alpine:3.19

# Install ca-certificates for HTTPS connections (needed for MongoDB Atlas)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Create non-root user for security
RUN adduser -D -g '' appuser

# Copy the binary from builder
COPY --from=builder /app/productivity-timer .

# Copy any static assets if needed (favicon, etc.)
COPY --from=builder /app/favicon.ico ./favicon.ico

# Use non-root user
USER appuser

# Expose port (Railway will set PORT env var)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8080}/health || exit 1

# Run the application
CMD ["./productivity-timer"]
