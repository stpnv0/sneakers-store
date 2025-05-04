#!/bin/sh

# Проверка наличия переменных окружения
if [ -z "$POSTGRES_HOST" ] || [ -z "$POSTGRES_USER" ] || [ -z "$POSTGRES_PASSWORD" ] || [ -z "$POSTGRES_DB" ]; then
  echo "Error: Required environment variables are not set"
  echo "Required: POSTGRES_HOST, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB"
  exit 1
fi

# Формирование строки подключения
CONNECTION_STRING="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:5432/${POSTGRES_DB}?sslmode=disable"

echo "Waiting for PostgreSQL to be ready..."
# Ожидание доступности PostgreSQL
for i in $(seq 1 30); do
  pg_isready -h ${POSTGRES_HOST} -U ${POSTGRES_USER} && break
  echo "Waiting for PostgreSQL to be ready... ${i}/30"
  sleep 1
done

# Проверка успешности подключения
if ! pg_isready -h ${POSTGRES_HOST} -U ${POSTGRES_USER}; then
  echo "Error: Could not connect to PostgreSQL after 30 attempts"
  exit 1
fi

echo "PostgreSQL is ready. Running migrations..."

# Run migrations using psql directly
PGPASSWORD=${POSTGRES_PASSWORD} psql -h ${POSTGRES_HOST} -U ${POSTGRES_USER} -d ${POSTGRES_DB} -f /app/scripts/migrations/001_create_fav_table.up.sql

# Check if the migration was successful
if [ $? -eq 0 ]; then
  echo "Migrations completed successfully"
else
  echo "Error: Migrations failed"
  exit 1
fi

exit 0
