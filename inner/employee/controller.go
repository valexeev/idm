package employee

import (
	"context"
	"errors"
	"fmt"
	"idm/inner/common"
	"idm/inner/web"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Controller структура контроллера для работы с сотрудниками
type Controller struct {
	server          *web.Server // экземпляр веб-сервера
	employeeService Svc         // сервис для работы с сотрудниками
	logger          *common.Logger
}

// Svc интерфейс сервиса для работы с сотрудниками
type Svc interface {
	FindById(ctx context.Context, id int64) (Response, error)                           // поиск сотрудника по ID
	AddTransactional(ctx context.Context, request AddEmployeeRequest) (Response, error) // добавление сотрудника в транзакции
	Add(ctx context.Context, name string) (Response, error)                             // простое добавление сотрудника
	FindAll(ctx context.Context) ([]Response, error)                                    // получение всех сотрудников
	FindByIds(ctx context.Context, ids []int64) ([]Response, error)                     // поиск сотрудников по списку ID
	DeleteById(ctx context.Context, id int64) error                                     // удаление сотрудника по ID
	DeleteByIds(ctx context.Context, ids []int64) error                                 // удаление сотрудников по списку ID
	ValidateRequest(request interface{}) error

	FindPage(ctx context.Context, req PageRequest) (PageResponse, error)
}

// NewController создает новый экземпляр контроллера сотрудников
func NewController(server *web.Server, employeeService Svc, logger *common.Logger) *Controller {
	return &Controller{server: server, employeeService: employeeService, logger: logger}
}

// RegisterRoutes регистрирует все маршруты для работы с сотрудниками
func (c *Controller) RegisterRoutes() {
	api := c.server.GroupApiV1 // группа маршрутов API v1

	// CRUD операции для сотрудников
	api.Post("/employees", c.CreateEmployee)                            // создание сотрудника
	api.Post("/employees/transactional", c.CreateEmployeeTransactional) // создание сотрудника в транзакции
	api.Get("/employees/page", c.GetEmployeesPage)
	api.Get("/employees/:id", c.GetEmployee)           // получение сотрудника по ID
	api.Get("/employees", c.GetAllEmployees)           // получение всех сотрудников
	api.Post("/employees/by-ids", c.GetEmployeesByIds) // получение сотрудников по списку ID
	api.Delete("/employees/:id", c.DeleteEmployee)     // удаление сотрудника по ID
	api.Delete("/employees", c.DeleteEmployeesByIds)   // удаление сотрудников по списку ID
}

// GetEmployeesPage получает страницу сотрудников
// GET /api/v1/employees/page?pageNumber=x&pageSize=y
func (c *Controller) GetEmployeesPage(ctx *fiber.Ctx) error {
	pageNumber, err := strconv.Atoi(ctx.Query("pageNumber", "0"))
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "invalid pageNumber")
	}
	pageSize, err := strconv.Atoi(ctx.Query("pageSize", "20"))
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "invalid pageSize")
	}
	textFilter := ctx.Query("textFilter", "")
	// Формируем запрос с фильтром
	req := PageRequest{PageNumber: pageNumber, PageSize: pageSize, TextFilter: textFilter}
	resp, err := c.employeeService.FindPage(context.Background(), req)
	if err != nil {
		fmt.Println("ERROR in GetEmployeesPage:", err)
		return handleError(ctx, err)
	}
	return common.OkResponse(ctx, resp)
}

// handleError централизованная обработка ошибок с соответствующими HTTP статусами
func handleError(ctx *fiber.Ctx, err error) error {
	switch {
	// Ошибки валидации и дублирования - 400 Bad Request
	case errors.As(err, &common.RequestValidationError{}),
		errors.As(err, &common.AlreadyExistsError{}):
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	// Ошибки транзакций и репозитория - 500 Internal Server Error
	case errors.As(err, &common.TransactionError{}),
		errors.As(err, &common.RepositoryError{}):
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	// Ошибки отсутствия данных - 404 Not Found
	case errors.As(err, &common.NotFoundError{}):
		return common.ErrResponse(ctx, fiber.StatusNotFound, err.Error())
	// Все остальные ошибки - 500 Internal Server Error
	default:
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
}

// CreateEmployeeTransactional создает нового сотрудника в рамках транзакции
// POST /api/v1/employees/transactional
func (c *Controller) CreateEmployeeTransactional(ctx *fiber.Ctx) error {
	var req AddEmployeeRequest

	// Парсинг JSON
	if err := ctx.BodyParser(&req); err != nil {
		c.logger.Error("create employee transactional: invalid JSON", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	// Логируем тело запроса
	c.logger.Debug("create employee transactional: received request", zap.Any("request", req))

	// Вызов сервиса
	resp, err := c.employeeService.AddTransactional(ctx.Context(), req)
	if err != nil {
		c.logger.Error("create employee transactional: failed to add employee", zap.Error(err))
		return handleError(ctx, err)
	}

	// Ответ
	if err := common.OkResponse(ctx, resp); err != nil {
		c.logger.Error("create employee transactional: failed to return response", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning created employee")
	}

	return nil
}

// CreateEmployee создает нового сотрудника (без транзакции)
// POST /api/v1/employees
func (c *Controller) CreateEmployee(ctx *fiber.Ctx) error {
	var req AddEmployeeRequest

	// Парсинг JSON
	if err := ctx.BodyParser(&req); err != nil {
		c.logger.Error("create employee: invalid JSON", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	// Логируем тело запроса
	c.logger.Debug("create employee: received request", zap.Any("request", req))

	// Вызов сервиса
	resp, err := c.employeeService.Add(ctx.Context(), req.Name)
	if err != nil {
		c.logger.Error("create employee: failed to add employee", zap.Error(err))

		if errors.As(err, &common.RequestValidationError{}) {
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		} else {
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}

	// Ответ
	if err := common.OkResponse(ctx, resp); err != nil {
		c.logger.Error("create employee: failed to return response", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning created employee")
	}

	return nil
}

// GetEmployee получает сотрудника по его ID
// GET /api/v1/employees/:id
func (c *Controller) GetEmployee(ctx *fiber.Ctx) error {
	// Извлечение и парсинг ID из параметров маршрута
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "invalid employee id")
	}

	// Поиск сотрудника по ID через сервис
	resp, err := c.employeeService.FindById(ctx.Context(), id)
	if err != nil {
		return handleError(ctx, err)
	}

	// Возврат данных найденного сотрудника
	if err := common.OkResponse(ctx, resp); err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning employee")
	}
	return nil
}

// GetAllEmployees получает список всех сотрудников
// GET /api/v1/employees
func (c *Controller) GetAllEmployees(ctx *fiber.Ctx) error {
	// Получение всех сотрудников через сервис
	resp, err := c.employeeService.FindAll(ctx.Context())
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}

	// Возврат списка всех сотрудников
	if err := common.OkResponse(ctx, resp); err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning employees")
	}
	return nil
}

// GetEmployeesByIds получает сотрудников по списку ID
// POST /api/v1/employees/by-ids
func (c *Controller) GetEmployeesByIds(ctx *fiber.Ctx) error {
	// Парсинг JSON тела запроса в структуру FindByIdsRequest
	var req FindByIdsRequest
	if err := ctx.BodyParser(&req); err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	if err := c.employeeService.ValidateRequest(req); err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	// Валидация: список ID не должен быть пустым
	if len(req.Ids) == 0 {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "ids list cannot be empty")
	}

	// Поиск сотрудников по списку ID через сервис
	resp, err := c.employeeService.FindByIds(ctx.Context(), req.Ids)
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}

	// Возврат найденных сотрудников
	if err := common.OkResponse(ctx, resp); err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning employees")
	}
	return nil
}

// DeleteEmployee удаляет сотрудника по его ID
// DELETE /api/v1/employees/:id
func (c *Controller) DeleteEmployee(ctx *fiber.Ctx) error {
	// Извлечение и парсинг ID из параметров маршрута
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "invalid employee id")
	}

	// Удаление сотрудника по ID через сервис
	if err = c.employeeService.DeleteById(ctx.Context(), id); err != nil {
		// Специальная обработка для случая "сотрудник не найден"
		if errors.As(err, &common.NotFoundError{}) {
			return common.ErrResponse(ctx, fiber.StatusNotFound, err.Error())
		} else {
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}

	// Возврат статуса 204 No Content при успешном удалении
	ctx.Status(fiber.StatusNoContent)
	return nil
}

// DeleteEmployeesByIds удаляет сотрудников по списку ID
// DELETE /api/v1/employees
func (c *Controller) DeleteEmployeesByIds(ctx *fiber.Ctx) error {
	// Парсинг JSON тела запроса в структуру DeleteByIdsRequest
	var req DeleteByIdsRequest
	if err := ctx.BodyParser(&req); err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	if err := c.employeeService.ValidateRequest(req); err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	// Валидация: список ID не должен быть пустым
	if len(req.Ids) == 0 {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "ids list cannot be empty")
	}

	// Удаление сотрудников по списку ID через сервис
	if err := c.employeeService.DeleteByIds(ctx.Context(), req.Ids); err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}

	// Возврат статуса 204 No Content при успешном удалении
	ctx.Status(fiber.StatusNoContent)
	return nil
}
