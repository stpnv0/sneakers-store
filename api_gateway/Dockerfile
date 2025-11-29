FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY api_gateway/go.mod api_gateway/go.sum ./api_gateway/
COPY protos/go.mod protos/go.sum ./protos/

WORKDIR /app/api_gateway

RUN go mod download

WORKDIR /app
COPY . .

WORKDIR /app/api_gateway

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api_gateway ./cmd/main.go

# --- Запуск (Runtime Stage) ---
FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/bin/api_gateway .

COPY --from=builder /app/api_gateway/config ./config

EXPOSE 8083 

CMD ["./api_gateway"]