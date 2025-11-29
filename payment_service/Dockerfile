FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git build-base

COPY payment_service .

RUN go mod download && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/payment_service ./cmd/api/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/payment_migrator ./cmd/migrator/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder --chmod=755 /app/payment_service /app/payment_service
COPY --from=builder --chmod=755 /app/payment_migrator /app/payment_migrator
COPY --from=builder /app/config /app/config
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8084

CMD ["/app/payment_service"]
