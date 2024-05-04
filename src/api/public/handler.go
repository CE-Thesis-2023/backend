package publicapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/CE-Thesis-2023/backend/src/biz/service"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	custhttp "github.com/CE-Thesis-2023/backend/src/internal/http"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/db"
	"github.com/CE-Thesis-2023/backend/src/models/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetTranscoderDevices(ctx *gin.Context) {
	logger.SDebug("GetTranscoderDevices: request")

	queries := ctx.Query("id")
	ids := strings.Split(queries, ",")
	if ids[0] == "" {
		ids = []string{}
	}

	resp, err := service.
		GetWebService().
		GetDevices(ctx, &web.GetTranscodersRequest{
			Ids: ids,
		})
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func GetCameras(ctx *gin.Context) {
	logger.SDebug("GetCameras: request")

	queries := ctx.Query("id")
	ids := strings.Split(queries, ",")
	if ids[0] == "" {
		ids = []string{}
	}

	resp, err := service.GetWebService().GetCameras(ctx, &web.GetCamerasRequest{Ids: ids})
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func CreateCamera(ctx *gin.Context) {
	var msg web.AddCameraRequest
	if err := ctx.BindJSON(&msg); err != nil {
		logger.SDebug("CreateCamera: unmarshal msg error", zap.Error(err))
		custhttp.ToHTTPErr(
			custerror.ErrorInvalidArgument,
			ctx)
		return
	}
	resp, err := service.GetWebService().AddCamera(ctx, &msg)
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func DeleteCamera(ctx *gin.Context) {
	logger.SDebug("DeleteCamera: request")

	id := ctx.Query("id")
	if len(id) == 0 {
		custhttp.ToHTTPErr(
			custerror.ErrorInvalidArgument,
			ctx)
		return
	}

	err := service.GetWebService().DeleteCamera(ctx, &web.DeleteCameraRequest{
		CameraId: id,
	})
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.Status(http.StatusAccepted)
}

func GetCamerasByGroupId(ctx *gin.Context) {
	logger.SDebug("GetCamerasInGroup: request")

	groupId, found := ctx.Params.Get("id")
	if !found {
		err := custerror.FormatInvalidArgument("missing groupId as parameter")
		custhttp.ToHTTPErr(err, ctx)
		return
	}
	if len(groupId) == 0 {
		err := custerror.FormatInvalidArgument("missing groupId as parameter")
		custhttp.ToHTTPErr(err, ctx)
	}

	resp, err := service.GetWebService().GetCameraByGroupId(ctx, &web.GetCamerasByGroupIdRequest{
		GroupId: groupId,
	})

	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func GetCameraGroups(ctx *gin.Context) {
	logger.SDebug("GetCameraGroups: request")

	queries := ctx.Query("ids")
	ids := strings.Split(queries, ",")

	if ids[0] == "" {
		ids = []string{}
	}

	resp, err := service.GetWebService().GetCameraGroupsByIds(ctx, &web.GetCameraGroupsRequest{Ids: ids})
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func AddCameraGroup(ctx *gin.Context) {
	logger.SDebug("AddCameraGroup: request")

	var msg web.AddCameraGroupRequest
	if err := ctx.BindJSON(&msg); err != nil {
		logger.SDebug("AddCameraGroup: unmarshal msg error", zap.Error(err))
		custhttp.ToHTTPErr(
			custerror.ErrorInvalidArgument,
			ctx)
		return
	}
	resp, err := service.GetWebService().AddCameraGroup(ctx, &msg)
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func DeleteCameraGroup(ctx *gin.Context) {
	logger.SDebug("DeleteCameraGroup: request")

	var msg web.DeleteCameraGroupRequest
	if err := ctx.BindJSON(&msg); err != nil {
		logger.SDebug("DeleteCameraGroup: unmarshal msg error", zap.Error(err))
		custhttp.ToHTTPErr(
			custerror.ErrorInvalidArgument,
			ctx)
		return
	}
	err := service.GetWebService().DeleteCameraGroup(ctx, &msg)
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.Status(http.StatusAccepted)
}

func AddCamerasToGroup(ctx *gin.Context) {
	logger.SDebug("AddCamerasToGroup: request")

	var msg web.AddCamerasToGroupRequest
	if err := ctx.BindJSON(&msg); err != nil {
		logger.SDebug("AddCamerasToGroup: unmarshal msg error", zap.Error(err))
		custhttp.ToHTTPErr(
			custerror.ErrorInvalidArgument,
			ctx)
		return
	}
	err := service.GetWebService().AddCamerasToGroup(ctx, &msg)
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, &web.AddCamerasToGroupResponse{GroupId: msg.GroupId})
}

func RemoveCamerasFromGroup(ctx *gin.Context) {
	logger.SDebug("RemoveCamerasFromGroup: request")

	var msg web.RemoveCamerasFromGroupRequest
	if err := ctx.BindJSON(&msg); err != nil {
		logger.SDebug("RemoveCamerasFromGroup: unmarshal msg error", zap.Error(err))
		custhttp.ToHTTPErr(
			custerror.ErrorInvalidArgument,
			ctx)
		return
	}
	err := service.GetWebService().DeleteCamerasFromGroup(ctx, &msg)
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, &web.RemoveCamerasFromGroupResponse{GroupId: msg.GroupId})
}

func UpdateTranscoder(ctx *gin.Context) {
	logger.SDebug("UpdateTranscoder: request")

	var msg web.UpdateTranscoderRequest
	if err := ctx.BindJSON(&msg); err != nil {
		logger.SDebug("UpdateTranscoder: unmarshal msg error", zap.Error(err))
		custhttp.ToHTTPErr(
			custerror.ErrorInvalidArgument,
			ctx)
		return
	}
	resp, err := service.GetWebService().UpdateTranscoder(ctx, &msg)
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func GetStreamInfo(ctx *gin.Context) {
	logger.SDebug("GetStreamInfo: request")

	cameraId, found := ctx.Params.Get("id")
	if !found {
		err := custerror.FormatInvalidArgument("cameraId not found")
		custhttp.ToHTTPErr(err, ctx)
		return
	}
	if cameraId == "" {
		err := custerror.FormatInvalidArgument("cameraId not found")
		custhttp.ToHTTPErr(err, ctx)
		return
	}
	resp, err := service.GetWebService().GetStreamInfo(ctx, &web.GetStreamInfoRequest{
		CameraId: cameraId,
	})
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func ToggleStream(ctx *gin.Context) {
	logger.SDebug("ToggleStream: request")
	cameraId, found := ctx.Params.Get("id")
	if !found {
		err := custerror.FormatInvalidArgument("missing cameraId as query string")
		custhttp.ToHTTPErr(err, ctx)
		return
	}
	if len(cameraId) == 0 {
		err := custerror.FormatInvalidArgument("missing cameraId as query string")
		custhttp.ToHTTPErr(err, ctx)
		return
	}
	enable := ctx.Query("enable")
	isEnable := true
	switch {
	case enable == "false":
		isEnable = false
	case enable == "true":
		isEnable = true
	default:
		err := custerror.FormatInvalidArgument("missing enable as query string")
		custhttp.ToHTTPErr(err, ctx)
		return
	}
	err := service.GetWebService().ToggleStream(ctx, &web.ToggleStreamRequest{
		CameraId: cameraId,
		Start:    isEnable,
	})
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}
	ctx.Status(200)
}

func RemoteControl(ctx *gin.Context) {
	logger.SDebug("RemoteControl: request")

	var msg web.RemoteControlRequest
	if err := ctx.BindJSON(&msg); err != nil {
		logger.SError("RemoteControl: unmarshal error", zap.Error(err))
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	err := service.GetWebService().RemoteControl(ctx, &msg)
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.Status(200)
}

func GetCameraDeviceInfo(ctx *gin.Context) {
	logger.SDebug("GetCameraDeviceInfo: request")

	cameraId, found := ctx.Params.Get("cameraId")
	if !found {
		err := custerror.FormatInvalidArgument("missing cameraId as parameter")
		custhttp.ToHTTPErr(err, ctx)
		return
	}
	if len(cameraId) == 0 {
		err := custerror.FormatInvalidArgument("missing cameraId as parameter")
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	resp, err := service.GetWebService().
		GetDeviceInfo(ctx, &web.GetCameraDeviceInfoRequest{
			CameraId: cameraId,
		})
	if err != nil {
		if errors.Is(err, custerror.ErrorFailedPrecondition) {
			ctx.Status(http.StatusExpectationFailed)
			custhttp.ToHTTPErr(err, ctx)
			return
		}
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func GetOpenGateSettings(ctx *gin.Context) {
	logger.SDebug("GetOpenGateSettings: request")

	openGateId, found := ctx.Params.Get("openGateId")
	if !found {
		err := custerror.FormatInvalidArgument("missing openGateId as parameter")
		custhttp.ToHTTPErr(err, ctx)
		return
	}
	if len(openGateId) == 0 {
		err := custerror.FormatInvalidArgument("missing openGateId as parameter")
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	resp, err := service.
		GetWebService().
		GetOpenGateIntegrationById(
			ctx,
			&web.GetOpenGateIntegrationByIdRequest{
				OpenGateId: openGateId,
			})
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func UpdateOpenGateSettings(ctx *gin.Context) {
	logger.SDebug("UpdateOpenGateSettings: request")

	openGateId, found := ctx.Params.Get("openGateId")
	if !found {
		err := custerror.FormatInvalidArgument("missing openGateId as parameter")
		custhttp.ToHTTPErr(err, ctx)
		return
	}
	if len(openGateId) == 0 {
		err := custerror.FormatInvalidArgument("missing openGateId as parameter")
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	var msg web.UpdateOpenGateIntegrationRequest
	if err := ctx.BindJSON(&msg); err != nil {
		logger.SDebug("UpdateOpenGateSettings: unmarshal msg error", zap.Error(err))
		custhttp.ToHTTPErr(
			custerror.ErrorInvalidArgument,
			ctx)
		return
	}

	if err := service.
		GetWebService().
		UpdateOpenGateIntegrationById(ctx, &msg); err != nil {
		logger.SDebug("UpdateOpenGateSettings: update error", zap.Error(err))
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.Status(http.StatusAccepted)
}

func GetObjectTrackingEvent(ctx *gin.Context) {
	logger.SDebug("GetObjectTrackingEvent: request")

	eventId := ctx.Query("event_id")
	openGateEventId := ctx.Query("open_gate_event_id")
	cameraId := ctx.Query("camera_id")
	limit := ctx.Query("limit")

	req := &web.GetObjectTrackingEventByIdRequest{}
	if eventId != "" {
		req.EventId = []string{eventId}
	}
	if openGateEventId != "" {
		req.OpenGateEventId = []string{openGateEventId}
	}
	if cameraId != "" {
		req.CameraId = cameraId
	}
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			custhttp.ToHTTPErr(custerror.ErrorInvalidArgument, ctx)
			return
		}
		req.Limit = l
	}

	resp, err := service.GetWebService().GetObjectTrackingEventById(ctx, req)
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}
	if resp.ObjectTrackingEvents == nil {
		resp.ObjectTrackingEvents = []db.ObjectTrackingEvent{}
	}

	ctx.JSON(200, resp)
}

func DeleteObjectTrackingEvent(ctx *gin.Context) {
	logger.SDebug("DeleteObjectTrackingEvent: request")

	eventId := ctx.Query("id")
	if len(eventId) == 0 {
		err := custerror.FormatInvalidArgument("missing eventId as query string")
		custhttp.ToHTTPErr(err, ctx)
		return
	}
	req := &web.DeleteObjectTrackingEventRequest{}
	if eventId != "" {
		req.EventId = eventId
	}

	err := service.GetWebService().DeleteObjectTrackingEvent(ctx, req)
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.Status(http.StatusAccepted)
}

func DoDeviceHealthcheck(ctx *gin.Context) {
	logger.SDebug("DoDeviceHealthcheck: request")

	deviceId := ctx.Query("transcoder_id")
	if len(deviceId) == 0 {
		err := custerror.FormatInvalidArgument("missing deviceId as parameter")
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	resp, err := service.GetWebService().DoDeviceHealthcheck(ctx, &web.DeviceHealthcheckRequest{
		TranscoderId: deviceId,
	})
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func AddDetectablePerson(ctx *gin.Context) {
	logger.SDebug("AddDetectablePerson: request")

	var msg web.AddDetectablePersonRequest
	if err := ctx.BindJSON(&msg); err != nil {
		logger.SDebug("AddDetectablePerson: unmarshal msg error", zap.Error(err))
		custhttp.ToHTTPErr(
			custerror.ErrorInvalidArgument,
			ctx)
		return
	}

	resp, err := service.
		GetWebService().
		AddDetectablePerson(ctx, &msg)
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func GetDetectablePeople(ctx *gin.Context) {
	logger.SDebug("GetDetectablePeople: request")

	queries := ctx.Query("ids")
	ids := []string{}
	if len(queries) > 0 {
		ids = strings.Split(queries, ",")
	}

	resp, err := service.
		GetWebService().
		GetDetectablePeople(ctx, &web.GetDetectablePeopleRequest{
			PersonIds: ids,
		})
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func DeleteDetectablePerson(ctx *gin.Context) {
	logger.SDebug("DeleteDetectablePerson: request")

	personId := ctx.Query("id")
	if len(personId) == 0 {
		err := custerror.FormatInvalidArgument("missing personId as query")
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	err := service.
		GetWebService().
		DeleteDetectablePerson(ctx, &web.DeleteDetectablePersonRequest{
			PersonId: personId,
		})
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.Status(http.StatusAccepted)
}

func GetDetectablePersonPresignedUrl(ctx *gin.Context) {
	logger.SDebug("GetDetectablePersonPresignedUrl: request")

	personId := ctx.Query("id")
	if len(personId) == 0 {
		err := custerror.FormatInvalidArgument("missing personId as parameter")
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	resp, err := service.
		GetWebService().
		GetDetectablePersonImagePresignedUrl(ctx, &web.GetDetectablePeopleImagePresignedUrlRequest{
			PersonId: personId,
		})
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func GetLatestOpenGateCameraStats(ctx *gin.Context) {
	logger.SDebug("GetStats: request")

	transcoderId := ctx.Query("transcoder_id")
	if len(transcoderId) == 0 {
		custhttp.ToHTTPErr(
			custerror.FormatInvalidArgument("missing transcoder_id"),
			ctx)
		return
	}

	cameraName := ctx.Query("camera_name")
	if len(cameraName) == 0 {
		custhttp.ToHTTPErr(
			custerror.FormatInvalidArgument("missing camera_name"),
			ctx)
		return
	}
	names := strings.Split(cameraName, ",")
	if len(names) == 0 {
		custhttp.ToHTTPErr(
			custerror.FormatInvalidArgument("missing camera_name"),
			ctx)
		return
	}

	resp, err := service.
		GetWebService().
		GetLatestOpenGateCameraStats(ctx, &web.GetLatestOpenGateCameraStatsRequest{
			TranscoderId: transcoderId,
			CameraNames:  names,
		})
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}
	ctx.JSON(200, resp)
}

func GetSnapshotPresignedUrl(ctx *gin.Context) {
	logger.SDebug("GetSnapshotPresignedUrl: request")

	ids := ctx.Query("snapshot_id")
	if len(ids) == 0 {
		custhttp.ToHTTPErr(
			custerror.FormatInvalidArgument("missing snapshot_id"),
			ctx)
		return
	}
	splittedIds := strings.Split(ids, ",")
	if len(splittedIds) == 0 {
		custhttp.ToHTTPErr(
			custerror.FormatInvalidArgument("missing snapshot_id"),
			ctx)
		return
	}

	resp, err := service.GetWebService().GetSnapshotPresignedUrl(ctx, &web.GetSnapshotPresignedUrlRequest{
		SnapshotId: splittedIds,
	})
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func GetPersonHistory(ctx *gin.Context) {
	logger.SDebug("GetPersonHistory: request")

	personId := ctx.Query("person_id")
	var personIds []string
	if len(personId) != 0 {
		personIds = strings.Split(personId, ",")
	}
	historyId := ctx.Query("history_id")
	var historyIds []string
	if len(historyId) != 0 {
		historyIds = strings.Split(historyId, ",")
	}

	resp, err := service.GetWebService().GetPersonHistory(ctx, &web.GetPersonHistoryRequest{
		PersonId:  personIds,
		HistoryId: historyIds,
	})
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}

func GetTranscoderStatus(ctx *gin.Context) {
	logger.SDebug("GetTranscoderStatus: request")

	transcoderId := ctx.Query("transcoder_id")
	if len(transcoderId) == 0 {
		custhttp.ToHTTPErr(
			custerror.FormatInvalidArgument("missing transcoder_id"),
			ctx)
		return
	}
	splitted := strings.Split(transcoderId, ",")
	if len(splitted) == 0 {
		custhttp.ToHTTPErr(
			custerror.FormatInvalidArgument("missing transcoder_id"),
			ctx)
		return
	}

	var cameraIds []string
	cameraId := ctx.Query("camera_id")
	if len(cameraId) != 0 {
		cameraIds = strings.Split(cameraId, ",")
	}

	resp, err := service.GetWebService().GetTranscoderStatus(ctx, &web.GetTranscoderStatusRequest{
		TranscoderId: splitted,
		CameraId:     cameraIds,
	})
	if err != nil {
		custhttp.ToHTTPErr(err, ctx)
		return
	}

	ctx.JSON(200, resp)
}
