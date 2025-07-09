package web

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	_ "idm/docs"
	"idm/inner/common"
)

// структура веб-сервера
type Server struct {
	App           *fiber.App
	GroupApi      fiber.Router
	GroupApiV1    fiber.Router
	GroupInternal fiber.Router
	logger        *common.Logger
}

type AuthMiddlewareInterface interface {
	ProtectWithJwt() func(*fiber.Ctx) error
}

// функция-конструктор
func NewServer(logger *common.Logger) *Server {
	// создаём новый веб-сервер
	app := fiber.New()
	// 👉 подключаем middleware
	RegisterMiddleware(app)
	// подключаем swagger
	app.Get("/swagger/*", swagger.HandlerDefault)
	// создаём группы
	groupInternal := app.Group("/internal")
	groupApi := app.Group("/api")

	// Применяем AuthMiddleware к API группе динамически (проверяется при каждом запросе)
	groupApi.Use(func(c *fiber.Ctx) error {
		return CreateAuthMiddleware(logger)(c)
	})

	groupApiV1 := groupApi.Group("/v1")

	return &Server{
		App:           app,
		GroupApi:      groupApi,
		GroupApiV1:    groupApiV1,
		GroupInternal: groupInternal,
		logger:        logger,
	}
}
