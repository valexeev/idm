package employee

import (
	"errors"
	"idm/inner/common"
	"idm/inner/web"
	"strconv"

	"github.com/gofiber/fiber"
)

type Controller struct {
	server          *web.Server
	employeeService Svc
}

// интерфейс сервиса employee.Service
type Svc interface {
	FindById(id int64) (Response, error)
	AddTransactional(request AddEmployeeRequest) (Response, error)
	Add(name string) (Response, error)
	FindAll() ([]Response, error)
	FindByIds(ids []int64) ([]Response, error)
	DeleteById(id int64) error
	DeleteByIds(ids []int64) error
}

func NewController(server *web.Server, employeeService Svc) *Controller {
	return &Controller{
		server:          server,
		employeeService: employeeService,
	}
}

// функция для регистрации маршрутов
func (c *Controller) RegisterRoutes() {
	// CRUD операции для employees
	c.server.GroupApiV1.Post("/employees", c.CreateEmployee)
	c.server.GroupApiV1.Post("/employees/transactional", c.CreateEmployeeTransactional)
	c.server.GroupApiV1.Get("/employees/:id", c.GetEmployee)
	c.server.GroupApiV1.Get("/employees", c.GetAllEmployees)
	c.server.GroupApiV1.Post("/employees/by-ids", c.GetEmployeesByIds)
	c.server.GroupApiV1.Delete("/employees/:id", c.DeleteEmployee)
	c.server.GroupApiV1.Delete("/employees", c.DeleteEmployeesByIds)
}

// функция-хендлер для создания сотрудника (транзакционно)
func (c *Controller) CreateEmployeeTransactional(ctx *fiber.Ctx) {
	// анмаршалим JSON body запроса в структуру AddEmployeeRequest
	var request AddEmployeeRequest
	if err := ctx.BodyParser(&request); err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		return
	}

	// вызываем метод AddTransactional сервиса employee.Service
	response, err := c.employeeService.AddTransactional(request)
	if err != nil {
		switch {
		// если сервис возвращает ошибку RequestValidationError или AlreadyExistsError,
		// то мы возвращаем ответ с кодом 400 (BadRequest)
		case errors.As(err, &common.RequestValidationError{}) || errors.As(err, &common.AlreadyExistsError{}):
			_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		// если сервис возвращает ошибку TransactionError,
		// то мы возвращаем ответ с кодом 500 (InternalServerError)
		case errors.As(err, &common.TransactionError{}):
			_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		// если сервис возвращает другую ошибку, то мы возвращаем ответ с кодом 500 (InternalServerError)
		default:
			_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
		return
	}

	// функция OkResponse() формирует и направляет ответ в случае успеха
	if err = common.OkResponse(ctx, response); err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning created employee")
		return
	}
}

// функция-хендлер для создания сотрудника (простая версия)
func (c *Controller) CreateEmployee(ctx *fiber.Ctx) {
	// анмаршалим JSON body запроса в структуру AddEmployeeRequest
	var request AddEmployeeRequest
	if err := ctx.BodyParser(&request); err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		return
	}

	// вызываем метод Add сервиса employee.Service
	response, err := c.employeeService.Add(request.Name)
	if err != nil {
		switch {
		// если сервис возвращает ошибку RequestValidationError,
		// то мы возвращаем ответ с кодом 400 (BadRequest)
		case errors.As(err, &common.RequestValidationError{}):
			_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		// если сервис возвращает другую ошибку, то мы возвращаем ответ с кодом 500 (InternalServerError)
		default:
			_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
		return
	}

	// функция OkResponse() формирует и направляет ответ в случае успеха
	if err = common.OkResponse(ctx, response); err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning created employee")
		return
	}
}

// функция-хендлер для получения сотрудника по ID
func (c *Controller) GetEmployee(ctx *fiber.Ctx) {
	// получаем ID из параметра маршрута
	idParam := ctx.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, "invalid employee id")
		return
	}

	// вызываем метод FindById сервиса employee.Service
	response, err := c.employeeService.FindById(id)
	if err != nil {
		switch {
		case errors.As(err, &common.RequestValidationError{}):
			_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		case errors.As(err, &common.RepositoryError{}):
			_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		case errors.As(err, &common.NotFoundError{}):
			_ = common.ErrResponse(ctx, fiber.StatusNotFound, err.Error())
		default:
			_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
		return
	}

	// возвращаем успешный ответ
	if err = common.OkResponse(ctx, response); err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning employee")
		return
	}
}

// функция-хендлер для получения всех сотрудников
func (c *Controller) GetAllEmployees(ctx *fiber.Ctx) {
	// вызываем метод FindAll сервиса employee.Service
	responses, err := c.employeeService.FindAll()
	if err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		return
	}

	// возвращаем успешный ответ
	if err = common.OkResponse(ctx, responses); err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning employees")
		return
	}
}

// функция-хендлер для получения сотрудников по списку ID
func (c *Controller) GetEmployeesByIds(ctx *fiber.Ctx) {
	// анмаршалим JSON body запроса в структуру FindByIdsRequest
	var request FindByIdsRequest
	if err := ctx.BodyParser(&request); err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		return
	}

	// валидация запроса
	if len(request.Ids) == 0 {
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, "ids list cannot be empty")
		return
	}

	// вызываем метод FindByIds сервиса employee.Service
	responses, err := c.employeeService.FindByIds(request.Ids)
	if err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		return
	}

	// возвращаем успешный ответ
	if err = common.OkResponse(ctx, responses); err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning employees")
		return
	}
}

// функция-хендлер для удаления сотрудника по ID
func (c *Controller) DeleteEmployee(ctx *fiber.Ctx) {
	// получаем ID из параметра маршрута
	idParam := ctx.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, "invalid employee id")
		return
	}

	// вызываем метод DeleteById сервиса employee.Service
	err = c.employeeService.DeleteById(id)
	if err != nil {
		switch {
		case errors.As(err, &common.NotFoundError{}):
			_ = common.ErrResponse(ctx, fiber.StatusNotFound, err.Error())
		default:
			_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
		return
	}

	// возвращаем успешный ответ (статус 204 No Content)
	ctx.Status(fiber.StatusNoContent)
}

// функция-хендлер для удаления сотрудников по списку ID
func (c *Controller) DeleteEmployeesByIds(ctx *fiber.Ctx) {
	// анмаршалим JSON body запроса в структуру DeleteByIdsRequest
	var request DeleteByIdsRequest
	if err := ctx.BodyParser(&request); err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		return
	}

	// валидация запроса
	if len(request.Ids) == 0 {
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, "ids list cannot be empty")
		return
	}

	// вызываем метод DeleteByIds сервиса employee.Service
	err := c.employeeService.DeleteByIds(request.Ids)
	if err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		return
	}

	// возвращаем успешный ответ (статус 204 No Content)
	ctx.Status(fiber.StatusNoContent)
}
