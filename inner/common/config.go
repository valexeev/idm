package common

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DbDriverName string
	Dsn          string
}

// GetConfig читает .env файл (если указан), подгружает переменные окружения
// и возвращает конфигурацию.
// Переменные окружения имеют приоритет над .env файлом.
func GetConfig(envFile string) (Config, error) {
	// Загружаем .env, если файл указан
	if envFile != "" {
		_ = godotenv.Load(envFile)
	}

	return Config{
		DbDriverName: os.Getenv("DB_DRIVER_NAME"),
		Dsn:          os.Getenv("DB_DSN"),
	}, nil
}

// GetConfigFromMap — вспомогательная функция для тестов,
// принимает map с ключами и возвращает Config,
// чтобы создавать конфиги без env и файлов.
func GetConfigFromMap(m map[string]string) Config {
	return Config{
		DbDriverName: m["DB_DRIVER_NAME"],
		Dsn:          m["DB_DSN"],
	}
}
