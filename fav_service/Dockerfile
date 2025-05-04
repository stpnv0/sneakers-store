FROM golang:1.21-alpine AS builder

# Install necessary dependencies including PostgreSQL client
RUN apk add --no-cache git postgresql-client

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fav_service ./cmd/api

# Final image
FROM alpine:latest

# Install necessary packages including PostgreSQL client
RUN apk --no-cache add ca-certificates tzdata postgresql-client

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/fav_service .

# Copy configuration files
COPY --from=builder /app/config /app/config

# Copy migration scripts
COPY --from=builder /app/scripts /app/scripts

# Make migration script executable
RUN chmod +x /app/scripts/run_migrations.sh

# Set environment variables
ENV GIN_MODE=release

# Expose port
EXPOSE 8082

# Run the application
CMD ["/app/fav_service"]
