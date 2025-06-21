package web

import "github.com/gofiber/fiber/v2"

// структура веб-сервера
type Server struct {
	App           *fiber.App
	GroupApiV1    fiber.Router
	GroupInternal fiber.Router
}

// функция-конструктор
func NewServer() *Server {

	// создаём новый веб-сервер
	app := fiber.New()

	// 👉 подключаем middleware
	RegisterMiddleware(app)

	// создаём группы
	groupInternal := app.Group("/internal")
	groupApi := app.Group("/api")
	groupApiV1 := groupApi.Group("/v1")

	return &Server{
		App:           app,
		GroupApiV1:    groupApiV1,
		GroupInternal: groupInternal,
	}
}
