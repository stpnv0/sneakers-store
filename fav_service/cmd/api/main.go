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

	"fav_service/internal/config"
	grpcapp "fav_service/internal/grpc"
	"fav_service/internal/repository"
	"fav_service/internal/services"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	log := setupLogger("fav_service")

	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "config/config.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + strconv.Itoa(cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return err
	}
	defer redisClient.Close()
	log.Info("connected to redis")

	// PostgreSQL
	db, err := repository.NewPostgresDB(cfg.Postgres.DSN())
	if err != nil {
		return err
	}
	defer db.Close()
	log.Info("connected to postgres")

	redisRepo := repository.NewRedisRepo(redisClient)
	pgRepo := repository.NewPostgresRepo(db)
	favService := services.NewFavService(pgRepo, redisRepo, 24*time.Hour, log)

	grpcApp := grpcapp.New(log, favService, cfg.GRPC.Port)

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
	log.Info("favourites service stopped")
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
