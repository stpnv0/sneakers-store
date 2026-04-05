package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api_gateway/internal/client/cart"
	"api_gateway/internal/client/favourites"
	"api_gateway/internal/client/order"
	"api_gateway/internal/client/product"
	"api_gateway/internal/client/sso"
	"api_gateway/internal/config"
	cart_handler "api_gateway/internal/handler/cart"
	fav_handler "api_gateway/internal/handler/favourites"
	order_handler "api_gateway/internal/handler/order"
	product_handler "api_gateway/internal/handler/product"
	auth_handler "api_gateway/internal/handler/sso"
	"api_gateway/internal/router"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	log := setupLogger("api_gateway")

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	const (
		dialTimeout = 2 * time.Second
		retries     = 3
	)

	productClient, err := product.New(ctx, log, cfg.Downstream.ProductgRPC, dialTimeout, retries)
	if err != nil {
		return err
	}
	defer productClient.Close()

	ssoClient, err := sso.New(ctx, log, cfg.Downstream.SSOgRPC, dialTimeout, retries)
	if err != nil {
		return err
	}
	defer ssoClient.Close()

	cartClient, err := cart.New(ctx, log, cfg.Downstream.CartgRPC, dialTimeout, retries)
	if err != nil {
		return err
	}
	defer cartClient.Close()

	favClient, err := favourites.New(ctx, log, cfg.Downstream.FavouritesgRPC, dialTimeout, retries)
	if err != nil {
		return err
	}
	defer favClient.Close()

	orderClient, err := order.New(ctx, log, cfg.Downstream.OrdergRPC, dialTimeout, retries)
	if err != nil {
		return err
	}
	defer orderClient.Close()

	// Создаём хендлеры (каждый принимает интерфейс, реализуемый конкретным клиентом).
	handlers := router.Handlers{
		Product:    product_handler.NewHandler(productClient, log),
		Auth:       auth_handler.NewHandler(ssoClient, log),
		Cart:       cart_handler.NewHandler(cartClient, log),
		Favourites: fav_handler.NewHandler(favClient, log),
		Order:      order_handler.New(orderClient, productClient, cartClient, log),
	}

	engine := router.New(cfg.AppSecret, log, handlers, ssoClient)

	httpServer := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: engine,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("api gateway starting", slog.String("addr", cfg.ListenAddr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-errCh:
		log.Error("http server failed", slog.String("error", err.Error()))
		stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Info("shutting down http server")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("http server shutdown error", slog.String("error", err.Error()))
	}

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
