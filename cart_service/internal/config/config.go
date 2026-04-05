package config

import (
	"fmt"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"
)

// Config содержит всю конфигурацию сервиса корзины.
type Config struct {
	Env      string         `yaml:"env"`
	GRPC     GRPCConfig     `yaml:"grpc"`
	Redis    RedisConfig    `yaml:"redis"`
	Postgres PostgresConfig `yaml:"postgres"`
}

// GRPCConfig содержит настройки gRPC-сервера.
type GRPCConfig struct {
	Port int `yaml:"port"`
}

// RedisConfig содержит параметры подключения к Redis.
type RedisConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Password   string `yaml:"password"`
	DB         int    `yaml:"db"`
	Expiration string `yaml:"expiration"`
}

// PostgresConfig содержит параметры подключения к PostgreSQL.
type PostgresConfig struct {
	Host               string `yaml:"host"`
	Port               int    `yaml:"port"`
	User               string `yaml:"user"`
	Password           string `yaml:"password"`
	DBName             string `yaml:"dbname"`
	SSLMode            string `yaml:"sslmode"`
	MaxConnections     int    `yaml:"max_connections"`
	ConnectionTimeoutS int    `yaml:"connection_timeout"`
}

// DSN возвращает строку подключения к PostgreSQL.
func (p PostgresConfig) DSN() string {
	host := p.Host
	if host == "" {
		host = "localhost"
	}
	port := p.Port
	if port == 0 {
		port = 5432
	}
	sslmode := p.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		url.QueryEscape(p.User),
		url.QueryEscape(p.Password),
		host, port, p.DBName, sslmode,
	)
}

// Load читает конфигурацию из указанного файла.
func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("config: read file %s: %w", configPath, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse yaml: %w", err)
	}

	return &cfg, nil
}
