FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY product_service/go.mod product_service/go.sum ./product_service/
COPY protos/go.mod protos/go.sum ./protos/

WORKDIR /app/product_service

RUN go mod download

WORKDIR /app
COPY . .

WORKDIR /app/product_service

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/product_service ./cmd/main.go


FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/product_service .

COPY --from=builder /app/product_service/config ./config

COPY --from=builder /app/product_service/migrations ./migrations

EXPOSE 44045
CMD ["./product_service"]