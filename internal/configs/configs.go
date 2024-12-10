package configs

import (
	"log"

	"github.com/spf13/viper"
)

// Config - структура конфигурации
type Config struct {
	Token  string // Telegram Bot Token
	DBPath string // Path to the SQLite database file
}

func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Не удалось загрузить конфигурацию из файла: %v", err)
	}

	viper.SetDefault("token", "")
	viper.SetDefault("db_path", "database.db")

	// Отладочный вывод для проверки токена
	log.Printf("Загруженный токен: %s", viper.GetString("token"))

	return &Config{
		Token:  viper.GetString("token"),
		DBPath: viper.GetString("db_path"),
	}, nil
}
