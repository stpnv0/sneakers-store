package main

import (
	"context"
	"fav_service/internal/config"
	"fav_service/internal/handlers"
	"fav_service/internal/repository"
	"fav_service/internal/router"
	"fav_service/internal/services"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
)

func main() {
	// Инициализация логгера
	slogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Загрузка конфигурации
	cfg := config.MustLoad()

	// Подключение к Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + strconv.Itoa(cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Проверка соединения с Redis
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		slogger.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}

	// Подключение к PostgreSQL
	pgDSN := cfg.Postgres.DSN()
	db, err := repository.NewPostgresDB(pgDSN)
	if err != nil {
		slogger.Error("Failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Инициализация репозиториев
	redisRepo := repository.NewRedisRepo(redisClient)
	pgRepo := repository.NewPostgresRepo(db)

	favService := services.NewFavService(
		pgRepo,
		redisRepo,
		24*time.Hour, // TTL для кэша - 24 часа
	)

	// Инициализация обработчиков HTTP с адаптацией интерфейса
	favHandler := handlers.NewFavHandler(favService)

	// Инициализация и настройка роутера
	r := router.InitRouter(favHandler)

	// Запуск HTTP-сервера
	server := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запускаем сервер в горутине
	go func() {
		slogger.Info("Starting favourites service", "address", cfg.HTTPServer.Address)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slogger.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Настраиваем грациозное завершение
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Блокируемся пока не получим сигнал
	<-quit
	slogger.Info("Shutting down server...")

	// Создаем контекст с таймаутом для грациозного завершения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slogger.Error("Server forced to shutdown", "error", err)
	}

	slogger.Info("Server exiting")
}
