package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Env        string        `yaml:"env"`
	AppSecret  string        `yaml:"app_secret"`
	HTTPServer HTTPServer    `yaml:"http_server"`
	Redis      RedisConfig   `yaml:"redis"`
	Clients    ClientsConfig `yaml:"clients"`
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

type ClientsConfig struct {
	MainService MainServiceConfig `yaml:"main_service"`
}

type MainServiceConfig struct {
	URL        string `yaml:"url"`
	Timeout    string `yaml:"timeout"`
	RetryCount int    `yaml:"retry_count"`
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
