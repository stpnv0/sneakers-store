FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY protos ./protos
COPY sso_service/go.mod sso_service/go.sum ./sso_service/
WORKDIR /app/sso_service
RUN go mod download

COPY sso_service ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/sso_service  ./cmd/sso && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/bin/sso_migrator ./cmd/migrator

# ---

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder --chmod=755 /app/bin/sso_service  ./sso_service
COPY --from=builder --chmod=755 /app/bin/sso_migrator ./sso_migrator
COPY --from=builder /app/sso_service/config     ./config
COPY --from=builder /app/sso_service/migrations ./migrations

USER appuser

EXPOSE 44044

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["true"]

CMD ["/app/sso_service"]
