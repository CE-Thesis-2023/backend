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
