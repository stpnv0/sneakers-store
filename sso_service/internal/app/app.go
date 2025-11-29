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
}

func New(
	log *slog.Logger,
	grpcPort int,
	dbHost string,
	dbPort int,
	dbUser string,
	dbPassword string,
	dbName string,
	tokenTTL time.Duration,
) *App {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	storage, err := postgres.New(context.Background(), connString)
	if err != nil {
		panic(err)
	}

	// Sync APP_SECRET from environment to database
	appSecret := os.Getenv("APP_SECRET")
	if appSecret == "" {
		log.Warn("APP_SECRET not set, using placeholder from database")
	} else {
		err = storage.UpdateAppSecret(context.Background(), 1, appSecret)
		if err != nil {
			log.Error("failed to update app secret", slog.String("error", err.Error()))
			// Don't panic - continue with existing secret
		} else {
			log.Info("app secret synchronized from environment variable")
		}
	}

	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
