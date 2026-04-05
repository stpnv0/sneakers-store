# Order Service

gRPC-микросервис управления заказами с интеграцией Kafka. Принимает заказы, публикует события, получает обновления статусов платежей.

## Ответственность

- Создание заказов из содержимого корзины
- Получение заказов пользователя с проверкой владения
- Управление статусами заказов с валидацией переходов
- Публикация событий `OrderCreated` в Kafka
- Потребление событий `PaymentProcessed` из Kafka (с retry + DLQ)

## Архитектура

```
gRPC-хендлер
    |
OrderService
    |
    +-- OrderRepository  (PostgreSQL)
    +-- EventPublisher   (Kafka Producer)

Kafka Consumer (PaymentProcessed)
    |
OrderService.HandlePaymentProcessed
```

Интерфейсы определены в `internal/service/interfaces.go`:
- `OrderRepository` — CRUD-операции с заказами
- `EventPublisher` — публикация событий в Kafka

## gRPC-эндпоинты

| RPC | Описание |
|-----|----------|
| `CreateOrder` | Создать заказ (items + суммы в копейках) |
| `GetOrder` | Получить заказ по ID (только свой) |
| `GetUserOrders` | Все заказы пользователя |
| `UpdateOrderStatus` | Обновить статус (внутренний, вызывается из Kafka consumer) |

## Статусы заказа

```
PENDING_PAYMENT  -->  PAID  -->  SHIPPED
       |                |
       v                v
 PAYMENT_FAILED    CANCELLED
       |
       v
 PENDING_PAYMENT  (повторная попытка)
```

| Статус | Описание |
|--------|----------|
| `PENDING_PAYMENT` | Ожидает оплаты |
| `PAID` | Оплачен |
| `PAYMENT_FAILED` | Ошибка оплаты |
| `SHIPPED` | Отправлен |
| `CANCELLED` | Отменён |

## Жизненный цикл заказа

```
CreateOrder --> Kafka: OrderCreated
                          |
                Payment Service создаёт платёж в ЮKassa
                          |
                Kafka: PaymentProcessed
                          |
                HandlePaymentProcessed --> PAID / PAYMENT_FAILED
```

## Схема базы данных

```sql
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL,        -- PENDING_PAYMENT, PAID, и т.д.
    total_amount INTEGER NOT NULL,      -- сумма в копейках
    payment_url TEXT,                   -- ссылка на страницу оплаты
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);

CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    sneaker_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    price_at_purchase INTEGER NOT NULL, -- цена на момент покупки в копейках
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);
```

## Конфигурация

| Переменная окружения | Описание |
|---------------------|----------|
| `CONFIG_PATH` | Путь к YAML-конфигу |

```yaml
env: "local"
grpc:
  port: 44048
  timeout: 10s
postgres:
  host: order_postgres
  port: 5432
  user: order_user
  password: order_password
  dbname: order_db
  sslmode: disable
kafka:
  brokers:
    - "kafka:9093"
  topic: "orders"
```

## Локальный запуск

```bash
POSTGRES_DSN="postgres://order_user:order_password@localhost:5438/order_db?sslmode=disable" \
  go run ./cmd/migrator -dir ./migrations -command up

CONFIG_PATH=./config/config.yaml go run ./cmd/api
```

## Тесты

```bash
go test ./... -v
```
