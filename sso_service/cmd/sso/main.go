package main

import (
	"log/slog"
	"os"
	"os/signal"
	"sso/internal/app"
	"sso/internal/config"
	"sso/internal/lib/logger/handlers/slogpretty"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting sso server", slog.Any("config", cfg))

	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)

	go application.GRPCServer.MustRun() //в горутине потому что дальше мы будем слушать сигналы из опепрационных систем

	//Graceful shutdown
	stop := make(chan os.Signal, 1)                      // перед завершением работы: операционная система отдает сигнал программе
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT) //и мы его получаем. программа сама занимается своим завершением

	sign := <-stop // записываем это в канал. получается в отдельной горутине работаем пока не придет один из сигналов от ос

	log.Info("stopping application", slog.String("signal", sign.String()))

	application.GRPCServer.Stop()

	log.Info("Application stopped")
	//подобные компоненты, например при добавлении бд (пулл соединений) надо тоже оборачивать в подобные приложения с методами Run, Stop
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
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

// go run cmd/sso/main.go --config=./config/local.yaml
