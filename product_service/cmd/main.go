package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"product_service/internal/app"
	"product_service/internal/config"
	grpc_handler "product_service/internal/grpc/product"
	"product_service/internal/repository"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	log := setupLogger("product_service")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	log.Info("starting service", slog.String("env", cfg.Env))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// PostgreSQL
	poolConfig, err := pgxpool.ParseConfig(cfg.DB.DSN())
	if err != nil {
		return fmt.Errorf("parse db config: %w", err)
	}
	poolConfig.MaxConns = cfg.DB.MaxConns
	poolConfig.MinConns = cfg.DB.MinConns
	poolConfig.MaxConnLifetime = 5 * time.Minute
	poolConfig.HealthCheckPeriod = 30 * time.Second
	dbPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	if err := dbPool.Ping(ctx); err != nil {
		dbPool.Close()
		return fmt.Errorf("ping db: %w", err)
	}
	defer dbPool.Close()
	log.Info("connected to postgres")

	// Redis
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr})
	defer redisClient.Close()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("connect to redis: %w", err)
	}
	log.Info("connected to redis")

	// S3 / MinIO
	fileStoreRepo, err := repository.NewFileStoreRepository(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("init file store: %w", err)
	}

	// Репозитории
	postgresRepo := repository.New(dbPool)
	redisRepo := repository.NewRedisRepository(redisClient)

	// Сервисный слой
	productService := app.NewService(postgresRepo, redisRepo, fileStoreRepo, log, cfg.CacheTTL)

	// gRPC-сервер
	grpcServer := grpc.NewServer()
	grpc_handler.Register(grpcServer, productService, log)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("gRPC server started", slog.Int("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	// Ожидание сигнала завершения или ошибки сервера
	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-errCh:
		log.Error("gRPC server failed", slog.String("error", err.Error()))
		stop()
	}

	grpcServer.GracefulStop()
	log.Info("product service stopped")
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
