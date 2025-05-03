# Этап сборки
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Копируем proto-файлы из корневой директории проекта
COPY ../protos ./protos

# Копируем исходный код
COPY . .

# Компилируем приложение
RUN go build -o sso_service ./cmd/sso

# Финальный этап
FROM alpine:latest

# Устанавливаем необходимые зависимости
RUN apk --no-cache add ca-certificates postgresql-client

WORKDIR /root/

# Копируем бинарный файл из этапа сборки
COPY --from=builder /app/sso_service .
COPY --from=builder /app/config ./config
COPY --from=builder /app/migrations ./migrations

# Создаем директорию для хранения данных
RUN mkdir -p /root/storage

# Создаем скрипт для запуска миграций
RUN echo '#!/bin/sh\n\
echo "Running SSO service migrations..."\n\
./sso_service --config=./config/prod.yaml\n\
echo "SSO service migrations completed."\n\
' > /root/run_migrations.sh && chmod +x /root/run_migrations.sh

# Создаем директорию для скриптов
RUN mkdir -p /app/scripts

# Создаем скрипт для запуска миграций в директории scripts
RUN echo '#!/bin/sh\n\
echo "Running SSO service migrations..."\n\
cd /root && ./sso_service --config=./config/prod.yaml\n\
echo "SSO service migrations completed."\n\
exit 0\n\
' > /app/scripts/run_migrations.sh && chmod +x /app/scripts/run_migrations.sh

# Открываем порт
EXPOSE 44044

# Запускаем приложение
CMD ["./sso_service", "--config=./config/prod.yaml"]