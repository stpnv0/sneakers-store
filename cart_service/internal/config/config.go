package config

import (
	"fmt"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Env        string         `yaml:"env"`
	HTTPServer HTTPServer     `yaml:"http_server"`
	Redis      RedisConfig    `yaml:"redis"`
	Postgres   PostgresConfig `yaml:"postgres"`
}

type HTTPServer struct {
	Address     string   `yaml:"address"`
	Timeout     string   `yaml:"timeout"`
	IdleTimeout string   `yaml:"idle_timeout"`
	CorsAllowed []string `yaml:"cors_allowed"`
}

type RedisConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Password   string `yaml:"password"`
	DB         int    `yaml:"db"`
	Expiration string `yaml:"expiration"`
}

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

// Load загружает конфигурацию из файла
func Load(configPath string) (*Config, error) {
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err = yaml.Unmarshal(configFile, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// MustLoad загружает конфигурацию из файла или паникует при ошибке
func MustLoad() *Config {
	cfg, err := Load("config/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}
