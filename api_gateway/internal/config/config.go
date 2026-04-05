package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	ListenAddr string           `mapstructure:"listen"`
	AppSecret  string           `mapstructure:"app_secret"`
	Downstream DownstreamConfig `mapstructure:"downstream"`
}

type DownstreamConfig struct {
	Cart           string `mapstructure:"cart"`
	CartgRPC       string `mapstructure:"cart_grpc"`
	Favourites     string `mapstructure:"favourites"`
	FavouritesgRPC string `mapstructure:"favourites_grpc"`
	Backend        string `mapstructure:"backend"`
	ProductgRPC    string `mapstructure:"product_grpc"`
	SSOgRPC        string `mapstructure:"sso_grpc"`
	OrdergRPC      string `mapstructure:"order_grpc"`
}

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.BindEnv("app_secret", "APP_SECRET"); err != nil {
		return nil, fmt.Errorf("config: bind env APP_SECRET: %w", err)
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config: read file %s: %w", path, err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshal: %w", err)
	}

	if cfg.AppSecret == "" {
		return nil, fmt.Errorf("config: APP_SECRET must be set via environment variable or config file")
	}

	return &cfg, nil
}
