package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"payment/internal/config"
	handlers "payment/internal/handler"
	"payment/internal/infrastructure/kafka"
	"payment/internal/infrastructure/yookassa"
	"payment/internal/repository"
	"payment/internal/router"
	services "payment/internal/service"
	"strings"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("ERROR: Failed to initialize logger: %v", err)
	}

	defer logger.Sync()

	cfg, err := config.LoadConfig(".")
	if err != nil {
		logger.Fatal("ERROR: Failde to load config: %v", zap.Error(err))
	}

	db, err := sqlx.Connect("postgres", cfg.DBHost)
	if err != nil {
		logger.Fatal("ERROR: Failed to connect to db: %v", zap.Error(err))
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxIdleTime(5 * time.Minute)

	paymentRepo := repository.NewPostgresRepository(db)

	yookassaClient := yookassa.NewClient(cfg.YooKassaShopID, cfg.YooKassaSecretKey, logger)

	kafkaBrokers := strings.Split(cfg.KafkaBrokers, ",")
	kafkaProducer := kafka.NewProducer(kafkaBrokers, logger)
	defer kafkaProducer.Close()

	paymentService := services.NewPaymentService(paymentRepo, yookassaClient, kafkaProducer, logger)

	paymentHandler := handlers.NewPaymentHandler(paymentService, logger)
	webhookHandler := handlers.NewWebhookHandler(paymentService, logger)

	r := router.SetupRouter(paymentHandler, webhookHandler, logger)

	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("INFO: Starting payment service", zap.String("port", cfg.ServerPort))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("ERROR: Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("INFO: Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("ERROR: Server forced to shutdown", zap.Error(err))
	}

	logger.Info("INFO: Server exiting")
}
