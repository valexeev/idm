package idm_test

import (
	"idm/inner/common"
	"idm/inner/common/validator"
	"idm/inner/employee"
	"idm/inner/web"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// setupTestApp создает Fiber-приложение с реальными зависимостями для интеграционных тестов.
func setupTestApp(db *sqlx.DB) *fiber.App {
	// Загружаем тестовую конфигурацию (например, из .env.tests)
	cfg := common.GetConfig(".env.tests")

	// Создаем zap-логгер
	logger := common.NewLogger(cfg)

	// Валидатор
	vld := validator.New()

	// Репозиторий и сервис
	employeeRepo := employee.NewRepository(db)
	// ВАЖНО: не создавать roleRepo, если не используется, чтобы не сбивать схему
	employeeService := employee.NewService(employeeRepo, vld)

	// Создаем сервер и контроллер
	server := web.NewServer()
	employeeController := employee.NewController(server, employeeService, logger)
	employeeController.RegisterRoutes()

	return server.App
}
