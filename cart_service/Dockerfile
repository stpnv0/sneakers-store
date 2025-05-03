FROM golang:1.23-alpine

WORKDIR /app

# Установка зависимостей
RUN apk add --no-cache redis postgresql-client

# Сделаем выполняемым скрипт миграций
COPY scripts/run_migrations.sh /app/scripts/run_migrations.sh
RUN chmod +x /app/scripts/run_migrations.sh

# Установка зависимостей Go
COPY go.mod go.sum ./
RUN go mod tidy && go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN go build -o main cmd/api/main.go

# Запуск приложения
CMD ["./main"]