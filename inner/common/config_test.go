package common_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"idm/inner/common"
)

const (
	dbDriverEnv = "DB_DRIVER_NAME"
	dsnEnv      = "DSN"
)

// helper: сброс переменных окружения
func unsetEnv() {
	os.Unsetenv(dbDriverEnv)
	os.Unsetenv(dsnEnv)
}

// helper: установка переменных окружения
func setEnv(driver, dsn string) {
	_ = os.Setenv(dbDriverEnv, driver)
	_ = os.Setenv(dsnEnv, dsn)
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
	setEnv("pgx", "postgres://user:pass@localhost/db")

	cfg, err := common.GetConfig("")
	require.NoError(t, err)
	require.Equal(t, "pgx", cfg.DbDriverName)
	require.Equal(t, "postgres://user:pass@localhost/db", cfg.Dsn)

	unsetEnv()
}

func Test_GetConfig_EmptyEverything(t *testing.T) {
	unsetEnv()
	cfg, err := common.GetConfig("")
	require.NoError(t, err)
	require.Equal(t, "", cfg.DbDriverName)
	require.Equal(t, "", cfg.Dsn)
}

func Test_GetConfig_EnvOverridesDotEnv(t *testing.T) {
	unsetEnv()

	envFile := writeTempEnvFile("DB_DRIVER_NAME=dotenv_driver\nDSN=dotenv_dsn\n")
	defer os.Remove(envFile)

	setEnv("env_driver", "env_dsn")

	cfg, err := common.GetConfig(envFile)
	require.NoError(t, err)
	require.Equal(t, "env_driver", cfg.DbDriverName)
	require.Equal(t, "env_dsn", cfg.Dsn)

	unsetEnv()
}

func Test_GetConfig_OnlyDotEnv(t *testing.T) {
	unsetEnv()

	envFile := writeTempEnvFile("DB_DRIVER_NAME=pgx\nDSN=postgres://user:pass@localhost/db\n")
	defer os.Remove(envFile)

	cfg, err := common.GetConfig(envFile)
	require.NoError(t, err)
	require.Equal(t, "pgx", cfg.DbDriverName)
	require.Equal(t, "postgres://user:pass@localhost/db", cfg.Dsn)
}

func Test_GetConfig_DotEnvMissingVars(t *testing.T) {
	unsetEnv()

	envFile := writeTempEnvFile("SOME_VAR=123\n")
	defer os.Remove(envFile)

	cfg, err := common.GetConfig(envFile)
	require.NoError(t, err)
	require.Equal(t, "", cfg.DbDriverName)
	require.Equal(t, "", cfg.Dsn)
}

func Test_GetConfig_DotEnvMissingVarsButEnvHasThem(t *testing.T) {
	unsetEnv()
	setEnv("pgx", "postgres://user:pass@localhost/db")

	envFile := writeTempEnvFile("SOME_VAR=123\n")
	defer os.Remove(envFile)

	cfg, err := common.GetConfig(envFile)
	require.NoError(t, err)
	require.Equal(t, "pgx", cfg.DbDriverName)
	require.Equal(t, "postgres://user:pass@localhost/db", cfg.Dsn)

	unsetEnv()
}
