package privateapi

import (
	custhttp "github.com/CE-Thesis-2023/backend/src/internal/http"
	"github.com/gofiber/fiber/v2"
)

func ServiceRegistration() func(app *fiber.App) {
	return func(app *fiber.App) {
		app.Use("/", custhttp.SetCors())
		priv := app.Group("/private")
		priv.Post("/registers", RegisterDevice)
		priv.Get("/transcoders/:id/cameras", GetTranscoderAssignedCameras)
	}
}
