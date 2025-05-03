#!/bin/bash

# Переменные окружения для режима разработки
export REDIS_HOST=localhost
export REDIS_PORT=6379
export MAIN_SERVICE_URL=http://localhost:8080

# Запуск Redis в Docker, если он еще не запущен
if ! docker ps | grep -q cart_redis; then
    echo "Запуск Redis..."
    docker run --name cart_redis -p 6379:6379 -d redis:alpine
    echo "Redis запущен"
else
    echo "Redis уже запущен"
fi

# Компиляция и запуск приложения
echo "Запуск микросервиса корзины..."
go run cmd/api/main.go 