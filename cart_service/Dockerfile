FROM golang:1.23-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o cart_service ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build -o cart_migrator ./cmd/migrator

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder --chmod=755 /app/cart_service /app/cart_service
COPY --from=builder --chmod=755 /app/cart_migrator /app/cart_migrator
COPY --from=builder /app/config /app/config
COPY --from=builder /app/migrations /app/migrations

ENV GIN_MODE=release

EXPOSE 8081 

CMD ["/app/cart_service"]