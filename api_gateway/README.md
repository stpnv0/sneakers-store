# API Gateway

Центральная HTTP-точка входа платформы Sneakers Store. Предоставляет REST/JSON API и транслирует запросы в gRPC-вызовы к нижестоящим микросервисам.

## Ответственность

- Аутентификация запросов через JWT (Bearer-токен)
- Маршрутизация публичных и защищённых эндпоинтов
- Контроль доступа администратора для управления товарами
- Трансляция HTTP-запросов в gRPC-вызовы
- Агрегация данных между сервисами (например, получение цен при создании заказа)
- Генерация и пропагация `request_id` через все downstream-сервисы

## Архитектура

```
Клиент (HTTP)
    |
API Gateway (Gin)
    |
    +-- Product Service  (gRPC :44045)
    +-- SSO Service      (gRPC :44044)
    +-- Cart Service     (gRPC :44046)
    +-- Favourites       (gRPC :44047)
    +-- Order Service    (gRPC :44048)
```

- **Интерфейсы на стороне потребителя**: каждый хендлер определяет нужный ему интерфейс, а не конкретный gRPC-клиент
- **Admin-мидлвар**: операции записи товаров требуют проверку `IsAdmin` через SSO Service
- **Без базы данных**: шлюз полностью stateless

## API-эндпоинты

### Публичные

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/v1/products` | Список товаров (с пагинацией) |
| GET | `/api/v1/products/:id` | Товар по ID |
| GET | `/api/v1/products/batch` | Товары по списку ID |
| POST | `/api/v1/auth/register` | Регистрация |
| POST | `/api/v1/auth/login` | Вход, возвращает JWT |

### Защищённые (требуется JWT)

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/v1/cart/` | Добавить товар в корзину |
| GET | `/api/v1/cart/` | Содержимое корзины |
| PUT | `/api/v1/cart/:id` | Изменить количество |
| DELETE | `/api/v1/cart/:id` | Удалить из корзины |
| POST | `/api/v1/favourites/` | Добавить в избранное |
| GET | `/api/v1/favourites/` | Список избранного |
| DELETE | `/api/v1/favourites/:id` | Удалить из избранного |
| GET | `/api/v1/favourites/:id` | Проверка избранного |
| GET | `/api/v1/favourites/batch` | Пакетное получение избранного |
| POST | `/api/v1/orders/` | Создать заказ |
| GET | `/api/v1/orders/` | Заказы пользователя |
| GET | `/api/v1/orders/:id` | Заказ по ID (только свои) |
| POST | `/api/v1/images/generate-upload-url` | Presigned URL для загрузки в S3 |

### Административные (требуется JWT + роль администратора)

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/v1/products` | Добавить новый товар |
| POST | `/api/v1/products/:id/image` | Обновить изображение товара |

## Конфигурация

| Переменная окружения | Описание |
|---------------------|----------|
| `CONFIG_PATH` | Путь к YAML-конфигу (по умолчанию `config.yaml`) |
| `APP_SECRET` | Секрет для верификации JWT (обязателен) |

```yaml
listen_addr: ":8083"
downstream:
  product_grpc: "product_service:44045"
  sso_grpc: "sso_service:44044"
  cart_grpc: "cart_service:44046"
  favourites_grpc: "fav_service:44047"
  order_grpc: "order_service:44048"
```

## Локальный запуск

```bash
APP_SECRET="your-secret" CONFIG_PATH=./config/config.yaml go run ./cmd
```

## Тесты

```bash
go test ./... -v
```
