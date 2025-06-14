package common_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"idm/inner/common"
)

const (
	dbDriverEnv   = "DB_DRIVER_NAME"
	dsnEnv        = "DB_DSN"
	appNameEnv    = "APP_NAME"
	appVersionEnv = "APP_VERSION"
)

// helper: сброс переменных окружения
func unsetEnv() {
	os.Unsetenv(dbDriverEnv)
	os.Unsetenv(dsnEnv)
	os.Unsetenv(appNameEnv)
	os.Unsetenv(appVersionEnv)
}

// helper: установка переменных окружения
func setEnv(driver, dsn, appName, appVersion string) {
	_ = os.Setenv(dbDriverEnv, driver)
	_ = os.Setenv(dsnEnv, dsn)
	_ = os.Setenv(appNameEnv, appName)
	_ = os.Setenv(appVersionEnv, appVersion)
}

// helper: создание временного .env-файла
func writeTempEnvFile(content string) string {
	tmpFile, err := os.CreateTemp("", ".env")
	if err != nil {
		panic("cannot create temp env file: " + err.Error())
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		panic("cannot write to temp env file: " + err.Error())
	}

	if err := tmpFile.Close(); err != nil {
		panic("cannot close temp env file: " + err.Error())
	}

	return tmpFile.Name()
}

func Test_GetConfig_OnlyEnvVars(t *testing.T) {
	unsetEnv()
	setEnv("pgx", "postgres://user:pass@localhost/db", "test-app", "1.0.0")

	cfg := common.GetConfig("")
	require.Equal(t, "pgx", cfg.DbDriverName)
	require.Equal(t, "postgres://user:pass@localhost/db", cfg.Dsn)
	require.Equal(t, "test-app", cfg.AppName)
	require.Equal(t, "1.0.0", cfg.AppVersion)

	unsetEnv()
}

func Test_GetConfig_EmptyEverything(t *testing.T) {
	unsetEnv()

	// Ожидаем panic из-за валидации
	require.Panics(t, func() {
		common.GetConfig("")
	})
}

func Test_GetConfig_EnvOverridesDotEnv(t *testing.T) {
	unsetEnv()
	envFile := writeTempEnvFile("DB_DRIVER_NAME=dotenv_driver\nDB_DSN=dotenv_dsn\nAPP_NAME=dotenv_app\nAPP_VERSION=dotenv_version\n")
	defer os.Remove(envFile)

	setEnv("env_driver", "env_dsn", "env_app", "env_version")

	cfg := common.GetConfig(envFile)

	require.Equal(t, "env_driver", cfg.DbDriverName)
	require.Equal(t, "env_dsn", cfg.Dsn)
	require.Equal(t, "env_app", cfg.AppName)
	require.Equal(t, "env_version", cfg.AppVersion)

	unsetEnv()
}

func Test_GetConfig_OnlyDotEnv(t *testing.T) {
	unsetEnv()
	envFile := writeTempEnvFile("DB_DRIVER_NAME=pgx\nDB_DSN=postgres://user:pass@localhost/db\nAPP_NAME=test-app\nAPP_VERSION=1.0.0\n")
	defer os.Remove(envFile)

	cfg := common.GetConfig(envFile)

	require.Equal(t, "pgx", cfg.DbDriverName)
	require.Equal(t, "postgres://user:pass@localhost/db", cfg.Dsn)
	require.Equal(t, "test-app", cfg.AppName)
	require.Equal(t, "1.0.0", cfg.AppVersion)
}

func Test_GetConfig_DotEnvMissingVars(t *testing.T) {
	unsetEnv()
	envFile := writeTempEnvFile("SOME_VAR=123\n")
	defer os.Remove(envFile)

	// Ожидаем panic из-за отсутствия обязательных переменных в .env файле
	require.Panics(t, func() {
		common.GetConfig(envFile)
	})
}

func Test_GetConfig_DotEnvMissingVarsButEnvHasThem(t *testing.T) {
	unsetEnv()
	setEnv("pgx", "postgres://user:pass@localhost/db", "test-app", "1.0.0") // Добавлены APP_NAME и APP_VERSION
	envFile := writeTempEnvFile("SOME_VAR=123\n")
	defer os.Remove(envFile)

	cfg := common.GetConfig(envFile)

	require.Equal(t, "pgx", cfg.DbDriverName)
	require.Equal(t, "postgres://user:pass@localhost/db", cfg.Dsn)
	require.Equal(t, "test-app", cfg.AppName)
	require.Equal(t, "1.0.0", cfg.AppVersion)

	unsetEnv()
}
