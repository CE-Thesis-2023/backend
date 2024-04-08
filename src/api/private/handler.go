package privateapi

import (
	"net/http"
	"strings"

	"github.com/CE-Thesis-2023/backend/src/biz/service"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/db"
	"github.com/CE-Thesis-2023/backend/src/models/events"
	"github.com/CE-Thesis-2023/backend/src/models/web"

	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func RegisterDevice(ctx *fiber.Ctx) error {
	logger.SInfo("RegisterDevice: request")

	var msg events.DeviceRegistrationRequest
	if err := json.Unmarshal(ctx.Body(), &msg); err != nil {
		logger.SError("RegisterDevice: request body marshal error", zap.Error(err))
		return custerror.FormatInvalidArgument(err.Error())
	}

	if err := service.GetPrivateService().
		RegisterDevice(ctx.Context(), &msg); err != nil {
		logger.SError("RegisterDevice: service.RegisterDevice error", zap.Error(err))
		return err
	}

	return ctx.SendStatus(200)
}

func GetTranscoderAssignedCameras(ctx *fiber.Ctx) error {
	logger.SInfo("GetListCameras: request")

	msg := events.UpdateCameraListRequest{}
	transcoderId := ctx.Params("id")
	msg.DeviceId = transcoderId
	if len(transcoderId) == 0 {
		return custerror.FormatInvalidArgument("missing transcoder_id params")
	}

	resp, err := service.GetPrivateService().UpdateCameraList(ctx.Context(), &msg)
	if err != nil {
		logger.SError("GetTranscoderAssignedCameras: service.UpdateCameraList error", zap.Error(err))
		return err
	}

	return ctx.JSON(resp)
}

func GetOpenGateCameraSettings(ctx *fiber.Ctx) error {
	logger.SInfo("GetOpenGateCameraSettings: request")

	cameraIds := ctx.Query("camera_id")
	if len(cameraIds) == 0 {
		return ctx.JSON(web.GetOpenGateCameraSettingsResponse{
			OpenGateCameraSettings: []db.OpenGateCameraSettings{},
		})
	}

	idList := strings.Split(cameraIds, ",")
	req := web.GetOpenGateCameraSettingsRequest{
		CameraId: idList,
	}

	resp, err := service.
		GetWebService().
		GetOpenGateCameraSettings(ctx.Context(), &req)
	if err != nil {
		logger.SError("GetOpenGateCameraSettings: service.GetOpenGateCameraSettings error", zap.Error(err))
		return err
	}

	return ctx.JSON(resp)
}

func GetOpenGateMqttSettings(ctx *fiber.Ctx) error {
	logger.SInfo("GetOpenGateMqttSettings: request")

	id := ctx.Params("id")
	if len(id) == 0 {
		return custerror.FormatInvalidArgument("missing id params")
	}

	resp, err := service.
		GetWebService().
		GetOpenGateMqttConfigurationById(ctx.Context(), &web.GetOpenGateMqttSettingsRequest{
			ConfigurationId: id,
		})
	if err != nil {
		logger.SError("GetOpenGateMqttSettings: service.GetOpenGateMqttSettings error", zap.Error(err))
		return err
	}

	return ctx.JSON(resp)
}

func DeleteTranscoder(ctx *fiber.Ctx) error {
	logger.SDebug("DeleteTranscoder: request")

	transcoderId := ctx.Query("id")
	if len(transcoderId) == 0 {
		return custerror.FormatInvalidArgument("missing transcoderId as query string")
	}

	err := service.GetPrivateService().DeleteTranscoder(
		ctx.UserContext(),
		&web.DeleteTranscoderRequest{
			DeviceId: transcoderId,
		})
	if err != nil {
		return err
	}

	return ctx.SendStatus(http.StatusAccepted)
}

func GetOpenGateIntegrationConfigurations(ctx *fiber.Ctx) error {
	logger.SInfo("GetOpenGateIntegrationConfigurations: request")

	openGateId := ctx.Params("id")
	if len(openGateId) == 0 {
		return custerror.FormatInvalidArgument("missing id params")
	}

	resp, err := service.
		GetWebService().
		GetOpenGateIntegrationById(
			ctx.Context(), &web.GetOpenGateIntegrationByIdRequest{
				OpenGateId: openGateId,
			})
	if err != nil {
		logger.SError("GetOpenGateIntegrationConfigurations: service.GetOpenGateIntegrationConfigurations error", zap.Error(err))
		return err
	}

	return ctx.JSON(resp)
}

func GetTranscoderStreamConfigurations(ctx *fiber.Ctx) error {
	logger.SInfo("GetTranscoderStreamConfigurations: request")

	cameraIds := ctx.Query("camera_id")
	if len(cameraIds) == 0 {
		return custerror.FormatInvalidArgument("missing camera_id as query string")
	}
	ids := strings.Split(cameraIds, ",")
	resp, err := service.GetPrivateService().GetStreamConfigurations(ctx.Context(), &web.GetStreamConfigurationsRequest{
		CameraId: ids,
	})
	if err != nil {
		logger.SError("GetTranscoderStreamConfigurations: service.GetTranscoderStreamConfigurations error", zap.Error(err))
		return err
	}

	return ctx.JSON(resp)
}

func GetTranscoderOpenGateConfiguration(ctx *fiber.Ctx) error {
	logger.SInfo("GetTranscoderOpenGateConfiguration: request")

	id := ctx.Params("id")
	if len(id) == 0 {
		return custerror.FormatInvalidArgument("missing id params")
	}

	resp, err := service.GetPrivateService().GetTranscoderOpenGateConfiguration(ctx.Context(), &web.GetTranscoderOpenGateConfigurationRequest{
		TranscoderId: id,
	})
	if err != nil {
		logger.SError("GetTranscoderOpenGateConfiguration: service.GetTranscoderOpenGateConfiguration error", zap.Error(err))
		return err
	}

	return ctx.JSON(resp)
}

func GetTranscoderMQTTConfigurations(ctx *fiber.Ctx) error {
	logger.SInfo("GetTranscoderMQTTConfig: request")

	transcoderId := ctx.Query("transcoder_id")
	if len(transcoderId) == 0 {
		return custerror.FormatInvalidArgument("missing transcoder_id as query string")
	}

	resp, err := service.GetPrivateService().GetMQTTEventEndpoint(ctx.Context(), &web.GetMQTTEventEndpointRequest{
		TranscoderId: transcoderId,
	})
	if err != nil {
		logger.SError("GetTranscoderMQTTConfig: service.GetTranscoderMQTTConfig error", zap.Error(err))
		return err
	}

	return ctx.JSON(resp)
}
