package role

import (
	"context"
	"errors"
	"idm/inner/common"
	"idm/inner/web"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Controller struct {
	server      *web.Server
	roleService Svc
}

// интерфейс сервиса role.Service
type Svc interface {
	FindById(ctx context.Context, id int64) (Response, error)
	Add(ctx context.Context, name string) (Response, error)
	FindAll(ctx context.Context) ([]Response, error)
	FindByIds(ctx context.Context, ids []int64) ([]Response, error)
	DeleteById(ctx context.Context, id int64) error
	DeleteByIds(ctx context.Context, ids []int64) error
	ValidateRequest(request any) error
}

func NewController(server *web.Server, roleService Svc) *Controller {
	return &Controller{
		server:      server,
		roleService: roleService,
	}
}

// функция для регистрации маршрутов
func (c *Controller) RegisterRoutes() {
	// Маршруты для администраторов (создание, изменение, удаление)
	c.server.GroupApiV1.Post("/roles", web.RequireRoles(web.IdmAdmin), c.CreateRole)
	c.server.GroupApiV1.Delete("/roles/:id", web.RequireRoles(web.IdmAdmin), c.DeleteRole)
	c.server.GroupApiV1.Delete("/roles", web.RequireRoles(web.IdmAdmin), c.DeleteRolesByIds)

	// Маршруты для администраторов и пользователей (чтение)
	c.server.GroupApiV1.Get("/roles/:id", web.RequireRoles(web.IdmAdmin, web.IdmUser), c.GetRole)
	c.server.GroupApiV1.Get("/roles", web.RequireRoles(web.IdmAdmin, web.IdmUser), c.GetAllRoles)
	c.server.GroupApiV1.Post("/roles/by-ids", web.RequireRoles(web.IdmAdmin, web.IdmUser), c.GetRolesByIds)
}

// функция-хендлер, которая будет вызываться при POST запросе по маршруту "/api/v1/roles"
// CreateRole создаёт новую роль
// @Summary Создать новую роль
// @Description Создать новую роль
// @Tags role
// @Accept json
// @Produce json
// @Param request body role.AddRoleRequest true "create role request"
// @Success 200 {object} common.ResponseExample
// @Failure 400 {object} common.ResponseExample
// @Router /roles [post]
func (c *Controller) CreateRole(ctx *fiber.Ctx) error {
	// анмаршалим JSON body запроса в структуру AddRoleRequest
	var request AddRoleRequest
	if err := ctx.BodyParser(&request); err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	if err := c.roleService.ValidateRequest(request); err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	// вызываем метод Add сервиса role.Service
	response, err := c.roleService.Add(ctx.Context(), request.Name)
	if err != nil {
		switch {
		// если сервис возвращает ошибку RequestValidationError или AlreadyExistsError,
		// то мы возвращаем ответ с кодом 400 (BadRequest)
		case errors.As(err, &common.RequestValidationError{}) || errors.As(err, &common.AlreadyExistsError{}):
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		// если сервис возвращает другую ошибку, то мы возвращаем ответ с кодом 500 (InternalServerError)
		default:
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}

	// функция OkResponse() формирует и направляет ответ в случае успеха
	if err = common.OkResponse(ctx, response); err != nil {
		// функция ErrorResponse() формирует и направляет ответ в случае ошибки
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning created role")
	}
	return nil
}

// функция-хендлер для получения роли по ID
// GetRole получает роль по ID
// @Summary Получить роль по ID
// @Description Получить роль по идентификатору
// @Tags role
// @Accept json
// @Produce json
// @Param id path int true "ID роли"
// @Success 200 {object} common.ResponseExample
// @Failure 400 {object} common.ResponseExample
// @Failure 404 {object} common.ResponseExample
// @Router /roles/{id} [get]
func (c *Controller) GetRole(ctx *fiber.Ctx) error {
	// получаем ID из параметра маршрута
	idParam := ctx.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || id <= 0 {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "invalid role id")
	}

	// вызываем метод FindById сервиса role.Service
	response, err := c.roleService.FindById(ctx.Context(), id)
	if err != nil {
		switch {
		case errors.As(err, &common.NotFoundError{}):
			return common.ErrResponse(ctx, fiber.StatusNotFound, err.Error())
		case errors.As(err, &common.RequestValidationError{}):
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}

	if err = common.OkResponse(ctx, response); err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning role")
	}
	return nil
}

// функция-хендлер для получения всех ролей
// GetAllRoles получает список всех ролей
// @Summary Получить все роли
// @Description Получить список всех ролей
// @Tags role
// @Accept json
// @Produce json
// @Success 200 {object} common.ResponseExample
// @Failure 500 {object} common.ResponseExample
// @Router /roles [get]
func (c *Controller) GetAllRoles(ctx *fiber.Ctx) error {
	// вызываем метод FindAll сервиса role.Service
	responses, err := c.roleService.FindAll(ctx.Context())
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}

	// возвращаем успешный ответ
	if err = common.OkResponse(ctx, responses); err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning roles")
	}
	return nil
}

// функция-хендлер для получения ролей по списку ID
// GetRolesByIds получает роли по списку ID
// @Summary Получить роли по списку ID
// @Description Получить роли по списку идентификаторов
// @Tags role
// @Accept json
// @Produce json
// @Param request body role.FindByIdsRequest true "список ID ролей"
// @Success 200 {object} common.ResponseExample
// @Failure 400 {object} common.ResponseExample
// @Router /roles/by-ids [post]
func (c *Controller) GetRolesByIds(ctx *fiber.Ctx) error {
	// анмаршалим JSON body запроса в структуру FindByIdsRequest
	var request FindByIdsRequest
	if err := ctx.BodyParser(&request); err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	if err := c.roleService.ValidateRequest(request); err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	// валидация запроса
	if len(request.Ids) == 0 {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "ids list cannot be empty")
	}

	// вызываем метод FindByIds сервиса role.Service
	responses, err := c.roleService.FindByIds(ctx.Context(), request.Ids)
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}

	// возвращаем успешный ответ
	if err = common.OkResponse(ctx, responses); err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning roles")
	}
	return nil
}

// функция-хендлер для удаления ролей по списку ID
// DeleteRolesByIds удаляет роли по списку ID
// @Summary Удалить роли по списку ID
// @Description Удалить роли по списку идентификаторов
// @Tags role
// @Accept json
// @Produce json
// @Param request body role.DeleteByIdsRequest true "список ID ролей для удаления"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} common.ResponseExample
// @Router /roles [delete]
func (c *Controller) DeleteRolesByIds(ctx *fiber.Ctx) error {
	// анмаршалим JSON body запроса в структуру DeleteByIdsRequest
	var request DeleteByIdsRequest
	if err := ctx.BodyParser(&request); err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	if err := c.roleService.ValidateRequest(request); err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	// валидация запроса
	if len(request.Ids) == 0 {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "ids list cannot be empty")
	}

	// вызываем метод DeleteByIds сервиса role.Service
	err := c.roleService.DeleteByIds(ctx.Context(), request.Ids)
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}

	// возвращаем успешный ответ (статус 204 No Content)
	ctx.Status(fiber.StatusNoContent)
	return nil
}

// функция-хендлер для удаления роли по ID
// DeleteRole удаляет роль по ID
// @Summary Удалить роль по ID
// @Description Удалить роль по идентификатору
// @Tags role
// @Accept json
// @Produce json
// @Param id path int true "ID роли"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} common.ResponseExample
// @Failure 404 {object} common.ResponseExample
// @Router /roles/{id} [delete]
func (c *Controller) DeleteRole(ctx *fiber.Ctx) error {
	// получаем ID из параметра маршрута
	idParam := ctx.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || id <= 0 {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "invalid role id")
	}

	// вызываем метод DeleteById сервиса role.Service
	err = c.roleService.DeleteById(ctx.Context(), id)
	if err != nil {
		switch {
		case errors.As(err, &common.NotFoundError{}):
			return common.ErrResponse(ctx, fiber.StatusNotFound, err.Error())
		case errors.As(err, &common.RequestValidationError{}):
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}

	return ctx.SendStatus(204)
}
