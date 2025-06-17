package employee

import (
	"bytes"
	"encoding/json"
	"idm/inner/common"
	"idm/inner/web"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockEmployeeService - полный мок для интерфейса Svc
type MockEmployeeService struct {
	mock.Mock
}

func (m *MockEmployeeService) FindById(id int64) (Response, error) {
	args := m.Called(id)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockEmployeeService) AddTransactional(request AddEmployeeRequest) (Response, error) {
	args := m.Called(request)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockEmployeeService) Add(name string) (Response, error) {
	args := m.Called(name)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockEmployeeService) FindAll() ([]Response, error) {
	args := m.Called()
	return args.Get(0).([]Response), args.Error(1)
}

func (m *MockEmployeeService) FindByIds(ids []int64) ([]Response, error) {
	args := m.Called(ids)
	return args.Get(0).([]Response), args.Error(1)
}

func (m *MockEmployeeService) DeleteById(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockEmployeeService) DeleteByIds(ids []int64) error {
	args := m.Called(ids)
	return args.Error(0)
}

// setupTest инициализирует тестовое окружение
func setupTest(t *testing.T) (*fiber.App, *MockEmployeeService) {
	t.Helper()

	app := fiber.New()

	// Создаем группу API V1
	groupApiV1 := app.Group("/api/v1")

	// Инициализируем сервер с правильно настроенными группами роутов
	server := &web.Server{
		App:        app,
		GroupApiV1: groupApiV1, // Убеждаемся, что группа создана корректно
	}

	mockService := new(MockEmployeeService)

	// Создаем тестовый логгер
	logger := zap.NewNop() // Логгер, который ничего не делает (подходит для тестов)
	commonLogger := &common.Logger{Logger: logger}

	// Добавляем логгер
	controller := NewController(server, mockService, commonLogger)

	// Явная проверка инициализации контроллера
	if controller == nil {
		t.Fatal("Controller is nil")
	}

	// Проверяем, что server.GroupApiV1 не nil перед регистрацией роутов
	if server.GroupApiV1 == nil {
		t.Fatal("GroupApiV1 is nil")
	}

	controller.RegisterRoutes()
	return app, mockService
}

// createTestRequest создает HTTP-запрос для тестов
func createTestRequest(t *testing.T, method, url string, body interface{}) *http.Request {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("Failed to encode request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, url, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// parseResponse обрабатывает HTTP-ответ
func parseResponse(t *testing.T, resp *http.Response, target interface{}) {
	t.Helper()
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
}

func TestCreateEmployee(t *testing.T) {
	app, mockService := setupTest(t)

	t.Run("Success", func(t *testing.T) {
		request := AddEmployeeRequest{Name: "John Doe"}
		expected := Response{Id: 1, Name: "John Doe"}
		mockService.On("Add", "John Doe").Return(expected, nil)

		req := createTestRequest(t, "POST", "/api/v1/employees", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		var result common.Response[Response]
		parseResponse(t, resp, &result)

		assert.True(t, result.Success)
		assert.Equal(t, expected, result.Data)
		mockService.AssertExpectations(t)
	})

	t.Run("Empty Name", func(t *testing.T) {
		// Для пустого имени мы НЕ настраиваем mock, потому что контроллер должен
		// вернуть ошибку валидации до вызова сервиса
		req := createTestRequest(t, "POST", "/api/v1/employees", AddEmployeeRequest{Name: ""})
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)

		var result common.Response[any]
		parseResponse(t, resp, &result)
		assert.False(t, result.Success)
		assert.Contains(t, result.Message, "cannot be empty")
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/employees", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestCreateEmployeeTransactional(t *testing.T) {
	app, mockService := setupTest(t)

	t.Run("Success", func(t *testing.T) {
		request := AddEmployeeRequest{Name: "John Transaction"}
		expected := Response{Id: 2, Name: "John Transaction"}
		mockService.On("AddTransactional", request).Return(expected, nil)

		req := createTestRequest(t, "POST", "/api/v1/employees/transactional", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		var result common.Response[Response]
		parseResponse(t, resp, &result)
		assert.Equal(t, expected, result.Data)
		mockService.AssertExpectations(t)
	})

	t.Run("Transaction Error", func(t *testing.T) {
		request := AddEmployeeRequest{Name: "Fail Transaction"}
		mockService.On("AddTransactional", request).Return(
			Response{},
			common.TransactionError{Message: "transaction failed"},
		)

		req := createTestRequest(t, "POST", "/api/v1/employees/transactional", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 500, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestGetEmployee(t *testing.T) {
	app, mockService := setupTest(t)

	t.Run("Success", func(t *testing.T) {
		expected := Response{Id: 1, Name: "Test User"}
		mockService.On("FindById", int64(1)).Return(expected, nil)

		req := httptest.NewRequest("GET", "/api/v1/employees/1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		var result common.Response[Response]
		parseResponse(t, resp, &result)
		assert.Equal(t, expected, result.Data)
		mockService.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockService.On("FindById", int64(999)).Return(
			Response{},
			common.NotFoundError{Message: "not found"},
		)

		req := httptest.NewRequest("GET", "/api/v1/employees/999", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 404, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/employees/invalid", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestGetAllEmployees(t *testing.T) {
	app, mockService := setupTest(t)

	t.Run("Success", func(t *testing.T) {
		expected := []Response{
			{Id: 1, Name: "User 1"},
			{Id: 2, Name: "User 2"},
		}
		mockService.On("FindAll").Return(expected, nil)

		req := httptest.NewRequest("GET", "/api/v1/employees", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		var result common.Response[[]Response]
		parseResponse(t, resp, &result)
		assert.Equal(t, expected, result.Data)
		mockService.AssertExpectations(t)
	})
}

func TestGetEmployeesByIds(t *testing.T) {
	app, mockService := setupTest(t)

	t.Run("Success", func(t *testing.T) {
		request := FindByIdsRequest{Ids: []int64{1, 2}}
		expected := []Response{
			{Id: 1, Name: "User 1"},
			{Id: 2, Name: "User 2"},
		}
		mockService.On("FindByIds", []int64{1, 2}).Return(expected, nil)

		req := createTestRequest(t, "POST", "/api/v1/employees/by-ids", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		var result common.Response[[]Response]
		parseResponse(t, resp, &result)
		assert.Equal(t, expected, result.Data)
		mockService.AssertExpectations(t)
	})

	t.Run("Empty IDs", func(t *testing.T) {
		request := FindByIdsRequest{Ids: []int64{}}

		req := createTestRequest(t, "POST", "/api/v1/employees/by-ids", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestDeleteEmployee(t *testing.T) {
	app, mockService := setupTest(t)

	t.Run("Success", func(t *testing.T) {
		mockService.On("DeleteById", int64(1)).Return(nil)

		req := httptest.NewRequest("DELETE", "/api/v1/employees/1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 204, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockService.On("DeleteById", int64(999)).Return(
			common.NotFoundError{Message: "employee not found"},
		)

		req := httptest.NewRequest("DELETE", "/api/v1/employees/999", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 404, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/employees/invalid", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestDeleteEmployeesByIds(t *testing.T) {
	app, mockService := setupTest(t)

	t.Run("Success", func(t *testing.T) {
		request := DeleteByIdsRequest{Ids: []int64{1, 2}}
		mockService.On("DeleteByIds", []int64{1, 2}).Return(nil)

		req := createTestRequest(t, "DELETE", "/api/v1/employees", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 204, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Empty IDs", func(t *testing.T) {
		request := DeleteByIdsRequest{Ids: []int64{}}

		req := createTestRequest(t, "DELETE", "/api/v1/employees", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
	})
}
