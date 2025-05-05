FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git protobuf-dev build-base

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /build

COPY protos /build/protos

COPY sso_service /build/sso_service

WORKDIR /build/sso_service

RUN sed -i 's|replace github.com/stpnv0/protos => ../protos|replace github.com/stpnv0/protos => /build/protos|g' go.mod

# Загружаем зависимости
RUN go mod download
RUN go mod tidy

RUN mkdir -p /build/protos/gen/go

RUN protoc --proto_path=/build/protos/proto \
           --go_out=/build/protos/gen/go --go_opt=paths=source_relative \
           --go-grpc_out=/build/protos/gen/go --go-grpc_opt=paths=source_relative \
           /build/protos/proto/sso/sso.proto

RUN CGO_ENABLED=1 go build -o /sso_service ./cmd/sso

RUN CGO_ENABLED=1 go build -o /sso_migrator ./cmd/migrator

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata sqlite

WORKDIR /app

COPY --from=builder /sso_service .
COPY --from=builder /sso_migrator .

COPY --from=builder /build/sso_service/config ./config
COPY --from=builder /build/sso_service/migrations ./migrations

RUN mkdir -p /root/storage

EXPOSE 44044

CMD ["./sso_service", "--config=./config/prod.yaml"]