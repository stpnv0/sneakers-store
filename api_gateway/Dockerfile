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

# ---

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder --chmod=755 /app/bin/api_gateway ./api_gateway
COPY --from=builder /app/api_gateway/config ./config

USER appuser

EXPOSE 8083

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8083/healthz || exit 1

CMD ["./api_gateway"]
