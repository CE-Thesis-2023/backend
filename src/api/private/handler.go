package privateapi

import (
	"net/http"
	"strings"

	"github.com/CE-Thesis-2023/backend/src/biz/service"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	custhttp "github.com/CE-Thesis-2023/backend/src/internal/http"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/events"
	"github.com/CE-Thesis-2023/backend/src/models/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RegisterDevice(ctx *gin.Context) {
	logger.SInfo("RegisterDevice: request")

	var msg events.DeviceRegistrationRequest
	if err := ctx.BindJSON(&msg); err != nil {
		logger.SError("RegisterDevice: request body marshal error", zap.Error(err))
		err := custerror.FormatInvalidArgument(err.Error())
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	if err := service.GetPrivateService().
		RegisterDevice(ctx, &msg); err != nil {
		logger.SError("RegisterDevice: service.RegisterDevice error", zap.Error(err))
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.Status(200)
}

func GetTranscoderAssignedCameras(ctx *gin.Context) {
	logger.SInfo("GetListCameras: request")

	msg := events.UpdateCameraListRequest{}
	transcoderId, found := ctx.Params.Get("id")
	if !found {
		err := custerror.FormatInvalidArgument("missing transcoder_id params")
		custhttp.ToHTTPErr(err, ctx)
		return
	}
	msg.DeviceId = transcoderId
	if len(transcoderId) == 0 {
		err := custerror.FormatInvalidArgument("missing transcoder_id params")
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	resp, err := service.GetPrivateService().UpdateCameraList(ctx, &msg)
	if err != nil {
		logger.SError("GetTranscoderAssignedCameras: service.UpdateCameraList error", zap.Error(err))
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func GetOpenGateCameraSettings(ctx *gin.Context) {
	logger.SInfo("GetOpenGateCameraSettings: request")

	cameraIds := ctx.Query("camera_id")
	if len(cameraIds) == 0 {
		err := custerror.FormatInvalidArgument("missing camera_id as query string")
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	idList := strings.Split(cameraIds, ",")
	req := web.GetOpenGateCameraSettingsRequest{
		CameraId: idList,
	}

	resp, err := service.
		GetWebService().
		GetOpenGateCameraSettings(ctx, &req)
	if err != nil {
		logger.SError("GetOpenGateCameraSettings: service.GetOpenGateCameraSettings error", zap.Error(err))
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func GetOpenGateMqttSettings(ctx *gin.Context) {
	logger.SInfo("GetOpenGateMqttSettings: request")

	id, found := ctx.Params.Get("id")
	if !found {
		err := custerror.FormatInvalidArgument("missing id params")
		custhttp.ToHTTPErr(err, ctx)
		return
	}
	if len(id) == 0 {
		err := custerror.FormatInvalidArgument("missing id params")
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	resp, err := service.
		GetWebService().
		GetOpenGateMqttConfigurationById(ctx, &web.GetOpenGateMqttSettingsRequest{
			ConfigurationId: id,
		})
	if err != nil {
		logger.SError("GetOpenGateMqttSettings: service.GetOpenGateMqttSettings error", zap.Error(err))
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func DeleteTranscoder(ctx *gin.Context) {
	logger.SDebug("DeleteTranscoder: request")

	transcoderId := ctx.Query("id")
	if len(transcoderId) == 0 {
		err := custerror.FormatInvalidArgument("missing transcoderId as query string")
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	err := service.GetPrivateService().DeleteTranscoder(
		ctx,
		&web.DeleteTranscoderRequest{
			DeviceId: transcoderId,
		})
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.Status(http.StatusAccepted)
}

func GetOpenGateIntegrationConfigurations(ctx *gin.Context) {
	logger.SInfo("GetOpenGateIntegrationConfigurations: request")

	openGateId, found := ctx.Params.Get("id")
	if !found {
		err := custerror.FormatInvalidArgument("missing id params")
		custhttp.ToHTTPErr(err, ctx)
		return
	}
	if len(openGateId) == 0 {
		err := custerror.FormatInvalidArgument("missing id params")
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	resp, err := service.
		GetWebService().
		GetOpenGateIntegrationById(
			ctx, &web.GetOpenGateIntegrationByIdRequest{
				OpenGateId: openGateId,
			})
	if err != nil {
		logger.SError("GetOpenGateIntegrationConfigurations: service.GetOpenGateIntegrationConfigurations error", zap.Error(err))
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func GetTranscoderStreamConfigurations(ctx *gin.Context) {
	logger.SInfo("GetTranscoderStreamConfigurations: request")

	cameraIds := ctx.Query("camera_id")
	if len(cameraIds) == 0 {
		err := custerror.FormatInvalidArgument("missing camera_id as query string")
		custhttp.ToHTTPErr(err, ctx)
		return
	}
	ids := strings.Split(cameraIds, ",")
	resp, err := service.GetPrivateService().GetStreamConfigurations(ctx, &web.GetStreamConfigurationsRequest{
		CameraId: ids,
	})
	if err != nil {
		logger.SError("GetTranscoderStreamConfigurations: service.GetTranscoderStreamConfigurations error", zap.Error(err))
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func GetTranscoderOpenGateConfiguration(ctx *gin.Context) {
	logger.SInfo("GetTranscoderOpenGateConfiguration: request")

	id, found := ctx.Params.Get("id")
	if !found {
		err := custerror.FormatInvalidArgument("missing id params")
		custhttp.ToHTTPErr(err, ctx)
		return
	}
	if len(id) == 0 {
		err := custerror.FormatInvalidArgument("missing id params")
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	resp, err := service.GetPrivateService().GetTranscoderOpenGateConfiguration(ctx, &web.GetTranscoderOpenGateConfigurationRequest{
		TranscoderId: id,
	})
	if err != nil {
		logger.SError("GetTranscoderOpenGateConfiguration: service.GetTranscoderOpenGateConfiguration error", zap.Error(err))
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func GetTranscoderMQTTConfigurations(ctx *gin.Context) {
	logger.SInfo("GetTranscoderMQTTConfig: request")

	transcoderId := ctx.Query("transcoder_id")
	if len(transcoderId) == 0 {
		err := custerror.FormatInvalidArgument("missing transcoder_id as query string")
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	resp, err := service.GetPrivateService().GetMQTTEventEndpoint(ctx, &web.GetMQTTEventEndpointRequest{
		TranscoderId: transcoderId,
	})
	if err != nil {
		logger.SError("GetTranscoderMQTTConfig: service.GetTranscoderMQTTConfig error", zap.Error(err))
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}
