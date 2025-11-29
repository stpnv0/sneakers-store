package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"payment_service/internal/api"
	"payment_service/internal/config"
	"payment_service/internal/kafka"
	"payment_service/internal/provider"
	"payment_service/internal/repository"
	"payment_service/internal/service"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg := config.MustLoad()

	// Debug logging for YooKassa config
	maskedKey := "empty"
	if len(cfg.YooKassa.SecretKey) > 4 {
		maskedKey = cfg.YooKassa.SecretKey[:4] + "..." + cfg.YooKassa.SecretKey[len(cfg.YooKassa.SecretKey)-4:]
	}
	log.Info("loaded yookassa config",
		slog.String("shop_id", cfg.YooKassa.ShopID),
		slog.String("secret_key_masked", maskedKey))

	// Connect to PostgreSQL
	pool, err := pgxpool.New(context.Background(), cfg.Postgres.DSN())
	if err != nil {
		log.Error("failed to connect to postgres", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	log.Info("connected to postgres")

	// Initialize repository
	repo := repository.NewPaymentRepository(pool)

	// Initialize YooKassa provider
	yooProvider := provider.NewYooKassaProvider(cfg, log)

	// Initialize Kafka producer
	producer := kafka.NewProducer(cfg.Kafka.Brokers, "payments", log)
	defer producer.Close()

	// Initialize service
	paymentService := service.NewPaymentService(repo, yooProvider, producer, log)

	// Initialize Kafka consumer for order events
	consumer := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.Topic, "payment-service-group", paymentService, log)

	// Start consumer in goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Error("kafka consumer error", slog.String("error", err.Error()))
		}
	}()

	// Initialize HTTP server
	router := gin.Default()
	handler := api.NewHandler(paymentService, log)
	handler.RegisterRoutes(router)

	// Start HTTP server in goroutine
	go func() {
		addr := fmt.Sprintf(":%d", cfg.HTTP.Port)
		log.Info("http server started", slog.String("addr", addr))
		if err := router.Run(addr); err != nil {
			log.Error("http server error", slog.String("error", err.Error()))
		}
	}()

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	log.Info("shutting down server...")
	cancel()
	consumer.Close()
}
