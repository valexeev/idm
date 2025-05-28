package database

import (
	"idm/inner/common"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Временная переменная, которая будет ссылаться на подключение к базе данных. Позже мы от неё избавимся
var DB *sqlx.DB

// ConnectDb получить конфиг и подключиться с ним к базе данных
func ConnectDb() *sqlx.DB {
	cfg, err := common.GetConfig(".env")
	if err != nil {
		panic("Ошибка при чтении конфига: " + err.Error())
	}
	return ConnectDbWithCfg(cfg)
}

// ConnectDbWithCfg подключиться к базе данных с переданным конфигом
func ConnectDbWithCfg(cfg common.Config) *sqlx.DB {
	DB = sqlx.MustConnect(cfg.DbDriverName, cfg.Dsn)
	// Настройки ниже конфигурируют пулл подключений к базе данных. Их названия стандартны для большинства библиотек.
	// Ознакомиться с их описанием можно на примере документации Hikari pool:
	// https://github.com/brettwooldridge/HikariCP?tab=readme-ov-file#gear-configuration-knobs-baby
	DB.SetMaxIdleConns(5)
	DB.SetMaxOpenConns(20)
	DB.SetConnMaxLifetime(1 * time.Minute)
	DB.SetConnMaxIdleTime(10 * time.Minute)
	return DB
}
