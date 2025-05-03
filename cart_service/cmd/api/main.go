package main

import (
	"cart_service/internal/config"
	"cart_service/internal/handlers"
	"cart_service/internal/repository"
	"cart_service/internal/router"
	"cart_service/internal/services"
	"context"
	"log"
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
	pgDSN := os.Getenv("POSTGRES_DSN")
	if pgDSN == "" {
		pgDSN = "postgres://root:password@postgres:5432/sneaker?sslmode=disable"
		slogger.Info("Using default PostgreSQL DSN")
	}

	db, err := repository.NewPostgresDB(pgDSN)
	if err != nil {
		slogger.Error("Failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Инициализация репозиториев
	redisRepo := repository.NewRedisRepository(redisClient)
	pgRepo := repository.NewPostgresRepository(db)

	// Настраиваем стандартный логгер для сервиса корзины
	logger := log.New(os.Stdout, "[CART] ", log.LstdFlags)

	// Инициализация сервиса корзины с Cache-Aside паттерном
	cartService := services.NewCartCacheAsideService(
		pgRepo,
		redisRepo,
		logger,
		24*time.Hour, // TTL для кэша - 24 часа
	)

	// Инициализация обработчиков HTTP с адаптацией интерфейса
	cartHandler := handlers.NewCartHandler(cartService)

	// Инициализация и настройка роутера
	r := router.InitRouter(cartHandler)

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
		slogger.Info("Starting cart service", "address", cfg.HTTPServer.Address)
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
