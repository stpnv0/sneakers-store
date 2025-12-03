FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git build-base

COPY protos ./protos
COPY order_service ./order_service

WORKDIR /app/order_service

RUN go mod download && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/bin/order_service ./cmd/api/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/bin/order_migrator ./cmd/migrator/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder --chmod=755 /app/bin/order_service /app/order_service
COPY --from=builder --chmod=755 /app/bin/order_migrator /app/order_migrator
COPY --from=builder /app/order_service/config /app/config
COPY --from=builder /app/order_service/migrations /app/migrations

EXPOSE 44048

CMD ["/app/order_service"]
