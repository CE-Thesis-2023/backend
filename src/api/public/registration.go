package publicapi

import (
	"time"

	custhttp "github.com/CE-Thesis-2023/backend/src/internal/http"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/timeout"
)

func ServiceRegistration() func(app *fiber.App) {
	return func(app *fiber.App) {
		app.Use("/", custhttp.SetCors())

		app.Get("/api/devices", GetTranscoderDevices)
		app.Put("/api/devices", UpdateTranscoder)

		app.Get("/api/cameras", GetCameras)
		app.Post("/api/cameras", CreateCamera)
		app.Delete("/api/cameras", DeleteCamera)

		app.Get("/api/opengate/:openGateId", GetOpenGateSettings)
		app.Put("/api/opengate/:openGateId", UpdateOpenGateSettings)

		app.Get("/api/groups/:id/cameras", GetCamerasByGroupId)
		app.Get("/api/groups", GetCameraGroups)
		app.Post("/api/groups", AddCameraGroup)
		app.Delete("/api/groups", DeleteCameraGroup)
		app.Post("/api/groups/cameras", AddCamerasToGroup)
		app.Delete("/api/groups/cameras", RemoveCamerasFromGroup)

		app.Get("/api/cameras/:id/streams", GetStreamInfo)
		app.Put("/api/cameras/:id/streams", ToggleStream)
		app.Post("/api/rc", RemoteControl)
		app.Get("/api/cameras/:cameraId/info", timeout.NewWithContext(
			GetCameraDeviceInfo, time.Second*3))

		app.Get("/api/events/object_tracking", GetObjectTrackingEvent)
		app.Delete("/api/events/object_tracking", DeleteObjectTrackingEvent)
	}
}
