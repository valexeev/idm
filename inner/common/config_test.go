package common_test

import (
	"errors"
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"idm/inner/common"
)

const (
	dbDriverEnv   = "DB_DRIVER_NAME"
	dsnEnv        = "DB_DSN"
	appNameEnv    = "APP_NAME"
	appVersionEnv = "APP_VERSION"
	sslCertEnv    = "SSL_SERT"
	sslKeyEnv     = "SSL_KEY"
)

// helper: сброс переменных окружения
func unsetEnv() {
	os.Unsetenv(dbDriverEnv)
	os.Unsetenv(dsnEnv)
	os.Unsetenv(appNameEnv)
	os.Unsetenv(appVersionEnv)
	os.Unsetenv(sslCertEnv)
	os.Unsetenv(sslKeyEnv)
}

// helper: установка переменных окружения
func setEnv(driver, dsn, appName, appVersion, sslCert, sslKey string) {
	_ = os.Setenv(dbDriverEnv, driver)
	_ = os.Setenv(dsnEnv, dsn)
	_ = os.Setenv(appNameEnv, appName)
	_ = os.Setenv(appVersionEnv, appVersion)
	_ = os.Setenv(sslCertEnv, sslCert)
	_ = os.Setenv(sslKeyEnv, sslKey)
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

func TestMain(m *testing.M) {
	_ = os.Setenv("KEYCLOAK_JWK_URL", "http://localhost/jwk")
	code := m.Run()
	os.Unsetenv("KEYCLOAK_JWK_URL")
	os.Exit(code)
}

func Test_GetConfig_OnlyEnvVars(t *testing.T) {
	os.Setenv("KEYCLOAK_JWK_URL", "http://localhost:9990/realms/idm/protocol/openid-connect/certs")
	defer os.Unsetenv("KEYCLOAK_JWK_URL")
	unsetEnv()
	setEnv("pgx", "postgres://user:pass@localhost/db", "test-app", "1.0.0", "test-cert", "test-key")

	cfg := common.GetConfig("")
	require.Equal(t, "pgx", cfg.DbDriverName)
	require.Equal(t, "postgres://user:pass@localhost/db", cfg.Dsn)
	require.Equal(t, "test-app", cfg.AppName)
	require.Equal(t, "1.0.0", cfg.AppVersion)
	require.Equal(t, "test-cert", cfg.SslSert)
	require.Equal(t, "test-key", cfg.SslKey)

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
	_ = os.Setenv("KEYCLOAK_JWK_URL", "http://localhost/jwk")
	envFile := writeTempEnvFile("DB_DRIVER_NAME=dotenv_driver\nDB_DSN=dotenv_dsn\nAPP_NAME=dotenv_app\nAPP_VERSION=dotenv_version\nSSL_SERT=dotenv_cert\nSSL_KEY=dotenv_key\n")
	defer os.Remove(envFile)

	setEnv("env_driver", "env_dsn", "env_app", "env_version", "env_cert", "env_key")

	cfg := common.GetConfig(envFile)

	require.Equal(t, "env_driver", cfg.DbDriverName)
	require.Equal(t, "env_dsn", cfg.Dsn)
	require.Equal(t, "env_app", cfg.AppName)
	require.Equal(t, "env_version", cfg.AppVersion)
	require.Equal(t, "env_cert", cfg.SslSert)
	require.Equal(t, "env_key", cfg.SslKey)

	unsetEnv()
	os.Unsetenv("KEYCLOAK_JWK_URL")
}

func Test_GetConfig_OnlyDotEnv(t *testing.T) {
	unsetEnv()
	_ = os.Setenv("KEYCLOAK_JWK_URL", "http://localhost/jwk")
	envFile := writeTempEnvFile("DB_DRIVER_NAME=pgx\nDB_DSN=postgres://user:pass@localhost/db\nAPP_NAME=test-app\nAPP_VERSION=1.0.0\nSSL_SERT=cert\nSSL_KEY=key\n")
	defer os.Remove(envFile)

	cfg := common.GetConfig(envFile)

	require.Equal(t, "pgx", cfg.DbDriverName)
	require.Equal(t, "postgres://user:pass@localhost/db", cfg.Dsn)
	require.Equal(t, "test-app", cfg.AppName)
	require.Equal(t, "1.0.0", cfg.AppVersion)
	require.Equal(t, "cert", cfg.SslSert)
	require.Equal(t, "key", cfg.SslKey)

	unsetEnv()
	os.Unsetenv("KEYCLOAK_JWK_URL")
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
	_ = os.Setenv("KEYCLOAK_JWK_URL", "http://localhost/jwk")
	setEnv("pgx", "postgres://user:pass@localhost/db", "test-app", "1.0.0", "cert", "key")
	envFile := writeTempEnvFile("SOME_VAR=123\n")
	defer os.Remove(envFile)

	cfg := common.GetConfig(envFile)

	require.Equal(t, "pgx", cfg.DbDriverName)
	require.Equal(t, "postgres://user:pass@localhost/db", cfg.Dsn)
	require.Equal(t, "test-app", cfg.AppName)
	require.Equal(t, "1.0.0", cfg.AppVersion)
	require.Equal(t, "cert", cfg.SslSert)
	require.Equal(t, "key", cfg.SslKey)

	unsetEnv()
	os.Unsetenv("KEYCLOAK_JWK_URL")
}

func Test_ConfigStruct_Validation(t *testing.T) {
	v := validator.New()

	t.Run("valid config", func(t *testing.T) {
		cfg := common.Config{
			DbDriverName:   "pgx",
			Dsn:            "postgres://user:pass@localhost/db",
			AppName:        "test-app",
			AppVersion:     "1.0.0",
			SslSert:        "cert-path",
			SslKey:         "key-path",
			KeycloakJwkUrl: "http://localhost/jwk",
		}
		err := v.Struct(cfg)
		assert.NoError(t, err)
	})

	t.Run("missing ssl cert", func(t *testing.T) {
		cfg := common.Config{
			DbDriverName: "pgx",
			Dsn:          "postgres://user:pass@localhost/db",
			AppName:      "test-app",
			AppVersion:   "1.0.0",
			SslSert:      "",
			SslKey:       "key-path",
		}
		err := v.Struct(cfg)
		assert.Error(t, err)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			assert.Equal(t, "SslSert", ve[0].Field())
			assert.Equal(t, "required", ve[0].Tag())
		}
	})

	t.Run("missing ssl key", func(t *testing.T) {
		cfg := common.Config{
			DbDriverName: "pgx",
			Dsn:          "postgres://user:pass@localhost/db",
			AppName:      "test-app",
			AppVersion:   "1.0.0",
			SslSert:      "cert-path",
			SslKey:       "",
		}
		err := v.Struct(cfg)
		assert.Error(t, err)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			assert.Equal(t, "SslKey", ve[0].Field())
			assert.Equal(t, "required", ve[0].Tag())
		}
	})

	t.Run("missing both ssl fields", func(t *testing.T) {
		cfg := common.Config{
			DbDriverName: "pgx",
			Dsn:          "postgres://user:pass@localhost/db",
			AppName:      "test-app",
			AppVersion:   "1.0.0",
			SslSert:      "",
			SslKey:       "",
		}
		err := v.Struct(cfg)
		assert.Error(t, err)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			fields := []string{ve[0].Field(), ve[1].Field()}
			assert.Contains(t, fields, "SslSert")
			assert.Contains(t, fields, "SslKey")
		}
	})
}
