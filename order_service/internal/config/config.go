package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config содержит всю конфигурацию приложения.
type Config struct {
	Env      string         `yaml:"env"`
	GRPC     GRPCConfig     `yaml:"grpc"`
	HTTP     HTTPConfig     `yaml:"http"`
	Postgres PostgresConfig `yaml:"postgres"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	YooKassa YooKassaConfig `yaml:"yookassa"`
	Shutdown ShutdownConfig `yaml:"shutdown"`
}

// GRPCConfig содержит настройки gRPC-сервера.
type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

// HTTPConfig содержит настройки HTTP-сервера (вебхуки YooKassa).
type HTTPConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

// PostgresConfig содержит параметры подключения к PostgreSQL.
type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

// KafkaConfig содержит настройки брокера Kafka.
type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
}

// YooKassaConfig содержит учётные данные платёжного шлюза.
type YooKassaConfig struct {
	ShopID          string `yaml:"shop_id"`
	SecretKey       string `yaml:"secret_key"`
	ReturnURL       string `yaml:"return_url"`
	NotificationURL string `yaml:"notification_url"`
}

// ShutdownConfig управляет поведением graceful shutdown.
type ShutdownConfig struct {
	Timeout time.Duration `yaml:"timeout"`
}

// DSN возвращает строку подключения к PostgreSQL.
func (p *PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.DBName, p.SSLMode,
	)
}

// Load читает конфигурацию из файла, указанного в CONFIG_PATH (или
// пути по умолчанию). В отличие от MustLoad, не вызывает panic.
func Load() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("config: read file %s: %w", configPath, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse yaml: %w", err)
	}

	if cfg.HTTP.Timeout == 0 {
		cfg.HTTP.Timeout = 10 * time.Second
	}
	if cfg.Shutdown.Timeout == 0 {
		cfg.Shutdown.Timeout = 15 * time.Second
	}

	return &cfg, nil
}
