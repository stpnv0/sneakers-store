package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"product_service/internal/app"
	"product_service/internal/config"
	grpc_handler "product_service/internal/grpc/product"
	"product_service/internal/repository"
	migration "product_service/internal/storage"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.MustLoad()
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})).With(slog.String("service", "product"))
	log.Info("starting service", slog.String("env", cfg.Env))

	migration.RunMigrations(cfg)

	// Инициализация зависимостей
	dbPool := mustInitDB(cfg)
	log.Info("migrations applied successfully")
	redisClient := mustInitRedis(cfg)
	fileStoreRepo, err := repository.NewFileStoreRepository(context.Background(), cfg, log)
	if err != nil {
		log.Error("failed to init file store", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Инициализация репозиториев
	postgresRepo := repository.New(dbPool)
	redisRepo := repository.NewRedisRepository(redisClient)

	// Инициализация сервисного слоя
	productService := app.NewService(postgresRepo, redisRepo, fileStoreRepo, log, cfg.CacheTTL)

	// Инициализация и запуск gRPC сервера
	grpcServer := grpc.NewServer()
	grpc_handler.Register(grpcServer, productService, log)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Error("failed to listen", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log.Info("gRPC server started", slog.Int("port", cfg.GRPC.Port))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("gRPC server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	log.Info("shutting down gRPC server")
	grpcServer.GracefulStop()
	log.Info("server stopped")
}

func mustInitDB(cfg *config.Config) *pgxpool.Pool {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.DBName)
	db, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		panic("failed to connect to db: " + err.Error())
	}
	if err := db.Ping(context.Background()); err != nil {
		panic("failed to ping db: " + err.Error())
	}
	return db
}

func mustInitRedis(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		panic("failed to connect to redis: " + err.Error())
	}
	return client
}
