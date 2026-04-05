package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"

	"cart_service/internal/config"
	grpcapp "cart_service/internal/grpc"
	"cart_service/internal/repository"
	"cart_service/internal/services"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	log := setupLogger("cart_service")

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Redis
	addr := cfg.Redis.Host + ":" + strconv.Itoa(cfg.Redis.Port)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return err
	}
	defer redisClient.Close()
	log.Info("connected to redis", slog.String("addr", addr))

	// PostgreSQL
	db, err := repository.NewPostgresDB(cfg.Postgres.DSN())
	if err != nil {
		return err
	}
	defer db.Close()
	log.Info("connected to postgres")

	// Связывание зависимостей
	redisRepo := repository.NewRedisRepository(redisClient)
	pgRepo := repository.NewPostgresRepository(db)

	expiration, err := time.ParseDuration(cfg.Redis.Expiration)
	if err != nil {
		expiration = 24 * time.Hour
	}
	cartService := services.NewCartCacheAsideService(pgRepo, redisRepo, log, expiration)

	grpcApp := grpcapp.New(log, cartService, cfg.GRPC.Port)

	errCh := make(chan error, 1)
	go func() {
		if err := grpcApp.Run(); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-errCh:
		log.Error("grpc server failed", slog.String("error", err.Error()))
		stop()
	}

	grpcApp.Stop()
	log.Info("cart service stopped")
	return nil
}

func setupLogger(serviceName string) *slog.Logger {
	env := os.Getenv("ENV")

	var level slog.Level
	switch env {
	case "prod":
		level = slog.LevelInfo
	default:
		level = slog.LevelDebug
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})).
		With(slog.String("service", serviceName))
}
