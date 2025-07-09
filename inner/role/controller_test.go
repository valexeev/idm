package role

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"idm/inner/common"
	"idm/inner/web"

	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRoleService - полный мок для интерфейса Svc
type MockRoleService struct {
	mock.Mock
}

func (m *MockRoleService) ValidateRequest(request any) error {
	args := m.Called(request)
	return args.Error(0)
}

func (m *MockRoleService) FindById(ctx context.Context, id int64) (Response, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockRoleService) Add(ctx context.Context, name string) (Response, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockRoleService) FindAll(ctx context.Context) ([]Response, error) {
	args := m.Called(ctx)
	return args.Get(0).([]Response), args.Error(1)
}

func (m *MockRoleService) FindByIds(ctx context.Context, ids []int64) ([]Response, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]Response), args.Error(1)
}

func (m *MockRoleService) DeleteById(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRoleService) DeleteByIds(ctx context.Context, ids []int64) error {
	args := m.Called(ctx, ids)
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

	// Добавляем middleware авторизации аналогично боевому приложению
	logger := &common.Logger{Logger: zap.NewNop()}
	groupApiV1.Use(web.AuthMiddleware(logger))

	controller := NewController(server, mockService, logger)

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

// createAuthRequest создает HTTP-запрос с авторизацией для тестов
func createAuthRequest(t *testing.T, method, url string, body interface{}) *http.Request {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("Failed to encode request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, url, &buf)
	req.Header.Set("Content-Type", "application/json")
	token := generateValidToken([]string{web.IdmAdmin})
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

// parseResponse обрабатывает HTTP-ответ
func parseResponse(t *testing.T, resp *http.Response, target interface{}) {
	t.Helper()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("failed to close response body: %v", err)
		}
	}()

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
}

func TestMain(m *testing.M) {
	os.Setenv("AUTH_TEST_SECRET", "testsecret")
	defer os.Unsetenv("AUTH_TEST_SECRET")
	_ = godotenv.Load(".env.tests")
	os.Exit(m.Run())
}

func TestCreateRole(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		request := AddRoleRequest{Name: "Admin"}
		expected := Response{Id: 1, Name: "Admin"}
		mockService.On("ValidateRequest", request).Return(nil).Once()
		mockService.On("Add", mock.Anything, "Admin").Return(expected, nil).Once()

		req := createAuthRequest(t, "POST", "/api/v1/roles", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

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
		mockService.On("ValidateRequest", request).Return(
			common.RequestValidationError{Message: "role name is invalid"}).Once()

		req := createAuthRequest(t, "POST", "/api/v1/roles", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

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
		mockService.On("ValidateRequest", request).Return(nil).Once()
		mockService.On("Add", mock.Anything, "Admin").Return(
			Response{},
			common.AlreadyExistsError{Message: "role already exists"},
		).Once()

		req := createAuthRequest(t, "POST", "/api/v1/roles", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

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
		mockService.On("ValidateRequest", request).Return(nil).Once()
		mockService.On("Add", mock.Anything, "Test Role").Return(
			Response{},
			errors.New("internal server error"),
		).Once()

		req := createAuthRequest(t, "POST", "/api/v1/roles", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

		assert.Equal(t, 500, resp.StatusCode)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		app, _ := setupTest(t)

		req := httptest.NewRequest("POST", "/api/v1/roles", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		token := generateValidToken([]string{web.IdmAdmin})
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		if resp != nil {
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("failed to close response body: %v", err)
				}
			}()
			assert.Equal(t, 400, resp.StatusCode)
		} else {
			t.Fatal("response is nil")
		}
	})
}

func TestGetRole(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		expected := Response{Id: 1, Name: "Admin"}
		mockService.On("FindById", mock.Anything, int64(1)).Return(expected, nil).Once()

		req := createAuthRequest(t, "GET", "/api/v1/roles/1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

		assert.Equal(t, 200, resp.StatusCode)

		var result common.Response[Response]
		parseResponse(t, resp, &result)
		assert.Equal(t, expected, result.Data)
	})

	t.Run("Not Found", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		mockService.On("FindById", mock.Anything, int64(999)).Return(
			Response{},
			common.NotFoundError{Message: "role not found"},
		).Once()

		req := createAuthRequest(t, "GET", "/api/v1/roles/999", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		mockService.On("FindById", mock.Anything, int64(1)).Return(
			Response{},
			errors.New("database error"),
		).Once()

		req := createAuthRequest(t, "GET", "/api/v1/roles/1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

		assert.Equal(t, 500, resp.StatusCode)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		app, _ := setupTest(t)

		req := createAuthRequest(t, "GET", "/api/v1/roles/invalid", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

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
		mockService.On("FindAll", mock.Anything).Return(expected, nil).Once()

		req := createAuthRequest(t, "GET", "/api/v1/roles", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

		assert.Equal(t, 200, resp.StatusCode)

		var result common.Response[[]Response]
		parseResponse(t, resp, &result)
		assert.Equal(t, expected, result.Data)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		mockService.On("FindAll", mock.Anything).Return(
			[]Response{},
			errors.New("database error"),
		).Once()

		req := createAuthRequest(t, "GET", "/api/v1/roles", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

		assert.Equal(t, 500, resp.StatusCode)
	})

	t.Run("Empty Result", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		expected := []Response{}
		mockService.On("FindAll", mock.Anything).Return(expected, nil).Once()

		req := createAuthRequest(t, "GET", "/api/v1/roles", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

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
		mockService.On("ValidateRequest", request).Return(nil).Once()
		mockService.On("FindByIds", mock.Anything, []int64{1, 2}).Return(expected, nil).Once()

		req := createAuthRequest(t, "POST", "/api/v1/roles/by-ids", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

		assert.Equal(t, 200, resp.StatusCode)

		var result common.Response[[]Response]
		parseResponse(t, resp, &result)
		assert.Equal(t, expected, result.Data)
	})

	t.Run("Empty IDs", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		request := FindByIdsRequest{Ids: []int64{}}
		mockService.On("ValidateRequest", request).Return(fmt.Errorf("ids list cannot be empty")).Once()

		req := createAuthRequest(t, "POST", "/api/v1/roles/by-ids", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

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
		mockService.On("ValidateRequest", request).Return(nil).Once()
		mockService.On("FindByIds", mock.Anything, []int64{1, 2}).Return(
			[]Response{},
			errors.New("database error"),
		).Once()

		req := createAuthRequest(t, "POST", "/api/v1/roles/by-ids", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

		assert.Equal(t, 500, resp.StatusCode)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		app, _ := setupTest(t)

		req := httptest.NewRequest("POST", "/api/v1/roles/by-ids", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		token := generateValidToken([]string{web.IdmAdmin})
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		if resp != nil {
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("failed to close response body: %v", err)
				}
			}()
			assert.Equal(t, 400, resp.StatusCode)
		} else {
			t.Fatal("response is nil")
		}
	})
}

func TestDeleteRole(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		mockService.On("DeleteById", mock.Anything, int64(1)).Return(nil).Once()

		req := createAuthRequest(t, "DELETE", "/api/v1/roles/1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

		assert.Equal(t, 204, resp.StatusCode)
	})

	t.Run("Not Found", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		mockService.On("DeleteById", mock.Anything, int64(999)).Return(
			common.NotFoundError{Message: "role not found"},
		).Once()

		req := createAuthRequest(t, "DELETE", "/api/v1/roles/999", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		mockService.On("DeleteById", mock.Anything, int64(1)).Return(
			errors.New("database error"),
		).Once()

		req := createAuthRequest(t, "DELETE", "/api/v1/roles/1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

		assert.Equal(t, 500, resp.StatusCode)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		app, _ := setupTest(t)

		req := createAuthRequest(t, "DELETE", "/api/v1/roles/invalid", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestDeleteRolesByIds(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		request := DeleteByIdsRequest{Ids: []int64{1, 2}}
		mockService.On("ValidateRequest", request).Return(nil).Once()
		mockService.On("DeleteByIds", mock.Anything, []int64{1, 2}).Return(nil).Once()

		req := createAuthRequest(t, "DELETE", "/api/v1/roles", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

		assert.Equal(t, 204, resp.StatusCode)
	})

	t.Run("Empty IDs", func(t *testing.T) {
		app, mockService := setupTest(t)
		defer mockService.AssertExpectations(t)

		request := DeleteByIdsRequest{Ids: []int64{}}
		mockService.On("ValidateRequest", request).Return(fmt.Errorf("ids list cannot be empty")).Once()

		req := createAuthRequest(t, "DELETE", "/api/v1/roles", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

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
		mockService.On("ValidateRequest", request).Return(nil).Once()
		mockService.On("DeleteByIds", mock.Anything, []int64{1, 2}).Return(
			errors.New("database error"),
		).Once()

		req := createAuthRequest(t, "DELETE", "/api/v1/roles", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

		assert.Equal(t, 500, resp.StatusCode)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		app, _ := setupTest(t)

		req := httptest.NewRequest("DELETE", "/api/v1/roles", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		token := generateValidToken([]string{web.IdmAdmin})
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		if resp != nil {
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("failed to close response body: %v", err)
				}
			}()
			assert.Equal(t, 400, resp.StatusCode)
		} else {
			t.Fatal("response is nil")
		}
	})
}

// TestRoleIncorrectDataHttpResponses - тесты правильных HTTP-ответов на некорректные данные
func TestRoleIncorrectDataHttpResponses(t *testing.T) {
	errorResponseTests := []struct {
		name           string
		method         string
		url            string
		body           interface{}
		expectedStatus int
		expectedError  string
		setupMock      func(*MockRoleService)
	}{
		{
			name:           "create_empty_name_400",
			method:         "POST",
			url:            "/api/v1/roles",
			body:           AddRoleRequest{Name: ""},
			expectedStatus: 400,
			expectedError:  "role name cannot be empty",
			setupMock: func(m *MockRoleService) {
				m.On("ValidateRequest", AddRoleRequest{Name: ""}).Return(
					common.RequestValidationError{Message: "role name cannot be empty"})
			},
		},
		{
			name:           "create_short_name_400",
			method:         "POST",
			url:            "/api/v1/roles",
			body:           AddRoleRequest{Name: "A"},
			expectedStatus: 400,
			expectedError:  "role name must be at least 2 characters long",
			setupMock: func(m *MockRoleService) {
				m.On("ValidateRequest", AddRoleRequest{Name: "A"}).Return(
					common.RequestValidationError{Message: "role name must be at least 2 characters long"})
			},
		},
		{
			name:           "create_long_name_400",
			method:         "POST",
			url:            "/api/v1/roles",
			body:           AddRoleRequest{Name: strings.Repeat("R", 101)},
			expectedStatus: 400,
			expectedError:  "role name must be at most 100 characters long",
			setupMock: func(m *MockRoleService) {
				longName := strings.Repeat("R", 101)
				m.On("ValidateRequest", AddRoleRequest{Name: longName}).Return(
					common.RequestValidationError{Message: "role name must be at most 100 characters long"})
			},
		},
		{
			name:           "get_role_invalid_id_400",
			method:         "GET",
			url:            "/api/v1/roles/invalid",
			expectedStatus: 400,
			expectedError:  "invalid role id",
			setupMock: func(m *MockRoleService) {

			},
		},
		{
			name:           "get_role_zero_id_400",
			method:         "GET",
			url:            "/api/v1/roles/0",
			expectedStatus: 400,
			expectedError:  "invalid role id",
			setupMock: func(m *MockRoleService) {

			},
		},
		{
			name:           "get_role_not_found_404",
			method:         "GET",
			url:            "/api/v1/roles/999",
			expectedStatus: 404,
			setupMock: func(m *MockRoleService) {
				m.On("FindById", mock.Anything, int64(999)).Return(
					Response{}, common.NotFoundError{Message: "role not found"})
			},
		},
		{
			name:           "get_by_ids_empty_list_400",
			method:         "POST",
			url:            "/api/v1/roles/by-ids",
			body:           FindByIdsRequest{Ids: []int64{}},
			expectedStatus: 400,
			expectedError:  "ids list cannot be empty",
			setupMock: func(m *MockRoleService) {
				m.On("ValidateRequest", FindByIdsRequest{Ids: []int64{}}).Return(
					common.RequestValidationError{Message: "ids list cannot be empty"})
			},
		},
		{
			name:           "get_by_ids_nil_list_400",
			method:         "POST",
			url:            "/api/v1/roles/by-ids",
			body:           FindByIdsRequest{Ids: nil},
			expectedStatus: 400,
			expectedError:  "ids list cannot be empty",
			setupMock: func(m *MockRoleService) {
				m.On("ValidateRequest", FindByIdsRequest{Ids: nil}).Return(
					common.RequestValidationError{Message: "ids list cannot be empty"})
			},
		},
		{
			name:           "get_by_ids_invalid_id_400",
			method:         "POST",
			url:            "/api/v1/roles/by-ids",
			body:           FindByIdsRequest{Ids: []int64{1, 0, 3}},
			expectedStatus: 400,
			setupMock: func(m *MockRoleService) {
				m.On("ValidateRequest", FindByIdsRequest{Ids: []int64{1, 0, 3}}).Return(
					common.RequestValidationError{Message: "id must be greater than 0"})
			},
		},
		{
			name:           "delete_invalid_id_400",
			method:         "DELETE",
			url:            "/api/v1/roles/invalid",
			expectedStatus: 400,
			expectedError:  "invalid role id",
			setupMock:      func(m *MockRoleService) {},
		},
		{
			name:           "delete_zero_id_400",
			method:         "DELETE",
			url:            "/api/v1/roles/0",
			expectedStatus: 400,
			expectedError:  "invalid role id",
			setupMock:      func(m *MockRoleService) {},
		},
		{
			name:           "delete_by_ids_validation_error_400",
			method:         "DELETE",
			url:            "/api/v1/roles",
			body:           DeleteByIdsRequest{Ids: []int64{1, -1, 3}},
			expectedStatus: 400,
			setupMock: func(m *MockRoleService) {
				m.On("ValidateRequest", DeleteByIdsRequest{Ids: []int64{1, -1, 3}}).Return(
					common.RequestValidationError{Message: "id must be greater than 0"})
			},
		},
		{
			name:           "delete_by_ids_empty_list_400",
			method:         "DELETE",
			url:            "/api/v1/roles",
			body:           DeleteByIdsRequest{Ids: []int64{}},
			expectedStatus: 400,
			expectedError:  "ids list cannot be empty",
			setupMock: func(m *MockRoleService) {
				m.On("ValidateRequest", DeleteByIdsRequest{Ids: []int64{}}).Return(
					common.RequestValidationError{Message: "ids list cannot be empty"})
			},
		},
		{
			name:           "repository_error_500",
			method:         "GET",
			url:            "/api/v1/roles/1",
			expectedStatus: 500,
			setupMock: func(m *MockRoleService) {
				m.On("FindById", mock.Anything, int64(1)).Return(
					Response{}, common.RepositoryError{Message: "database query failed"})
			},
		},
		{
			name:           "already_exists_error_400",
			method:         "POST",
			url:            "/api/v1/roles",
			body:           AddRoleRequest{Name: "Admin"},
			expectedStatus: 400,
			expectedError:  "already exists",
			setupMock: func(m *MockRoleService) {
				m.On("ValidateRequest", AddRoleRequest{Name: "Admin"}).Return(nil)
				m.On("Add", mock.Anything, "Admin").Return(
					Response{}, common.AlreadyExistsError{Message: "role already exists"})
			},
		},
	}

	for _, test := range errorResponseTests {
		t.Run(test.name, func(t *testing.T) {
			app, mockService := setupTest(t)

			// Сбрасываем мок для каждого теста
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			// Настраиваем мок
			test.setupMock(mockService)

			// Создаем запрос
			var req *http.Request
			if test.body != nil {
				req = createAuthRequest(t, test.method, test.url, test.body)
			} else {
				req = createAuthRequest(t, test.method, test.url, nil)
			}

			// Выполняем запрос
			resp, err := app.Test(req)
			assert.NoError(t, err)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("failed to close response body: %v", err)
				}
			}()

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

// TestRoleValidationDoesNotReachService - тесты что валидация на уровне контроллера не достигает сервиса
func TestRoleValidationDoesNotReachService(t *testing.T) {
	app, mockService := setupTest(t)

	controllerValidationTests := []struct {
		name   string
		method string
		url    string
		body   interface{}
	}{
		{
			name:   "invalid_json_body_create",
			method: "POST",
			url:    "/api/v1/roles",
			body:   "invalid json",
		},
		{
			name:   "invalid_json_body_get_by_ids",
			method: "POST",
			url:    "/api/v1/roles/by-ids",
			body:   "invalid json",
		},
		{
			name:   "invalid_json_body_delete_by_ids",
			method: "DELETE",
			url:    "/api/v1/roles",
			body:   "invalid json",
		},
		{
			name:   "invalid_id_parameter_get",
			method: "GET",
			url:    "/api/v1/roles/not_a_number",
		},
		{
			name:   "invalid_id_parameter_delete",
			method: "DELETE",
			url:    "/api/v1/roles/not_a_number",
		},
		{
			name:   "negative_id_parameter_get",
			method: "GET",
			url:    "/api/v1/roles/-1",
		},
		{
			name:   "negative_id_parameter_delete",
			method: "DELETE",
			url:    "/api/v1/roles/-1",
		},
	}

	for _, test := range controllerValidationTests {
		t.Run(test.name, func(t *testing.T) {
			// НЕ настраиваем никаких ожиданий для мока
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			var req *http.Request
			if test.body != nil {
				if strBody, ok := test.body.(string); ok {
					req = httptest.NewRequest(test.method, test.url, strings.NewReader(strBody))
					req.Header.Set("Content-Type", "application/json")
					// Добавляем валидный токен для прохождения auth middleware
					token := generateValidToken([]string{web.IdmAdmin})
					req.Header.Set("Authorization", "Bearer "+token)
				} else {
					req = createTestRequest(t, test.method, test.url, test.body)
				}
			} else {
				req = httptest.NewRequest(test.method, test.url, nil)
				// Для тестов с некорректным id тоже нужен ток��н
				token := generateValidToken([]string{web.IdmAdmin})
				req.Header.Set("Authorization", "Bearer "+token)
			}

			resp, err := app.Test(req)
			assert.NoError(t, err)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("failed to close response body: %v", err)
				}
			}()

			// Должен быть статус 400
			assert.Equal(t, 400, resp.StatusCode)

			// Убеждаемся, что сервис НЕ вызывался
			assert.Empty(t, mockService.Calls, "Service should not be called for controller-level validation errors")
		})
	}
}

// TestRoleValidationWithDifferentErrorTypes - тесты различных типов ошибок валидации
func TestRoleValidationWithDifferentErrorTypes(t *testing.T) {
	app, mockService := setupTest(t)

	validationTests := []struct {
		name           string
		method         string
		url            string
		body           interface{}
		expectedStatus int
		setupMock      func(*MockRoleService)
	}{
		{
			name:           "validation_error_during_create",
			method:         "POST",
			url:            "/api/v1/roles",
			body:           AddRoleRequest{Name: ""},
			expectedStatus: 400,
			setupMock: func(m *MockRoleService) {
				m.On("ValidateRequest", AddRoleRequest{Name: ""}).Return(
					common.RequestValidationError{Message: "role name cannot be empty"})
			},
		},
		{
			name:           "validation_error_during_get_by_ids",
			method:         "POST",
			url:            "/api/v1/roles/by-ids",
			body:           FindByIdsRequest{Ids: []int64{}},
			expectedStatus: 400,
			setupMock: func(m *MockRoleService) {
				m.On("ValidateRequest", FindByIdsRequest{Ids: []int64{}}).Return(
					common.RequestValidationError{Message: "ids list cannot be empty"})
			},
		},
		{
			name:           "validation_error_during_delete_by_ids",
			method:         "DELETE",
			url:            "/api/v1/roles",
			body:           DeleteByIdsRequest{Ids: []int64{0}},
			expectedStatus: 400,
			setupMock: func(m *MockRoleService) {
				m.On("ValidateRequest", DeleteByIdsRequest{Ids: []int64{0}}).Return(
					common.RequestValidationError{Message: "id must be greater than 0"})
			},
		},
	}

	for _, test := range validationTests {
		t.Run(test.name, func(t *testing.T) {
			// Сбрасываем мок
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			// Настраиваем мок
			test.setupMock(mockService)

			// Создаем запрос
			req := createAuthRequest(t, test.method, test.url, test.body)

			// Выполняем запрос
			resp, err := app.Test(req)
			assert.NoError(t, err)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("failed to close response body: %v", err)
				}
			}()

			// Проверяем статус
			assert.Equal(t, test.expectedStatus, resp.StatusCode)

			// Проверяем, что ответ содержит информацию об ошибке
			var result common.Response[any]
			parseResponse(t, resp, &result)
			assert.False(t, result.Success)
			assert.NotEmpty(t, result.Message)

			// Проверяем выполнение ожиданий мока
			mockService.AssertExpectations(t)
		})
	}
}

// TestRoleBoundaryValues - тесты ��раничных значений
func TestRoleBoundaryValues(t *testing.T) {
	boundaryTests := []struct {
		name           string
		method         string
		url            string
		body           interface{}
		expectedStatus int
		setupMock      func(*MockRoleService)
	}{
		{
			name:           "minimum_valid_name_length",
			method:         "POST",
			url:            "/api/v1/roles",
			body:           AddRoleRequest{Name: "AB"}, // 2 символа - минимум
			expectedStatus: 200,
			setupMock: func(m *MockRoleService) {
				m.On("ValidateRequest", AddRoleRequest{Name: "AB"}).Return(nil)
				m.On("Add", mock.Anything, "AB").Return(Response{Id: 1, Name: "AB"}, nil)
			},
		},
		{
			name:           "maximum_valid_name_length",
			method:         "POST",
			url:            "/api/v1/roles",
			body:           AddRoleRequest{Name: strings.Repeat("R", 100)}, // 100 символов - максимум
			expectedStatus: 200,
			setupMock: func(m *MockRoleService) {
				longName := strings.Repeat("R", 100)
				m.On("ValidateRequest", AddRoleRequest{Name: longName}).Return(nil)
				m.On("Add", mock.Anything, longName).Return(Response{Id: 1, Name: longName}, nil)
			},
		},
		{
			name:           "maximum_valid_id",
			method:         "GET",
			url:            "/api/v1/roles/9223372036854775807", // max int64
			expectedStatus: 200,
			setupMock: func(m *MockRoleService) {
				maxId := int64(9223372036854775807)
				m.On("FindById", mock.Anything, maxId).Return(Response{Id: maxId, Name: "Test"}, nil)
			},
		},
		{
			name:           "single_valid_id_in_list",
			method:         "POST",
			url:            "/api/v1/roles/by-ids",
			body:           FindByIdsRequest{Ids: []int64{1}}, // Минимум 1 элемент
			expectedStatus: 200,
			setupMock: func(m *MockRoleService) {
				m.On("ValidateRequest", FindByIdsRequest{Ids: []int64{1}}).Return(nil)
				m.On("FindByIds", mock.Anything, []int64{1}).Return([]Response{{Id: 1, Name: "Test"}}, nil)
			},
		},
	}

	for _, test := range boundaryTests {
		t.Run(test.name, func(t *testing.T) {
			app, mockService := setupTest(t)

			// Сбрасываем мок
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			// Настраиваем мок
			test.setupMock(mockService)

			// Создаем запрос
			var req *http.Request
			if test.body != nil {
				req = createAuthRequest(t, test.method, test.url, test.body)
			} else {
				req = createAuthRequest(t, test.method, test.url, nil)
			}

			// Выполняем запрос
			resp, err := app.Test(req)
			assert.NoError(t, err)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("failed to close response body: %v", err)
				}
			}()

			// Проверяем статус
			assert.Equal(t, test.expectedStatus, resp.StatusCode)

			// Проверяем выполнение ожиданий мока
			mockService.AssertExpectations(t)
		})
	}
}

// --- helpers for authz tests ---

// Генерирует валидный JWT-токен с нужными ролями
func generateValidToken(roles []string) string {
	claims := &web.IdmClaims{
		RealmAccess: web.RealmAccessClaims{Roles: roles},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte("testsecret"))
	return signed
}

// Генерирует просроченный JWT-токен
func generateExpiredToken(roles []string) string {
	claims := &web.IdmClaims{
		RealmAccess: web.RealmAccessClaims{Roles: roles},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte("testsecret"))
	return signed
}

// Генерирует токен с неверной подписью
func generateTokenWithWrongSignature(roles []string) string {
	claims := &web.IdmClaims{
		RealmAccess: web.RealmAccessClaims{Roles: roles},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte("wrongsecret"))
	return signed
}

// --- Тесты на аутентификацию и авторизацию для всех эндпоинтов ---
func TestRole_AuthZ(t *testing.T) {
	endpoints := []struct {
		name   string
		method string
		url    string
		body   interface{}
		roles  []string // роли, которые нужны для доступа
	}{
		{"CreateRole", "POST", "/api/v1/roles", AddRoleRequest{Name: "Test"}, []string{web.IdmAdmin}},
		{"GetRole", "GET", "/api/v1/roles/1", nil, []string{web.IdmAdmin, web.IdmUser}},
		{"GetAllRoles", "GET", "/api/v1/roles", nil, []string{web.IdmAdmin, web.IdmUser}},
		{"GetRolesByIds", "POST", "/api/v1/roles/by-ids", FindByIdsRequest{Ids: []int64{1}}, []string{web.IdmAdmin, web.IdmUser}},
		{"DeleteRole", "DELETE", "/api/v1/roles/1", nil, []string{web.IdmAdmin}},
		{"DeleteRolesByIds", "DELETE", "/api/v1/roles", DeleteByIdsRequest{Ids: []int64{1}}, []string{web.IdmAdmin}},
	}

	for _, ep := range endpoints {
		t.Run(ep.name+"_NoToken", func(t *testing.T) {
			app, mockService := setupTest(t)
			// Настроить мок для всех методов, которые могут быть вызваны
			mockService.On("ValidateRequest", mock.Anything).Return(nil).Maybe()
			mockService.On("Add", mock.Anything, mock.Anything).Return(Response{}, nil).Maybe()
			mockService.On("FindById", mock.Anything, mock.Anything).Return(Response{}, nil).Maybe()
			mockService.On("FindAll", mock.Anything).Return([]Response{}, nil).Maybe()
			mockService.On("FindByIds", mock.Anything, mock.Anything).Return([]Response{}, nil).Maybe()
			mockService.On("DeleteById", mock.Anything, mock.Anything).Return(nil).Maybe()
			mockService.On("DeleteByIds", mock.Anything, mock.Anything).Return(nil).Maybe()
			var req *http.Request
			if ep.body != nil {
				req = createTestRequest(t, ep.method, ep.url, ep.body)
			} else {
				req = httptest.NewRequest(ep.method, ep.url, nil)
			}
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, 401, resp.StatusCode)
		})

		t.Run(ep.name+"_InvalidToken", func(t *testing.T) {
			app, mockService := setupTest(t)
			mockService.On("ValidateRequest", mock.Anything).Return(nil).Maybe()
			mockService.On("Add", mock.Anything, mock.Anything).Return(Response{}, nil).Maybe()
			mockService.On("FindById", mock.Anything, mock.Anything).Return(Response{}, nil).Maybe()
			mockService.On("FindAll", mock.Anything).Return([]Response{}, nil).Maybe()
			mockService.On("FindByIds", mock.Anything, mock.Anything).Return([]Response{}, nil).Maybe()
			mockService.On("DeleteById", mock.Anything, mock.Anything).Return(nil).Maybe()
			mockService.On("DeleteByIds", mock.Anything, mock.Anything).Return(nil).Maybe()
			var req *http.Request
			if ep.body != nil {
				req = createTestRequest(t, ep.method, ep.url, ep.body)
			} else {
				req = httptest.NewRequest(ep.method, ep.url, nil)
			}
			req.Header.Set("Authorization", "Bearer invalid.token.here")
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, 401, resp.StatusCode)
		})

		t.Run(ep.name+"_NoRole", func(t *testing.T) {
			app, mockService := setupTest(t)
			// НЕ настраиваем моки, так как до сервиса дойти не должно
			var req *http.Request
			if ep.body != nil {
				req = createTestRequest(t, ep.method, ep.url, ep.body)
			} else {
				req = httptest.NewRequest(ep.method, ep.url, nil)
			}
			// Токен без нужной роли
			token := generateValidToken([]string{"SOME_OTHER_ROLE"})
			req.Header.Set("Authorization", "Bearer "+token)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, 403, resp.StatusCode)
			// Проверяем, что сервис не был вызван
			assert.Empty(t, mockService.Calls, "Service should not be called for forbidden access")
		})

		t.Run(ep.name+"_ExpiredToken", func(t *testing.T) {
			app, mockService := setupTest(t)
			mockService.On("ValidateRequest", mock.Anything).Return(nil).Maybe()
			mockService.On("Add", mock.Anything, mock.Anything).Return(Response{}, nil).Maybe()
			mockService.On("FindById", mock.Anything, mock.Anything).Return(Response{}, nil).Maybe()
			mockService.On("FindAll", mock.Anything).Return([]Response{}, nil).Maybe()
			mockService.On("FindByIds", mock.Anything, mock.Anything).Return([]Response{}, nil).Maybe()
			mockService.On("DeleteById", mock.Anything, mock.Anything).Return(nil).Maybe()
			mockService.On("DeleteByIds", mock.Anything, mock.Anything).Return(nil).Maybe()
			var req *http.Request
			if ep.body != nil {
				req = createTestRequest(t, ep.method, ep.url, ep.body)
			} else {
				req = httptest.NewRequest(ep.method, ep.url, nil)
			}
			token := generateExpiredToken(ep.roles)
			req.Header.Set("Authorization", "Bearer "+token)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, 401, resp.StatusCode)
		})

		t.Run(ep.name+"_WrongSignature", func(t *testing.T) {
			app, mockService := setupTest(t)
			mockService.On("ValidateRequest", mock.Anything).Return(nil).Maybe()
			mockService.On("Add", mock.Anything, mock.Anything).Return(Response{}, nil).Maybe()
			mockService.On("FindById", mock.Anything, mock.Anything).Return(Response{}, nil).Maybe()
			mockService.On("FindAll", mock.Anything).Return([]Response{}, nil).Maybe()
			mockService.On("FindByIds", mock.Anything, mock.Anything).Return([]Response{}, nil).Maybe()
			mockService.On("DeleteById", mock.Anything, mock.Anything).Return(nil).Maybe()
			mockService.On("DeleteByIds", mock.Anything, mock.Anything).Return(nil).Maybe()
			var req *http.Request
			if ep.body != nil {
				req = createTestRequest(t, ep.method, ep.url, ep.body)
			} else {
				req = httptest.NewRequest(ep.method, ep.url, nil)
			}
			token := generateTokenWithWrongSignature(ep.roles)
			req.Header.Set("Authorization", "Bearer "+token)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, 401, resp.StatusCode)
		})
	}
}
