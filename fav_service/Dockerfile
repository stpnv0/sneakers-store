# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o fav_service ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build -o fav_migrator ./cmd/migrator

# Final image
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/fav_service /app/fav_service
COPY --from=builder /app/fav_migrator /app/fav_migrator
COPY --from=builder /app/config /app/config
COPY --from=builder /app/migrations /app/migrations

ENV GIN_MODE=release

EXPOSE 8082

CMD ["/app/fav_service"]