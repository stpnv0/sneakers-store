package config

import (
	"log"

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

func Load(path string) *Config {
	viper.SetConfigFile(path)

	viper.AutomaticEnv()
	viper.BindEnv("app_secret", "APP_SECRET")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("[ERROR] Can't read config %s: %v", path, err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("[ERROR] Can't unmarshal config: %v", err)
	}

	if config.AppSecret == "" {
		log.Fatal("[ERROR] APP_SECRET must be set via environment variable or config file")
	}

	return &config
}
