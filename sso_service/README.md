# SSO Service

gRPC-сервис аутентификации и авторизации. Регистрация пользователей, логин, выпуск JWT-токенов, проверка роли администратора.

## Ответственность

- Регистрация пользователей с хэшированием паролей (bcrypt)
- Логин с выпуском JWT-токена (HMAC-SHA256)
- Проверка роли администратора
- Управление секретами приложений через таблицу `apps`

## Архитектура

```
gRPC-хендлер (authgrpc)
    |
Auth Service (services/auth)
    |
    +-- UserSaver     (postgres)
    +-- UserProvider   (postgres)
    +-- AppProvider    (postgres)
    +-- JWT-библиотека
```

Интерфейсы определены на стороне потребителя в `internal/services/auth/auth.go`:
- `UserSaver` — сохранение новых пользователей
- `UserProvider` — получение пользователей, проверка admin-статуса
- `AppProvider` — получение записей приложений (секреты для подписи JWT)

## gRPC-эндпоинты

| RPC          | Описание                     |
|--------------|------------------------------|
| `Register`   | Создание учётной записи      |
| `Login`      | Аутентификация, возврат JWT  |
| `IsAdmin`    | Проверка роли администратора |

## Структура JWT-токена

```json
{
  "uid": 1,
  "email": "user@example.com",
  "app_id": 1,
  "exp": 1700000000
}
```

Подписывается HMAC-SHA256 с секретом приложения из таблицы `apps`.

## Схема базы данных

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    pass_hash BYTEA NOT NULL,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_email ON users (email);

CREATE TABLE apps (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    secret TEXT NOT NULL UNIQUE
);
```

При первом запуске миграция `00002_insert_app_secret.sql` создаёт запись приложения `sneakers` с автоматически сгенерированным секретом (`gen_random_uuid()`).

## Конфигурация

| Переменная окружения | Описание                                          |
|----------------------|---------------------------------------------------|
| `CONFIG_PATH`        | Путь к YAML-конфигу                               |
| `APP_SECRET`         | Секрет для верификации JWT на стороне API Gateway |

```yaml
env: "local"
db:
  host: sso_postgres
  port: 5432
  user: sso_user
  password: sso_password
  dbname: sso_db
token_ttl: 24h
grpc:
  port: 44044
  timeout: 10s
```

## Локальный запуск

```bash
POSTGRES_DSN="postgres://sso_user:sso_password@localhost:5436/sso_db?sslmode=disable" \
  go run ./cmd/migrator -dir ./migrations -command up

CONFIG_PATH=./config/config.yaml go run ./cmd/sso
```

## Тесты

```bash
go test ./... -v
```
