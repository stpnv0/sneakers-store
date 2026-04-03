# Protos

Общий Go-модуль с определениями Protocol Buffers и сгенерированным gRPC-кодом для всех микросервисов.

## Контракты

```
protos/
├── proto/
│   ├── cart/        # .proto-файлы Cart
│   ├── favourites/  # .proto-файлы Favourites
│   ├── order/       # .proto-файлы Order
│   ├── product/     # .proto-файлы Product
│   └── sso/         # .proto-файлы SSO
├── go.mod
└── go.sum
```

## Сервисы

### SSO (Auth)

| RPC          | Описание                     |
|--------------|------------------------------|
| `Register`   | Создание пользователя        |
| `Login`      | Аутентификация, возврат JWT  |
| `IsAdmin`    | Проверка роли администратора |

### Product

| RPC                  | Описание               |
|----------------------|------------------------|
| `GetSneakerByID`     | Товар по ID            |
| `GetAllSneakers`     | Список с пагинацией    |
| `GetSneakersByIDs`   | Пакетное получение     |
| `AddSneaker`         | Добавление товара      |
| `DeleteSneaker`      | Удаление товара        |
| `GenerateUploadURL`  | Presigned URL для S3   |
| `UpdateProductImage` | Обновление изображения |

### Cart

| RPC                      | Описание                   |
|--------------------------|----------------------------|
| `AddToCart`              | Добавить товар в корзину   | 
| `GetCart`                | Содержимое корзины         |
| `UpdateCartItemQuantity` | Изменить количество        |
| `RemoveFromCart`         | Удалить товар              |
| `ClearCart`              | Очистить корзину           |

### Favourites

| RPC                    | Описание                 |
|------------------------|--------------------------|
| `AddToFavourites`      | Добавить в избранное     |
| `RemoveFromFavourites` | Удалить из избранного    |
| `GetFavourites`        | Список избранного        |
| `IsFavourite`          | Проверка наличия         |
| `GetFavouritesByIDs`   | Пакетное получение по ID |

### Order

| RPC                 | Описание                |
|---------------------|-------------------------|
| `CreateOrder`       | Создать заказ           | 
| `GetOrder`          | Заказ по ID             |
| `GetUserOrders`     | Все заказы пользователя |
| `UpdateOrderStatus` | Обновить статус заказа  |

## Генерация кода

Необходимые инструменты:
- `protoc` (Protocol Buffer compiler)
- `protoc-gen-go` (генерация Go-структур)
- `protoc-gen-go-grpc` (генерация gRPC-кода)

```bash
protoc --go_out=./gen/go --go-grpc_out=./gen/go \
  proto/order/order.proto
```
