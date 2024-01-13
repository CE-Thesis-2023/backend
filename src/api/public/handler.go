package publicapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/CE-Thesis-2023/backend/src/biz/service"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/web"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func GetTranscoderDevices(ctx *fiber.Ctx) error {
	logger.SDebug("GetTranscoderDevices: request")

	queries := ctx.Query("id")
	ids := strings.Split(queries, ",")
	if ids[0] == "" {
		ids = []string{}
	}

	resp, err := service.
		GetWebService().
		GetDevices(ctx.Context(), &web.GetTranscodersRequest{
			Ids: ids,
		})
	if err != nil {
		return err
	}

	return ctx.JSON(resp)
}

func GetCameras(ctx *fiber.Ctx) error {
	logger.SDebug("GetCameras: request")

	queries := ctx.Query("id")
	ids := strings.Split(queries, ",")
	if ids[0] == "" {
		ids = []string{}
	}

	resp, err := service.GetWebService().GetCameras(ctx.Context(), &web.GetCamerasRequest{Ids: ids})
	if err != nil {
		return err
	}

	return ctx.JSON(resp)
}

func CreateCamera(ctx *fiber.Ctx) error {
	logger.SDebug("CreateCamera: request",
		zap.String("request", string(ctx.Body())))

	var msg web.AddCameraRequest
	if err := sonic.Unmarshal(ctx.Body(), &msg); err != nil {
		logger.SDebug("CreateCamera: unmarshal msg error", zap.Error(err))
		return custerror.ErrorInvalidArgument
	}
	resp, err := service.GetWebService().AddCamera(ctx.Context(), &msg)
	if err != nil {
		return err
	}

	return ctx.JSON(resp)
}

func DeleteCamera(ctx *fiber.Ctx) error {
	logger.SDebug("DeleteCamera: request")

	id := ctx.Query("id")
	if len(id) == 0 {
		return custerror.ErrorInvalidArgument
	}

	err := service.GetWebService().DeleteCamera(ctx.Context(), &web.DeleteCameraRequest{
		CameraId: id,
	})
	if err != nil {
		return err
	}

	return ctx.SendStatus(http.StatusAccepted)
}

func UpdateTranscoder(ctx *fiber.Ctx) error {
	logger.SDebug("UpdateTranscoder: request")

	var msg web.UpdateTranscoderRequest
	if err := sonic.Unmarshal(ctx.Body(), &msg); err != nil {
		logger.SDebug("UpdateTranscoder: unmarshal msg error", zap.Error(err))
		return custerror.ErrorInvalidArgument
	}
	resp, err := service.GetWebService().UpdateTranscoder(ctx.Context(), &msg)
	if err != nil {
		return err
	}

	return ctx.JSON(resp)
}

func GetStreamInfo(ctx *fiber.Ctx) error {
	logger.SDebug("GetStreamInfo: request")

	cameraId := ctx.Params("id")
	if cameraId == "" {
		return custerror.FormatInvalidArgument("cameraId not found")
	}
	resp, err := service.GetWebService().GetStreamInfo(ctx.Context(), &web.GetStreamInfoRequest{
		CameraId: cameraId,
	})
	if err != nil {
		return err
	}

	return ctx.JSON(resp)
}

func ToggleStream(ctx *fiber.Ctx) error {
	logger.SDebug("ToggleStream: request")
	cameraId := ctx.Params("id")
	if len(cameraId) == 0 {
		return custerror.FormatInvalidArgument("missing cameraId as query string")
	}
	enable := ctx.QueryBool("enable")
	err := service.GetWebService().ToggleStream(ctx.Context(), &web.ToggleStreamRequest{
		CameraId: cameraId,
		Start:    enable,
	})
	if err != nil {
		return err
	}
	return ctx.SendStatus(200)
}

func RemoteControl(ctx *fiber.Ctx) error {
	logger.SDebug("RemoteControl: request")

	var msg web.RemoteControlRequest
	if err := sonic.Unmarshal(ctx.Body(), &msg); err != nil {
		logger.SError("RemoteControl: unmarshal error", zap.Error(err))
		return err
	}

	err := service.GetWebService().RemoteControl(ctx.Context(), &msg)
	if err != nil {
		return err
	}

	return ctx.SendStatus(200)
}

func GetCameraDeviceInfo(ctx *fiber.Ctx) error {
	logger.SDebug("GetCameraDeviceInfo: request")

	cameraId := ctx.Params("cameraId")
	if len(cameraId) == 0 {
		return custerror.FormatInvalidArgument("missing cameraId as parameter")
	}

	resp, err := service.GetWebService().GetDeviceInfo(ctx.Context(), &web.GetCameraDeviceInfoRequest{
		CameraId: cameraId,
	})
	if err != nil {
		if errors.Is(err, custerror.ErrorFailedPrecondition) {
			return ctx.SendStatus(http.StatusExpectationFailed)
		}
		return err
	}

	return ctx.JSON(resp)
}
