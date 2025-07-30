package main

import (
	"cart_service/internal/config"
	"cart_service/internal/handlers"
	"cart_service/internal/repository"
	"cart_service/internal/router"
	"cart_service/internal/services"
	"context"
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
	cfg := config.MustLoad()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})).With(slog.String("service", "cart"))

	redisClient := mustInitRedis(cfg, log)

	db, err := repository.NewPostgresDB(cfg.Postgres.DSN)
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

	cartHandler := handlers.NewCartHandler(cartService, log)

	r := router.InitRouter(cartHandler)
	runServer(r, cfg, log)
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

func runServer(handler http.Handler, cfg *config.Config, log *slog.Logger) {
	server := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Info("starting cart service", slog.String("address", cfg.HTTPServer.Address))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed to start", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", slog.String("error", err.Error()))
	}
	log.Info("server exiting")
}
