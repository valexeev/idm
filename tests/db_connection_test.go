package idm_test

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	_ = godotenv.Load("../.env")
	os.Exit(m.Run())
}

func TestDBConnectionFailsWithBadConfig(t *testing.T) {

	dsn := "host=wronghost port=5432 user=wronguser password=wrongpass dbname=wrongdb sslmode=disable"

	db, err := sqlx.Open("postgres", dsn)
	assert.NoError(t, err)

	err = db.Ping()
	assert.Error(t, err, "Ожидается ошибка подключения с неверным конфигом")
}

func TestDBConnectionWorksWithValidConfig(t *testing.T) {
	dsn := os.Getenv("DB_DSN")

	db, err := sqlx.Open("postgres", dsn)
	assert.NoError(t, err, "не удалось открыть соединение с БД")

	err = db.Ping()
	assert.NoError(t, err, "не удалось подключиться к БД — проверь конфиг, хост, порт, логин, пароль")
}
