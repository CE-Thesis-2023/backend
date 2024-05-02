package publicapi

import (
	"github.com/gin-gonic/gin"
)

func ServiceRegistration() func(app *gin.Engine) {
	return func(app *gin.Engine) {
		app.GET("/api/devices", GetTranscoderDevices)
		app.PUT("/api/devices", UpdateTranscoder)
		app.GET("/api/devices/status", GetTranscoderStatus)
		app.GET("/api/devices/healthcheck", DoDeviceHealthcheck)

		app.GET("/api/cameras", GetCameras)
		app.POST("/api/cameras", CreateCamera)
		app.DELETE("/api/cameras", DeleteCamera)

		app.GET("/api/opengate/:openGateId", GetOpenGateSettings)

		app.GET("/api/cameras/:id/streams", GetStreamInfo)
		app.PUT("/api/cameras/:id/streams", ToggleStream)
		app.POST("/api/rc", RemoteControl)
		app.GET("/api/cameras/info/:cameraId", GetCameraDeviceInfo)

		app.GET("/api/events/object_tracking", GetObjectTrackingEvent)
		app.DELETE("/api/events/object_tracking", DeleteObjectTrackingEvent)
		app.GET("/api/snapshots", GetSnapshotPresignedUrl)

		app.GET("/api/people", GetDetectablePeople)
		app.POST("/api/people", AddDetectablePerson)
		app.DELETE("/api/people", DeleteDetectablePerson)
		app.GET("/api/people/presigned", GetDetectablePersonPresignedUrl)
		app.GET("/api/people/history", GetPersonHistory)

		app.GET("/api/stats", GetLatestOpenGateCameraStats)
	}
}
