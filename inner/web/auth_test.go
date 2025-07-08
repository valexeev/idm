package web

import (
	"encoding/json"
	"idm/inner/common"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestRequireRoles(t *testing.T) {
	// Устанавливаем тестовый секрет для JWT
	os.Setenv("AUTH_TEST_SECRET", "testsecret")
	defer os.Unsetenv("AUTH_TEST_SECRET")

	// Создаем тестовое приложение
	app := fiber.New()
	logger := common.NewTestLogger()

	// Применяем AuthMiddleware
	app.Use(AuthMiddleware(logger))

	// Тестовый эндпоинт, требующий роль IDM_ADMIN
	app.Get("/admin-only", RequireRoles(IdmAdmin), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Тестовый эндпоинт, требующий роли IDM_ADMIN или IDM_USER
	app.Get("/admin-or-user", RequireRoles(IdmAdmin, IdmUser), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	t.Run("should allow access with admin role", func(t *testing.T) {
		token := GenerateTestToken([]string{IdmAdmin})

		req := httptest.NewRequest("GET", "/admin-only", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("should allow access with user role for mixed endpoint", func(t *testing.T) {
		token := GenerateTestToken([]string{IdmUser})

		req := httptest.NewRequest("GET", "/admin-or-user", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("should deny access without admin role", func(t *testing.T) {
		token := GenerateTestToken([]string{IdmUser})

		req := httptest.NewRequest("GET", "/admin-only", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 403, resp.StatusCode)

		// Проверяем тело ответа
		defer resp.Body.Close()

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "insufficient permissions", response["error"]) // Изменено с "message" на "error"
	})

	t.Run("should deny access without token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin-only", nil)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 401, resp.StatusCode)
	})

	t.Run("should deny access with invalid token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin-only", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 401, resp.StatusCode)
	})

	t.Run("should deny access with no roles", func(t *testing.T) {
		token := GenerateTestToken([]string{})

		req := httptest.NewRequest("GET", "/admin-only", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 403, resp.StatusCode)
	})

	t.Run("should allow access with multiple roles", func(t *testing.T) {
		token := GenerateTestToken([]string{IdmUser, IdmAdmin, "OTHER_ROLE"})

		req := httptest.NewRequest("GET", "/admin-only", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}

func TestGenerateTestToken(t *testing.T) {
	t.Run("should generate valid token with roles", func(t *testing.T) {
		roles := []string{IdmAdmin, IdmUser}
		token := GenerateTestToken(roles)

		assert.NotEmpty(t, token)
		assert.True(t, len(token) > 50) // JWT токены обычно довольно длинные
	})

	t.Run("should generate token with empty roles", func(t *testing.T) {
		token := GenerateTestToken([]string{})

		assert.NotEmpty(t, token)
	})
}

func TestAuthMiddleware(t *testing.T) {
	// Устанавливаем тестовый секрет для JWT
	os.Setenv("AUTH_TEST_SECRET", "testsecret")
	defer os.Unsetenv("AUTH_TEST_SECRET")

	app := fiber.New()
	logger := common.NewTestLogger()

	// Применяем AuthMiddleware
	app.Use(AuthMiddleware(logger))

	// Простой тестовый эндпоинт
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	t.Run("should authenticate with valid token", func(t *testing.T) {
		token := GenerateTestToken([]string{IdmUser})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("should reject invalid token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 401, resp.StatusCode)
	})

	t.Run("should reject request without authorization header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 401, resp.StatusCode)
	})
}
