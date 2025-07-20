package main

import (
	"api_gateway/internal/config"
	"api_gateway/internal/router"
	"log"
	"os"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	cfg := config.Load(configPath)

	r := router.New(cfg)

	log.Printf("api_gateway is starting on %s", cfg.ListenAddr)
	if err := r.Run(cfg.ListenAddr); err != nil {
		log.Fatal("Failed to start server: %v", err)
	}
}
