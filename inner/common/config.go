package common

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2/log"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// Config общая конфигурация всего приложения
type Config struct {
	DbDriverName   string `validate:"required"`
	Dsn            string `validate:"required"`
	AppName        string `validate:"required"`
	AppVersion     string `validate:"required"`
	SslSert        string `validate:"required"`
	SslKey         string `validate:"required"`
	LogLevel       string
	LogDevelopMode bool
}

// GetConfig получение конфигурации из .env файла или переменных окружения
func GetConfig(envFile string) Config {
	var err = godotenv.Load(envFile)
	// если нет файла, то залогируем это и попробуем получить конфиг из переменных окружения
	if err != nil {
		log.Info("Error loading .env file: %v\n", zap.Error(err))
	}
	var cfg = Config{
		DbDriverName:   os.Getenv("DB_DRIVER_NAME"),
		Dsn:            os.Getenv("DB_DSN"),
		AppName:        os.Getenv("APP_NAME"),
		AppVersion:     os.Getenv("APP_VERSION"),
		SslSert:        os.Getenv("SSL_SERT"),
		SslKey:         os.Getenv("SSL_KEY"),
		LogLevel:       os.Getenv("LOG_LEVEL"),
		LogDevelopMode: os.Getenv("LOG_DEVELOP_MODE") == "true",
	}
	err = validator.New().Struct(cfg)
	if err != nil {
		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			// если конфиг не прошел валидацию, то паникуем
			panic(fmt.Sprintf("config validation error: %v", err))
		}
	}
	return cfg
}
