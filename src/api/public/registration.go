package publicapi

import (
	"time"

	"github.com/CE-Thesis-2023/backend/src/biz/handlers"
	custhttp "github.com/CE-Thesis-2023/backend/src/internal/http"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/timeout"
)

func ServiceRegistration() func(app *fiber.App) {
	return func(app *fiber.App) {
		handlers.WebsocketInit(
			handlers.WithAuthorizer(WsAuthorizeLtd()),
			handlers.WithConnectionHandler(WsListenToMessages()),
		)
		communicator := handlers.GetWebsocketCommunicator()
		app.Use("/", custhttp.SetCors())

		app.Get("/api/devices", GetTranscoderDevices)
		app.Put("/api/devices", UpdateTranscoder)

		app.Get("/api/cameras", GetCameras)
		app.Post("/api/cameras", CreateCamera)
		app.Delete("/api/cameras", DeleteCamera)

		app.Get("/api/cameras/:id/streams", GetStreamInfo)
		app.Put("/api/cameras/:id/streams", ToggleStream)
		app.Post("/api/rc", RemoteControl)
		app.Get("/api/cameras/:cameraId/info", timeout.NewWithContext(
			GetCameraDeviceInfo, time.Second*3))

		app.Use("/ws/ltd/:id", communicator.HandleRegisterRequest)
		app.Get("/ws/ltd/:id", communicator.CreateWebsocketHandler())
	}
}
