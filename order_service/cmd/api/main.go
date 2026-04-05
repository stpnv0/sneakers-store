package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	pb "github.com/stpnv0/protos/gen/go/order"

	"order_service/internal/api"
	"order_service/internal/config"
	grpcserver "order_service/internal/grpc"
	orderhandler "order_service/internal/grpc/order"
	"order_service/internal/kafka"
	"order_service/internal/provider"
	"order_service/internal/repository"
	"order_service/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	log := setupLogger("order_service")

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// ---- Инфраструктура ----

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.Postgres.DSN())
	if err != nil {
		return err
	}
	defer pool.Close()
	log.Info("connected to postgres")

	// ---- Зависимости ----

	orderRepo := repository.NewOrderRepository(pool)
	paymentRepo := repository.NewPaymentRepository(pool)

	yooProvider := provider.NewYooKassaProvider(
		cfg.YooKassa.ShopID,
		cfg.YooKassa.SecretKey,
		cfg.YooKassa.ReturnURL,
		cfg.YooKassa.NotificationURL,
		cfg.HTTP.Timeout,
		log,
	)

	// Авто-регистрация вебхуков в ЮKassa (best-effort, ошибки логируются).
	yooProvider.RegisterWebhooks(ctx)

	producer := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic, log)
	orderService := service.NewOrderService(orderRepo, paymentRepo, yooProvider, producer, log)

	// ---- gRPC-сервер ----

	grpcSrv := grpcserver.NewServer(log)
	pb.RegisterOrderServiceServer(grpcSrv.GRPCServer(), orderhandler.NewHandler(orderService, log))

	// ---- HTTP-сервер (вебхуки YooKassa) ----

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	adminAPIKey := os.Getenv("ADMIN_API_KEY")
	webhookHandler := api.NewWebhookHandler(orderService, log, adminAPIKey)
	webhookHandler.RegisterRoutes(router)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler: router,
	}

	// ---- Запуск серверов ----

	errCh := make(chan error, 2)

	go func() {
		if err := grpcSrv.Run(cfg.GRPC.Port); err != nil {
			errCh <- fmt.Errorf("grpc server: %w", err)
		}
	}()

	go func() {
		log.Info("http server started", slog.String("addr", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("http server: %w", err)
		}
	}()

	// ---- Ожидание сигнала завершения ----

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-errCh:
		log.Error("component failed, initiating shutdown", slog.String("error", err.Error()))
		stop()
	}

	// ---- Плавное завершение с таймаутом ----

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Shutdown.Timeout)
	defer shutdownCancel()

	log.Info("shutting down grpc server")
	grpcSrv.Stop()

	log.Info("shutting down http server")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("http server shutdown error", slog.String("error", err.Error()))
	}

	log.Info("closing kafka producer")
	if err := producer.Close(); err != nil {
		log.Error("kafka producer close error", slog.String("error", err.Error()))
	}

	// pool.Close() вызывается через defer выше.

	log.Info("shutdown complete")
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
