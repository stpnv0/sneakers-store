package main

import (
	"cart_service/internal/config"
	grpcapp "cart_service/internal/grpc"
	"cart_service/internal/repository"
	"cart_service/internal/services"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
)

func main() {
	cfg := config.MustLoad()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})).With(slog.String("service", "cart"))

	redisClient := mustInitRedis(cfg, log)

	db, err := repository.NewPostgresDB(cfg.Postgres.DSN())
	if err != nil {
		log.Error("failed to connect to PostgreSQL", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	redisRepo := repository.NewRedisRepository(redisClient)
	pgRepo := repository.NewPostgresRepository(db)

	cartService := services.NewCartCacheAsideService(
		pgRepo,
		redisRepo,
		log,
		24*time.Hour,
	)

	// Initialize gRPC server
	grpcApp := grpcapp.New(log, cartService, cfg.GRPC.Port)

	go grpcApp.MustRun()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down cart service...")
	grpcApp.Stop()
	log.Info("cart service stopped")
}

func mustInitRedis(cfg *config.Config, log *slog.Logger) *redis.Client {
	addr := cfg.Redis.Host + ":" + strconv.Itoa(cfg.Redis.Port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Error("failed to connect to Redis", slog.String("error", err.Error()))
		os.Exit(1)
	}
	log.Info("successfully connected to Redis", slog.String("addr", addr))
	return client
}
