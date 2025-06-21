package web

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// RegisterMiddleware — подключает все глобальные middleware
func RegisterMiddleware(app *fiber.App) {
	app.Use(recover.New())
	app.Use(requestid.New())
}
