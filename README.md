# Sneakers Store

Микросервисный интернет-магазин кроссовок, построенный на Go.

## Архитектура

<img width="1416" height="764" alt="image" src="https://github.com/user-attachments/assets/0d360f19-9247-48b4-ba13-c4714bb1dc47" />

### Ключевые Компоненты:

*   **`nginx`**: Входная точка. Выполняет роль обратного прокси, отдает статические файлы фронтенда и направляет все API-запросы на `api-gateway`.
*   **`api-gateway` (Go, Gin)**: Центральный шлюз.
    *   Предоставляет единый REST/JSON API для клиентских приложений.
    *   Транслирует HTTP-запросы в gRPC-вызовы к микросервисам.
    *   Оркестрирует создание заказа: получает цены из Product Service, создаёт заказ в Order Service, очищает корзину.
*   **`product_service` (Go, gRPC)**: Каталог товаров (Sneaker = Product). Кэширование через Redis (L1 — отдельный товар, L2 — списки). Хранение изображений в MinIO.
*   **`sso_service` (Go, gRPC)**: Сервис единого входа (Single Sign-On). Регистрация, аутентификация и выпуск JWT-токенов.
*   **`order_service` (Go, gRPC + HTTP)**: Сервис заказов и платежей. Синхронно создаёт платёж через YooKassa API и возвращает ссылку на оплату. Принимает вебхуки YooKassa по HTTP (:8084). Публикует события в Kafka для будущих потребителей.
*   **`cart_service` (Go, gRPC)**: Управляет корзиной пользователя. Паттерн cache-aside (PostgreSQL + Redis).
*   **`favourites_service` (Go, gRPC)**: Управляет списком избранных товаров. Паттерн cache-aside (PostgreSQL + Redis).
*   **`kafka`**: Брокер сообщений. Order Service публикует события (`OrderCreated`, `OrderPaymentUpdated`) для будущих потребителей (например, notification service).
*   **`minio`**: S3-совместимое объектное хранилище для изображений товаров.
*   **`postgres` & `redis`**: Отдельная БД на каждый сервис (database-per-service). Redis для кэширования в Product, Cart и Favourites.

### Поток данных

**Процесс заказа:**
1. Пользователь просматривает товары через Product Service
2. Добавляет товары в корзину через Cart Service
3. Создаёт заказ — API Gateway получает цены batch-запросом из Product Service, создаёт заказ в Order Service, очищает корзину
4. Order Service синхронно вызывает YooKassa API, создаёт платёж и сразу возвращает `payment_url` пользователю
5. Пользователь переходит по ссылке и оплачивает на странице YooKassa
6. YooKassa отправляет вебхук на Order Service (:8084) — статус заказа обновляется на «Оплачен»

## Технологический стек

| Категория                   | Технология                                 |
|-----------------------------|--------------------------------------------|
| Язык                        | Go 1.24                                    |
| Межсервисное взаимодействие | gRPC + Protocol Buffers                    |
| Публичный API               | REST/JSON (Gin)                            |
| Базы данных                 | PostgreSQL 15                              |
| Кэш                         | Redis                                      |
| Брокер сообщений            | Apache Kafka                               |
| Объектное хранилище         | MinIO                                      |
| Аутентификация              | JWT (HMAC-SHA256)                          |
| Платёжный шлюз              | ЮKassa                                     |
| Миграции                    | [goose](https://github.com/pressly/goose)  |
| Логирование                 | `log/slog` (структурированный JSON)        |
| Тестирование                | `testify` + `mockery v3`                   |
| Валидация                   | `go-playground/validator`                  |
| Контейнеризация             | Docker + Docker Compose                    |
| Обратный прокси             | Nginx                                      |

## Быстрый старт

### Требования

- Docker и Docker Compose
- Node.js и npm (для сборки фронтенда)
- Go 1.24+ (для локальной разработки)

### 1. Клонировать репозиторий

```bash
git clone https://github.com/stpnv0/sneakers-store.git
cd sneakers-store
```

### 2. Создать конфигурационные файлы

Скопируйте пример конфига для Order Service и укажите учётные данные YooKassa:

```bash
cp order_service/config/config.example.yaml order_service/config/config.yaml
# Отредактируйте order_service/config/config.yaml — укажите shop_id и secret_key от ЮKassa
```

### 3. Собрать фронтенд

```bash
cd frontend && npm install && npm run build && cd ..
```

### 4. Запустить все сервисы

```bash
docker compose up -d
```

Приложение будет доступно по адресу http://localhost.

### Настройка ЮKassa (оплата)

Для работы оплаты необходимо:

#### 1. Получить учётные данные
1. Зарегистрируйтесь в [ЮKassa](https://yookassa.ru/) и создайте **тестовый магазин**
2. В личном кабинете получите **shopId** и **Секретный ключ** (начинается с `test_`)
3. Пропишите их в `order_service/config/config.yaml`:

```yaml
yookassa:
  shop_id: "ВАШ_SHOP_ID"
  secret_key: "test_XXXXXXXX"
  return_url: "http://localhost/"
  notification_url: ""  # заполняется на шаге 2
```

#### 2. Настроить вебхуки

После оплаты ЮKassa должна уведомить ваш сервер об успешном платеже. Для этого ей нужен **публичный URL**, доступный из интернета

**Вариант через ngrok (рекомендуется)**

```bash
ngrok http 8084
```

Выдаст URL вида `https://xxxx-xx-xx.ngrok-free.app`.


**После получения публичного URL:**

1. Пропишите его в `order_service/config/config.yaml`:

```yaml
yookassa:
  notification_url: "https://xxxx-xx-xx.ngrok-free.app/webhook/yookassa"
```

2. Перезапустите Order Service:

```bash
docker compose restart order_service
```

3. Либо настройте вебхуки вручную в [личном кабинете YooKassa](https://yookassa.ru/my/) → Интеграция → HTTP-уведомления:
   - URL: ваш публичный URL + `/webhook/yookassa`
   - События: `payment.succeeded`, `payment.canceled`

## Структура проекта

```
sneakers-store/
├── api_gateway/         # API Gateway (Gin, REST → gRPC)
├── cart_service/        # Сервис корзины (gRPC, cache-aside)
├── fav_service/         # Сервис избранного (gRPC, cache-aside)
├── frontend/            # React-приложение
├── nginx/               # Конфигурация Nginx
├── order_service/       # Сервис заказов и платежей (gRPC + HTTP + Kafka)
├── product_service/     # Сервис товаров (gRPC, двухуровневый кэш)
├── protos/              # Общие Protocol Buffer определения
├── seed-images/         # Изображения товаров для начальной загрузки в MinIO
├── sso_service/         # Сервис авторизации (gRPC)
└── docker-compose.yml   # Оркестрация всего стека
```

Каждый сервис следует единообразной внутренней структуре:

```
service/
├── cmd/
│   ├── api/main.go      # Точка входа
│   └── migrator/main.go # Инструмент миграций
├── internal/
│   ├── config/          # Загрузка конфигурации
│   ├── models/          # Модели данных
│   ├── repository/      # Слой доступа к данным
│   ├── service/         # Бизнес-логика (интерфейсы + реализация)
│   └── grpc/ или api/   # Транспортный слой (gRPC или HTTP хендлеры)
├── migrations/          # SQL-миграции (формат goose)
├── config/              # YAML-конфиги
├── Dockerfile           # Многоэтапная Docker-сборка
├── .mockery.yaml        # Конфиг генерации моков
└── go.mod
```

## Разработка

### Запуск тестов

```bash
# Тесты конкретного сервиса
cd order_service && go test ./... -v

# Тесты всех сервисов
for svc in order_service product_service sso_service cart_service fav_service; do
  echo "=== $svc ===" && cd $svc && go test ./... -v && cd ..
done
```

### Запуск миграций локально

У каждого сервиса есть отдельный бинарник-мигратор:

```bash
cd order_service
POSTGRES_DSN="postgres://user:pass@localhost:5438/order_db?sslmode=disable" \
  go run ./cmd/migrator -dir ./migrations -command up
```

### Генерация моков

У каждого сервиса есть конфиг `.mockery.yaml`. Установите mockery v3 и запустите:

```bash
cd order_service && mockery
```
