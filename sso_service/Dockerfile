FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY protos ./protos

COPY sso_service .

# Заменяем относительный путь в replace директиве
RUN sed -i 's|replace github.com/stpnv0/protos => ../protos|replace github.com/stpnv0/protos => ./protos|g' go.mod

# Устанавливаем необходимые зависимости для сборки
RUN apk add --no-cache git build-base

# Собираем приложения
RUN go mod download && \
    go mod tidy && \
    CGO_ENABLED=1 go build -o sso_service ./cmd/sso && \
    CGO_ENABLED=1 go build -o sso_migrator ./cmd/migrator

FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite tzdata

WORKDIR /root/

COPY --from=builder /app/sso_service .
COPY --from=builder /app/sso_migrator .
COPY --from=builder /app/config ./config
COPY --from=builder /app/migrations ./migrations

RUN mkdir -p /root/storage

EXPOSE 44044

CMD ["./sso_service", "--config=./config/prod.yaml"]