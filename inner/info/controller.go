package info

import (
	"context"
	"idm/inner/common"
	"idm/inner/web"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Database интерфейс для работы с базой данных
type Database interface {
	PingContext(ctx context.Context) error
}

type Controller struct {
	server *web.Server
	cfg    common.Config
	db     Database // используем интерфейс вместо конкретного типа
}

func NewController(server *web.Server, cfg common.Config, database Database) *Controller {
	return &Controller{
		server: server,
		cfg:    cfg,
		db:     database,
	}
}

type InfoResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Message  string `json:"message,omitempty"`
}

func (c *Controller) RegisterRoutes() {
	// полный путь будет "/internal/info"
	c.server.GroupInternal.Get("/info", c.GetInfo)
	// полный путь будет "/internal/health"
	c.server.GroupInternal.Get("/health", c.GetHealth)
}

// GetInfo получение информации о приложении
func (c *Controller) GetInfo(ctx *fiber.Ctx) error {
	if err := ctx.Status(fiber.StatusOK).JSON(&InfoResponse{
		Name:    c.cfg.AppName,
		Version: c.cfg.AppVersion,
	}); err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning info")
	}
	return nil
}

// GetHealth проверка работоспособности приложения
func (c *Controller) GetHealth(ctx *fiber.Ctx) error {
	// Создаем контекст с таймаутом для проверки БД
	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Проверяем подключение к базе данных
	if err := c.db.PingContext(dbCtx); err != nil {
		// База данных недоступна - возвращаем 500 для перезапуска
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, "Database connection failed")
	}

	// Все проверки прошли успешно
	return ctx.Status(fiber.StatusOK).SendString("OK")

}

// GetHealthDetailed возвращает детальную информацию о состоянии системы
// Можно использовать для более подробной диагностики
func (c *Controller) GetHealthDetailed(ctx *fiber.Ctx) error {
	// Создаем контекст с таймаутом для проверки БД
	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Проверяем подключение к базе данных
	dbStatus := "ok"
	var healthStatus string
	var statusCode int
	if err := c.db.PingContext(dbCtx); err != nil {
		dbStatus = "error"
		healthStatus = "unhealthy"
		statusCode = fiber.StatusInternalServerError
	} else {
		healthStatus = "healthy"
		statusCode = fiber.StatusOK
	}

	if err := ctx.Status(statusCode).JSON(&HealthResponse{
		Status:   healthStatus,
		Database: dbStatus,
	}); err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning health status")
	}
	return nil
}
