package employee

import (
	"bytes"
	"encoding/json"
	"idm/inner/common"
	"idm/inner/web"
	"net/http"
	"net/http/httptest"
	"strings"
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
func (m *MockEmployeeService) ValidateRequest(request interface{}) error {
	args := m.Called(request)
	return args.Error(0)
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
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTest(t)
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
		app, mockService := setupTest(t)
		mockService.On("Add", "").Return(
			Response{},
			common.RequestValidationError{Message: "name cannot be empty"},
		)

		req := createTestRequest(t, "POST", "/api/v1/employees", AddEmployeeRequest{Name: ""})
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)

		var result common.Response[any]
		parseResponse(t, resp, &result)
		assert.False(t, result.Success)
		assert.Contains(t, result.Message, "cannot be empty")
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		app, _ := setupTest(t)
		req := httptest.NewRequest("POST", "/api/v1/employees", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestCreateEmployeeTransactional(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTest(t)
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
		app, mockService := setupTest(t)
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
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTest(t)
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
		app, mockService := setupTest(t)
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
		app, _ := setupTest(t)
		req := httptest.NewRequest("GET", "/api/v1/employees/invalid", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestGetAllEmployees(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTest(t)
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
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTest(t)
		request := FindByIdsRequest{Ids: []int64{1, 2}}
		expected := []Response{
			{Id: 1, Name: "User 1"},
			{Id: 2, Name: "User 2"},
		}
		mockService.On("ValidateRequest", request).Return(nil)
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
		app, mockService := setupTest(t)
		request := FindByIdsRequest{Ids: []int64{}}
		mockService.On("ValidateRequest", request).Return(
			common.RequestValidationError{Message: "ids cannot be empty"})

		req := createTestRequest(t, "POST", "/api/v1/employees/by-ids", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestDeleteEmployee(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTest(t)
		mockService.On("DeleteById", int64(1)).Return(nil)

		req := httptest.NewRequest("DELETE", "/api/v1/employees/1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 204, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		app, mockService := setupTest(t)
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
		app, _ := setupTest(t)
		req := httptest.NewRequest("DELETE", "/api/v1/employees/invalid", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestDeleteEmployeesByIds(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTest(t)
		request := DeleteByIdsRequest{Ids: []int64{1, 2}}
		mockService.On("ValidateRequest", request).Return(nil)
		mockService.On("DeleteByIds", []int64{1, 2}).Return(nil)

		req := createTestRequest(t, "DELETE", "/api/v1/employees", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 204, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Empty IDs", func(t *testing.T) {
		app, mockService := setupTest(t)
		request := DeleteByIdsRequest{Ids: []int64{}}
		mockService.On("ValidateRequest", request).Return(
			common.RequestValidationError{Message: "ids cannot be empty"})

		req := createTestRequest(t, "DELETE", "/api/v1/employees", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

// TestIncorrectDataHttpResponses - тесты правильных HTTP-ответов на некорректные данные
func TestIncorrectDataHttpResponses(t *testing.T) {
	errorResponseTests := []struct {
		name           string
		method         string
		url            string
		body           interface{}
		expectedStatus int
		expectedError  string
		setupMock      func(*MockEmployeeService)
	}{
		{
			name:           "create_empty_name_400",
			method:         "POST",
			url:            "/api/v1/employees",
			body:           AddEmployeeRequest{Name: ""},
			expectedStatus: 400,
			expectedError:  "cannot be empty",
			setupMock: func(m *MockEmployeeService) {
				m.On("Add", "").Return(
					Response{}, common.RequestValidationError{Message: "name cannot be empty"})
			},
		},
		{
			name:           "create_transactional_validation_error_400",
			method:         "POST",
			url:            "/api/v1/employees/transactional",
			body:           AddEmployeeRequest{Name: "A"},
			expectedStatus: 400,
			setupMock: func(m *MockEmployeeService) {
				m.On("AddTransactional", AddEmployeeRequest{Name: "A"}).Return(
					Response{}, common.RequestValidationError{Message: "name must be at least 2 characters long"})
			},
		},
		{
			name:           "get_employee_invalid_id_400",
			method:         "GET",
			url:            "/api/v1/employees/invalid",
			expectedStatus: 400,
			expectedError:  "invalid employee id",
			setupMock:      func(m *MockEmployeeService) {},
		},
		{
			name:           "get_employee_not_found_404",
			method:         "GET",
			url:            "/api/v1/employees/999",
			expectedStatus: 404,
			setupMock: func(m *MockEmployeeService) {
				m.On("FindById", int64(999)).Return(
					Response{}, common.NotFoundError{Message: "employee not found"})
			},
		},
		{
			name:           "get_by_ids_empty_list_400",
			method:         "POST",
			url:            "/api/v1/employees/by-ids",
			body:           FindByIdsRequest{Ids: []int64{}},
			expectedStatus: 400,
			expectedError:  "ids cannot be empty",
			setupMock: func(m *MockEmployeeService) {
				m.On("ValidateRequest", FindByIdsRequest{Ids: []int64{}}).Return(
					common.RequestValidationError{Message: "ids cannot be empty"})
			},
		},
		{
			name:           "delete_invalid_id_400",
			method:         "DELETE",
			url:            "/api/v1/employees/invalid",
			expectedStatus: 400,
			expectedError:  "invalid employee id",
			setupMock:      func(m *MockEmployeeService) {},
		},
		{
			name:           "delete_by_ids_validation_error_400",
			method:         "DELETE",
			url:            "/api/v1/employees",
			body:           DeleteByIdsRequest{Ids: []int64{1, 0, 3}},
			expectedStatus: 400,
			setupMock: func(m *MockEmployeeService) {
				m.On("ValidateRequest", DeleteByIdsRequest{Ids: []int64{1, 0, 3}}).Return(
					common.RequestValidationError{Message: "id must be greater than 0"})
			},
		},
		{
			name:           "internal_server_error_500",
			method:         "POST",
			url:            "/api/v1/employees/transactional",
			body:           AddEmployeeRequest{Name: "Valid Name"},
			expectedStatus: 500,
			setupMock: func(m *MockEmployeeService) {
				m.On("AddTransactional", AddEmployeeRequest{Name: "Valid Name"}).Return(
					Response{}, common.TransactionError{Message: "database connection failed"})
			},
		},
		{
			name:           "repository_error_500",
			method:         "GET",
			url:            "/api/v1/employees/1",
			expectedStatus: 500,
			setupMock: func(m *MockEmployeeService) {
				m.On("FindById", int64(1)).Return(
					Response{}, common.RepositoryError{Message: "database query failed"})
			},
		},
	}

	for _, test := range errorResponseTests {
		t.Run(test.name, func(t *testing.T) {
			app, mockService := setupTest(t)

			// Настраиваем мок
			test.setupMock(mockService)

			// Создаем запрос
			var req *http.Request
			if test.body != nil {
				req = createTestRequest(t, test.method, test.url, test.body)
			} else {
				req = httptest.NewRequest(test.method, test.url, nil)
			}

			// Выполняем запрос
			resp, err := app.Test(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			// Проверяем статус
			assert.Equal(t, test.expectedStatus, resp.StatusCode)

			// Проверяем тело ответа
			var result common.Response[any]
			parseResponse(t, resp, &result)

			assert.False(t, result.Success, "Response should indicate failure")

			if test.expectedError != "" {
				assert.Contains(t, result.Message, test.expectedError, "Error message should contain expected text")
			}

			// Проверяем выполнение ожиданий мока
			mockService.AssertExpectations(t)
		})
	}
}

// TestValidationDoesNotReachService - тесты что валидация на уровне контроллера не достигает сервиса
func TestValidationDoesNotReachService(t *testing.T) {
	controllerValidationTests := []struct {
		name   string
		method string
		url    string
		body   interface{}
	}{
		{
			name:   "invalid_json_body",
			method: "POST",
			url:    "/api/v1/employees",
			body:   "invalid json",
		},
		{
			name:   "invalid_id_parameter",
			method: "GET",
			url:    "/api/v1/employees/not_a_number",
		},
		{
			name:   "invalid_delete_id_parameter",
			method: "DELETE",
			url:    "/api/v1/employees/not_a_number",
		},
	}

	for _, test := range controllerValidationTests {
		t.Run(test.name, func(t *testing.T) {
			app, mockService := setupTest(t)
			// НЕ настраиваем никаких ожиданий для мока

			var req *http.Request
			if test.body != nil {
				if strBody, ok := test.body.(string); ok {
					req = httptest.NewRequest(test.method, test.url, strings.NewReader(strBody))
					req.Header.Set("Content-Type", "application/json")
				} else {
					req = createTestRequest(t, test.method, test.url, test.body)
				}
			} else {
				req = httptest.NewRequest(test.method, test.url, nil)
			}

			resp, err := app.Test(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			// Должен быть статус 400
			assert.Equal(t, 400, resp.StatusCode)

			// Убеждаемся, что сервис НЕ вызывался
			assert.Empty(t, mockService.Calls, "Service should not be called for controller-level validation errors")
		})
	}
}
