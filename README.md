# Sneakers Store - Микросервисный Проект Интернет-Магазина

## О Проекте

**Sneakers Store** — это fullstack-приложение для интернет-магазина кроссовок, построенное на микросервисной архитектуре.

## Архитектура проекта:
<img width="1049" height="632" alt="image" src="https://github.com/user-attachments/assets/940a9529-f9c9-441a-a924-9d3634d616d7" />



### Ключевые Компоненты:

*   **`nginx`**: Входная точка. Выполняет роль обратного прокси, отдает статические файлы фронтенда и направляет все API-запросы на `api-gateway`.
*   **`api-gateway` (Go, Gin)**: Центральный шлюз.
    *   Предоставляет единый REST/JSON API для клиентских приложений.
    *   Транслирует запросы в микросервисы.
*   **`product_service` (Go, gRPC)**: Отвечает за каталог товаров. Является "источником правды" для всей информации о продуктах.
*   **`sso_service` (Go, gRPC)**: Сервис единого входа (Single Sign-On). Отвечает за регистрацию, аутентификацию пользователей и выпуск JWT-токенов.
*   **`order_service` (Go, gRPC)**: Сервис заказов. Отвечает за создание и обработку заказов.
*   **`payment_service` (Go, gRPC)**: Сервис оплаты. Интегрирован с YooKassa для обработки платежей.
*   **`cart_service` (Go, gRPC)**: Управляет состоянием корзины пользователя.
*   **`favourites_service` (Go, gRPC)**: Управляет списком избранных товаров пользователя.
*   **`kafka`**: Брокер сообщений для асинхронного взаимодействия между сервисами (order service, payment service).
*   **`minio`**: S3-совместимое объектное хранилище для всех изображений товаров.
*   **`postgres` & `redis`**

## Технологический Стек

*   **Язык**: Go
*   **API**: gRPC (для межсервисного взаимодействия), REST/JSON (публичный API)
*   **Веб**: Gin (API Gateway), gRPC (Microservices)
*   **Базы Данных**: PostgreSQL (основная), Redis (кэш)
*   **Брокер сообщений**: Apache Kafka
*   **Хранилище Файлов**: MinIO
*   **Оркестрация**: Docker & Docker Compose
*   **Миграции БД**: `golang-migrate/migrate`
*   **Протоколы**: Protobuf

## Конфигурация

Проект использует файлы `config.yaml` для настройки сервисов.

### Настройка YooKassa (Оплата)

Для работы сервиса оплаты (`payment_service`) необходимо:
1. Зарегистрировать тестовый магазин в [YooKassa](https://yookassa.ru/).
2. Получить `shop_id` и `secret_key`.
3. Создать файл `payment_service/config/config.yaml` на основе `payment_service/config/config.example.yaml`:
   ```bash
   cp payment_service/config/config.example.yaml payment_service/config/config.yaml
   ```
4. Вставить свои `shop_id` и `secret_key` в `config.yaml`.

## Быстрый старт
### Установка и запуск

1. Клонируйте репозиторий:
```bash
git clone https://github.com/stpnv0/sneakers-store.git
cd sneakers-store
```

2. Создайте необходимые конфигурационные файлы.

3. Переходим в папку с фронтендом и устанавливаем все зависимости
```bash
cd frontend && npm install
```

4. Запускаем сборку и возвращаемся в корень
```bash
npm run build && cd ..
```

5. Запустите все сервисы через Docker Compose:
```bash
docker compose up -d
```

6. Приложение будет доступно на http://localhost

### Структура проекта
Проект организован как монорепозиторий. Каждый сервис находится в своей директории:

    sneakers-store/
    ├── api_gateway/      # API Gateway сервис
    ├── cart_service/     # Сервис корзины
    ├── fav_service/      # Сервис избранного
    ├── frontend/         # React приложение
    ├── nginx/            # Nginx конфигурация
    ├── order_service/    # Сервис заказов
    ├── payment_service/  # Сервис оплаты
    ├── product_service/  # Сервис продуктов
    ├── protos/           # Protobuf файлы
    └── sso_service/      # Сервис авторизации


