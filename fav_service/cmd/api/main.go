package main

import (
	"context"
	"fav_service/internal/config"
	grpcapp "fav_service/internal/grpc"
	"fav_service/internal/repository"
	"fav_service/internal/services"
	"log/slog"
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

	// Инициализация gRPC сервера
	grpcApp := grpcapp.New(slogger, favService, cfg.GRPC.Port)

	go grpcApp.MustRun()

	// Настраиваем грациозное завершение
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Блокируемся пока не получим сигнал
	<-quit
	slogger.Info("Shutting down favourites service...")

	grpcApp.Stop()
	slogger.Info("Favourites service stopped")
}
