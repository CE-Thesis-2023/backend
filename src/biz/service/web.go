package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/CE-Thesis-2023/backend/src/biz/handlers"
	"github.com/CE-Thesis-2023/backend/src/helper"
	"github.com/CE-Thesis-2023/backend/src/internal/cache"
	custcon "github.com/CE-Thesis-2023/backend/src/internal/concurrent"
	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	custdb "github.com/CE-Thesis-2023/backend/src/internal/db"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	custmqtt "github.com/CE-Thesis-2023/backend/src/internal/mqtt"
	"github.com/CE-Thesis-2023/backend/src/models/db"
	"github.com/CE-Thesis-2023/backend/src/models/web"
	"github.com/mitchellh/mapstructure"

	events "github.com/CE-Thesis-2023/ltd/src/models/events"
	"github.com/Masterminds/squirrel"
	"github.com/bytedance/sonic"
	"github.com/dgraph-io/ristretto"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

type WebService struct {
	db         *custdb.LayeredDb
	cache      *ristretto.Cache
	mqttClient *autopaho.ConnectionManager
	pool       *ants.Pool
}

func NewWebService() *WebService {
	return &WebService{
		db:         custdb.Layered(),
		cache:      cache.Cache(),
		mqttClient: custmqtt.Client(),
		pool:       custcon.New(10),
	}
}

func (s *WebService) GetDevices(ctx context.Context, req *web.GetTranscodersRequest) (*web.GetTranscodersResponse, error) {
	logger.SDebug("GetDevices: request", zap.Any("request", req))

	devices, err := s.getDeviceById(ctx, req.Ids)
	if err != nil {
		logger.SDebug("GetDevices: error", zap.Error(err))
		return nil, err
	}

	logger.SDebug("GetDevices: devices", zap.Any("devices", devices))
	resp := web.GetTranscodersResponse{
		Transcoders: devices,
	}

	return &resp, nil
}

func (s *WebService) GetCameras(ctx context.Context, req *web.GetCamerasRequest) (*web.GetCamerasResponse, error) {
	logger.SDebug("GetCameras: request", zap.Any("request", req))

	cameras, err := s.getCameraById(ctx, req.Ids)
	if err != nil {
		logger.SError("GetCameras: error", zap.Error(err))
		return nil, err
	}

	logger.SDebug("GetCameras: cameras", zap.Any("cameras", cameras))
	resp := web.GetCamerasResponse{
		Cameras: cameras,
	}
	return &resp, nil
}

func (s *WebService) GetCamerasByOpenGateId(ctx context.Context, req *web.GetCameraByOpenGateIdRequest) (*web.GetCameraByOpenGateIdResponse, error) {
	logger.SDebug("GetCamerasByOpenGateId: request", zap.Any("request", req))

	camera, err := s.getCameraByOpenGateId(ctx, req.OpenGateId)
	if err != nil {
		logger.SError("GetCamerasByOpenGateId: error", zap.Error(err))
		return nil, err
	}

	return &web.GetCameraByOpenGateIdResponse{
		Camera: camera,
	}, nil
}

func (s *WebService) AddCamera(ctx context.Context, req *web.AddCameraRequest) (*web.AddCameraResponse, error) {
	logger.SDebug("AddCamera: request", zap.Any("request", req))

	existing, err := s.getCameraByName(ctx, []string{req.Name})
	if err != nil {
		logger.SError("AddCamera: getCameraByName", zap.Error(err))
		return nil, err
	}

	if len(existing) > 0 {
		logger.SDebug("AddCamera: camera already exists")
		return nil, custerror.ErrorAlreadyExists
	}

	_, err = s.getDeviceById(ctx, []string{req.TranscoderId})
	if err != nil {
		logger.SError("AddCamera: transcoder device not found")
		return nil, custerror.FormatNotFound("transcoder device not found")
	}

	var entry db.Camera
	if err := copier.Copy(&entry, req); err != nil {
		logger.SError("AddCamera: copier.Copy error", zap.Error(err))
		return nil, err
	}
	entry.CameraId = uuid.NewString()

	if err := s.addCamera(ctx, &entry); err != nil {
		logger.SError("AddCamera: addCamera error", zap.Error(err))
		return nil, err
	}

	if err = s.requestAddCamera(ctx, &entry); err != nil {
		logger.SError("AddCamera: requestAddCamera error", zap.Error(err))
		return nil, err
	}

	logger.SInfo("AddCamera: success", zap.String("id", entry.CameraId))
	return &web.AddCameraResponse{CameraId: entry.CameraId}, err
}

func (s *WebService) DeleteCamera(ctx context.Context, req *web.DeleteCameraRequest) error {
	logger.SDebug("DeleteCamera: request", zap.Any("request", req))

	c, err := s.getCameraById(ctx, []string{req.CameraId})
	if err != nil {
		logger.SError("DeleteCamera: getCameraByName", zap.Error(err))
		return err
	}
	if len(c) == 0 {
		logger.SError("DeleteCamera: camera not found")
		return custerror.ErrorNotFound
	}

	if err := s.deleteCameraById(ctx, req.CameraId); err != nil {
		logger.SError("DeleteCamera: deleteCameraById", zap.Error(err))
		return err
	}

	if err := s.requestRemoveCamera(ctx, &c[0]); err != nil {
		logger.SError("requestRemoveCamera: error = %s", zap.Error(err))
		return err
	}

	return nil
}

func (s *WebService) GetCameraByGroupId(ctx context.Context, req *web.GetCamerasByGroupIdRequest) (*web.GetCamerasByGroupIdResponse, error) {
	logger.SDebug("getCameraByGroupId: request", zap.Any("request", req))

	cameras, err := s.getCameraByGroupId(ctx, req.GroupId)
	if err != nil {
		logger.SError("getCameraByGroupId: getCameraByGroupId", zap.Error(err))
		return nil, err
	}

	logger.SDebug("getCameraByGroupId: cameras", zap.Any("cameras", cameras))
	resp := web.GetCamerasByGroupIdResponse{
		Cameras: cameras,
	}
	return &resp, nil
}

func (s *WebService) getCameraByGroupId(ctx context.Context, groupId string) ([]db.Camera, error) {
	q := squirrel.Select("*").From("cameras").Where("group_id = ?", groupId)

	sql, args, _ := q.ToSql()
	logger.SDebug("getCameraByGroupId: SQL",
		zap.Any("q", sql),
		zap.Any("args", args))

	var cameras []db.Camera
	if err := s.db.Select(ctx, q.PlaceholderFormat(squirrel.Dollar), &cameras); err != nil {
		return nil, err
	}

	return cameras, nil
}

func (s *WebService) AddCamerasToGroup(ctx context.Context, req *web.AddCamerasToGroupRequest) error {
	logger.SDebug("AddCamerasToGroup: request", zap.Any("request", req))

	group, err := s.getCameraGroupById(ctx, []string{req.GroupId})
	if err != nil {
		logger.SError("AddCamerasToGroup: getCameraGroupById", zap.Error(err))
		return err
	}
	if len(group) == 0 {
		logger.SError("AddCamerasToGroup: group not found")
		return custerror.ErrorNotFound
	}

	for _, cameraId := range req.CameraIds {
		camera, err := s.getCameraById(ctx, []string{cameraId})
		if err != nil {
			logger.SError("AddCamerasToGroup: getCameraById", zap.Error(err))
			return err
		}
		if len(camera) == 0 {
			logger.SError("AddCamerasToGroup: camera not found")
			return custerror.ErrorNotFound
		}

		if err := s.addCameraToGroup(ctx, camera, req.GroupId); err != nil {
			logger.SError("AddCamerasToGroup: addCameraToGroup", zap.Error(err))
			return err
		}
	}

	return nil
}

func (s *WebService) DeleteCamerasFromGroup(ctx context.Context, req *web.RemoveCamerasFromGroupRequest) error {
	logger.SDebug("DeleteCamerasFromGroup: request", zap.Any("request", req))

	group, err := s.getCameraGroupById(ctx, []string{req.GroupId})
	if err != nil {
		logger.SError("DeleteCamerasFromGroup: getCameraGroupById", zap.Error(err))
		return err
	}
	if len(group) == 0 {
		logger.SError("DeleteCamerasFromGroup: group not found")
		return custerror.ErrorNotFound
	}

	for _, cameraId := range req.CameraIds {
		camera, err := s.getCameraById(ctx, []string{cameraId})
		if err != nil {
			logger.SError("DeleteCamerasFromGroup: getCameraById", zap.Error(err))
			return err
		}
		if len(camera) == 0 {
			logger.SError("DeleteCamerasFromGroup: camera not found")
			return custerror.ErrorNotFound
		}

		if err := s.deleteCameraFromGroup(ctx, camera); err != nil {
			logger.SError("DeleteCamerasFromGroup: deleteCameraFromGroup", zap.Error(err))
			return err
		}
	}
	return nil
}

func (s *WebService) GetCameraGroupsByIds(ctx context.Context, req *web.GetCameraGroupsRequest) (*web.GetCameraGroupsResponse, error) {
	logger.SDebug("GetCameraGroups: request", zap.Any("request", req))

	groups, err := s.getCameraGroupById(ctx, req.Ids)

	if err != nil {
		logger.SError("GetCameraGroups: getCameraGroupById", zap.Error(err))
		return nil, err
	}

	logger.SDebug("GetCameraGroups: groups", zap.Any("groups", groups))
	resp := web.GetCameraGroupsResponse{
		CameraGroups: groups,
	}
	return &resp, nil
}

func (s *WebService) AddCameraGroup(ctx context.Context, req *web.AddCameraGroupRequest) (*web.AddCameraGroupResponse, error) {
	logger.SDebug("AddCameraGroup: request", zap.Any("request", req))

	existing, err := s.getCameraGroupByName(ctx, []string{req.Name})
	if err != nil {
		logger.SError("AddCameraGroup: getCameraGroupByName", zap.Error(err))
		return nil, err
	}

	if len(existing) > 0 {
		logger.SDebug("AddCameraGroup: group already exists")
		return nil, custerror.ErrorAlreadyExists
	}

	var entry db.CameraGroup
	if err := copier.Copy(&entry, req); err != nil {
		logger.SError("AddCameraGroup: copier.Copy error", zap.Error(err))
		return nil, err
	}
	entry.GroupId = uuid.NewString()
	entry.CreatedDate = time.Now()

	if err := s.addCameraGroup(ctx, &entry); err != nil {
		logger.SError("AddCameraGroup: addCameraGroup error", zap.Error(err))
		return nil, err
	}

	logger.SInfo("AddCameraGroup: success", zap.String("id", entry.GroupId))
	return &web.AddCameraGroupResponse{GroupId: entry.GroupId}, nil
}

func (s *WebService) DeleteCameraGroup(ctx context.Context, req *web.DeleteCameraGroupRequest) error {
	logger.SDebug("DeleteCameraGroup: request", zap.Any("request", req))

	group, err := s.getCameraGroupById(ctx, []string{req.GroupId})
	if err != nil {
		logger.SError("DeleteCameraGroup: getCameraGroupById", zap.Error(err))
		return err
	}
	if len(group) == 0 {
		logger.SError("DeleteCameraGroup: group not found")
		return custerror.ErrorNotFound
	}

	if err := s.deleteCameraGroup(ctx, req.GroupId); err != nil {
		logger.SError("DeleteCameraGroup: deleteCameraGroup", zap.Error(err))
		return err
	}

	return nil
}

func (s *WebService) deleteCameraById(ctx context.Context, id string) error {
	return s.db.Delete(ctx,
		squirrel.Delete("cameras").
			Where("camera_id = $1", id))
}

func (s *WebService) getCameraById(ctx context.Context, id []string) ([]db.Camera, error) {
	q := squirrel.Select("*").From("cameras")

	if len(id) > 0 {
		or := squirrel.Or{}
		for _, i := range id {
			or = append(or, squirrel.Eq{"camera_id": i})
		}
		q = q.Where(or)
	}

	sql, args, _ := q.ToSql()
	logger.SDebug("getCameraById: SQL",
		zap.Any("q", sql),
		zap.Any("args", args))

	cameras := []db.Camera{}
	if err := s.db.Select(ctx, q.PlaceholderFormat(squirrel.Dollar), &cameras); err != nil {
		return nil, err
	}

	return cameras, nil
}

func (s *WebService) getCameraByOpenGateId(ctx context.Context, openGateId string) (*db.Camera, error) {
	q := squirrel.Select("*").
		From("open_gate_integration").
		Where("opengate_id = $1", openGateId)
	sql, args, _ := q.ToSql()
	logger.SDebug("getCameraByOpenGateId: SQL",
		zap.Any("q", sql),
		zap.Any("args", args))

	var camera db.Camera
	if err := s.db.Get(ctx, q, &camera); err != nil {
		return nil, err
	}

	return &camera, nil
}

func (s *WebService) getCameraByName(ctx context.Context, names []string) ([]db.Camera, error) {
	q := squirrel.Select("*").From("cameras")

	if len(names) > 0 {
		or := squirrel.Or{}
		for _, i := range names {
			or = append(or, squirrel.Eq{"name": i})
		}
		q = q.Where(or)
	}

	sql, args, _ := q.ToSql()
	logger.SDebug("getCameraByName: SQL",
		zap.Any("q", sql),
		zap.Any("args", args))

	cameras := []db.Camera{}
	if err := s.db.Select(ctx, q, &cameras); err != nil {
		return nil, err
	}

	return cameras, nil
}

func (s *WebService) addCamera(ctx context.Context, camera *db.Camera) error {
	q := squirrel.Insert("cameras").
		Columns(camera.Fields()...).
		Values(camera.Values()...)
	if err := s.db.Insert(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) addCameraToGroup(ctx context.Context, cameras []db.Camera, groupId string) error {
	for _, camera := range cameras {
		q := squirrel.Update("cameras").Where("camera_id = ?", camera.CameraId).SetMap(map[string]interface{}{
			"group_id": groupId,
		})
		sql, args, _ := q.ToSql()
		logger.SDebug("addCameraToGroup: SQL query",
			zap.String("query", sql),
			zap.Any("args", args))
		if err := s.db.Update(ctx, q.PlaceholderFormat(squirrel.Dollar)); err != nil {
			return err
		}
	}
	return nil
}

func (s *WebService) deleteCameraFromGroup(ctx context.Context, cameras []db.Camera) error {
	for _, camera := range cameras {
		q := squirrel.Update("cameras").Where("camera_id = ?", camera.CameraId).Where("group_id = ?", camera.GroupId).SetMap(map[string]interface{}{
			"group_id": nil,
		})
		sql, args, _ := q.ToSql()
		logger.SDebug("deleteCameraFromGroup: SQL query",
			zap.String("query", sql),
			zap.Any("args", args))
		if err := s.db.Update(ctx, q.PlaceholderFormat(squirrel.Dollar)); err != nil {
			return err
		}
	}
	return nil
}

func (s *WebService) getCameraGroupById(ctx context.Context, ids []string) ([]db.CameraGroup, error) {
	q := squirrel.Select("*").From("camera_groups")

	if len(ids) > 0 {
		or := squirrel.Or{}
		for _, i := range ids {
			or = append(or, squirrel.Eq{"group_id": i})
		}
		q = q.Where(squirrel.Eq{"group_id": ids[0]})
	}

	sql, args, _ := q.ToSql()
	logger.SDebug("getGroupByIds: SQL",
		zap.Any("q", sql),
		zap.Any("args", args))

	groups := []db.CameraGroup{}
	if err := s.db.Select(ctx, q.PlaceholderFormat(squirrel.Dollar), &groups); err != nil {
		return nil, err
	}

	return groups, nil
}

func (s *WebService) getCameraGroupByName(ctx context.Context, names []string) ([]db.CameraGroup, error) {
	q := squirrel.Select("*").From("camera_groups")

	if len(names) > 0 {
		or := squirrel.Or{}
		for _, i := range names {
			eq := squirrel.Eq{"name": i}
			or = append(or, eq)
		}
		q = q.Where(or)
	}

	sql, args, _ := q.ToSql()
	logger.SDebug("getGroupByName: SQL",
		zap.Any("q", sql),
		zap.Any("args", args))

	var groups []db.CameraGroup
	if err := s.db.Select(ctx, q.PlaceholderFormat(squirrel.Dollar), &groups); err != nil {
		return nil, err
	}

	return groups, nil
}

func (s *WebService) addCameraGroup(ctx context.Context, group *db.CameraGroup) error {
	q := squirrel.Insert("camera_groups").
		Columns(group.Fields()...).
		Values(group.Values()...)
	if err := s.db.Insert(ctx, q.PlaceholderFormat(squirrel.Dollar)); err != nil {
		return err
	}
	return nil
}

func (s *WebService) deleteCameraGroup(ctx context.Context, id string) error {
	return s.db.Delete(ctx,
		squirrel.Delete("camera_groups").
			Where("group_id = $1", id))
}

func (s *WebService) updateCameraGroup(ctx context.Context, group *db.CameraGroup) error {
	q := squirrel.Update("camera_groups").Where("group_id = $1", group.GroupId).SetMap(map[string]interface{}{
		"name": group.Name,
	})
	sql, args, _ := q.ToSql()
	logger.SDebug("updateGroup: SQL query",
		zap.String("query", sql),
		zap.Any("args", args))
	if err := s.db.Update(ctx, q.PlaceholderFormat(squirrel.Dollar)); err != nil {
		return err
	}
	return nil
}

func (s *WebService) getDeviceById(ctx context.Context, id []string) ([]db.Transcoder, error) {
	query := squirrel.
		Select("*").
		From("transcoders")

	if len(id) > 0 {
		or := squirrel.Or{}
		for _, i := range id {
			or = append(or, squirrel.Eq{"device_id": i})
		}
		query = query.Where(or)
	}

	sql, args, _ := query.ToSql()
	logger.SDebug("getDeviceById: SQL",
		zap.Any("q", sql),
		zap.Any("args", args))

	var transcoders []db.Transcoder
	if err := s.db.Select(ctx, query, &transcoders); err != nil {
		return nil, err
	}

	return transcoders, nil
}

func (s *WebService) addDevice(ctx context.Context, d *db.Transcoder) error {
	query := squirrel.Insert("transcoders").
		Columns(d.Fields()...).
		Values(d.Values()...)
	if err := s.db.Insert(ctx, query); err != nil {
		return err
	}
	return nil
}

func (s *WebService) UpdateTranscoder(ctx context.Context, req *web.UpdateTranscoderRequest) (*db.Transcoder, error) {
	logger.SInfo("UpdateTranscoder: request", zap.Any("request", req))

	transcoders, err := s.getDeviceById(ctx, []string{req.Id})
	if err != nil {
		logger.SError("UpdateTranscoder: getDeviceById",
			zap.Error(err))
		return nil, err
	}
	if len(transcoders) == 0 {
		logger.SError("UpdateTranscoder: transcoder not found")
		return nil, custerror.ErrorNotFound
	}
	transcoder := transcoders[0]

	logger.SDebug("UpdateTranscoder: original",
		zap.Any("original", transcoder))

	if err := copier.Copy(transcoder, req); err != nil {
		logger.SError("UpdateTranscoder: copy error", zap.Error(err))
		return nil, err
	}
	if err := s.updateDevice(ctx, &transcoder); err != nil {
		logger.SError("UpdateTranscoder: update error", zap.Error(err))
		return nil, err
	}
	logger.SDebug("UpdatedTranscoder: updated", zap.Any("updated", transcoder))
	return &transcoder, nil
}

func (s *WebService) updateDevice(ctx context.Context, d *db.Transcoder) error {
	q := squirrel.Update("transcoders").Where("device_id = $1", d.DeviceId).SetMap(map[string]interface{}{
		"name": d.Name,
	})
	sql, args, _ := q.ToSql()
	logger.SDebug("updateDevice: SQL query",
		zap.String("query", sql),
		zap.Any("args", args))
	if err := s.db.Update(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) GetStreamInfo(ctx context.Context, req *web.GetStreamInfoRequest) (*web.GetStreamInfoResponse, error) {
	camera, err := s.getCameraById(ctx, []string{req.CameraId})
	if err != nil {
		logger.SDebug("GetStreamInfo: error", zap.Error(err))
		return nil, err
	}
	if len(camera) == 0 {
		return nil, custerror.FormatNotFound("camera not found")
	}

	streamUrl := s.buildStreamUrl(ctx, &camera[0])
	logger.SDebug("GetStreamInfo: streamUrl", zap.String("url", streamUrl))

	transcoder, err := s.getDeviceById(ctx, []string{camera[0].TranscoderId})
	if err != nil {
		logger.SDebug("GetStreamInfo: getDeviceById error", zap.Error(err))
		return nil, err
	}

	if len(transcoder) == 0 {
		logger.SDebug("GetStreamInfo: transcoder not found")
		return nil, custerror.FormatInternalError("transcoder device not found")
	}

	return &web.GetStreamInfoResponse{
		StreamUrl:      streamUrl,
		Protocol:       "webrtc",
		TranscoderId:   transcoder[0].DeviceId,
		TranscoderName: transcoder[0].Name,
		Started:        camera[0].Started,
	}, nil
}

func (s *WebService) buildStreamUrl(ctx context.Context, camera *db.Camera) string {
	configs := configs.Get().MediaEngine
	url := &url.URL{}
	url.Scheme = "http"
	url.Host = fmt.Sprintf("%s:%d", configs.Host, configs.PublishPorts.WebRTC)
	url = url.JoinPath(camera.CameraId)
	return url.String()
}

func (s *WebService) ToggleStream(ctx context.Context, req *web.ToggleStreamRequest) error {
	camera, err := s.getCameraById(ctx, []string{req.CameraId})
	if err != nil {
		logger.SError("ToggleStream: getCameraById error", zap.Error(err))
		return err
	}

	if len(camera) == 0 {
		logger.SError("ToggleStream: camera not found")
		return custerror.FormatNotFound("camera not found")
	}

	logger.SDebug("ToggleStream: camera", zap.Any("camera", camera[0]))

	if camera[0].Started == req.Start {
		logger.SDebug("ToggleStream: stream already started")
		return nil
	}

	var newCamera db.Camera
	if err := copier.Copy(&newCamera, &camera[0]); err != nil {
		logger.SError("ToggleStream: copy error", zap.Error(err))
		return err
	}

	newCamera.Started = req.Start

	err = s.updateCamera(ctx, &newCamera)
	if err != nil {
		logger.SError("ToggleStream: camera status update failed")
		return nil
	}

	err = s.requestLtdStreamControl(ctx, &newCamera)
	if err != nil {
		logger.SError("ToggleStream: requestLtdStreamControl: error", zap.Error(err))
		return err
	}

	logger.SDebug("ToggleStream: success", zap.String("cameraId", req.CameraId))
	return nil
}

func (s *WebService) updateCamera(ctx context.Context, camera *db.Camera) error {
	valueMap := map[string]interface{}{}
	fields := camera.Fields()
	values := camera.Values()
	for i := 0; i < len(fields); i += 1 {
		valueMap[fields[i]] = values[i]
	}

	q := squirrel.Update("cameras").Where("camera_id = $1", camera.CameraId).SetMap(valueMap)
	sql, args, _ := q.ToSql()
	logger.SDebug("updateCamera: SQL query",
		zap.String("query", sql),
		zap.Any("args", args))
	if err := s.db.Update(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) requestLtdStreamControl(ctx context.Context, camera *db.Camera) error {
	logger.SDebug("requestLtdStreamControl: request",
		zap.String("cameraId", camera.CameraId),
		zap.String("transcoderId", camera.TranscoderId))
	cmd := events.CommandRequest{}
	if camera.Started {
		logger.SDebug("requestLtdStreamControl: stream start")
		cmd.CommandType = events.Command_StartFfmpegStream
		cmd.Info = map[string]interface{}{
			"cameraId":  camera.CameraId,
			"channelId": "",
		}
	} else {
		logger.SDebug("requestLtdStreamControl: stream end")
		cmd.CommandType = events.Command_EndFfmpegStream
		cmd.Info = map[string]interface{}{
			"cameraId": camera.CameraId,
		}
	}
	msg, err := sonic.Marshal(&cmd)
	if err != nil {
		logger.SError("requestLtdStreamControl: error", zap.Error(err))
		return err
	}

	logger.SDebug("requestLtdStreamControl: msg", zap.String("message", string(msg)))
	_, err = s.mqttClient.Publish(ctx, &paho.Publish{
		Topic:   fmt.Sprintf("commands/%s", camera.TranscoderId),
		QoS:     1,
		Payload: msg,
	})
	if err != nil {
		logger.SError("requestLtdStreamControl: mqtt publish error", zap.Error(err))
		return err
	}
	logger.SDebug("requestLtdStreamControl: message published successfully")
	return nil
}

func (s *WebService) requestAddCamera(ctx context.Context, camera *db.Camera) error {
	logger.SDebug("requestAddCamera: request",
		zap.String("cameraId", camera.CameraId),
		zap.String("transcoderId", camera.TranscoderId))
	cmd := events.CommandRequest{
		CommandType: events.Command_AddCamera,
	}
	info := events.CommandAddCameraInfo{
		CameraId: camera.CameraId,
		Name:     camera.Name,
		Ip:       camera.Ip,
		Port:     camera.Port,
		Username: camera.Username,
		Password: camera.Password,
	}
	mapped, err := helper.ToMap(info)
	if err != nil {
		logger.SDebug("requestAddCamera: ToMap error", zap.Error(err))
		return err
	}
	cmd.Info = mapped

	pl, err := sonic.Marshal(cmd)
	if err != nil {
		logger.SError("requestAddCamera: marshal error", zap.Error(err))
		return err
	}
	logger.SDebug("requestAddCamera: info", zap.String("info", string(pl)))

	_, err = s.mqttClient.Publish(ctx, &paho.Publish{
		Topic:   fmt.Sprintf("commands/%s", camera.TranscoderId),
		QoS:     1,
		Payload: pl,
	})
	if err != nil {
		logger.SError("requestAddCamera: publish error", zap.Error(err))
		return err
	}
	logger.SInfo("requestAddCamera: success")
	return nil
}

func (s *WebService) requestRemoveCamera(ctx context.Context, camera *db.Camera) error {
	logger.SDebug("requestRemoveCamera: request",
		zap.String("cameraId", camera.CameraId),
		zap.String("transcoderId", camera.TranscoderId))
	cmd := events.CommandRequest{
		CommandType: events.Command_DeleteCamera,
	}
	mapped, err := helper.ToMap(events.CommandDeleteCameraRequest{
		CameraId: camera.CameraId,
	})
	if err != nil {
		logger.SError("requestRemoveCamera: ToMap error", zap.Error(err))
		return err
	}
	cmd.Info = mapped

	pl, err := sonic.Marshal(cmd)
	if err != nil {
		logger.SError("requestRemoveCamera: marshal error", zap.Error(err))
		return err
	}
	logger.SDebug("requestRemoveCamera: message", zap.String("msg", string(pl)))

	_, err = s.mqttClient.Publish(ctx, &paho.Publish{
		Topic:   fmt.Sprintf("commands/%s", camera.TranscoderId),
		QoS:     1,
		Payload: pl,
	})
	if err != nil {
		logger.SError("requestRemoveCamera: publish error", zap.Error(err))
		return err
	}
	logger.SInfo("requestRemoveCamera: success")
	return nil
}

func (s *WebService) getCamerasByTranscoderId(ctx context.Context, transcoderId string) ([]db.Camera, error) {
	q := squirrel.Select("*").From("cameras").Where("transcoder_id = $1", transcoderId)

	results := []db.Camera{}
	if err := s.db.Select(ctx, q, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (s *WebService) RemoteControl(ctx context.Context, req *web.RemoteControlRequest) error {
	logger.SDebug("biz.RemoteControl: request")

	c, err := s.getCameraByIdCached(ctx, req.CameraId)
	if err != nil {
		logger.SError("RemoteControl: getCameraByIdCached error", zap.Error(err))
		return err
	}

	if err := s.sendRemoteControlCommand(ctx, req, c); err != nil {
		logger.SError("RemoteControl: sendRemoteControlCommand error", zap.Error(err))
		return err
	}

	return nil
}

func (s *WebService) getCameraByIdCached(ctx context.Context, cameraId string) (*db.Camera, error) {
	result, found := s.cache.Get(fmt.Sprintf("rc-Camera-cameraId=%s", cameraId))
	if found {
		logger.SDebug("getCameraByIdCached: cache hit")
		c := result.(db.Camera)
		return &c, nil
	}

	camera, err := s.getCameraById(ctx, []string{cameraId})
	if err != nil {
		return nil, err
	}

	logger.SDebug("getCameraByIdCached: db hit")
	if len(camera) == 0 {
		return nil, custerror.ErrorNotFound
	}

	s.pool.Submit(func() {
		set := s.cache.Set(fmt.Sprintf("rc-Camera-cameraId=%s", cameraId), camera[0], 100)
		if set {
			logger.SDebug("getCameraByIdCached: camera info cache set")
		}
	})
	return &camera[0], nil
}

func (s *WebService) sendRemoteControlCommand(ctx context.Context, req *web.RemoteControlRequest, camera *db.Camera) error {
	msg := events.PtzCtrlRequest{
		CameraId:         camera.CameraId,
		Pan:              req.Pan,
		Tilt:             req.Tilt,
		StopAfterSeconds: helper.Int(2),
	}
	pl, err := sonic.Marshal(msg)
	if err != nil {
		logger.SDebug("sendRemoteControlCommand: marshal error", zap.Error(err))
		return err
	}
	if _, err := s.mqttClient.Publish(ctx, &paho.Publish{
		Topic:   fmt.Sprintf("ptzctrl/%s", camera.TranscoderId),
		QoS:     1,
		Payload: pl,
	}); err != nil {
		logger.SDebug("sendRemoteControlCommand: Publish error", zap.Error(err))
		return err
	}

	return nil
}

func (s *WebService) GetDeviceInfo(ctx context.Context, req *web.GetCameraDeviceInfoRequest) (*web.GetCameraDeviceInfo, error) {
	logger.SInfo("GetDeviceInfo: request", zap.Any("request", req))

	cameras, err := s.getCameraById(ctx, []string{req.CameraId})
	if err != nil {
		if errors.Is(err, custerror.ErrorNotFound) {
			logger.SError("GetDeviceInfo: cameraId not found")
			return nil, custerror.ErrorNotFound
		}
		logger.SError("GetDeviceInfo: getCameraById error", zap.Error(err))
		return nil, err
	}

	if len(cameras) == 0 {
		logger.SError("GetDeviceInfo: cameraId not found")
	}
	cam := cameras[0]
	msg := s.prepareGetDeviceInfoMessage(req)

	rr, err := handlers.GetWebsocketCommunicator().
		RequestReply(cam.TranscoderId)
	if err != nil {
		if errors.Is(err, custerror.ErrorFailedPrecondition) {
			logger.SError("GetDeviceInfo: local transcoder device not connected")
			return nil, err
		}
		logger.SError("GetDeviceInfo: error", zap.Error(err))
		return nil, err
	}

	resp, err := rr.Request(ctx, &events.CommandRequest{
		CommandType: events.Command_GetDeviceInfo,
		Info: map[string]interface{}{
			"cameraId": msg.CameraId,
		},
	})
	if err != nil {
		logger.SError("GetDeviceInfo: rr.Request error", zap.Error(err))
		return nil, err
	}

	logger.SDebug("GetDeviceInfo: response", zap.Any("response", resp))
	if resp == nil {
		logger.SError("GetDeviceInfo: ltd responded with null")
		return nil, custerror.ErrorInternal
	}
	var info web.GetCameraDeviceInfo
	if err := mapstructure.Decode(resp.Info, &info); err != nil {
		logger.SError("GetDeviceInfo: mapStructure.Decode error", zap.Error(err))
		return nil, err
	}

	logger.SInfo("GetDeviceInfo: info", zap.Any("deviceInfo", info))
	return &info, nil
}

func (s *WebService) prepareGetDeviceInfoMessage(req *web.GetCameraDeviceInfoRequest) *events.CommandRetrieveDeviceInfo {
	return &events.CommandRetrieveDeviceInfo{
		CameraId: req.CameraId,
	}
}

func (s *WebService) SendEventToMqtt(ctx context.Context, request *web.SendEventToMqttRequest) error {
	logger.SDebug("SendCameraEvent: request", zap.Any("request", request))

	cameras, err := s.getCameraById(ctx, []string{request.CameraId})
	if err != nil {
		logger.SError("SendCameraEvent: getCameraById error", zap.Error(err))
		return err
	}

	if len(cameras) == 0 {
		logger.SError("SendCameraEvent: camera not found", zap.String("cameraId", request.CameraId))
		return custerror.FormatNotFound("camera not found")
	}

	msg := &web.EventRequest{
		Event: request.Event,
	}

	pl, err := sonic.Marshal(msg)

	if _, err := s.mqttClient.Publish(ctx, &paho.Publish{
		Topic:   fmt.Sprintf("events/%s/%s", cameras[0].GroupId, cameras[0].CameraId),
		QoS:     1,
		Payload: pl,
	}); err != nil {
		logger.SError("SendCameraEvent: Publish error", zap.Error(err))
		return err
	}

	logger.SDebug("SendCameraEvent: success")
	return nil
}

func (s *WebService) PublicEventToOtherCamerasInGroup(ctx context.Context, req *web.PublicEventToOtherCamerasInGroupRequest) error {
	logger.SDebug("PublicEventToOtherCamerasInGroup: request", zap.Any("request", req))

	camera, err := s.getCameraById(ctx, []string{req.CameraId})
	if err != nil {
		logger.SError("PublicEventToOtherCamerasInGroup: getCameraById error", zap.Error(err))
		return err
	}

	if len(camera) == 0 {
		logger.SError("PublicEventToOtherCamerasInGroup: camera not found", zap.String("cameraId", req.CameraId))
		return custerror.FormatNotFound("camera not found")
	}

	if err := s.publicEventToOtherCamerasInGroup(ctx, camera[0], req.Event); err != nil {
		logger.SError("PublicEventToOtherCamerasInGroup: publicEventToOtherCamerasInGroup error", zap.Error(err))
		return err
	}

	logger.SDebug("PublicEventToOtherCamerasInGroup: success")
	return nil
}

// This function is to public event to other topics of cameras in the same group
func (s *WebService) publicEventToOtherCamerasInGroup(ctx context.Context, camera db.Camera, event string) error {

	if camera.GroupId == "" {
		logger.SError("PublicEventToOtherCamerasInGroup: camera is not in any group")
		return custerror.FormatInternalError("camera is not in any group")
	}

	cameras, err := s.getCameraByGroupId(ctx, camera.GroupId)
	if err != nil {
		logger.SError("PublicEventToOtherCamerasInGroup: getCameraGroupById error", zap.Error(err))
		return err
	}

	for _, c := range cameras {
		if c.CameraId == camera.CameraId {
			continue
		}

		msg := &web.EventRequest{
			Event: fmt.Sprintf("{cameraId: %s, event: %s}", camera.CameraId, event),
		}

		pl, err := sonic.Marshal(msg)

		if err != nil {
			logger.SError("PublicEventToOtherCamerasInGroup: Marshal error", zap.Error(err))
			return err
		}

		if _, err := s.mqttClient.Publish(ctx, &paho.Publish{
			Topic:   fmt.Sprintf("events/%s/%s", c.GroupId, c.CameraId),
			QoS:     1,
			Payload: pl,
		}); err != nil {
			logger.SError("PublicEventToOtherCamerasInGroup: Publish error", zap.Error(err))
			return err
		}
	}
	return nil
}
