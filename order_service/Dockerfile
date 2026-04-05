FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY protos ./protos
COPY order_service/go.mod order_service/go.sum ./order_service/
WORKDIR /app/order_service
RUN go mod download

COPY order_service ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/order_service  ./cmd/api/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/bin/order_migrator ./cmd/migrator/main.go

# ---

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder --chmod=755 /app/bin/order_service  ./order_service
COPY --from=builder --chmod=755 /app/bin/order_migrator ./order_migrator
COPY --from=builder /app/order_service/config     ./config
COPY --from=builder /app/order_service/migrations ./migrations

USER appuser

EXPOSE 44048 8084

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8084/healthz || exit 1

CMD ["/app/order_service"]
