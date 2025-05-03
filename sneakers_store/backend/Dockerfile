FROM golang:1.23-alpine

WORKDIR /app

# Установка зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Изменение настроек конфигурации для Docker
RUN sed -i 's/host: "localhost"/host: "postgres"/' config/config.yaml
# Изменение адреса SSO сервиса для доступа к хосту из контейнера
RUN sed -i 's/address: "localhost:44044"/address: "host.docker.internal:44044"/' config/config.yaml
# Не изменяем адрес сервиса корзины, он подставляется из переменной окружения

# Сборка приложения
RUN go build -o /go/bin/app cmd/main.go

EXPOSE 8080

CMD ["/go/bin/app"] 