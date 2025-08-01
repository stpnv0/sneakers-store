# Sneakers Store - Микросервисный Проект Интернет-Магазина

## 🚀 О Проекте

**Sneakers Store** — это полноценное fullstack-приложение для интернет-магазина кроссовок, построенное на основе микросервисной архитектуры. 

## 🏛️ Архитектура проекта (на данный момент):
<img width="1223" height="781" alt="image" src="https://github.com/user-attachments/assets/448c76b7-0a86-437e-8920-b5d6126d7493" />





### Ключевые Компоненты:

*   **`nginx`**: Входная точка. Выполняет роль обратного прокси, отдает статические файлы фронтенда и направляет все API-запросы на `api-gateway`.
*   **`api-gateway` (Go, Gin)**: Центральный шлюз. 
    *   Предоставляет единый REST/JSON API для клиентских приложений.
    *   Выполняет централизованную аутентификацию и авторизацию.
    *   Транслирует запросы в другие микросервисы .
*   **`product_service` (Go, gRPC)**: Отвечает за каталог товаров. Является "источником правды" для всей информации о продуктах. 
*   **`sso_service` (Go, gRPC)**: Сервис единого входа (Single Sign-On). Отвечает за регистрацию, аутентификацию пользователей и выпуск JWT-токенов.
*   **`cart_service` (Go, Gin)**: Управляет состоянием корзины пользователя. 
*   **`favourites_service` (Go, Gin)**: Управляет списком избранных товаров пользователя. 
*   **`minio`**: S3-совместимое объектное хранилище для всех изображений товаров. 
*   **`postgres` & `redis`**: Каждый сервис за исключением api-gateway имеет свой собственный экземпляр PostgreSQL для персистентного хранения и Redis для кэширования.

## 🛠️ Технологический Стек

*   **Язык**: Go
*   **API**: gRPC (для межсервисного взаимодействия), REST/JSON (публичный API)
*   **Веб-фреймворки**: Gin Gonic
*   **Базы Данных**: PostgreSQL (основная), Redis (кэш), SQLite (для SSO)
*   **Хранилище Файлов**: MinIO (S3-совместимое)
*   **Оркестрация**: Docker & Docker Compose
*   **Миграции БД**: `golang-migrate/migrate`
*   **Протоколы**: Protobuf

## 🚀 Быстрый старт
### Установка и запуск

1. Клонируйте репозиторий:
```bash
git clone https://github.com/stpnv0/sneakers-store.git
cd sneakers-store
```

2. Переходим в папку с фронтендом и устанавливаем все зависимости
```bash
cd frontend && npm install
```

3. Запускаем сборку и возвращаемся в корень
```bash
npm run build && cd ..
```

4. Запустите все сервисы через Docker Compose:
```bash
docker compose up -d
```

5. Приложение будет доступно на http://localhost

### Структура проекта
Проект организован как монорепозиторий. Каждый сервис находится в своей директории:

    sneakers-store/
    ├── api_gateway/      # API Gateway сервис
    ├── cart_service/     # Сервис корзины
    ├── fav_service/      # Сервис избранного
    ├── frontend/         # React приложение
    ├── nginx/            # Nginx конфигурация
    ├── product_service/  # Сервис продуктов
    ├── protos/           # Protobuf файлы
    └── sso_service/      # Сервис авторизации


