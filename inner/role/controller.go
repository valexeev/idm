package role

import (
	"errors"
	"idm/inner/common"
	"idm/inner/web"
	"strconv"

	"github.com/gofiber/fiber"
)

type Controller struct {
	server      *web.Server
	roleService Svc
}

// интерфейс сервиса role.Service
type Svc interface {
	FindById(id int64) (Response, error)
	Add(name string) (Response, error)
	FindAll() ([]Response, error)
	FindByIds(ids []int64) ([]Response, error)
	DeleteById(id int64) error
	DeleteByIds(ids []int64) error
}

func NewController(server *web.Server, roleService Svc) *Controller {
	return &Controller{
		server:      server,
		roleService: roleService,
	}
}

// функция для регистрации маршрутов
func (c *Controller) RegisterRoutes() {
	// полный маршрут получится "/api/v1/roles"
	c.server.GroupApiV1.Post("/roles", c.CreateRole)
	c.server.GroupApiV1.Get("/roles/:id", c.GetRole)
	c.server.GroupApiV1.Get("/roles", c.GetAllRoles)
	c.server.GroupApiV1.Post("/roles/by-ids", c.GetRolesByIds)
	c.server.GroupApiV1.Delete("/roles/:id", c.DeleteRole)
	c.server.GroupApiV1.Delete("/roles", c.DeleteRolesByIds)
}

// функция-хендлер, которая будет вызываться при POST запросе по маршруту "/api/v1/roles"
func (c *Controller) CreateRole(ctx *fiber.Ctx) {
	// анмаршалим JSON body запроса в структуру AddRoleRequest
	var request AddRoleRequest
	if err := ctx.BodyParser(&request); err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		return
	}

	// вызываем метод Add сервиса role.Service
	response, err := c.roleService.Add(request.Name)
	if err != nil {
		switch {
		// если сервис возвращает ошибку RequestValidationError или AlreadyExistsError,
		// то мы возвращаем ответ с кодом 400 (BadRequest)
		case errors.As(err, &common.RequestValidationError{}) || errors.As(err, &common.AlreadyExistsError{}):
			_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		// если сервис возвращает другую ошибку, то мы возвращаем ответ с кодом 500 (InternalServerError)
		default:
			_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
		return
	}

	// функция OkResponse() формирует и направляет ответ в случае успеха
	if err = common.OkResponse(ctx, response); err != nil {
		// функция ErrorResponse() формирует и направляет ответ в случае ошибки
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning created role")
		return
	}
}

// функция-хендлер для получения роли по ID
func (c *Controller) GetRole(ctx *fiber.Ctx) {
	// получаем ID из параметра маршрута
	idParam := ctx.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, "invalid role id")
		return
	}

	// вызываем метод FindById сервиса role.Service
	response, err := c.roleService.FindById(id)
	if err != nil {
		switch {
		case errors.As(err, &common.NotFoundError{}):
			_ = common.ErrResponse(ctx, fiber.StatusNotFound, err.Error())
		default:
			_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
		return
	}

	// возвращаем успешный ответ
	if err = common.OkResponse(ctx, response); err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning role")
		return
	}
}

// функция-хендлер для получения всех ролей
func (c *Controller) GetAllRoles(ctx *fiber.Ctx) {
	// вызываем метод FindAll сервиса role.Service
	responses, err := c.roleService.FindAll()
	if err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		return
	}

	// возвращаем успешный ответ
	if err = common.OkResponse(ctx, responses); err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning roles")
		return
	}
}

// функция-хендлер для получения ролей по списку ID
func (c *Controller) GetRolesByIds(ctx *fiber.Ctx) {
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

	// вызываем метод FindByIds сервиса role.Service
	responses, err := c.roleService.FindByIds(request.Ids)
	if err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		return
	}

	// возвращаем успешный ответ
	if err = common.OkResponse(ctx, responses); err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning roles")
		return
	}
}

// функция-хендлер для удаления ролей по списку ID
func (c *Controller) DeleteRolesByIds(ctx *fiber.Ctx) {
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

	// вызываем метод DeleteByIds сервиса role.Service
	err := c.roleService.DeleteByIds(request.Ids)
	if err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		return
	}

	// возвращаем успешный ответ (статус 204 No Content)
	ctx.Status(fiber.StatusNoContent)
}

// функция-хендлер для удаления роли по ID
func (c *Controller) DeleteRole(ctx *fiber.Ctx) {
	// получаем ID из параметра маршрута
	idParam := ctx.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, "invalid role id")
		return
	}

	// вызываем метод DeleteById сервиса role.Service
	err = c.roleService.DeleteById(id)
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
