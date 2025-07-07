package idm_test

import (
	"idm/inner/common"
	"idm/inner/common/validator"
	"idm/inner/employee"
	"idm/inner/web"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
)

// helper: middleware для подстановки валидного JWT-токена с ролями IDM_ADMIN и IDM_USER
func addTestAuthMiddleware(app *fiber.App) {
	token := &jwt.Token{Claims: &web.IdmClaims{
		RealmAccess: web.RealmAccessClaims{Roles: []string{web.IdmAdmin, web.IdmUser}},
	}}
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(web.JwtKey, token)
		return c.Next()
	})
}

// setupTestApp создает Fiber-приложение с реальными зависимостями для интеграционных тестов.
func setupTestApp(db *sqlx.DB) *fiber.App {
	app := fiber.New()
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

	// Добавляем middleware авторизации ПЕРЕД регистрацией маршрутов
	addTestAuthMiddleware(app)

	// Создаем сервер и ко��троллер
	server := web.NewServer()
	server.App = app // ВАЖНО: маршруты будут регистрироваться на этом app

	// Создаём группу API V1 (как в реальном сервере)
	server.GroupApiV1 = app.Group("/api/v1")

	employeeController := employee.NewController(server, employeeService, logger)
	employeeController.RegisterRoutes()

	return app
}
