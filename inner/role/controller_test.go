package role

import (
	"bytes"
	"encoding/json"
	"errors"
	"idm/inner/common"
	"idm/inner/web"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRoleService - полный мок для интерфейса Svc
type MockRoleService struct {
	mock.Mock
}

func (m *MockRoleService) FindById(id int64) (Response, error) {
	args := m.Called(id)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockRoleService) Add(name string) (Response, error) {
	args := m.Called(name)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockRoleService) FindAll() ([]Response, error) {
	args := m.Called()
	return args.Get(0).([]Response), args.Error(1)
}

func (m *MockRoleService) FindByIds(ids []int64) ([]Response, error) {
	args := m.Called(ids)
	return args.Get(0).([]Response), args.Error(1)
}

func (m *MockRoleService) DeleteById(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRoleService) DeleteByIds(ids []int64) error {
	args := m.Called(ids)
	return args.Error(0)
}

// setupTest инициализирует тестовое окружение
func setupTest(t *testing.T) (*fiber.App, *MockRoleService) {
	t.Helper()

	app := fiber.New()

	// Создаем группу API V1
	groupApiV1 := app.Group("/api/v1")

	// Инициализируем сервер с правильно настроенными группами роутов
	server := &web.Server{
		App:        app,
		GroupApiV1: groupApiV1,
	}

	mockService := new(MockRoleService)
	controller := NewController(server, mockService)

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

func TestCreateRole(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		request := AddRoleRequest{Name: "Admin"}
		expected := Response{Id: 1, Name: "Admin"}
		mockService.On("Add", "Admin").Return(expected, nil).Once()

		req := createTestRequest(t, "POST", "/api/v1/roles", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		var result common.Response[Response]
		parseResponse(t, resp, &result)

		assert.True(t, result.Success)
		assert.Equal(t, expected, result.Data)
	})

	t.Run("Validation Error", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		request := AddRoleRequest{Name: "Invalid Role"}
		mockService.On("Add", "Invalid Role").Return(
			Response{},
			common.RequestValidationError{Message: "role name is invalid"},
		).Once()

		req := createTestRequest(t, "POST", "/api/v1/roles", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)

		var result common.Response[any]
		parseResponse(t, resp, &result)
		assert.False(t, result.Success)
		assert.Contains(t, result.Message, "role name is invalid")
	})

	t.Run("Already Exists Error", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		request := AddRoleRequest{Name: "Admin"}
		mockService.On("Add", "Admin").Return(
			Response{},
			common.AlreadyExistsError{Message: "role already exists"},
		).Once()

		req := createTestRequest(t, "POST", "/api/v1/roles", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)

		var result common.Response[any]
		parseResponse(t, resp, &result)
		assert.False(t, result.Success)
		assert.Contains(t, result.Message, "role already exists")
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		request := AddRoleRequest{Name: "Test Role"}
		mockService.On("Add", "Test Role").Return(
			Response{},
			errors.New("internal server error"),
		).Once()

		req := createTestRequest(t, "POST", "/api/v1/roles", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 500, resp.StatusCode)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		app, _ := setupTest(t)

		req := httptest.NewRequest("POST", "/api/v1/roles", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestGetRole(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		expected := Response{Id: 1, Name: "Admin"}
		mockService.On("FindById", int64(1)).Return(expected, nil).Once()

		req := httptest.NewRequest("GET", "/api/v1/roles/1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		var result common.Response[Response]
		parseResponse(t, resp, &result)
		assert.Equal(t, expected, result.Data)
	})

	t.Run("Not Found", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		mockService.On("FindById", int64(999)).Return(
			Response{},
			common.NotFoundError{Message: "role not found"},
		).Once()

		req := httptest.NewRequest("GET", "/api/v1/roles/999", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		mockService.On("FindById", int64(1)).Return(
			Response{},
			errors.New("database error"),
		).Once()

		req := httptest.NewRequest("GET", "/api/v1/roles/1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 500, resp.StatusCode)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		app, _ := setupTest(t)

		req := httptest.NewRequest("GET", "/api/v1/roles/invalid", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestGetAllRoles(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		expected := []Response{
			{Id: 1, Name: "Admin"},
			{Id: 2, Name: "User"},
		}
		mockService.On("FindAll").Return(expected, nil).Once()

		req := httptest.NewRequest("GET", "/api/v1/roles", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		var result common.Response[[]Response]
		parseResponse(t, resp, &result)
		assert.Equal(t, expected, result.Data)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		mockService.On("FindAll").Return(
			[]Response{},
			errors.New("database error"),
		).Once()

		req := httptest.NewRequest("GET", "/api/v1/roles", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 500, resp.StatusCode)
	})

	t.Run("Empty Result", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		expected := []Response{}
		mockService.On("FindAll").Return(expected, nil).Once()

		req := httptest.NewRequest("GET", "/api/v1/roles", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		var result common.Response[[]Response]
		parseResponse(t, resp, &result)
		assert.Equal(t, expected, result.Data)
	})
}

func TestGetRolesByIds(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		request := FindByIdsRequest{Ids: []int64{1, 2}}
		expected := []Response{
			{Id: 1, Name: "Admin"},
			{Id: 2, Name: "User"},
		}
		mockService.On("FindByIds", []int64{1, 2}).Return(expected, nil).Once()

		req := createTestRequest(t, "POST", "/api/v1/roles/by-ids", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		var result common.Response[[]Response]
		parseResponse(t, resp, &result)
		assert.Equal(t, expected, result.Data)
	})

	t.Run("Empty IDs", func(t *testing.T) {
		app, _ := setupTest(t)

		request := FindByIdsRequest{Ids: []int64{}}

		req := createTestRequest(t, "POST", "/api/v1/roles/by-ids", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)

		var result common.Response[any]
		parseResponse(t, resp, &result)
		assert.False(t, result.Success)
		assert.Contains(t, result.Message, "ids list cannot be empty")
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		request := FindByIdsRequest{Ids: []int64{1, 2}}
		mockService.On("FindByIds", []int64{1, 2}).Return(
			[]Response{},
			errors.New("database error"),
		).Once()

		req := createTestRequest(t, "POST", "/api/v1/roles/by-ids", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 500, resp.StatusCode)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		app, _ := setupTest(t)

		req := httptest.NewRequest("POST", "/api/v1/roles/by-ids", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestDeleteRole(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		mockService.On("DeleteById", int64(1)).Return(nil).Once()

		req := httptest.NewRequest("DELETE", "/api/v1/roles/1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 204, resp.StatusCode)
	})

	t.Run("Not Found", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		mockService.On("DeleteById", int64(999)).Return(
			common.NotFoundError{Message: "role not found"},
		).Once()

		req := httptest.NewRequest("DELETE", "/api/v1/roles/999", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		mockService.On("DeleteById", int64(1)).Return(
			errors.New("database error"),
		).Once()

		req := httptest.NewRequest("DELETE", "/api/v1/roles/1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 500, resp.StatusCode)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		app, _ := setupTest(t)

		req := httptest.NewRequest("DELETE", "/api/v1/roles/invalid", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestDeleteRolesByIds(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		request := DeleteByIdsRequest{Ids: []int64{1, 2}}
		mockService.On("DeleteByIds", []int64{1, 2}).Return(nil).Once()

		req := createTestRequest(t, "DELETE", "/api/v1/roles", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 204, resp.StatusCode)
	})

	t.Run("Empty IDs", func(t *testing.T) {
		app, _ := setupTest(t)

		request := DeleteByIdsRequest{Ids: []int64{}}

		req := createTestRequest(t, "DELETE", "/api/v1/roles", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)

		var result common.Response[any]
		parseResponse(t, resp, &result)
		assert.False(t, result.Success)
		assert.Contains(t, result.Message, "ids list cannot be empty")
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		request := DeleteByIdsRequest{Ids: []int64{1, 2}}
		mockService.On("DeleteByIds", []int64{1, 2}).Return(
			errors.New("database error"),
		).Once()

		req := createTestRequest(t, "DELETE", "/api/v1/roles", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 500, resp.StatusCode)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		app, _ := setupTest(t)

		req := httptest.NewRequest("DELETE", "/api/v1/roles", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
	})
}
