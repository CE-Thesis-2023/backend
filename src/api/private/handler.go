package privateapi

import (
	"github.com/CE-Thesis-2023/backend/src/biz/service"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/events"

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

	if err := service.GetCommandService().
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

	resp, err := service.GetCommandService().UpdateCameraList(ctx.Context(), &msg)
	if err != nil {
		logger.SError("GetTranscoderAssignedCameras: service.UpdateCameraList error", zap.Error(err))
		return err
	}

	return ctx.JSON(resp)
}
