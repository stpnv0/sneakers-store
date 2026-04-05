FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY protos ./protos
COPY product_service/go.mod product_service/go.sum ./product_service/
WORKDIR /app/product_service
RUN go mod download

COPY product_service ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/product_service  ./cmd && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/bin/product_migrator ./cmd/migrator

# ---

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder --chmod=755 /app/bin/product_service  ./product_service
COPY --from=builder --chmod=755 /app/bin/product_migrator ./product_migrator
COPY --from=builder /app/product_service/config     ./config
COPY --from=builder /app/product_service/migrations ./migrations

USER appuser

EXPOSE 44045

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["true"]

CMD ["/app/product_service"]
