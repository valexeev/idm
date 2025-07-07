package web

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	_ "idm/docs"
)

// структура веб-сервера
type Server struct {
	App           *fiber.App
	GroupApi      fiber.Router
	GroupApiV1    fiber.Router
	GroupInternal fiber.Router
}

type AuthMiddlewareInterface interface {
	ProtectWithJwt() func(*fiber.Ctx) error
}

// функция-конструктор
func NewServer() *Server {
	// создаём новый веб-сервер
	app := fiber.New()
	// 👉 подключаем middleware
	RegisterMiddleware(app)
	// подключаем swagger
	app.Get("/swagger/*", swagger.HandlerDefault)
	// создаём группы
	groupInternal := app.Group("/internal")
	groupApi := app.Group("/api")
	groupApiV1 := groupApi.Group("/v1")

	return &Server{
		App:           app,
		GroupApi:      groupApi,
		GroupApiV1:    groupApiV1,
		GroupInternal: groupInternal,
	}
}
