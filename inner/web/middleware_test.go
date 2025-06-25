package web

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecoverMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		setupRoute     func(app *fiber.App)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "should recover from panic and return 500",
			setupRoute: func(app *fiber.App) {
				app.Get("/panic", func(c *fiber.Ctx) error {
					panic("test panic")
				})
			},
			expectedStatus: 500,
			expectedBody:   "test panic",
		},
		{
			name: "should not interfere with normal requests",
			setupRoute: func(app *fiber.App) {
				app.Get("/normal", func(c *fiber.Ctx) error {
					return c.JSON(fiber.Map{"message": "success"})
				})
			},
			expectedStatus: 200,
			expectedBody:   `{"message":"success"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем новый сервер для каждого теста
			server := NewServer()

			// Настраиваем роут для теста
			tt.setupRoute(server.App)

			// Создаем тестовый запрос
			req := httptest.NewRequest("GET", getTestPath(tt.name), nil)
			resp, err := server.App.Test(req)

			// Проверяем результат
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Читаем тело ответа
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Contains(t, string(body), tt.expectedBody)
		})
	}
}

func TestRequestIdMiddleware(t *testing.T) {
	tests := []struct {
		name              string
		requestId         string
		expectedRequestId string
		description       string
	}{
		{
			name:              "should use provided request ID",
			requestId:         "test-request-id-123",
			expectedRequestId: "test-request-id-123",
			description:       "When X-Request-ID header is provided, it should be used",
		},
		{
			name:              "should generate request ID when not provided",
			requestId:         "",
			expectedRequestId: "", // Будем проверять, что ID не пустой
			description:       "When X-Request-ID header is not provided, it should generate one",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем новый сервер
			server := NewServer()

			// Добавляем тестовый роут, который возвращает Request ID
			server.App.Get("/test", func(c *fiber.Ctx) error {
				requestId := c.Locals("requestid")
				return c.JSON(fiber.Map{
					"request_id": requestId,
				})
			})

			// Создаем запрос
			req := httptest.NewRequest("GET", "/test", nil)

			// Устанавливаем заголовок, если он предоставлен
			if tt.requestId != "" {
				req.Header.Set("X-Request-ID", tt.requestId)
			}

			// Выполняем запрос
			resp, err := server.App.Test(req)
			require.NoError(t, err)

			// Проверяем статус
			assert.Equal(t, 200, resp.StatusCode)

			// Читаем тело ответа
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// Проверяем Request ID
			if tt.expectedRequestId != "" {
				// Если ожидаем конкретный ID
				assert.Contains(t, string(body), tt.expectedRequestId)
			} else {
				// Если ожидаем сгенерированный ID (не пустой)
				assert.Contains(t, string(body), `"request_id":`)
				assert.NotContains(t, string(body), `"request_id":""`)
				assert.NotContains(t, string(body), `"request_id":null`)
			}

			// Проверяем, что заголовок X-Request-ID установлен в ответе
			responseRequestId := resp.Header.Get("X-Request-ID")
			if tt.requestId != "" {
				assert.Equal(t, tt.requestId, responseRequestId)
			} else {
				assert.NotEmpty(t, responseRequestId)
			}
		})
	}
}

func TestMiddlewareIntegration(t *testing.T) {
	t.Run("should handle panic with request ID", func(t *testing.T) {
		server := NewServer()

		// Роут, который паникует
		server.App.Get("/panic-with-id", func(c *fiber.Ctx) error {
			// Проверяем, что Request ID есть перед паникой
			requestId := c.Locals("requestid")
			if requestId == nil {
				t.Error("Request ID should be set before panic")
			}
			panic("test panic with request id")
		})

		// Запрос с Request ID
		req := httptest.NewRequest("GET", "/panic-with-id", nil)
		req.Header.Set("X-Request-ID", "panic-test-123")

		resp, err := server.App.Test(req)
		require.NoError(t, err)

		// Проверяем, что сервер восстановился после паники
		assert.Equal(t, 500, resp.StatusCode)

		// Проверяем, что Request ID сохранен в заголовке ответа
		assert.Equal(t, "panic-test-123", resp.Header.Get("X-Request-ID"))
	})
}

// Вспомогательная функция для получения пути для теста
func getTestPath(testName string) string {
	switch testName {
	case "should recover from panic and return 500":
		return "/panic"
	case "should not interfere with normal requests":
		return "/normal"
	default:
		return "/test"
	}
}
