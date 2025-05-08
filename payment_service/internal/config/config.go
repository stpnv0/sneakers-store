// internal/config/config.go
package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DBHost            string `mapstructure:"DB_HOST"`
	DBPort            string `mapstructure:"DB_PORT"`
	DBUser            string `mapstructure:"DB_USER"`
	DBPassword        string `mapstructure:"DB_PASSWORD"`
	DBName            string `mapstructure:"DB_NAME"`
	ServerPort        string `mapstructure:"SERVER_PORT"`
	KafkaBrokers      string `mapstructure:"KAFKA_BROKERS"`
	YooKassaShopID    string `mapstructure:"YOOKASSA_SHOP_ID"`
	YooKassaSecretKey string `mapstructure:"YOOKASSA_SECRET_KEY"`
	AppBaseURL        string `mapstructure:"APP_BASE_URL"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv() // Читать также переменные окружения

	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return
		}
	}

	err = viper.Unmarshal(&config)
	return
}
