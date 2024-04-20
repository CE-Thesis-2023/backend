package privateapi

import (
	"github.com/gofiber/fiber/v2"
)

func ServiceRegistration() func(app *fiber.App) {
	return func(app *fiber.App) {
		app.Post("/private/registers", RegisterDevice)
		app.Delete("/private/deregisters", DeleteTranscoder)
		app.Get("/private/transcoders/:id/cameras", GetTranscoderAssignedCameras)
		app.Get("/private/opengate/cameras", GetOpenGateCameraSettings)
		app.Get("/private/opengate/:id/mqtt", GetOpenGateMqttSettings)
		app.Get("/private/opengate/:id", GetOpenGateIntegrationConfigurations)
		app.Get("/private/transcoders/streams", GetTranscoderStreamConfigurations)
		app.Get("/private/opengate/configurations/:id", GetTranscoderOpenGateConfiguration)
		app.Get("/private/transcoders/mqtt", GetTranscoderMQTTConfigurations)
	}
}
