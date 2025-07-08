package web

import (
	"idm/inner/common"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddlewareDebug(t *testing.T) {
	// Устанавливаем переменную окружения в начале теста
	os.Setenv("AUTH_TEST_SECRET", "testsecret")
	defer os.Unsetenv("AUTH_TEST_SECRET")

	// Проверяем, что переменная установлена
	secret := os.Getenv("AUTH_TEST_SECRET")
	t.Logf("AUTH_TEST_SECRET: %s", secret)
	assert.Equal(t, "testsecret", secret)

	// Создаем приложение
	app := fiber.New()
	logger := common.NewTestLogger()

	// Применяем наш middleware динамически
	app.Use(func(c *fiber.Ctx) error {
		return CreateAuthMiddleware(logger)(c)
	})

	// Простой тестовый эндпоинт
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	t.Run("should work with valid token", func(t *testing.T) {
		// Генерируем тестовый токен
		token := GenerateTestToken([]string{IdmUser})
		t.Logf("Generated token: %s...", token[:50])

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		assert.NoError(t, err)

		t.Logf("Response status: %d", resp.StatusCode)
		if resp.StatusCode != 200 {
			// Читаем тело ответа для отладки
			body := make([]byte, 1024)
			n, _ := resp.Body.Read(body)
			t.Logf("Response body: %s", string(body[:n]))
		}

		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("should fail without token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 401, resp.StatusCode)
	})
}
