# Cart Service

gRPC-микросервис управления корзиной покупок с паттерном cache-aside на Redis.

## Ответственность

- Добавление, обновление, удаление товаров в корзине пользователя
- Получение содержимого корзины
- Очистка корзины (после создания заказа)
- Cache-aside: сначала чтение из Redis, при промахе — из PostgreSQL

## Архитектура

```
gRPC-хендлер
    |
CartCacheAsideService
    |
    +-- CartRepository (PostgreSQL)
    +-- CartCache      (Redis)
```

Интерфейсы определены в `internal/services/cart_interfaces.go`:
- `CartRepository` — CRUD в PostgreSQL
- `CartCache` — кэш-операции в Redis

## gRPC-эндпоинты

| RPC | Описание |
|-----|----------|
| `AddToCart` | Добавить товар в корзину |
| `GetCart` | Получить все товары корзины |
| `UpdateCartItemQuantity` | Изменить количество |
| `RemoveFromCart` | Удалить товар |
| `ClearCart` | Очистить корзину пользователя |

## Схема базы данных

```sql
CREATE TABLE carts (
    user_sso_id INTEGER PRIMARY KEY,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE cart_items (
    id SERIAL PRIMARY KEY,
    cart_id INTEGER NOT NULL,
    user_sso_id INTEGER NOT NULL,
    sneaker_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    added_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    FOREIGN KEY (cart_id) REFERENCES carts(user_sso_id) ON DELETE CASCADE
);

CREATE INDEX idx_cart_items_cart_id ON cart_items(cart_id);
CREATE INDEX idx_cart_items_sneaker_id ON cart_items(sneaker_id);
```

## Конфигурация

| Переменная окружения | Описание |
|---------------------|----------|
| `CONFIG_PATH` | Путь к YAML-конфигу |

```yaml
env: "local"
grpc:
  port: 44046
postgres:
  host: cart_postgres
  port: 5432
  user: cart_user
  password: cart_password
  dbname: cart_db
  sslmode: disable
  max_connections: 10
  connection_timeout: 5
redis:
  host: cart_redis
  port: 6379
  password: ""
  db: 0
  expiration: "168h"
```

## Локальный запуск

```bash
POSTGRES_DSN="postgres://cart_user:cart_password@localhost:5433/cart_db?sslmode=disable" \
  go run ./cmd/migrator -dir ./migrations -command up

CONFIG_PATH=./config/config.yaml go run ./cmd/api
```

## Тесты

```bash
go test ./... -v
```
