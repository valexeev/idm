package common

import (
	"os"

	"github.com/joho/godotenv"
)

// Config общая конфигурация всего приложения
type Config struct {
	DbDriverName string `validate:"required"`
	Dsn          string `validate:"required"`
}

// GetConfig получение конфигурации из .env файла или переменных окружения
func GetConfig(envFile string) (Config, error) {
	err := godotenv.Load(envFile)
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		DbDriverName: os.Getenv("DB_DRIVER_NAME"),
		Dsn:          os.Getenv("DB_DSN"),
	}
	return cfg, nil
}
