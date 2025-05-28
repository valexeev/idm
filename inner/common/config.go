package common

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DbDriverName string
	Dsn          string
}

func GetConfig(envFile string) (Config, error) {
	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			return Config{}, err
		}
	}

	return Config{
		DbDriverName: os.Getenv("DB_DRIVER_NAME"),
		Dsn:          os.Getenv("DSN"),
	}, nil
}
