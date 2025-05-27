package common_test

import (
	"idm/inner/common"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyEnvAndNoEnvVars(t *testing.T) {
	// гарантируем отсутствие переменных окружения
	os.Unsetenv("DB_DRIVER_NAME")
	os.Unsetenv("DB_DSN")

	// .env лежит в корне и НЕ содержит этих переменных
	cfg, err := common.GetConfig(".env.empty")
	if err != nil {
		t.Fatalf("Failed to load .env file: %v", err)
	}
	assert.Equal(t, "", cfg.DbDriverName)
	assert.Equal(t, "", cfg.Dsn)
}
func TestEnvFileWithoutVarsButHasEnvVars(t *testing.T) {
	// Устанавливаем переменные окружения
	os.Setenv("DB_DRIVER_NAME", "postgres")
	os.Setenv("DB_DSN", "postgres://postgres:@localhost:5432/idm_service?sslmode=disable")

	// Очищаем после теста
	defer os.Unsetenv("DB_DRIVER_NAME")
	defer os.Unsetenv("DB_DSN")

	// Загружаем конфиг (в .env нужных переменных нет)
	cfg, err := common.GetConfig(".env")
	if err != nil {
		t.Fatalf("Failed to load .env file: %v", err)
	}
	// Проверяем, что переменные из окружения подставились
	assert.Equal(t, "postgres", cfg.DbDriverName)
	assert.Equal(t, "postgres://postgres:@localhost:5432/idm_service?sslmode=disable", cfg.Dsn)
}
func TestEnvFileWithVarsAndNoEnvVars(t *testing.T) {
	// Удаляем переменные окружения, чтобы они не мешали
	os.Unsetenv("DB_DRIVER_NAME")
	os.Unsetenv("DB_DSN")

	// Загружаем конфиг из .env
	cfg, err := common.GetConfig(".env")
	if err != nil {
		t.Fatalf("Failed to load .env file: %v", err)
	}
	// Проверяем, что подставились значения из .env
	assert.Equal(t, "postgres", cfg.DbDriverName)
	assert.Equal(t, "host=127.0.0.1 port=5432 user=postgres password= dbname=idm_service sslmode=disable", cfg.Dsn)
}
func TestEnvOverridesEnvFile(t *testing.T) {
	// Задаём переменные окружения с "другими" значениями
	os.Setenv("DB_DRIVER_NAME", "env_postgres")
	os.Setenv("DB_DSN", "env_dsn_value")

	defer os.Unsetenv("DB_DRIVER_NAME")
	defer os.Unsetenv("DB_DSN")

	// Загружаем конфиг (godotenv.Load не должен перезаписать эти переменные)
	cfg, err := common.GetConfig(".env")
	if err != nil {
		t.Fatalf("Failed to load .env file: %v", err)
	}

	// Проверяем, что взялись значения из переменных окружения, а не из .env
	assert.Equal(t, "env_postgres", cfg.DbDriverName)
	assert.Equal(t, "env_dsn_value", cfg.Dsn)
}
