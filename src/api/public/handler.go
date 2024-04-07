package publicapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/CE-Thesis-2023/backend/src/biz/service"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/web"

	"encoding/json"
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
		GetDevices(ctx.UserContext(), &web.GetTranscodersRequest{
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

	resp, err := service.GetWebService().GetCameras(ctx.UserContext(), &web.GetCamerasRequest{Ids: ids})
	if err != nil {
		return err
	}

	return ctx.JSON(resp)
}

func CreateCamera(ctx *fiber.Ctx) error {
	logger.SDebug("CreateCamera: request",
		zap.String("request", string(ctx.Body())))

	var msg web.AddCameraRequest
	if err := json.Unmarshal(ctx.Body(), &msg); err != nil {
		logger.SDebug("CreateCamera: unmarshal msg error", zap.Error(err))
		return custerror.ErrorInvalidArgument
	}
	resp, err := service.GetWebService().AddCamera(ctx.UserContext(), &msg)
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

	err := service.GetWebService().DeleteCamera(ctx.UserContext(), &web.DeleteCameraRequest{
		CameraId: id,
	})
	if err != nil {
		return err
	}

	return ctx.SendStatus(http.StatusAccepted)
}

func GetCamerasByGroupId(ctx *fiber.Ctx) error {
	logger.SDebug("GetCamerasInGroup: request")

	groupId := ctx.Params("id")
	if len(groupId) == 0 {
		return custerror.FormatInvalidArgument("missing groupId as parameter")
	}

	resp, err := service.GetWebService().GetCameraByGroupId(ctx.UserContext(), &web.GetCamerasByGroupIdRequest{
		GroupId: groupId,
	})

	if err != nil {
		return err
	}

	return ctx.JSON(resp)
}

func GetCameraGroups(ctx *fiber.Ctx) error {
	logger.SDebug("GetCameraGroups: request")

	queries := ctx.Query("ids")
	ids := strings.Split(queries, ",")

	if ids[0] == "" {
		ids = []string{}
	}

	resp, err := service.GetWebService().GetCameraGroupsByIds(ctx.UserContext(), &web.GetCameraGroupsRequest{Ids: ids})
	if err != nil {
		return err
	}

	return ctx.JSON(resp)
}

func AddCameraGroup(ctx *fiber.Ctx) error {
	logger.SDebug("AddCameraGroup: request")

	var msg web.AddCameraGroupRequest
	if err := json.Unmarshal(ctx.Body(), &msg); err != nil {
		logger.SDebug("AddCameraGroup: unmarshal msg error", zap.Error(err))
		return custerror.ErrorInvalidArgument
	}
	resp, err := service.GetWebService().AddCameraGroup(ctx.UserContext(), &msg)
	if err != nil {
		return err
	}

	return ctx.JSON(resp)
}

func DeleteCameraGroup(ctx *fiber.Ctx) error {
	logger.SDebug("DeleteCameraGroup: request")

	var msg web.DeleteCameraGroupRequest
	if err := json.Unmarshal(ctx.Body(), &msg); err != nil {
		logger.SDebug("DeleteCameraGroup: unmarshal msg error", zap.Error(err))
		return custerror.ErrorInvalidArgument
	}
	err := service.GetWebService().DeleteCameraGroup(ctx.UserContext(), &msg)
	if err != nil {
		return err
	}

	return ctx.SendStatus(http.StatusAccepted)
}

func AddCamerasToGroup(ctx *fiber.Ctx) error {
	logger.SDebug("AddCamerasToGroup: request")

	var msg web.AddCamerasToGroupRequest
	if err := json.Unmarshal(ctx.Body(), &msg); err != nil {
		logger.SDebug("AddCamerasToGroup: unmarshal msg error", zap.Error(err))
		return custerror.ErrorInvalidArgument
	}
	err := service.GetWebService().AddCamerasToGroup(ctx.UserContext(), &msg)
	if err != nil {
		return err
	}

	return ctx.JSON(&web.AddCamerasToGroupResponse{GroupId: msg.GroupId})
}

func RemoveCamerasFromGroup(ctx *fiber.Ctx) error {
	logger.SDebug("RemoveCamerasFromGroup: request")

	var msg web.RemoveCamerasFromGroupRequest
	if err := json.Unmarshal(ctx.Body(), &msg); err != nil {
		logger.SDebug("RemoveCamerasFromGroup: unmarshal msg error", zap.Error(err))
		return custerror.ErrorInvalidArgument
	}
	err := service.GetWebService().DeleteCamerasFromGroup(ctx.UserContext(), &msg)
	if err != nil {
		return err
	}

	return ctx.JSON(&web.RemoveCamerasFromGroupResponse{GroupId: msg.GroupId})

}

func UpdateTranscoder(ctx *fiber.Ctx) error {
	logger.SDebug("UpdateTranscoder: request")

	var msg web.UpdateTranscoderRequest
	if err := json.Unmarshal(ctx.Body(), &msg); err != nil {
		logger.SDebug("UpdateTranscoder: unmarshal msg error", zap.Error(err))
		return custerror.ErrorInvalidArgument
	}
	resp, err := service.GetWebService().UpdateTranscoder(ctx.UserContext(), &msg)
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
	resp, err := service.GetWebService().GetStreamInfo(ctx.UserContext(), &web.GetStreamInfoRequest{
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
	err := service.GetWebService().ToggleStream(ctx.UserContext(), &web.ToggleStreamRequest{
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
	if err := json.Unmarshal(ctx.Body(), &msg); err != nil {
		logger.SError("RemoteControl: unmarshal error", zap.Error(err))
		return err
	}

	err := service.GetWebService().RemoteControl(ctx.UserContext(), &msg)
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

	resp, err := service.GetWebService().
		GetDeviceInfo(ctx.UserContext(), &web.GetCameraDeviceInfoRequest{
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

func GetOpenGateSettings(ctx *fiber.Ctx) error {
	logger.SDebug("GetOpenGateSettings: request")

	openGateId := ctx.Params("openGateId")
	if len(openGateId) == 0 {
		return custerror.FormatInvalidArgument("missing openGateId as parameter")
	}

	resp, err := service.
		GetWebService().
		GetOpenGateIntegrationById(
			ctx.UserContext(),
			&web.GetOpenGateIntegrationByIdRequest{
				OpenGateId: openGateId,
			})
	if err != nil {
		return err
	}

	return ctx.JSON(resp)
}

func UpdateOpenGateSettings(ctx *fiber.Ctx) error {
	logger.SDebug("UpdateOpenGateSettings: request")

	openGateId := ctx.Params("openGateId")
	if len(openGateId) == 0 {
		return custerror.FormatInvalidArgument("missing openGateId as parameter")
	}

	var msg web.UpdateOpenGateIntegrationRequest
	if err := json.Unmarshal(ctx.Body(), &msg); err != nil {
		logger.SDebug("UpdateOpenGateSettings: unmarshal msg error", zap.Error(err))
		return custerror.ErrorInvalidArgument
	}

	if err := service.
		GetWebService().
		UpdateOpenGateIntegrationById(ctx.UserContext(), &msg); err != nil {
		logger.SDebug("UpdateOpenGateSettings: update error", zap.Error(err))
		return err
	}

	return ctx.SendStatus(http.StatusAccepted)
}

func GetObjectTrackingEvent(ctx *fiber.Ctx) error {
	logger.SDebug("GetObjectTrackingEvent: request")

	eventId := ctx.Query("eventId")
	openGateEventId := ctx.Query("openGateEventId")
	if eventId == "" && openGateEventId == "" {
		return custerror.FormatInvalidArgument("missing eventId or openGateEventId as query string")
	}
	req := &web.GetObjectTrackingEventByIdRequest{}
	if eventId != "" {
		req.EventId = []string{eventId}
	}
	if openGateEventId != "" {
		req.OpenGateEventId = []string{openGateEventId}
	}

	resp, err := service.GetWebService().GetObjectTrackingEventById(ctx.UserContext(), req)
	if err != nil {
		return err
	}

	return ctx.JSON(resp)
}

func DeleteObjectTrackingEvent(ctx *fiber.Ctx) error {
	logger.SDebug("DeleteObjectTrackingEvent: request")

	eventId := ctx.Query("eventId")
	if len(eventId) == 0 {
		return custerror.FormatInvalidArgument("missing eventId as query string")
	}
	req := &web.DeleteObjectTrackingEventRequest{}
	if eventId != "" {
		req.EventId = eventId
	}

	err := service.GetWebService().DeleteObjectTrackingEvent(ctx.UserContext(), req)
	if err != nil {
		return err
	}

	return ctx.SendStatus(http.StatusAccepted)
}
