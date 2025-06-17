package main

import (
	"idm/inner/common"
	"idm/inner/database"

	"go.uber.org/zap"
)

func main() {
	// Читаем конфиг
	cfg := common.GetConfig(".env")

	// Создаём логгер
	logger := common.NewLogger(cfg)
	defer func() { _ = logger.Sync() }()

	logger.Info("Hello, Go.")

	// Подключаемся к БД
	db := database.ConnectDbWithCfg(cfg)
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("error closing db", zap.Error(err))
		}
	}()

	logger.Info("DB connected")
}
