package main

import (
	"api_gateway/internal/client/product"
	"api_gateway/internal/client/sso"
	"api_gateway/internal/config"
	"api_gateway/internal/router"
	"context"
	"log/slog"
	"os"
	"time"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}
	cfg := config.Load(configPath)

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	productClient, err := product.New(context.Background(), log, cfg.Downstream.ProductgRPC, 2*time.Second, 3)
	if err != nil {
		log.Error("failed to connect to product service", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ssoClient, err := sso.New(context.Background(), log, cfg.Downstream.SSOgRPC, 2*time.Second, 3)
	if err != nil {
		log.Error("failed to connect to sso service", slog.String("error", err.Error()))
		os.Exit(1)
	}

	r := router.New(cfg, log, productClient, ssoClient)

	log.Info("api_gateway is starting on %s", cfg.ListenAddr)
	if err := r.Run(cfg.ListenAddr); err != nil {
		log.Error("Failed to start server: %v", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
