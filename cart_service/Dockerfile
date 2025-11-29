FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy protos first
COPY protos ./protos

# Copy cart_service code
COPY cart_service .

# Update replace directive to match Docker structure
RUN sed -i 's|replace github.com/stpnv0/protos => ../protos|replace github.com/stpnv0/protos => ./protos|g' go.mod

# Install git and build-base for dependencies
RUN apk add --no-cache git build-base

# Build the service
RUN go mod download && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -o cart_service ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build -o cart_migrator ./cmd/migrator

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder --chmod=755 /app/cart_service /app/cart_service
COPY --from=builder --chmod=755 /app/cart_migrator /app/cart_migrator
COPY --from=builder /app/config /app/config
COPY --from=builder /app/migrations /app/migrations

ENV GIN_MODE=release

EXPOSE 44046

CMD ["/app/cart_service"]