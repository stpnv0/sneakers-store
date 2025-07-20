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
	Cart       string `mapstructure:"cart"`
	Favourites string `mapstructure:"favourites"`
	Backend    string `mapstructure:"backend"`
}

func Load(path string) *Config {
	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("[ERROR] Can't read config %s: %v", path, err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("[ERROR] Can't unmarshal config: %v", err)
	}

	return &config
}
