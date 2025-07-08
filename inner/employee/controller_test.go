package employee

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
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
	"go.uber.org/zap"
)

// MockEmployeeService - полный мок для интерфейса Svc
type MockEmployeeService struct {
	mock.Mock
}

func (m *MockEmployeeService) FindById(ctx context.Context, id int64) (Response, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockEmployeeService) AddTransactional(ctx context.Context, request AddEmployeeRequest) (Response, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(Response), args.Error(1)
}
func (m *MockEmployeeService) ValidateRequest(request interface{}) error {
	args := m.Called(request)
	return args.Error(0)
}
func (m *MockEmployeeService) Add(ctx context.Context, name string) (Response, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockEmployeeService) FindAll(ctx context.Context) ([]Response, error) {
	args := m.Called(ctx)
	return args.Get(0).([]Response), args.Error(1)
}

func (m *MockEmployeeService) FindByIds(ctx context.Context, ids []int64) ([]Response, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]Response), args.Error(1)
}

func (m *MockEmployeeService) DeleteById(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEmployeeService) DeleteByIds(ctx context.Context, ids []int64) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *MockEmployeeService) FindPage(ctx context.Context, req PageRequest) (PageResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(PageResponse), args.Error(1)
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

	// Добавляем реальный middleware авторизации
	groupApiV1.Use(web.AuthMiddleware(commonLogger))

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
	defer func() { _ = resp.Body.Close() }()

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

func TestCreateEmployee(t *testing.T) {
	t.Run("Success with admin role", func(t *testing.T) {
		svc := new(MockEmployeeService)
		app := setupAppWithAuth(t, svc, []string{web.IdmAdmin})
		request := AddEmployeeRequest{Name: "John Doe"}
		expected := Response{Id: 1, Name: "John Doe"}
		svc.On("Add", mock.Anything, "John Doe").Return(expected, nil)

		req := createTestRequest(t, "POST", "/api/v1/employees", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 200, resp.StatusCode)
		var result common.Response[Response]
		parseResponse(t, resp, &result)
		assert.True(t, result.Success)
		assert.Equal(t, expected, result.Data)
		svc.AssertExpectations(t)
	})

	t.Run("Forbidden for non-admin role", func(t *testing.T) {
		svc := new(MockEmployeeService)
		app := setupAppWithAuth(t, svc, []string{"user"})
		request := AddEmployeeRequest{Name: "John Doe"}

		req := createTestRequest(t, "POST", "/api/v1/employees", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 403, resp.StatusCode)
		var result common.Response[any]
		parseResponse(t, resp, &result)
		assert.False(t, result.Success)
		assert.Contains(t, result.Message, "Permission denied")
	})

	t.Run("Empty Name", func(t *testing.T) {
		svc := new(MockEmployeeService)
		app := setupAppWithAuth(t, svc, []string{web.IdmAdmin})
		svc.On("Add", mock.Anything, "").Return(
			Response{},
			common.RequestValidationError{Message: "name cannot be empty"},
		)

		req := createTestRequest(t, "POST", "/api/v1/employees", AddEmployeeRequest{Name: ""})
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 400, resp.StatusCode)

		var result common.Response[any]
		parseResponse(t, resp, &result)
		assert.False(t, result.Success)
		assert.Contains(t, result.Message, "cannot be empty")
		svc.AssertExpectations(t)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		svc := new(MockEmployeeService)
		app := setupAppWithAuth(t, svc, []string{web.IdmAdmin})
		req := httptest.NewRequest("POST", "/api/v1/employees", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestCreateEmployeeTransactional(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc := new(MockEmployeeService)
		app := setupAppWithAuth(t, svc, []string{web.IdmAdmin})
		request := AddEmployeeRequest{Name: "John Transaction"}
		expected := Response{Id: 2, Name: "John Transaction"}
		svc.On("AddTransactional", mock.Anything, request).Return(expected, nil)

		req := createTestRequest(t, "POST", "/api/v1/employees/transactional", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 200, resp.StatusCode)

		var result common.Response[Response]
		parseResponse(t, resp, &result)
		assert.Equal(t, expected, result.Data)
		svc.AssertExpectations(t)
	})

	t.Run("Transaction Error", func(t *testing.T) {
		svc := new(MockEmployeeService)
		app := setupAppWithAuth(t, svc, []string{web.IdmAdmin})
		request := AddEmployeeRequest{Name: "Fail Transaction"}
		svc.On("AddTransactional", mock.Anything, request).Return(
			Response{},
			common.TransactionError{Message: "transaction failed"},
		)

		req := createTestRequest(t, "POST", "/api/v1/employees/transactional", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 500, resp.StatusCode)
		svc.AssertExpectations(t)
	})
}

func TestGetEmployee(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc := new(MockEmployeeService)
		app := setupAppWithAuth(t, svc, []string{web.IdmAdmin, web.IdmUser})
		expected := Response{Id: 1, Name: "Test User"}
		svc.On("FindById", mock.Anything, int64(1)).Return(expected, nil)

		req := httptest.NewRequest("GET", "/api/v1/employees/1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 200, resp.StatusCode)

		var result common.Response[Response]
		parseResponse(t, resp, &result)
		assert.Equal(t, expected, result.Data)
		svc.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		svc := new(MockEmployeeService)
		app := setupAppWithAuth(t, svc, []string{web.IdmAdmin, web.IdmUser})
		svc.On("FindById", mock.Anything, int64(999)).Return(
			Response{},
			common.NotFoundError{Message: "not found"},
		)

		req := httptest.NewRequest("GET", "/api/v1/employees/999", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 404, resp.StatusCode)
		svc.AssertExpectations(t)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		svc := new(MockEmployeeService)
		app := setupAppWithAuth(t, svc, []string{web.IdmAdmin, web.IdmUser})
		req := httptest.NewRequest("GET", "/api/v1/employees/invalid", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestGetAllEmployees(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc := new(MockEmployeeService)
		app := setupAppWithAuth(t, svc, []string{web.IdmAdmin, web.IdmUser})
		expected := []Response{
			{Id: 1, Name: "User 1"},
			{Id: 2, Name: "User 2"},
		}
		svc.On("FindAll", mock.Anything).Return(expected, nil)

		req := httptest.NewRequest("GET", "/api/v1/employees", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 200, resp.StatusCode)

		var result common.Response[[]Response]
		parseResponse(t, resp, &result)
		assert.Equal(t, expected, result.Data)
		svc.AssertExpectations(t)
	})
}

func TestGetEmployeesByIds(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc := new(MockEmployeeService)
		app := setupAppWithAuth(t, svc, []string{web.IdmAdmin, web.IdmUser})
		request := FindByIdsRequest{Ids: []int64{1, 2}}
		expected := []Response{
			{Id: 1, Name: "User 1"},
			{Id: 2, Name: "User 2"},
		}
		svc.On("ValidateRequest", request).Return(nil)
		svc.On("FindByIds", mock.Anything, []int64{1, 2}).Return(expected, nil)

		req := createTestRequest(t, "POST", "/api/v1/employees/by-ids", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 200, resp.StatusCode)

		var result common.Response[[]Response]
		parseResponse(t, resp, &result)
		assert.Equal(t, expected, result.Data)
		svc.AssertExpectations(t)
	})

	t.Run("Empty IDs", func(t *testing.T) {
		svc := new(MockEmployeeService)
		app := setupAppWithAuth(t, svc, []string{web.IdmAdmin, web.IdmUser})
		request := FindByIdsRequest{Ids: []int64{}}
		svc.On("ValidateRequest", request).Return(
			common.RequestValidationError{Message: "ids cannot be empty"})

		req := createTestRequest(t, "POST", "/api/v1/employees/by-ids", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 400, resp.StatusCode)
		svc.AssertExpectations(t)
	})
}

func TestDeleteEmployee(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc := new(MockEmployeeService)
		app := setupAppWithAuth(t, svc, []string{web.IdmAdmin})
		svc.On("DeleteById", mock.Anything, int64(1)).Return(nil)

		req := httptest.NewRequest("DELETE", "/api/v1/employees/1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 204, resp.StatusCode)
		svc.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		svc := new(MockEmployeeService)
		app := setupAppWithAuth(t, svc, []string{web.IdmAdmin})
		svc.On("DeleteById", mock.Anything, int64(999)).Return(
			common.NotFoundError{Message: "employee not found"},
		)

		req := httptest.NewRequest("DELETE", "/api/v1/employees/999", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 404, resp.StatusCode)
		svc.AssertExpectations(t)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		svc := new(MockEmployeeService)
		app := setupAppWithAuth(t, svc, []string{web.IdmAdmin})
		req := httptest.NewRequest("DELETE", "/api/v1/employees/invalid", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestDeleteEmployeesByIds(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc := new(MockEmployeeService)
		app := setupAppWithAuth(t, svc, []string{web.IdmAdmin})
		request := DeleteByIdsRequest{Ids: []int64{1, 2}}
		svc.On("ValidateRequest", request).Return(nil)
		svc.On("DeleteByIds", mock.Anything, []int64{1, 2}).Return(nil)

		req := createTestRequest(t, "DELETE", "/api/v1/employees", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 204, resp.StatusCode)
		svc.AssertExpectations(t)
	})

	t.Run("Empty IDs", func(t *testing.T) {
		svc := new(MockEmployeeService)
		app := setupAppWithAuth(t, svc, []string{web.IdmAdmin})
		request := DeleteByIdsRequest{Ids: []int64{}}
		svc.On("ValidateRequest", request).Return(
			common.RequestValidationError{Message: "ids cannot be empty"})

		req := createTestRequest(t, "DELETE", "/api/v1/employees", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 400, resp.StatusCode)
		svc.AssertExpectations(t)
	})
}

// --- Тесты на аутентификацию и авторизацию для всех эндпоинтов ---
func TestEmployee_AuthZ(t *testing.T) {
	endpoints := []struct {
		name   string
		method string
		url    string
		body   interface{}
		roles  []string // роли, которые нужны для доступа
	}{
		{"CreateEmployee", "POST", "/api/v1/employees", AddEmployeeRequest{Name: "Test"}, []string{web.IdmAdmin}},
		{"CreateEmployeeTransactional", "POST", "/api/v1/employees/transactional", AddEmployeeRequest{Name: "Test"}, []string{web.IdmAdmin}},
		{"GetEmployee", "GET", "/api/v1/employees/1", nil, []string{web.IdmAdmin, web.IdmUser}},
		{"GetAllEmployees", "GET", "/api/v1/employees", nil, []string{web.IdmAdmin, web.IdmUser}},
		{"GetEmployeesPage", "GET", "/api/v1/employees/page?pageNumber=0&pageSize=1", nil, []string{web.IdmAdmin, web.IdmUser}},
		{"GetEmployeesByIds", "POST", "/api/v1/employees/by-ids", FindByIdsRequest{Ids: []int64{1}}, []string{web.IdmAdmin, web.IdmUser}},
		{"DeleteEmployee", "DELETE", "/api/v1/employees/1", nil, []string{web.IdmAdmin}},
		{"DeleteEmployeesByIds", "DELETE", "/api/v1/employees", DeleteByIdsRequest{Ids: []int64{1}}, []string{web.IdmAdmin}},
	}

	for _, ep := range endpoints {
		t.Run(ep.name+"_NoToken", func(t *testing.T) {
			app, _ := setupTest(t)
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
			app, _ := setupTest(t)
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
			app, _ := setupTest(t)
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
		})

		t.Run(ep.name+"_ExpiredToken", func(t *testing.T) {
			app, _ := setupTest(t)
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
			app, _ := setupTest(t)
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

// helper для создания jwt.Token с нужными ролями
func makeJWTToken(roles []string) *jwt.Token {
	claims := &web.IdmClaims{
		RealmAccess: web.RealmAccessClaims{Roles: roles},
	}
	return &jwt.Token{Claims: claims}
}

// helper для Fiber-приложения с подменой middleware авторизации
func setupAppWithAuth(_ *testing.T, svc *MockEmployeeService, roles []string) *fiber.App {
	app := fiber.New()
	groupApiV1 := app.Group("/api/v1")
	server := &web.Server{App: app, GroupApiV1: groupApiV1}
	logger := &common.Logger{Logger: zap.NewNop()}
	controller := NewController(server, svc, logger)
	// middleware, который кладёт jwt.Token с нужными ролями
	auth := func(c *fiber.Ctx) error {
		c.Locals(web.JwtKey, makeJWTToken(roles))
		return c.Next()
	}
	server.GroupApiV1.Use(auth)
	controller.RegisterRoutes()
	return app
}

func TestCreateEmployee_WithAuth(t *testing.T) {
	t.Run("Success with admin role", func(t *testing.T) {
		svc := new(MockEmployeeService)
		app := setupAppWithAuth(t, svc, []string{web.IdmAdmin})
		request := AddEmployeeRequest{Name: "John Doe"}
		expected := Response{Id: 1, Name: "John Doe"}
		svc.On("Add", mock.Anything, "John Doe").Return(expected, nil)

		req := createTestRequest(t, "POST", "/api/v1/employees", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 200, resp.StatusCode)
		var result common.Response[Response]
		parseResponse(t, resp, &result)
		assert.True(t, result.Success)
		assert.Equal(t, expected, result.Data)
		svc.AssertExpectations(t)
	})

	t.Run("Forbidden for non-admin role", func(t *testing.T) {
		svc := new(MockEmployeeService)
		app := setupAppWithAuth(t, svc, []string{"user"})
		request := AddEmployeeRequest{Name: "John Doe"}

		req := createTestRequest(t, "POST", "/api/v1/employees", request)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 403, resp.StatusCode)
		var result common.Response[any]
		parseResponse(t, resp, &result)
		assert.False(t, result.Success)
		assert.Contains(t, result.Message, "Permission denied")
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
				m.On("Add", mock.Anything, "").Return(
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
				m.On("AddTransactional", mock.Anything, AddEmployeeRequest{Name: "A"}).Return(
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
				m.On("FindById", mock.Anything, int64(999)).Return(
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
				m.On("AddTransactional", mock.Anything, AddEmployeeRequest{Name: "Valid Name"}).Return(
					Response{}, common.TransactionError{Message: "database connection failed"})
			},
		},
		{
			name:           "repository_error_500",
			method:         "GET",
			url:            "/api/v1/employees/1",
			expectedStatus: 500,
			setupMock: func(m *MockEmployeeService) {
				m.On("FindById", mock.Anything, int64(1)).Return(
					Response{}, common.RepositoryError{Message: "database query failed"})
			},
		},
	}

	for _, test := range errorResponseTests {
		t.Run(test.name, func(t *testing.T) {
			var app *fiber.App
			var mockService *MockEmployeeService
			// Для защищённых эндпоинтов используем setupAppWithAuth
			if (test.method == "POST" && (test.url == "/api/v1/employees" || test.url == "/api/v1/employees/transactional" || test.url == "/api/v1/employees/by-ids")) ||
				(test.method == "GET" && strings.HasPrefix(test.url, "/api/v1/employees")) ||
				(test.method == "DELETE" && strings.HasPrefix(test.url, "/api/v1/employees")) {
				mockService = new(MockEmployeeService)
				app = setupAppWithAuth(t, mockService, []string{web.IdmAdmin, web.IdmUser})
			} else {
				app, mockService = setupTest(t)
			}

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

			// Проверяем выполненное ожиданий мока
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
			var app *fiber.App
			var mockService *MockEmployeeService
			// Для защищённых эндпоинтов используем setupAppWithAuth
			if (test.method == "POST" && strings.HasPrefix(test.url, "/api/v1/employees")) ||
				(test.method == "GET" && strings.HasPrefix(test.url, "/api/v1/employees")) ||
				(test.method == "DELETE" && strings.HasPrefix(test.url, "/api/v1/employees")) {
				mockService = new(MockEmployeeService)
				app = setupAppWithAuth(t, mockService, []string{web.IdmAdmin, web.IdmUser})
			} else {
				app, mockService = setupTest(t)
			}

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

// TestEmployeeAuthenticationAndAuthorization тестирует аутентификацию и авторизацию для всех эндпоинтов
func TestEmployeeAuthenticationAndAuthorization(t *testing.T) {
	// Тестовые данные
	testEmployee := Response{Id: 1, Name: "Test Employee"}

	// Список всех эндпоинтов для тестирования
	testCases := []struct {
		name          string
		method        string
		path          string
		body          interface{}
		requiredRoles []string // роли, которые должны иметь доступ
		setupMock     func(*MockEmployeeService)
	}{
		{
			name:          "CreateEmployee",
			method:        "POST",
			path:          "/api/v1/employees",
			body:          AddEmployeeRequest{Name: "New Employee"},
			requiredRoles: []string{web.IdmAdmin},
			setupMock: func(svc *MockEmployeeService) {
				svc.On("Add", mock.Anything, "New Employee").Return(testEmployee, nil)
			},
		},
		{
			name:          "CreateEmployeeTransactional",
			method:        "POST",
			path:          "/api/v1/employees/transactional",
			body:          AddEmployeeRequest{Name: "New Employee"},
			requiredRoles: []string{web.IdmAdmin},
			setupMock: func(svc *MockEmployeeService) {
				svc.On("AddTransactional", mock.Anything, mock.AnythingOfType("AddEmployeeRequest")).Return(testEmployee, nil)
			},
		},
		{
			name:          "GetEmployee",
			method:        "GET",
			path:          "/api/v1/employees/1",
			body:          nil,
			requiredRoles: []string{web.IdmAdmin, web.IdmUser},
			setupMock: func(svc *MockEmployeeService) {
				svc.On("FindById", mock.Anything, int64(1)).Return(testEmployee, nil)
			},
		},
		{
			name:          "GetAllEmployees",
			method:        "GET",
			path:          "/api/v1/employees",
			body:          nil,
			requiredRoles: []string{web.IdmAdmin, web.IdmUser},
			setupMock: func(svc *MockEmployeeService) {
				svc.On("FindAll", mock.Anything).Return([]Response{testEmployee}, nil)
			},
		},
		{
			name:          "GetEmployeesPage",
			method:        "GET",
			path:          "/api/v1/employees/page?pageNumber=0&pageSize=20",
			body:          nil,
			requiredRoles: []string{web.IdmAdmin, web.IdmUser},
			setupMock: func(svc *MockEmployeeService) {
				svc.On("FindPage", mock.Anything, mock.AnythingOfType("PageRequest")).Return(
					PageResponse{Result: []Response{testEmployee}, Total: 1}, nil)
			},
		},
		{
			name:          "GetEmployeesByIds",
			method:        "POST",
			path:          "/api/v1/employees/by-ids",
			body:          FindByIdsRequest{Ids: []int64{1}},
			requiredRoles: []string{web.IdmAdmin, web.IdmUser},
			setupMock: func(svc *MockEmployeeService) {
				svc.On("ValidateRequest", mock.AnythingOfType("FindByIdsRequest")).Return(nil)
				svc.On("FindByIds", mock.Anything, []int64{1}).Return([]Response{testEmployee}, nil)
			},
		},
		{
			name:          "DeleteEmployee",
			method:        "DELETE",
			path:          "/api/v1/employees/1",
			body:          nil,
			requiredRoles: []string{web.IdmAdmin},
			setupMock: func(svc *MockEmployeeService) {
				svc.On("DeleteById", mock.Anything, int64(1)).Return(nil)
			},
		},
		{
			name:          "DeleteEmployeesByIds",
			method:        "DELETE",
			path:          "/api/v1/employees",
			body:          DeleteByIdsRequest{Ids: []int64{1}},
			requiredRoles: []string{web.IdmAdmin},
			setupMock: func(svc *MockEmployeeService) {
				svc.On("ValidateRequest", mock.AnythingOfType("DeleteByIdsRequest")).Return(nil)
				svc.On("DeleteByIds", mock.Anything, []int64{1}).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test 1: Unauthorized - нет токена (должен возвращать 401)
			t.Run("Unauthorized_NoToken", func(t *testing.T) {
				svc := new(MockEmployeeService)
				app := setupAppWithoutAuth(t, svc)

				var req *http.Request
				if tc.body != nil {
					req = createTestRequest(t, tc.method, tc.path, tc.body)
				} else {
					req = httptest.NewRequest(tc.method, tc.path, nil)
				}

				resp, err := app.Test(req)
				assert.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, 401, resp.StatusCode, "Should return 401 when no token provided")
			})

			// Test 2: Unauthorized - неверный токен (должен возвращать 401)
			t.Run("Unauthorized_InvalidToken", func(t *testing.T) {
				svc := new(MockEmployeeService)
				app := setupAppWithoutAuth(t, svc)

				var req *http.Request
				if tc.body != nil {
					req = createTestRequest(t, tc.method, tc.path, tc.body)
				} else {
					req = httptest.NewRequest(tc.method, tc.path, nil)
				}
				req.Header.Set("Authorization", "Bearer invalid-token")

				resp, err := app.Test(req)
				assert.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, 401, resp.StatusCode, "Should return 401 when invalid token provided")
			})

			// Test 3: Forbidden - недостаточные права (должен возвращать 403)
			t.Run("Forbidden_InsufficientRoles", func(t *testing.T) {
				// Используем роль, которая НЕ входит в requiredRoles
				wrongRole := "WRONG_ROLE"
				if len(tc.requiredRoles) > 0 {
					// Если требуется только IDM_ADMIN, используем IDM_USER
					if tc.requiredRoles[0] == web.IdmAdmin && len(tc.requiredRoles) == 1 {
						wrongRole = web.IdmUser
					} else {
						// Для остальных случаев используем произвольную роль
						wrongRole = "SOME_OTHER_ROLE"
					}
				}

				svc := new(MockEmployeeService)
				app := setupAppWithAuth(t, svc, []string{wrongRole})

				var req *http.Request
				if tc.body != nil {
					req = createTestRequest(t, tc.method, tc.path, tc.body)
				} else {
					req = createAuthenticatedRequest(t, tc.method, tc.path, nil, []string{wrongRole})
				}

				resp, err := app.Test(req)
				assert.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, 403, resp.StatusCode, "Should return 403 when insufficient roles")

				var result common.Response[any]
				parseResponse(t, resp, &result)
				assert.False(t, result.Success)
				assert.Contains(t, result.Message, "Permission denied")
			})

			// Test 4: Success - правильные права (должен возвращать успех)
			for _, role := range tc.requiredRoles {
				t.Run("Success_With_"+role, func(t *testing.T) {
					svc := new(MockEmployeeService)
					app := setupAppWithAuth(t, svc, []string{role})

					// Настраиваем мок
					if tc.setupMock != nil {
						tc.setupMock(svc)
					}

					var req *http.Request
					if tc.body != nil {
						req = createTestRequest(t, tc.method, tc.path, tc.body)
					} else {
						req = createAuthenticatedRequest(t, tc.method, tc.path, nil, []string{role})
					}

					resp, err := app.Test(req)
					assert.NoError(t, err)
					defer resp.Body.Close()

					// Проверяем, что запрос успешен (не 401/403)
					assert.NotEqual(t, 401, resp.StatusCode, "Should not return 401 with valid token and role")
					assert.NotEqual(t, 403, resp.StatusCode, "Should not return 403 with sufficient roles")

					// Для DELETE операций ожидаем 204, для остальных - 200
					if tc.method == "DELETE" {
						assert.Equal(t, 204, resp.StatusCode, "DELETE should return 204")
					} else {
						assert.Equal(t, 200, resp.StatusCode, "Should return 200 for successful operation")
					}

					svc.AssertExpectations(t)
				})
			}
		})
	}
}

// setupAppWithoutAuth создает приложение БЕЗ middleware аутентификации для тестирования 401 ошибок
func setupAppWithoutAuth(t *testing.T, svc *MockEmployeeService) *fiber.App {
	t.Helper()

	app := fiber.New()
	groupApiV1 := app.Group("/api/v1")

	// НЕ добавляем AuthMiddleware - это позволит тестировать 401 ошибки

	server := &web.Server{
		App:        app,
		GroupApiV1: groupApiV1,
	}

	logger := common.NewTestLogger()
	controller := NewController(server, svc, logger)
	controller.RegisterRoutes()

	return app
}

// createAuthenticatedRequest создает HTTP запрос с валидным токеном аутентификации
func createAuthenticatedRequest(t *testing.T, method, path string, body interface{}, roles []string) *http.Request {
	t.Helper()

	var req *http.Request
	if body != nil {
		jsonBody, err := json.Marshal(body)
		assert.NoError(t, err)
		req = httptest.NewRequest(method, path, bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	// Генерируем валидный тестовый токен с указанными ролями
	token := web.GenerateTestToken(roles)
	req.Header.Set("Authorization", "Bearer "+token)

	return req
}
