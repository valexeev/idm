package idm_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	_ = godotenv.Load("../.env") // ../ — из папки tests выйти в корень проекта
	os.Exit(m.Run())
}

func TestDBConnectionFailsWithBadConfig(t *testing.T) {
	// Конфиг с неправильными параметрами
	dsn := "host=wronghost port=5432 user=wronguser password=wrongpass dbname=wrongdb sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	assert.NoError(t, err)

	err = db.Ping() // Пингуем базу - пытаемся реально подключиться
	assert.Error(t, err, "Ожидается ошибка подключения с неверным конфигом")
}

func TestDBConnectionWorksWithValidConfig(t *testing.T) {
	dsn := os.Getenv("DB_DSN") // берём корректный DSN из переменной окружения

	db, err := sql.Open("postgres", dsn)
	assert.NoError(t, err, "не удалось открыть соединение с БД")

	err = db.Ping()
	assert.NoError(t, err, "не удалось подключиться к БД — проверь конфиг, хост, порт, логин, пароль")
}
