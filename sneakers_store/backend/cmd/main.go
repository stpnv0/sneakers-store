package main

import (
	"context"
	"log/slog"
	"os"
	"sneakers-store/db"
	"sneakers-store/internal/auth"
	// clients "sneakers-store/internal/clients/favourites"
	ssogrpc "sneakers-store/internal/clients/sso/grpc"
	"sneakers-store/internal/config"
	// fav "sneakers-store/internal/favourites"
	"sneakers-store/internal/logger/handlers/slogpretty"
	"sneakers-store/internal/logger/sl"
	"sneakers-store/internal/sneakers"
	"sneakers-store/router"
)

func main() {
	cfg := config.MustLoad()

	log := setupPrettySlog()
	log.Info("starting sneakers-store", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	database, err := db.ConnectDB(cfg)
	if err != nil {
		log.Error("[ERROR] can't connect to database", sl.Err(err))
	}
	defer database.Close()

	ssoClient, err := ssogrpc.New(
		context.Background(),
		log,
		cfg.Clients.SSO.Address,
		cfg.Clients.SSO.Timeout,
		cfg.Clients.SSO.RetriesCount,
	)
	if err != nil {
		log.Error("failed to init sso client", sl.Err(err))
		os.Exit(1)
	}

	//Инициализация слоев
	repo := sneakers.NewRepository(database)       // Репозиторий
	service := sneakers.NewService(repo)           // Сервис
	sneakerHandler := sneakers.NewHandler(service) // Хендлер

	// Инициализация обработчика auth
	authHandler := auth.New(ssoClient, log)
	// favClient := clients.NewFavClient(cfg.Clients.Favourites.Address, time.Second*5)
	// favHandler := fav.NewHandler(favClient)
	router.InitRouter(sneakerHandler, authHandler, cfg)
	// Настройка роутера и запуск сервера
	if err := router.Start("0.0.0.0:8080"); err != nil {
		log.Error("[ERROR] failed to start server", sl.Err(err))
	}
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
