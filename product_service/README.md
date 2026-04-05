# Product Service

gRPC-микросервис каталога товаров с двухуровневым Redis-кэшированием и управлением изображениями через S3/MinIO.

## Ответственность

- Хранение и выдача товаров (название, цена в копейках, изображение)
- Двухуровневое Redis-кэширование (L1 — отдельный товар, L2 — страницы списка)
- Генерация presigned URL для загрузки изображений в MinIO/S3
- Обновление ключей изображений с валидацией формата

## Архитектура

```
gRPC-хендлер
    |
Сервисный слой (app.Service)
    |
    +-- ProductPostgres (репозиторий)
    +-- ProductCache    (Redis)
    +-- FileStore       (MinIO/S3)
```

Интерфейсы определены на стороне потребителя в `internal/app/interfaces.go`:
- `ProductPostgres` — операции с БД
- `ProductCache` — универсальный кэш Get/Set/Delete
- `FileStore` — генерация presigned URL

## gRPC-эндпоинты

| RPC | Описание |
|-----|----------|
| `GetSneakerByID` | Получить товар по ID (L1-кэш) |
| `GetAllSneakers` | Список с пагинацией (L2-кэш) |
| `GetSneakersByIDs` | Пакетное получение по списку ID |
| `AddSneaker` | Добавить новый товар |
| `DeleteSneaker` | Удалить товар |
| `GenerateUploadURL` | Получить presigned S3 PUT URL |
| `UpdateProductImage` | Обновить ключ изображения товара |

## Схема базы данных

```sql
CREATE TABLE sneakers (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    price BIGINT NOT NULL,            -- цена в копейках
    image_key VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE INDEX idx_sneakers_title ON sneakers (title);
```

> Цена хранится в копейках (`BIGINT`). Миграция `00003_change_price_to_bigint.sql` конвертировала из `REAL` в `BIGINT` с умножением на 100.

## Конфигурация

| Переменная окружения | Описание |
|---------------------|----------|
| `CONFIG_PATH` | Путь к YAML-конфигу (по умолчанию `./config/docker.yaml`) |

```yaml
env: "local"
grpc:
  port: 44045
  timeout: 10s
database:
  host: product_postgres
  port: 5432
  user: product_user
  password: product_password
  dbname: product_db
s3:
  endpoint: "minio:9000"
  access_key: "admin"
  secret_key: "admin123"
  bucket: "products"
redis:
  addr: "product_redis:6379"
cache_ttl: 10m
```

## Локальный запуск

```bash
POSTGRES_DSN="postgres://product_user:product_password@localhost:5435/product_db?sslmode=disable" \
  go run ./cmd/migrator -dir ./migrations -command up

CONFIG_PATH=./config/local.yaml go run ./cmd
```

## Тесты

```bash
go test ./... -v
```
