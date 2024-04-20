package privateapi

import (
	"github.com/gin-gonic/gin"
)

func ServiceRegistration() func(app *gin.Engine) {
	return func(app *gin.Engine) {
		app.POST("/private/registers", RegisterDevice)
		app.DELETE("/private/deregisters", DeleteTranscoder)
		app.GET("/private/transcoders/:id/cameras", GetTranscoderAssignedCameras)
		app.GET("/private/opengate/cameras", GetOpenGateCameraSettings)
		app.GET("/private/opengate/:id/mqtt", GetOpenGateMqttSettings)
		app.GET("/private/opengate/:id", GetOpenGateIntegrationConfigurations)
		app.GET("/private/transcoders/streams", GetTranscoderStreamConfigurations)
		app.GET("/private/opengate/configurations/:id", GetTranscoderOpenGateConfiguration)
		app.GET("/private/transcoders/mqtt", GetTranscoderMQTTConfigurations)
	}
}
