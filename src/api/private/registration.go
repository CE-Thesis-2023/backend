package privateapi

import "github.com/gofiber/fiber/v2"

func ServiceRegistration() func(app *fiber.App) {
	return func(app *fiber.App) {
		priv := app.Group("/api/priv")
		priv.Post("/registers", RegisterDevice)
		priv.Get("/transcoders/:id/cameras", GetTranscoderAssignedCameras)
	}
}
