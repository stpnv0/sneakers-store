package main

import (
	"context"
	"log/slog"
	"order_service/internal/config"
	grpcServer "order_service/internal/grpc"
	"order_service/internal/grpc/order"
	"order_service/internal/kafka"
	"order_service/internal/repository"
	"order_service/internal/service"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/stpnv0/protos/gen/go/order"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg := config.MustLoad()

	pool, err := pgxpool.New(context.Background(), cfg.Postgres.DSN())
	if err != nil {
		log.Error("failed to connect to postgres", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	log.Info("connected to postgres")

	repo := repository.NewOrderRepository(pool)

	producer := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic, log)
	defer producer.Close()

	orderService := service.NewOrderService(repo, producer, log)

	consumer := kafka.NewConsumer(cfg.Kafka.Brokers, "payments", "order-service-group", orderService, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Error("kafka consumer error", slog.String("error", err.Error()))
		}
	}()

	grpcSrv := grpcServer.NewServer(log)
	orderHandler := order.NewHandler(orderService, log)
	pb.RegisterOrderServiceServer(grpcSrv.GetGRPCServer(), orderHandler)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop
		log.Info("shutting down server...")
		cancel()
		grpcSrv.Stop()
		consumer.Close()
	}()

	if err := grpcSrv.Run(cfg.GRPC.Port); err != nil {
		log.Error("failed to start server", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
