# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata build-base

# Set working directory
WORKDIR /app

# Copy source code
COPY . .

# Download dependencies
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o pocketbase ./cmd/pocketbase

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata sqlite curl

# Create app user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup -u 1000

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/pocketbase .

# Create directories
RUN mkdir -p pb-data pb-backups

# Change ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

EXPOSE 8090

# Set entrypoint
ENTRYPOINT ["/app/pocketbase"]