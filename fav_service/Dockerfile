FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY protos ./protos
COPY fav_service/go.mod fav_service/go.sum ./fav_service/
WORKDIR /app/fav_service
RUN go mod download

COPY fav_service ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/fav_service  ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/bin/fav_migrator ./cmd/migrator

# ---

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder --chmod=755 /app/bin/fav_service  ./fav_service
COPY --from=builder --chmod=755 /app/bin/fav_migrator ./fav_migrator
COPY --from=builder /app/fav_service/config     ./config
COPY --from=builder /app/fav_service/migrations ./migrations

USER appuser

EXPOSE 44047

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["true"]

CMD ["/app/fav_service"]
