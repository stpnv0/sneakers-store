FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY protos ./protos
COPY cart_service/go.mod cart_service/go.sum ./cart_service/
WORKDIR /app/cart_service
RUN go mod download

COPY cart_service ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/cart_service  ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/bin/cart_migrator ./cmd/migrator

# ---

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder --chmod=755 /app/bin/cart_service  ./cart_service
COPY --from=builder --chmod=755 /app/bin/cart_migrator ./cart_migrator
COPY --from=builder /app/cart_service/config     ./config
COPY --from=builder /app/cart_service/migrations ./migrations

USER appuser

EXPOSE 44046

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["true"]

CMD ["/app/cart_service"]
