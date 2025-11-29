FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git build-base

COPY protos ./protos
COPY order_service .

RUN sed -i 's|replace github.com/stpnv0/protos => ../protos|replace github.com/stpnv0/protos => /app/protos|g' go.mod

RUN go mod download && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/order_service ./cmd/api/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/order_migrator ./cmd/migrator/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder --chmod=755 /app/order_service /app/order_service
COPY --from=builder --chmod=755 /app/order_migrator /app/order_migrator
COPY --from=builder /app/config /app/config
COPY --from=builder /app/migrations /app/migrations

EXPOSE 44048

CMD ["/app/order_service"]
