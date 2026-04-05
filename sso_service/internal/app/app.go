package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/services/auth"
	"sso/internal/storage/postgres"
	"time"
)

type App struct {
	GRPCServer *grpcapp.App
	storage    *postgres.Storage
}

// New creates a new App instance. Returns an error instead of panicking.
func New(
	log *slog.Logger,
	grpcPort int,
	dbHost string,
	dbPort int,
	dbUser string,
	dbPassword string,
	dbName string,
	tokenTTL time.Duration,
) (*App, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	storage, err := postgres.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("init storage: %w", err)
	}

	appSecret := os.Getenv("APP_SECRET")
	if appSecret == "" {
		log.Warn("APP_SECRET not set, using placeholder from database")
	} else {
		if err := storage.UpdateAppSecret(context.Background(), 1, appSecret); err != nil {
			log.Error("failed to update app secret", slog.String("error", err.Error()))
			// Don't fail — continue with existing secret.
		} else {
			log.Info("app secret synchronized from environment variable")
		}
	}

	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
		storage:    storage,
	}, nil
}

// Close releases all resources held by the application.
func (a *App) Close() {
	if a.storage != nil {
		a.storage.Close()
	}
}
