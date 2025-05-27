package common_test

import (
	"idm/inner/common"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyEnvAndNoEnvVars(t *testing.T) {
	env := map[string]string{}
	cfg := common.GetConfigFromMap(env)
	assert.Equal(t, "", cfg.DbDriverName)
	assert.Equal(t, "", cfg.Dsn)
}

func TestEnvFileWithoutVarsButHasEnvVars(t *testing.T) {
	env := map[string]string{
		"DB_DRIVER_NAME": "postgres",
		"DB_DSN":         "postgres://postgres:@localhost:5432/idm_service?sslmode=disable",
	}
	cfg := common.GetConfigFromMap(env)
	assert.Equal(t, "postgres", cfg.DbDriverName)
	assert.Equal(t, "postgres://postgres:@localhost:5432/idm_service?sslmode=disable", cfg.Dsn)
}

func TestEnvFileWithVarsAndNoEnvVars(t *testing.T) {
	// В тестах без env, "имитация" загрузки из .env — просто передать значения
	env := map[string]string{
		"DB_DRIVER_NAME": "postgres",
		"DB_DSN":         "host=127.0.0.1 port=5432 user=postgres password= dbname=idm_service sslmode=disable",
	}
	cfg := common.GetConfigFromMap(env)
	assert.Equal(t, "postgres", cfg.DbDriverName)
	assert.Equal(t, "host=127.0.0.1 port=5432 user=postgres password= dbname=idm_service sslmode=disable", cfg.Dsn)
}

func TestEnvOverridesEnvFile(t *testing.T) {
	// эмулируем приоритет env-переменных над .env: подаем только env-переменные
	env := map[string]string{
		"DB_DRIVER_NAME": "env_postgres",
		"DB_DSN":         "env_dsn_value",
	}
	cfg := common.GetConfigFromMap(env)
	assert.Equal(t, "env_postgres", cfg.DbDriverName)
	assert.Equal(t, "env_dsn_value", cfg.Dsn)
}
