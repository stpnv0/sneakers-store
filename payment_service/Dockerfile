FROM golang:1.23-alpine

WORKDIR /app

RUN apk add --no-cache redis postgresql-client

COPY scripts/run_migrations.sh /app/scripts/run_migrations.sh
RUN chmod +x /app/scripts/run_migrations.sh

COPY go.mod go.sum ./
RUN go mod tidy && go mod download

COPY . .

RUN go build -o main cmd/api/main.go

CMD ["./main"]