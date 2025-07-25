package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string        `yaml:"env" env:"ENV" env-default:"local"`
	GRPC     GRPCConfig    `yaml:"grpc"`
	DB       DBConfig      `yaml:"database"`
	S3       S3Config      `yaml:"s3"`
	Redis    RedisConfig   `yaml:"redis"`
	CacheTTL time.Duration `yaml:"cache_ttl" env:"CACHE_TTL" env-default:"10m"`
}

type RedisConfig struct {
	Addr string `yaml:"addr" env:"REDIS_ADDR"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port" env:"GRPC_PORT" env-default:"44045"`
	Timeout time.Duration `yaml:"timeout" env:"GRPC_TIMEOUT" env-default:"10s"`
}

type DBConfig struct {
	Host     string `yaml:"host" env:"DB_HOST"`
	Port     int    `yaml:"port" env:"DB_PORT"`
	User     string `yaml:"user" env:"DB_USER"`
	Password string `yaml:"password" env:"DB_PASSWORD"`
	DBName   string `yaml:"dbname" env:"DB_NAME"`
}

type S3Config struct {
	Endpoint  string `yaml:"endpoint" env:"S3_ENDPOINT"`
	AccessKey string `yaml:"access_key" env:"S3_ACCESS_KEY"`
	SecretKey string `yaml:"secret_key" env:"S3_SECRET_KEY"`
	Bucket    string `yaml:"bucket" env:"S3_BUCKET"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/local.yaml"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
