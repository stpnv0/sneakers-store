# Favourites Service

gRPC-микросервис для управления списком избранных товаров с кэшированием через Redis.

## Ответственность

- Добавление и удаление товаров из избранного
- Получение списка избранного пользователя
- Проверка, находится ли товар в избранном
- Пакетное получение избранных по списку ID
- Cache-aside: Redis как первый уровень, PostgreSQL как источник истины

## Архитектура

```
gRPC-хендлер
    |
FavService
    |
    +-- FavouritesRepo (PostgreSQL)
    +-- CacheRepo      (Redis)
```

Интерфейсы определены в `internal/services/fav.go`:
- `FavouritesRepo` — CRUD-операции с PostgreSQL
- `CacheRepo` — кэширование через Redis

## gRPC-эндпоинты

| RPC | Описание |
|-----|----------|
| `AddToFavourites` | Добавить товар в избранное |
| `RemoveFromFavourites` | Удалить из избранного |
| `GetFavourites` | Все избранные товары пользователя |
| `IsFavourite` | Проверить наличие в избранном |
| `GetFavouritesByIDs` | Пакетное получение по списку sneaker ID |

## Схема базы данных

```sql
CREATE TABLE favourites_items (
    id SERIAL PRIMARY KEY,
    user_sso_id INTEGER NOT NULL,
    sneaker_id INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_user_sneaker UNIQUE (user_sso_id, sneaker_id)
);

CREATE INDEX idx_favourites_user_sso_id ON favourites_items (user_sso_id);
```

## Конфигурация

| Переменная окружения | Описание |
|---------------------|----------|
| `CONFIG_PATH` | Путь к YAML-конфигу |

```yaml
env: "local"
grpc:
  port: 44047
postgres:
  host: favourites_postgres
  port: 5432
  user: favourites_user
  password: favourites_password
  dbname: favourites_db
  sslmode: disable
redis:
  host: favourites_redis
  port: 6379
```

## Локальный запуск

```bash
POSTGRES_DSN="postgres://favourites_user:favourites_password@localhost:5434/favourites_db?sslmode=disable" \
  go run ./cmd/migrator -dir ./migrations -command up

CONFIG_PATH=./config/config.yaml go run ./cmd/api
```

## Тесты

```bash
go test ./... -v
```
