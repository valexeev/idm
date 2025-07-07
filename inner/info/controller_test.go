package info

import (
	"context"
	"encoding/json"
	"errors"
	"idm/inner/common"
	"idm/inner/web"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDatabase - мок для интерфейса Database
type MockDatabase struct {
	mock.Mock
}

// PingContext мокает метод PingContext для проверки подключения к БД
// Реализует интерфейс Database
func (m *MockDatabase) PingContext(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// setupTest инициализирует тестовое окружение
func setupTest(t *testing.T) (*fiber.App, *MockDatabase) {
	t.Helper()

	// Используем конструктор веб-сервера из вашего кода
	server := web.NewServer()

	// Создаем мок базы данных
	mockDB := new(MockDatabase)

	// Создаем тестовый конфиг
	cfg := common.Config{
		DbDriverName: "postgres",
		Dsn:          "test_dsn",
		AppName:      "Test App",
		AppVersion:   "1.0.0",
	}

	// Создаем контроллер используя ваш конструктор
	controller := NewController(server, cfg, mockDB)

	// Явная проверка инициализации контроллера
	if controller == nil {
		t.Fatal("Controller is nil")
	}

	// Проверяем, что server.GroupInternal не nil перед регистрацией роутов
	if server.GroupInternal == nil {
		t.Fatal("GroupInternal is nil")
	}

	controller.RegisterRoutes()
	return server.App, mockDB
}

// parseResponse обрабатывает HTTP-ответ
func parseResponse(t *testing.T, resp *http.Response, target interface{}) {
	t.Helper()
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
}

func TestMain(m *testing.M) {
	_ = godotenv.Load(".env.tests")
	os.Exit(m.Run())
}

func TestGetInfo(t *testing.T) {
	app, _ := setupTest(t)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/internal/info", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		var result InfoResponse
		parseResponse(t, resp, &result)

		assert.Equal(t, "Test App", result.Name)
		assert.Equal(t, "1.0.0", result.Version)
	})
}

func TestGetHealth(t *testing.T) {
	app, mockDB := setupTest(t)

	t.Run("Success - Database Available", func(t *testing.T) {
		// Настраиваем мок: база данных доступна
		mockDB.On("PingContext", mock.AnythingOfType("*context.timerCtx")).Return(nil)

		req := httptest.NewRequest("GET", "/internal/health", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		// Проверяем, что ответ содержит "OK"
		body := make([]byte, 1024) // Увеличиваем размер буфера
		n, err := resp.Body.Read(body)
		if err != nil && err.Error() != "EOF" {
			t.Fatalf("Failed to read response body: %v", err)
		}
		responseBody := string(body[:n])
		assert.Equal(t, "OK", responseBody)

		mockDB.AssertExpectations(t)
	})

	t.Run("Failure - Database Unavailable", func(t *testing.T) {
		// Создаем новый мок для этого теста, чтобы избежать конфликтов
		app2, mockDB2 := setupTest(t)

		// Настраиваем мок: база данных недоступна
		mockDB2.On("PingContext", mock.AnythingOfType("*context.timerCtx")).Return(errors.New("database connection failed"))

		req := httptest.NewRequest("GET", "/internal/health", nil)
		resp, err := app2.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		// Отладочная информация
		if resp.StatusCode != 500 {
			t.Logf("Expected status 500, got %d", resp.StatusCode)

			// Читаем тело ответа для отладки
			body := make([]byte, 1024)
			n, _ := resp.Body.Read(body)
			t.Logf("Response body: %s", string(body[:n]))

			// Сбрасываем позицию в теле ответа
			resp.Body.Close()
			req2 := httptest.NewRequest("GET", "/internal/health", nil)
			resp, _ = app2.Test(req2)
		}

		assert.Equal(t, 500, resp.StatusCode)

		// Проверяем, что ответ содержит сообщение об ошибке
		var errorResponse common.Response[any]
		err = json.NewDecoder(resp.Body).Decode(&errorResponse)
		if err != nil {
			// Если не удается декодировать как JSON, читаем как строку
			resp.Body.Close()
			req3 := httptest.NewRequest("GET", "/internal/health", nil)
			resp3, _ := app2.Test(req3)
			defer resp3.Body.Close()

			body := make([]byte, 1024)
			n, _ := resp3.Body.Read(body)
			t.Logf("Raw response: %s", string(body[:n]))
			t.Fatalf("Failed to decode JSON response: %v", err)
		}

		assert.False(t, errorResponse.Success)
		assert.Contains(t, errorResponse.Message, "Database connection failed")

		mockDB2.AssertExpectations(t)
	})
}

func TestGetHealthDetailed(t *testing.T) {
	// Отдельный setup для детального health check
	server := web.NewServer()
	mockDB := new(MockDatabase)
	cfg := common.Config{
		DbDriverName: "postgres",
		Dsn:          "test_dsn",
		AppName:      "Test App",
		AppVersion:   "1.0.0",
	}

	controller := NewController(server, cfg, mockDB)
	controller.RegisterRoutes()

	// Регистрируем дополнительный роут для детального health check
	server.GroupInternal.Get("/health/detailed", controller.GetHealthDetailed)

	t.Run("Success - Database Available", func(t *testing.T) {
		// Настраиваем мок: база данных доступна
		mockDB.On("PingContext", mock.AnythingOfType("*context.timerCtx")).Return(nil)

		req := httptest.NewRequest("GET", "/internal/health/detailed", nil)
		resp, err := server.App.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		var result HealthResponse
		parseResponse(t, resp, &result)

		assert.Equal(t, "healthy", result.Status)
		assert.Equal(t, "ok", result.Database)

		mockDB.AssertExpectations(t)
	})

	t.Run("Failure - Database Unavailable", func(t *testing.T) {
		// Создаем новый мок для этого теста
		mockDB2 := new(MockDatabase)
		server2 := web.NewServer()
		controller2 := NewController(server2, cfg, mockDB2)
		controller2.RegisterRoutes()
		server2.GroupInternal.Get("/health/detailed", controller2.GetHealthDetailed)

		// Настраиваем мок: база данных недоступна
		mockDB2.On("PingContext", mock.AnythingOfType("*context.timerCtx")).Return(errors.New("database connection failed"))

		req := httptest.NewRequest("GET", "/internal/health/detailed", nil)
		resp, err := server2.App.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 500, resp.StatusCode)

		var result HealthResponse
		parseResponse(t, resp, &result)

		assert.Equal(t, "unhealthy", result.Status)
		assert.Equal(t, "error", result.Database)

		mockDB2.AssertExpectations(t)
	})
}
