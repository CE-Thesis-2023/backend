package publicapi

import (
	"github.com/gin-gonic/gin"
)

func ServiceRegistration() func(app *gin.Engine) {
	return func(app *gin.Engine) {
		app.GET("/api/devices", GetTranscoderDevices)
		app.PUT("/api/devices", UpdateTranscoder)
		app.GET("/api/devices/healthcheck", DoDeviceHealthcheck)

		app.GET("/api/cameras", GetCameras)
		app.POST("/api/cameras", CreateCamera)
		app.DELETE("/api/cameras", DeleteCamera)

		app.GET("/api/opengate/:openGateId", GetOpenGateSettings)
		app.PUT("/api/opengate/:openGateId", UpdateOpenGateSettings)

		app.GET("/api/groups/:id/cameras", GetCamerasByGroupId)
		app.GET("/api/groups", GetCameraGroups)
		app.POST("/api/groups", AddCameraGroup)
		app.DELETE("/api/groups", DeleteCameraGroup)
		app.POST("/api/groups/cameras", AddCamerasToGroup)
		app.DELETE("/api/groups/cameras", RemoveCamerasFromGroup)

		app.GET("/api/cameras/:id/streams", GetStreamInfo)
		app.PUT("/api/cameras/:id/streams", ToggleStream)
		app.POST("/api/rc", RemoteControl)
		app.GET("/api/cameras/info/:cameraId", GetCameraDeviceInfo)

		app.GET("/api/events/object_tracking", GetObjectTrackingEvent)
		app.DELETE("/api/events/object_tracking", DeleteObjectTrackingEvent)

		app.GET("/api/people", GetDetectablePeople)
		app.POST("/api/people", AddDetectablePerson)
		app.DELETE("/api/people", DeleteDetectablePerson)
		app.GET("/api/people/presigned", GetDetectablePersonPresignedUrl)

		app.GET("/api/stats", GetLatestOpenGateCameraStats)

		app.GET("api/snapshot/:id", GetSnapshotPresignedUrl)
	}
}
