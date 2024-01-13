package publicapi

import (
	"github.com/CE-Thesis-2023/backend/src/biz/handlers"
	"github.com/gofiber/fiber/v2"
)

func ServiceRegistration() func(app *fiber.App) {
	return func(app *fiber.App) {
		handlers.WebsocketInit(
			handlers.WithAuthorizer(nil),
			handlers.WithConnectionHandler(nil),
			handlers.WithChannelSize(128),
		)
		communicator := handlers.GetWebsocketCommunicator()

		app.Get("/api/devices", GetTranscoderDevices)
		app.Put("/api/devices", UpdateTranscoder)

		app.Get("/api/cameras", GetCameras)
		app.Post("/api/cameras", CreateCamera)
		app.Delete("/api/cameras", DeleteCamera)

		app.Get("/api/cameras/:id/streams", GetStreamInfo)
		app.Put("/api/cameras/:id/streams", ToggleStream)
		app.Post("/api/rc", RemoteControl)

		app.Get("/cameras/:id/viewer", HtmlVideoPlayer)
		app.Use("/ws/ltd", communicator.HandleRegisterRequest)
		app.Get("/ws/ltd/:id", communicator.CreateWebsocketHandler())
	}
}
