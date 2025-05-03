#!/bin/sh

set -e

# Функция для проверки доступности PostgreSQL
wait_for_postgres() {
  echo "Waiting for PostgreSQL to start..."
  # Увеличим количество попыток и интервал
  retries=10
  interval=3
  # Используем pg_isready для надежной проверки
  until pg_isready -h "${POSTGRES_HOST:-postgres}" -U "${POSTGRES_USER:-root}" -d "${POSTGRES_DB:-sneaker}" -q || [ $retries -eq 0 ]; do
    echo "PostgreSQL is unavailable - sleeping ($retries retries left)"
    sleep $interval
    retries=$((retries - 1))
  done

  if [ $retries -eq 0 ]; then
    echo "PostgreSQL did not become available in time."
    exit 1
  fi
  echo "PostgreSQL is up - executing command"
}

# Вызываем функцию ожидания
wait_for_postgres

echo "Running migrations..."
PGPASSWORD=${POSTGRES_PASSWORD:-password} psql -h "${POSTGRES_HOST:-postgres}" -U "${POSTGRES_USER:-root}" -d "${POSTGRES_DB:-sneaker}" -f /app/scripts/migrations/001_create_cart_tables.up.sql

echo "Migrations completed successfully!"