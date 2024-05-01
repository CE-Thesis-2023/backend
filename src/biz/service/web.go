package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"

	"strings"
	"time"

	"github.com/CE-Thesis-2023/backend/src/helper"
	"github.com/CE-Thesis-2023/backend/src/helper/media"
	custdb "github.com/CE-Thesis-2023/backend/src/internal/db"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	custmqtt "github.com/CE-Thesis-2023/backend/src/internal/mqtt"
	"github.com/CE-Thesis-2023/backend/src/models/db"
	"github.com/CE-Thesis-2023/backend/src/models/events"
	"github.com/CE-Thesis-2023/backend/src/models/web"
	"github.com/Kagami/go-face"

	"github.com/Masterminds/squirrel"
	"github.com/dgraph-io/ristretto"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
)

type WebService struct {
	db          *custdb.LayeredDb
	cache       *ristretto.Cache
	builder     squirrel.StatementBuilderType
	MediaHelper *media.MediaHelper
	reqreply    *custmqtt.MQTTSession
	cvs         *ComputerVisionService
}

func NewWebService(reqreply *custmqtt.MQTTSession, mediaHelper *media.MediaHelper, cvs *ComputerVisionService) *WebService {
	return &WebService{
		db: custdb.Layered(),
		builder: squirrel.
			StatementBuilder.
			PlaceholderFormat(squirrel.Dollar),
		MediaHelper: mediaHelper,
		reqreply:    reqreply,
		cvs:         cvs,
	}
}

func (s *WebService) validateGetDevices(_ *web.GetTranscodersRequest) error {
	return nil
}

func (s *WebService) GetDevices(ctx context.Context, req *web.GetTranscodersRequest) (*web.GetTranscodersResponse, error) {
	logger.SDebug("GetDevices: request", zap.Reflect("request", req))

	if err := s.validateGetDevices(req); err != nil {
		logger.SDebug("GetDevices: validateGetDevices", zap.Error(err))
		return nil, err
	}

	devices, err := s.getDeviceById(ctx, req.Ids)
	if err != nil {
		logger.SDebug("GetDevices: error", zap.Error(err))
		return nil, err
	}

	logger.SDebug("GetDevices: devices", zap.Reflect("devices", devices))
	resp := web.GetTranscodersResponse{
		Transcoders: devices,
	}

	return &resp, nil
}

func (s *WebService) validateGetCameras(_ *web.GetCamerasRequest) error {
	return nil
}

func (s *WebService) GetCameras(ctx context.Context, req *web.GetCamerasRequest) (*web.GetCamerasResponse, error) {
	logger.SDebug("GetCameras: request", zap.Reflect("request", req))

	if err := s.validateGetCameras(req); err != nil {
		logger.SDebug("GetCameras: validateGetCameras", zap.Error(err))
		return nil, err
	}

	cameras, err := s.getCameraById(ctx, req.Ids)
	if err != nil {
		logger.SError("GetCameras: error", zap.Error(err))
		return nil, err
	}

	logger.SDebug("GetCameras: cameras", zap.Reflect("cameras", cameras))
	resp := web.GetCamerasResponse{
		Cameras: cameras,
	}
	return &resp, nil
}

func (s *WebService) validateGetCamerasByTranscoderId(req *web.GetCameraByTranscoderId) error {
	if len(req.TranscoderId) == 0 {
		return custerror.FormatInvalidArgument("missing client")
	}
	return nil
}

func (s *WebService) GetCamerasByTranscoderId(ctx context.Context, req *web.GetCameraByTranscoderId) (*web.GetCameraByOpenGateIdResponse, error) {
	logger.SDebug("GetCamerasByClientId: request",
		zap.Reflect("request", req))

	if err := s.validateGetCamerasByTranscoderId(req); err != nil {
		logger.SDebug("GetCamerasByTranscoderId: validateGetCamerasByTranscoderId", zap.Error(err))
		return nil, err
	}

	transcoder, err := s.getTranscoderById(ctx, req.TranscoderId)
	if err != nil {
		logger.SError("GetCamerasByTranscoderId: getTranscoderById", zap.Error(err))
		return nil, err
	}

	cameras, err := s.getCamerasByTranscoderId(
		ctx,
		transcoder.DeviceId,
		req.OpenGateCameraNames)
	if err != nil {
		logger.SError("GetCamerasByClientId: error", zap.Error(err))
		return nil, err
	}

	return &web.GetCameraByOpenGateIdResponse{
		Cameras: cameras,
	}, nil
}

func (s *WebService) getOpenGateIntegrationById(ctx context.Context, id string) (*db.OpenGateIntegration, error) {
	q := s.builder.Select("*").
		From("open_gate_integrations").
		Where("open_gate_id = $1", id)
	sql, args, _ := q.ToSql()
	logger.SDebug("getOpenGateIntegrationById: SQL",
		zap.Reflect("q", sql),
		zap.Reflect("args", args))

	var openGate db.OpenGateIntegration
	if err := s.db.Get(ctx, q, &openGate); err != nil {
		return nil, err
	}
	return &openGate, nil
}

func (s *WebService) getTranscoderById(ctx context.Context, id string) (*db.Transcoder, error) {
	q := s.builder.Select("*").
		From("transcoders").
		Where("device_id = ?", id)
	sql, args, _ := q.ToSql()
	logger.SDebug("getTranscoderById: SQL",
		zap.Reflect("q", sql),
		zap.Reflect("args", args))

	var transcoder db.Transcoder
	if err := s.db.Get(ctx, q, &transcoder); err != nil {
		return nil, err
	}
	return &transcoder, nil
}

func (s *WebService) validateAddCameraRequest(req *web.AddCameraRequest) error {
	if len(req.Name) == 0 {
		return custerror.FormatInvalidArgument("missing name")
	}
	if len(req.TranscoderId) == 0 {
		return custerror.FormatInvalidArgument("missing transcoder id")
	}
	if len(req.Ip) == 0 {
		return custerror.FormatInvalidArgument("missing ip")
	}
	if req.Port <= 0 || req.Port > 65535 {
		return custerror.FormatInvalidArgument("missing port")
	}
	if req.Username == "" {
		return custerror.FormatInvalidArgument("missing username")
	}
	if req.Password == "" {
		return custerror.FormatInvalidArgument("missing password")
	}
	return nil
}

func (s *WebService) AddCamera(ctx context.Context, req *web.AddCameraRequest) (*web.AddCameraResponse, error) {
	logger.SDebug("AddCamera: request", zap.Reflect("request", req))

	if err := s.validateAddCameraRequest(req); err != nil {
		logger.SDebug("AddCamera: validateAddCamera", zap.Error(err))
		return nil, err
	}

	existing, err := s.getCameraByName(ctx, []string{req.Name})
	if err != nil {
		logger.SError("AddCamera: getCameraByName", zap.Error(err))
		return nil, err
	}

	if len(existing) > 0 {
		logger.SDebug("AddCamera: camera already exists")
		return nil, custerror.ErrorAlreadyExists
	}

	transcoder, err := s.getDeviceById(ctx, []string{req.TranscoderId})
	if err != nil {
		logger.SError("AddCamera: transcoder device not found")
		return nil, custerror.FormatNotFound("transcoder device not found")
	}

	if transcoder == nil {
		logger.SError("AddCamera: transcoder device not found")
		return nil, custerror.FormatNotFound("transcoder device not found")
	}
	if len(transcoder) == 0 {
		logger.SError("AddCamera: transcoder device not found")
		return nil, custerror.FormatNotFound("transcoder device not found")
	}

	var entry db.Camera
	if err := copier.Copy(&entry, req); err != nil {
		logger.SError("AddCamera: copier.Copy error", zap.Error(err))
		return nil, err
	}
	entry.CameraId = uuid.NewString()
	openGateCameraName := strings.ReplaceAll(req.Name, " ", "_")
	openGateCameraName = strings.ToLower(openGateCameraName)

	entry.OpenGateCameraName = openGateCameraName

	openGateCameraSettings, err := s.initializeDefaultOpenGateCameraSettings(ctx, &entry, &transcoder[0])
	if err != nil {
		logger.SError("AddCamera: initializeDefaultOpenGateCameraSettings error", zap.Error(err))
		return nil, err
	}
	entry.SettingsId = openGateCameraSettings.SettingsId
	logger.SDebug("AddCamera: openGateCameraSettings",
		zap.Reflect("settings", openGateCameraSettings),
		zap.Reflect("entry", entry))

	if err := s.addCamera(ctx, &entry); err != nil {
		logger.SError("AddCamera: addCamera error", zap.Error(err))
		return nil, err
	}

	logger.SInfo("AddCamera: success", zap.String("id", entry.CameraId))
	return &web.AddCameraResponse{CameraId: entry.CameraId}, err
}

func (s *WebService) initializeDefaultOpenGateCameraSettings(ctx context.Context, camera *db.Camera, transcoder *db.Transcoder) (*db.OpenGateCameraSettings, error) {
	settings := db.OpenGateCameraSettings{
		SettingsId:  uuid.NewString(),
		CameraId:    camera.CameraId,
		Height:      480,
		Width:       640,
		Fps:         5,
		MqttEnabled: true,
		Timestamp:   true,
		BoundingBox: true,
		Crop:        true,
		OpenGateId:  transcoder.OpenGateIntegrationId,
	}

	return &settings,
		s.addOpenGateCameraSettings(ctx, &settings)
}

func (s *WebService) addOpenGateCameraSettings(ctx context.Context, settings *db.OpenGateCameraSettings) error {
	q := s.builder.Insert("open_gate_camera_settings").
		Columns(settings.Fields()...).
		Values(settings.Values()...)
	if err := s.db.Insert(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) validateDeleteCameraRequest(req *web.DeleteCameraRequest) error {
	if len(req.CameraId) == 0 {
		return custerror.FormatInvalidArgument("missing camera id")
	}
	return nil
}

func (s *WebService) DeleteCamera(ctx context.Context, req *web.DeleteCameraRequest) error {
	logger.SDebug("DeleteCamera: request", zap.Reflect("request", req))

	if err := s.validateDeleteCameraRequest(req); err != nil {
		logger.SDebug("DeleteCamera: validateDeleteCameraRequest", zap.Error(err))
		return err
	}

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

	if err := s.deleteOpenGateCameraSettings(ctx, req.CameraId); err != nil {
		logger.SError("DeleteCamera: deleteOpenGateCameraSettings", zap.Error(err))
		return err
	}
	return nil
}

func (s *WebService) deleteOpenGateCameraSettings(ctx context.Context, cameraId string) error {
	return s.db.Delete(ctx,
		s.builder.Delete("open_gate_camera_settings").
			Where("camera_id = ?", cameraId))
}

func (s *WebService) validateGetCameraByGroupId(req *web.GetCamerasByGroupIdRequest) error {
	if len(req.GroupId) == 0 {
		return custerror.FormatInvalidArgument("missing group id")
	}
	return nil
}

func (s *WebService) GetCameraByGroupId(ctx context.Context, req *web.GetCamerasByGroupIdRequest) (*web.GetCamerasByGroupIdResponse, error) {
	logger.SDebug("getCameraByGroupId: request", zap.Reflect("request", req))

	if err := s.validateGetCameraByGroupId(req); err != nil {
		logger.SDebug("getCameraByGroupId: validateGetCameraByGroupId", zap.Error(err))
		return nil, err
	}

	cameras, err := s.getCameraByGroupId(ctx, req.GroupId)
	if err != nil {
		logger.SError("getCameraByGroupId: getCameraByGroupId", zap.Error(err))
		return nil, err
	}

	logger.SDebug("getCameraByGroupId: cameras", zap.Reflect("cameras", cameras))
	resp := web.GetCamerasByGroupIdResponse{
		Cameras: cameras,
	}
	return &resp, nil
}

func (s *WebService) getCameraByGroupId(ctx context.Context, groupId string) ([]db.Camera, error) {
	q := s.builder.Select("*").
		From("cameras").
		Where("group_id = ?", groupId)

	sql, args, _ := q.ToSql()
	logger.SDebug("getCameraByGroupId: SQL",
		zap.Reflect("q", sql),
		zap.Reflect("args", args))

	var cameras []db.Camera
	if err := s.db.Select(ctx, q, &cameras); err != nil {
		return nil, err
	}
	if cameras == nil {
		return []db.Camera{}, nil
	}

	return cameras, nil
}

func (s *WebService) validateAddCamerasToGroup(req *web.AddCamerasToGroupRequest) error {
	if len(req.GroupId) == 0 {
		return custerror.FormatInvalidArgument("missing group id")
	}
	if len(req.CameraIds) == 0 {
		return custerror.FormatInvalidArgument("missing camera ids")
	}
	return nil
}

func (s *WebService) AddCamerasToGroup(ctx context.Context, req *web.AddCamerasToGroupRequest) error {
	logger.SDebug("AddCamerasToGroup: request", zap.Reflect("request", req))

	if err := s.validateAddCamerasToGroup(req); err != nil {
		logger.SDebug("AddCamerasToGroup: validateAddCamerasToGroup", zap.Error(err))
		return err
	}

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

func (s *WebService) validateDeleteCamerasFromGroupRequest(req *web.RemoveCamerasFromGroupRequest) error {
	if len(req.GroupId) == 0 {
		return custerror.FormatInvalidArgument("missing group id")
	}
	if len(req.CameraIds) == 0 {
		return custerror.FormatInvalidArgument("missing camera ids")
	}
	return nil
}

func (s *WebService) DeleteCamerasFromGroup(ctx context.Context, req *web.RemoveCamerasFromGroupRequest) error {
	logger.SDebug("DeleteCamerasFromGroup: request", zap.Reflect("request", req))

	if err := s.validateDeleteCamerasFromGroupRequest(req); err != nil {
		return err
	}

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

func (s *WebService) validateGetCameraGroupsByIdsRequest(_ *web.GetCameraGroupsRequest) error {
	return nil
}

func (s *WebService) GetCameraGroupsByIds(ctx context.Context, req *web.GetCameraGroupsRequest) (*web.GetCameraGroupsResponse, error) {
	logger.SDebug("GetCameraGroups: request", zap.Reflect("request", req))

	if err := s.validateGetCameraGroupsByIdsRequest(req); err != nil {
		logger.SDebug("GetCameraGroups: validateGetCameraGroupsByIdsRequest", zap.Error(err))
		return nil, err
	}

	groups, err := s.getCameraGroupById(ctx, req.Ids)

	if err != nil {
		logger.SError("GetCameraGroups: getCameraGroupById", zap.Error(err))
		return nil, err
	}

	logger.SDebug("GetCameraGroups: groups", zap.Reflect("groups", groups))
	resp := web.GetCameraGroupsResponse{
		CameraGroups: groups,
	}
	return &resp, nil
}

func (s *WebService) validateAddCameraGroupRequest(req *web.AddCameraGroupRequest) error {
	if len(req.Name) == 0 {
		return custerror.FormatInvalidArgument("missing name")
	}
	return nil
}

func (s *WebService) AddCameraGroup(ctx context.Context, req *web.AddCameraGroupRequest) (*web.AddCameraGroupResponse, error) {
	logger.SDebug("AddCameraGroup: request", zap.Reflect("request", req))

	if err := s.validateAddCameraGroupRequest(req); err != nil {
		logger.SDebug("AddCameraGroup: validateAddCameraGroupRequest", zap.Error(err))
		return nil, err
	}

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

func (s *WebService) validateDeleteCameraGroupRequest(req *web.DeleteCameraGroupRequest) error {
	if len(req.GroupId) == 0 {
		return custerror.FormatInvalidArgument("missing group id")
	}
	return nil
}

func (s *WebService) DeleteCameraGroup(ctx context.Context, req *web.DeleteCameraGroupRequest) error {
	logger.SDebug("DeleteCameraGroup: request", zap.Reflect("request", req))

	if err := s.validateDeleteCameraGroupRequest(req); err != nil {
		return err
	}

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
		s.builder.Delete("cameras").
			Where("camera_id = ?", id))
}

func (s *WebService) getCameraById(ctx context.Context, id []string) ([]db.Camera, error) {
	q := s.builder.
		Select("*").
		From("cameras")

	if len(id) > 0 {
		or := squirrel.Or{}
		for _, i := range id {
			or = append(or, squirrel.Eq{"camera_id": i})
		}
		q = q.Where(or)
	}

	sql, args, _ := q.ToSql()
	logger.SDebug("getCameraById: SQL",
		zap.Reflect("q", sql),
		zap.Reflect("args", args))

	cameras := []db.Camera{}
	if err := s.db.Select(ctx, q, &cameras); err != nil {
		return nil, err
	}

	return cameras, nil
}

func (s *WebService) getCameraByName(ctx context.Context, names []string) ([]db.Camera, error) {
	q := s.builder.Select("*").
		From("cameras")

	if len(names) > 0 {
		or := squirrel.Or{}
		for _, i := range names {
			or = append(or, squirrel.Eq{"name": i})
		}
		q = q.Where(or)
	}

	sql, args, _ := q.ToSql()
	logger.SDebug("getCameraByName: SQL",
		zap.Reflect("q", sql),
		zap.Reflect("args", args))

	cameras := []db.Camera{}
	if err := s.db.Select(ctx, q, &cameras); err != nil {
		return nil, err
	}

	return cameras, nil
}

func (s *WebService) addCamera(ctx context.Context, camera *db.Camera) error {
	q := s.builder.Insert("cameras").
		Columns(camera.Fields()...).
		Values(camera.Values()...)
	if err := s.db.Insert(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) addCameraToGroup(ctx context.Context, cameras []db.Camera, groupId string) error {
	for _, camera := range cameras {
		q := s.builder.Update("cameras").
			Where("camera_id = ?", camera.CameraId).
			SetMap(map[string]interface{}{
				"group_id": groupId,
			})
		sql, args, _ := q.ToSql()
		logger.SDebug("addCameraToGroup: SQL query",
			zap.String("query", sql),
			zap.Reflect("args", args))
		if err := s.db.Update(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

func (s *WebService) deleteCameraFromGroup(ctx context.Context, cameras []db.Camera) error {
	for _, camera := range cameras {
		q := s.builder.Update("cameras").
			Where("camera_id = ?", camera.CameraId).
			Where("group_id = ?", camera.GroupId).
			SetMap(map[string]interface{}{
				"group_id": nil,
			})
		sql, args, _ := q.ToSql()
		logger.SDebug("deleteCameraFromGroup: SQL query",
			zap.String("query", sql),
			zap.Reflect("args", args))
		if err := s.db.Update(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

func (s *WebService) getCameraGroupById(ctx context.Context, ids []string) ([]db.CameraGroup, error) {
	q := s.builder.Select("*").
		From("camera_groups")

	if len(ids) > 0 {
		or := squirrel.Or{}
		for _, i := range ids {
			or = append(or, squirrel.Eq{"group_id": i})
		}
		q = q.Where(or)
	}

	sql, args, _ := q.ToSql()
	logger.SDebug("getGroupByIds: SQL",
		zap.Reflect("q", sql),
		zap.Reflect("args", args))

	groups := []db.CameraGroup{}
	if err := s.db.Select(ctx, q, &groups); err != nil {
		return nil, err
	}

	return groups, nil
}

func (s *WebService) getCameraGroupByName(ctx context.Context, names []string) ([]db.CameraGroup, error) {
	q := s.builder.
		Select("*").
		From("camera_groups")

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
		zap.Reflect("q", sql),
		zap.Reflect("args", args))

	var groups []db.CameraGroup
	if err := s.db.Select(ctx, q, &groups); err != nil {
		return nil, err
	}

	return groups, nil
}

func (s *WebService) addCameraGroup(ctx context.Context, group *db.CameraGroup) error {
	q := s.builder.Insert("camera_groups").
		Columns(group.Fields()...).
		Values(group.Values()...)
	if err := s.db.Insert(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) deleteCameraGroup(ctx context.Context, id string) error {
	return s.db.Delete(ctx,
		s.builder.
			Delete("camera_groups").
			Where("group_id = ?", id))
}

func (s *WebService) getDeviceById(ctx context.Context, id []string) ([]db.Transcoder, error) {
	query := s.builder.
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
		zap.Reflect("q", sql),
		zap.Reflect("args", args))

	var transcoders []db.Transcoder
	if err := s.db.Select(ctx, query, &transcoders); err != nil {
		return nil, err
	}
	if transcoders == nil {
		return []db.Transcoder{}, nil
	}

	return transcoders, nil
}

func (s *WebService) addOpenGateIntegration(ctx context.Context, integration *db.OpenGateIntegration) error {
	q := s.builder.Insert("open_gate_integrations").
		Columns(integration.Fields()...).
		Values(integration.Values()...)
	if err := s.db.Insert(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) addOpenGateMqttConfigurations(ctx context.Context, config *db.OpenGateMqttConfiguration) error {
	q := s.builder.Insert("open_gate_mqtt_configurations").
		Columns(config.Fields()...).
		Values(config.Values()...)
	if err := s.db.Insert(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) addDevice(ctx context.Context, d *db.Transcoder) error {
	query := s.builder.Insert("transcoders").
		Columns(d.Fields()...).
		Values(d.Values()...)
	if err := s.db.Insert(ctx, query); err != nil {
		return err
	}
	return nil
}

func (s *WebService) validateUpdateTranscoderRequest(req *web.UpdateTranscoderRequest) error {
	if len(req.Id) == 0 {
		return custerror.FormatInvalidArgument("missing id")
	}
	if len(req.Name) == 0 {
		return custerror.FormatInvalidArgument("missing name")
	}
	return nil
}

func (s *WebService) UpdateTranscoder(ctx context.Context, req *web.UpdateTranscoderRequest) (*db.Transcoder, error) {
	logger.SInfo("UpdateTranscoder: request", zap.Reflect("request", req))

	if err := s.validateUpdateTranscoderRequest(req); err != nil {
		logger.SError("UpdateTranscoder: validateUpdateTranscoderRequest",
			zap.Error(err))
		return nil, err
	}

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
		zap.Reflect("original", transcoder))

	if err := copier.Copy(transcoder, req); err != nil {
		logger.SError("UpdateTranscoder: copy error", zap.Error(err))
		return nil, err
	}
	if err := s.updateDevice(ctx, &transcoder); err != nil {
		logger.SError("UpdateTranscoder: update error", zap.Error(err))
		return nil, err
	}
	logger.SDebug("UpdatedTranscoder: updated", zap.Reflect("updated", transcoder))
	return &transcoder, nil
}

func (s *WebService) updateDevice(ctx context.Context, d *db.Transcoder) error {
	q := s.builder.
		Update("transcoders").
		Where("device_id = ?", d.DeviceId).
		SetMap(map[string]interface{}{
			"name": d.Name,
		})
	sql, args, _ := q.ToSql()
	logger.SDebug("updateDevice: SQL query",
		zap.String("query", sql),
		zap.Reflect("args", args))
	if err := s.db.Update(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) validateGetStreamInfoRequest(req *web.GetStreamInfoRequest) error {
	if len(req.CameraId) == 0 {
		return custerror.FormatInvalidArgument("missing camera id")
	}
	return nil
}

func (s *WebService) GetStreamInfo(ctx context.Context, req *web.GetStreamInfoRequest) (*web.GetStreamInfoResponse, error) {
	logger.SDebug("GetStreamInfo: request", zap.Reflect("request", req))

	if err := s.validateGetStreamInfoRequest(req); err != nil {
		logger.SDebug("GetStreamInfo: validateGetStreamInfoRequest", zap.Error(err))
		return nil, err
	}

	camera, err := s.getCameraById(ctx, []string{req.CameraId})
	if err != nil {
		logger.SDebug("GetStreamInfo: error", zap.Error(err))
		return nil, err
	}
	if len(camera) == 0 {
		return nil, custerror.FormatNotFound("camera not found")
	}

	streamUrl := s.MediaHelper.BuildWebRTCViewStream(camera[0].CameraId)
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
		Started:        camera[0].Enabled,
	}, nil
}

func (s *WebService) validateToggleStreamRequest(req *web.ToggleStreamRequest) error {
	if len(req.CameraId) == 0 {
		return custerror.FormatInvalidArgument("missing camera id")
	}
	return nil
}

func (s *WebService) ToggleStream(ctx context.Context, req *web.ToggleStreamRequest) error {
	logger.SDebug("ToggleStream: request", zap.Reflect("request", req))

	if err := s.validateToggleStreamRequest(req); err != nil {
		logger.SDebug("ToggleStream: validateToggleStreamRequest", zap.Error(err))
		return err
	}

	camera, err := s.getCameraById(ctx, []string{req.CameraId})
	if err != nil {
		logger.SError("ToggleStream: getCameraById error", zap.Error(err))
		return err
	}

	if len(camera) == 0 {
		logger.SError("ToggleStream: camera not found")
		return custerror.FormatNotFound("camera not found")
	}

	logger.SDebug("ToggleStream: camera", zap.Reflect("camera", camera[0]))

	if camera[0].Enabled == req.Start {
		logger.SDebug("ToggleStream: stream already started")
		return nil
	}

	var newCamera db.Camera
	if err := copier.Copy(&newCamera, &camera[0]); err != nil {
		logger.SError("ToggleStream: copy error", zap.Error(err))
		return err
	}

	newCamera.Enabled = req.Start

	err = s.updateCamera(ctx, &newCamera)
	if err != nil {
		logger.SError("ToggleStream: camera status update failed")
		return nil
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

	q := s.builder.Update("cameras").
		Where("camera_id = ?", camera.CameraId).
		SetMap(valueMap)
	sql, args, _ := q.ToSql()
	logger.SDebug("updateCamera: SQL query",
		zap.String("query", sql),
		zap.Reflect("args", args))
	if err := s.db.Update(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) getCamerasByTranscoderId(ctx context.Context, transcoderId string, openGateCameraNames []string) ([]db.Camera, error) {
	q := s.builder.Select("*").
		From("cameras")

	if len(transcoderId) > 0 {
		q = q.Where("transcoder_id = ?", transcoderId)
	}

	if len(openGateCameraNames) > 0 {
		or := squirrel.Or{}
		for _, i := range openGateCameraNames {
			or = append(or, squirrel.Eq{"open_gate_camera_name": i})
		}
		q = q.Where(or)
	}

	sql, args, _ := q.ToSql()
	logger.SDebug("getCamerasByTranscoderId: SQL query",
		zap.String("query", sql),
		zap.Reflect("args", args))

	var results []db.Camera
	if err := s.db.Select(ctx, q.PlaceholderFormat(squirrel.Dollar), &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (s *WebService) validateRemoteControlRequest(req *web.RemoteControlRequest) error {
	if len(req.CameraId) == 0 {
		return custerror.FormatInvalidArgument("missing camera id")
	}
	if req.Pan > 360 || req.Pan < -360 {
		return custerror.FormatInvalidArgument("pan value out of range between -360 to 360")
	}
	if req.Tilt > 360 || req.Tilt < -360 {
		return custerror.FormatInvalidArgument("tilt value out of range between -360 to 360")
	}
	return nil
}

func (s *WebService) RemoteControl(ctx context.Context, req *web.RemoteControlRequest) error {
	logger.SDebug("biz.RemoteControl: request", zap.Reflect("request", req))

	if err := s.validateRemoteControlRequest(req); err != nil {
		logger.SDebug("RemoteControl: validateRemoteControlRequest error", zap.Error(err))
		return err
	}

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

	go func() {
		set := s.cache.Set(fmt.Sprintf("rc-Camera-cameraId=%s", cameraId), camera[0], 100)
		if set {
			logger.SDebug("getCameraByIdCached: camera info cache set")
		}
	}()
	return &camera[0], nil
}

func (s *WebService) sendRemoteControlCommand(ctx context.Context, req *web.RemoteControlRequest, camera *db.Camera) error {
	ptzCtrlRequest := events.PTZCtrlRequest{
		CameraId: camera.CameraId,
		Pan:      req.Pan,
		Tilt:     req.Tilt,
	}
	response := map[string]string{}
	reqReplyRequest := custmqtt.RequestReplyRequest{
		Topic: events.Event{
			Prefix:    "commands",
			ID:        camera.TranscoderId,
			Type:      "ptz",
			Arguments: []string{camera.CameraId},
		},
		Request:    ptzCtrlRequest,
		Reply:      response,
		MaxTimeout: time.Second * 3,
	}
	if err := s.reqreply.Request(ctx, &reqReplyRequest); err != nil {
		return custerror.FormatInternalError("request-reply error: %s", err)
	}
	return nil
}

func (s *WebService) validateGetDeviceInfoRequest(req *web.GetCameraDeviceInfoRequest) error {
	if len(req.CameraId) == 0 {
		return custerror.FormatInvalidArgument("missing camera id")
	}
	return nil
}

func (s *WebService) GetDeviceInfo(ctx context.Context, req *web.GetCameraDeviceInfoRequest) (*web.GetCameraDeviceInfo, error) {
	logger.SInfo("GetDeviceInfo: request", zap.Reflect("request", req))

	if err := s.validateGetDeviceInfoRequest(req); err != nil {
		logger.SError("GetDeviceInfo: validateGetDeviceInfoRequest error", zap.Error(err))
		return nil, err
	}

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

	var resp interface{}
	if err := s.sendDeviceInfoRequest(ctx, req, &cameras[0], &resp); err != nil {
		logger.SError("GetDeviceInfo: sendDeviceInfoRequest error",
			zap.Error(err))
		return nil, err
	}
	apiResp := &web.GetCameraDeviceInfo{}
	if err := copier.Copy(apiResp, resp); err != nil {
		logger.SError("GetDeviceInfo: copier.Copy error", zap.Error(err))
		return nil, err
	}
	return apiResp, nil
}

func (s *WebService) sendDeviceInfoRequest(ctx context.Context, _ *web.GetCameraDeviceInfoRequest, camera *db.Camera, res interface{}) error {
	reqreplyRequest := custmqtt.RequestReplyRequest{
		Topic: events.Event{
			Prefix:    "commands",
			ID:        camera.TranscoderId,
			Type:      "info",
			Arguments: []string{camera.CameraId},
		},
		Request:    map[string]interface{}{},
		Reply:      res,
		MaxTimeout: time.Second * 3,
	}
	if err := s.reqreply.Request(ctx, &reqreplyRequest); err != nil {
		return custerror.FormatInternalError("request-reply error: %s", err)
	}
	return nil
}

func (s *WebService) validateGetOpenGateIntegrationById(req *web.GetOpenGateIntegrationByIdRequest) error {
	if req.OpenGateId == "" {
		return custerror.FormatInvalidArgument("missing open gate id")
	}
	return nil
}

func (s *WebService) GetOpenGateIntegrationById(ctx context.Context, req *web.GetOpenGateIntegrationByIdRequest) (*web.GetOpenGateIntegrationByIdResponse, error) {
	logger.SDebug("GetOpenGateIntegrationById: request",
		zap.Reflect("request", req))

	if err := s.validateGetOpenGateIntegrationById(req); err != nil {
		logger.SError("GetOpenGateIntegrationById: validateGetOpenGateIntegrationById error", zap.Error(err))
		return nil, err
	}

	opengate, err := s.getOpenGateIntegrationById(ctx, req.OpenGateId)
	if err != nil {
		logger.SError("GetOpenGateIntegrationById: getOpenGateIntegrationById error", zap.Error(err))
		return nil, err
	}
	return &web.GetOpenGateIntegrationByIdResponse{
		OpenGateIntegration: opengate,
	}, nil
}

func (s *WebService) validateUpdateOpenGateIntegrationById(req *web.UpdateOpenGateIntegrationRequest) error {
	if req.OpenGateId == "" {
		return custerror.FormatInvalidArgument("missing open gate id")
	}
	switch req.LogLevel {
	case "debug":
	case "info":
	case "warning":
	default:
		return custerror.FormatInvalidArgument("invalid log level, must be one of debug, info, warning")
	}
	if req.SnapshotRetentionDays <= 0 {
		return custerror.FormatInvalidArgument("invalid snapshot retention days")
	}
	mqtt := req.Mqtt
	if mqtt == nil {
		return custerror.FormatInvalidArgument("missing mqtt configuration")
	}
	if mqtt.Host == "" {
		return custerror.FormatInvalidArgument("missing mqtt host")
	}
	if mqtt.Port <= 0 {
		return custerror.FormatInvalidArgument("invalid mqtt port")
	}
	if mqtt.Username == "" {
		return custerror.FormatInvalidArgument("missing mqtt username")
	}
	if mqtt.Password == "" {
		return custerror.FormatInvalidArgument("missing mqtt password")
	}
	return nil
}

func (s *WebService) UpdateOpenGateIntegrationById(ctx context.Context, req *web.UpdateOpenGateIntegrationRequest) error {
	logger.SDebug("UpdateOpenGateIntegrationById: request",
		zap.Reflect("request", req))

	if err := s.validateUpdateOpenGateIntegrationById(req); err != nil {
		logger.SError("UpdateOpenGateIntegrationById: validateUpdateOpenGateIntegrationById error", zap.Error(err))
		return err
	}

	integration, err := s.getOpenGateIntegrationById(ctx, req.OpenGateId)
	if err != nil {
		logger.SError("updateOpenGateIntegrationById: getOpenGateIntegrationById error", zap.Error(err))
		return err
	}

	if integration == nil {
		logger.SError("updateOpenGateIntegrationById: integration not found")
		return custerror.FormatNotFound("integration not found")
	}

	s.updateOpenGateIntegrationDto(req, integration)
	if err := s.updateOpenGateIntegration(ctx, integration); err != nil {
		logger.SError("updateOpenGateIntegrationById: updateOpenGateMqttConfiguration error", zap.Error(err))
		return err
	}

	mqttConfiguration, err := s.getOpenGateMqttConfigurationById(ctx, integration.MqttId)
	if err != nil {
		logger.SError("updateOpenGateIntegrationById: getOpenGateMqttConfigurationById error",
			zap.Error(err))
		return err
	}

	if mqttConfiguration == nil {
		logger.SError("updateOpenGateIntegrationById: mqtt configuration not found")
		return custerror.FormatNotFound("mqtt configuration not found")
	}

	s.updateOpenGateMqttConfigurationDto(req.Mqtt, mqttConfiguration)
	if err := s.updateOpenGateMqttConfiguration(ctx, mqttConfiguration); err != nil {
		logger.SError("updateOpenGateIntegrationById: updateOpenGateMqttConfiguration error", zap.Error(err))
		return err
	}

	return nil
}

func (s *WebService) updateOpenGateIntegrationDto(req *web.UpdateOpenGateIntegrationRequest, integration *db.OpenGateIntegration) {
	if req.LogLevel != "" {
		integration.LogLevel = req.LogLevel
	}
	if req.SnapshotRetentionDays > 0 {
		integration.SnapshotRetentionDays = req.SnapshotRetentionDays
	}
}

func (s *WebService) updateOpenGateMqttConfigurationDto(req *web.UpdateOpenGateIntegrationMqttRequest, mqtt *db.OpenGateMqttConfiguration) {
	if req.Host != "" {
		mqtt.Host = req.Host

	}
	if req.Port > 0 {
		mqtt.Port = req.Port
	}
	if req.Username != "" {
		mqtt.Username = req.Username
	}
	if req.Password != "" {
		mqtt.Password = req.Password
	}
	mqtt.Enabled = req.Enabled
}

func (s *WebService) getOpenGateMqttConfigurationById(ctx context.Context, id string) (*db.OpenGateMqttConfiguration, error) {
	q := s.builder.Select("*").
		From("open_gate_mqtt_configurations").
		Where("configuration_id = ?", id)

	sql, args, _ := q.ToSql()
	logger.SDebug("getOpenGateMqttConfigurationById: SQL",
		zap.Reflect("q", sql),
		zap.Reflect("args", args))

	var config db.OpenGateMqttConfiguration
	if err := s.db.Get(ctx, q, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (s *WebService) updateOpenGateIntegration(ctx context.Context, integration *db.OpenGateIntegration) error {
	valueMap := map[string]interface{}{}
	fields := integration.Fields()
	values := integration.Values()
	for i := 0; i < len(fields); i += 1 {
		valueMap[fields[i]] = values[i]
	}

	q := s.builder.Update("open_gate_integrations").
		Where("integration_id = ?", integration.OpenGateId).
		SetMap(valueMap)
	sql, args, _ := q.ToSql()
	logger.SDebug("updateOpenGateIntegration: SQL query",
		zap.String("query", sql),
		zap.Reflect("args", args))
	if err := s.db.Update(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) updateOpenGateMqttConfiguration(ctx context.Context, mqtt *db.OpenGateMqttConfiguration) error {
	valueMap := map[string]interface{}{}
	fields := mqtt.Fields()
	values := mqtt.Values()
	for i := 0; i < len(fields); i += 1 {
		valueMap[fields[i]] = values[i]
	}

	q := s.builder.Update("open_gate_mqtt_configurations").
		Where("configuration_id = ?", mqtt.ConfigurationId).
		SetMap(valueMap)
	sql, args, _ := q.ToSql()
	logger.SDebug("updateOpenGateMqttConfiguration: SQL query",
		zap.String("query", sql),
		zap.Reflect("args", args))
	if err := s.db.Update(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) updateGetOpenGateCameraSettingsRequest(req *web.GetOpenGateCameraSettingsRequest) error {
	if len(req.CameraId) == 0 {
		return custerror.FormatInvalidArgument("missing camera id")
	}
	return nil
}

func (s *WebService) GetOpenGateCameraSettings(ctx context.Context, req *web.GetOpenGateCameraSettingsRequest) (*web.GetOpenGateCameraSettingsResponse, error) {
	logger.SDebug("GetOpenGateCameraSettings: request", zap.Reflect("request", req))

	if err := s.updateGetOpenGateCameraSettingsRequest(req); err != nil {
		logger.SError("GetOpenGateCameraSettings: updateGetOpenGateCameraSettingsRequest error", zap.Error(err))
		return nil, err
	}

	cameras, err := s.getCameraById(ctx, req.CameraId)
	if err != nil {
		logger.SError("GetOpenGateCameraSettings: getCameraById error", zap.Error(err))
		return nil, err
	}

	if len(cameras) == 0 {
		logger.SError("GetOpenGateCameraSettings: camera not found")
		return &web.GetOpenGateCameraSettingsResponse{
			OpenGateCameraSettings: []db.OpenGateCameraSettings{},
		}, nil
	}

	allowedIds := make([]string, 0)
	for _, c := range cameras {
		allowedIds = append(allowedIds, c.SettingsId)
	}

	settings, err := s.getOpenGateCameraSettings(ctx, allowedIds)
	if err != nil {
		logger.SError("GetOpenGateCameraSettings: getOpenGateCameraSettings error", zap.Error(err))
		return nil, err
	}

	return &web.GetOpenGateCameraSettingsResponse{
		OpenGateCameraSettings: settings,
	}, nil
}

func (s *WebService) getOpenGateCameraSettings(ctx context.Context, ids []string) ([]db.OpenGateCameraSettings, error) {
	q := s.builder.Select("*").
		From("open_gate_camera_settings")

	if len(ids) > 0 {
		or := squirrel.Or{}
		for _, i := range ids {
			or = append(or, squirrel.Eq{"settings_id": i})
		}
		q = q.Where(or)
	}

	sql, args, _ := q.ToSql()
	logger.SDebug("getOpenGateCameraSettings: SQL",
		zap.Reflect("q", sql),
		zap.Reflect("args", args))

	var settings []db.OpenGateCameraSettings
	if err := s.db.Select(ctx, q, &settings); err != nil {
		return nil, err
	}

	return settings, nil
}

func (s *WebService) validateGetOpenGateMqttSettingsRequest(req *web.GetOpenGateMqttSettingsRequest) error {
	if req.ConfigurationId == "" {
		return custerror.FormatInvalidArgument("missing configuration id")
	}
	return nil
}

func (s *WebService) GetOpenGateMqttConfigurationById(ctx context.Context, req *web.GetOpenGateMqttSettingsRequest) (*web.GetOpenGateMqttSettingsResponse, error) {
	logger.SDebug("GetOpenGateMqttConfigurationById: request", zap.Reflect("request", req))

	if err := s.validateGetOpenGateMqttSettingsRequest(req); err != nil {
		logger.SError("GetOpenGateMqttConfigurationById: validateGetOpenGateMqttSettingsRequest error", zap.Error(err))
		return nil, err
	}

	config, err := s.getOpenGateMqttConfigurationById(ctx, req.ConfigurationId)
	if err != nil {
		logger.SError("GetOpenGateMqttConfigurationById: getOpenGateMqttConfigurationById error", zap.Error(err))
		return nil, err
	}

	return &web.GetOpenGateMqttSettingsResponse{
		OpenGateMqttConfiguration: config,
	}, nil
}

func (s *WebService) deleteOpenGateIntegration(ctx context.Context, id string) error {
	return s.db.Delete(ctx,
		s.builder.Delete("open_gate_integrations").
			Where("open_gate_id = ?", id))
}

func (s *WebService) deleteOpenGateMqttConfiguration(ctx context.Context, id string) error {
	return s.db.Delete(ctx,
		s.builder.Delete("open_gate_mqtt_configurations").
			Where("configuration_id = ?", id))
}

func (s *WebService) deleteDeviceById(ctx context.Context, id string) error {
	return s.db.Delete(ctx,
		s.builder.Delete("transcoders").
			Where("device_id = ?", id))
}

func (s *WebService) validateGetObjectTrackingEventByIdRequest(_ *web.GetObjectTrackingEventByIdRequest) error {
	return nil
}

func (s *WebService) GetObjectTrackingEventById(ctx context.Context, req *web.GetObjectTrackingEventByIdRequest) (*web.GetObjectTrackingEventByIdResponse, error) {
	logger.SDebug("GetObjectTrackingEventById: request", zap.Reflect("request", req))

	if err := s.validateGetObjectTrackingEventByIdRequest(req); err != nil {
		logger.SError("GetObjectTrackingEventById: validateGetObjectTrackingEventByIdRequest error", zap.Error(err))
		return nil, err
	}

	trackingEvents, err := s.getObjectTrackingEventById(ctx, req.EventId, req.OpenGateEventId)
	if err != nil {
		logger.SError("GetObjectTrackingEventById: getObjectTrackingEventById error", zap.Error(err))
		return nil, err
	}

	if trackingEvents == nil {
		logger.SError("GetObjectTrackingEventById: trackingEvents not found")
		return &web.GetObjectTrackingEventByIdResponse{
			ObjectTrackingEvents: []db.ObjectTrackingEvent{},
		}, nil
	}

	return &web.GetObjectTrackingEventByIdResponse{
		ObjectTrackingEvents: trackingEvents,
	}, nil
}

func (s *WebService) getObjectTrackingEventById(ctx context.Context, ids []string, openGateIds []string) ([]db.ObjectTrackingEvent, error) {
	q := s.builder.Select("*").
		From("object_tracking_events").
		OrderBy("frame_time DESC")

	if len(ids) > 0 {
		or := squirrel.Or{}
		for _, i := range ids {
			or = append(or, squirrel.Eq{"event_id": i})
		}
		q = q.Where(or)
	}

	if len(openGateIds) > 0 {
		or := squirrel.Or{}
		for _, i := range openGateIds {
			or = append(or, squirrel.Eq{"open_gate_event_id": i})
		}
		q = q.Where(or)
	}

	sql, args, _ := q.ToSql()
	logger.SDebug("getObjectTrackingEventById: SQL",
		zap.Reflect("q", sql),
		zap.Reflect("args", args))

	var trackingEvent []db.ObjectTrackingEvent
	if err := s.db.Select(ctx, q, &trackingEvent); err != nil {
		return nil, err
	}

	return trackingEvent, nil
}
func (s *WebService) addObjectTrackingEvent(ctx context.Context, event *db.ObjectTrackingEvent) error {
	q := s.builder.Insert("object_tracking_events").
		Columns(event.Fields()...).
		Values(event.Values()...)
	if err := s.db.Insert(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) updateObjectTrackingEvent(ctx context.Context, event *db.ObjectTrackingEvent) error {
	valueMap := map[string]interface{}{}
	fields := event.Fields()
	values := event.Values()
	for i := 0; i < len(fields); i += 1 {
		valueMap[fields[i]] = values[i]
	}

	q := s.builder.Update("object_tracking_events").
		Where("event_id = ?", event.EventId).
		SetMap(valueMap)
	sql, args, _ := q.ToSql()
	logger.SDebug("updateEvent: SQL query",
		zap.String("query", sql),
		zap.Reflect("args", args))
	if err := s.db.Update(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) validateDeleteObjectTrackingEventRequest(req *web.DeleteObjectTrackingEventRequest) error {
	if req.EventId == "" {
		return custerror.FormatInvalidArgument("missing event id")
	}
	return nil
}

func (s *WebService) DeleteObjectTrackingEvent(ctx context.Context, req *web.DeleteObjectTrackingEventRequest) error {
	logger.SInfo("commandService.DeleteObjectTrackingEvent: request", zap.Any("request", req))

	if err := s.validateDeleteObjectTrackingEventRequest(req); err != nil {
		logger.SError("commandService.DeleteObjectTrackingEvent: validateDeleteObjectTrackingEventRequest error",
			zap.Error(err))
		return err
	}

	_, err := s.getObjectTrackingEventById(ctx, []string{req.EventId}, nil)
	if err != nil {
		logger.SError("commandService.DeleteObjectTrackingEvent: getObjectTrackingEventById error",
			zap.Error(err))
		return err
	}

	if req.EventId == "" {
		logger.SError("DeleteObjectTrackingEvent: missing event_id")
		return custerror.FormatInvalidArgument("missing event_id")
	}

	if err := s.deleteObjectTrackingEvent(ctx, req.EventId); err != nil {
		logger.SDebug("DeleteObjectTrackingEvent: deleteEvent", zap.Error(err))
		return err
	}

	return nil
}

func (s *WebService) deleteObjectTrackingEvent(ctx context.Context, id string) error {
	return s.db.Delete(ctx,
		s.builder.Delete("object_tracking_events").
			Where("event_id = ?", id))
}

func (s *WebService) fromObjectTrackingEventToDto(event *events.DetectionEventStatus) *db.ObjectTrackingEvent {
	frameTimeSec, frameTimeDec := math.Modf(event.FrameTime)
	frameTime := time.Unix(int64(frameTimeSec), int64(frameTimeDec*1e9))

	startTimeSec, startTimeDec := math.Modf(event.StartTime)
	startTime := time.Unix(int64(startTimeSec), int64(startTimeDec*1e9))

	dto := &db.ObjectTrackingEvent{
		OpenGateEventId: event.ID,
		CameraName:      event.Camera,
		FrameTime:       &frameTime,
		Label:           event.Label,
		TopScore:        event.TopScore,
		Score:           event.Score,
		HasSnapshot:     event.HasSnapshot,
		HasClip:         event.HasClip,
		Stationary:      event.Stationary,
		FalsePositive:   event.FalsePositive,
		StartTime:       &startTime,
	}

	if event.EndTime != nil {
		endTimeSec, endTimeDec := math.Modf(*event.EndTime)
		endTime := time.Unix(int64(endTimeSec), int64(endTimeDec*1e9))
		dto.EndTime = &endTime
	}

	return dto
}

func (s *WebService) DoDeviceHealthcheck(ctx context.Context, req *web.DeviceHealthcheckRequest) (*web.DeviceHealthcheckResponse, error) {
	logger.SDebug("DoDeviceHealthcheck: request", zap.Reflect("request", req))

	if err := s.validateDeviceHealthcheckRequest(req); err != nil {
		logger.SError("DoDeviceHealthcheck: validateDeviceHealthcheckRequest error", zap.Error(err))
	}

	transcoder, err := s.getDeviceById(ctx, []string{req.TranscoderId})
	if err != nil {
		logger.SError("DoDeviceHealthcheck: getDeviceById error", zap.Error(err))
		return nil, err
	}
	if len(transcoder) == 0 {
		logger.SError("DoDeviceHealthcheck: transcoder not found")
		return nil, custerror.FormatNotFound("transcoder not found")
	}

	t := transcoder[0]
	r := &custmqtt.RequestReplyRequest{
		Topic: events.Event{
			Prefix:    "commands",
			ID:        t.DeviceId,
			Type:      "healthcheck",
			Arguments: []string{},
		},
		Request:    nil,
		Reply:      &web.DeviceHealthcheckResponse{},
		MaxTimeout: 5 * time.Second,
	}
	err = s.reqreply.Request(ctx, r)
	if err != nil {
		logger.SError("DoDeviceHealthcheck: request error", zap.Error(err))
		return nil, err
	}
	return r.Reply.(*web.DeviceHealthcheckResponse), nil
}

func (s *WebService) validateDeviceHealthcheckRequest(req *web.DeviceHealthcheckRequest) error {
	if req.TranscoderId == "" {
		return custerror.FormatInvalidArgument("missing transcoder id")
	}
	return nil
}

func (s *WebService) validateOpenGateStatsRequest(req *web.UpsertOpenGateCameraStatsRequest) error {
	if req.CameraName == "" {
		return custerror.FormatInvalidArgument("missing camera name")
	}
	if req.CameraFPS < 0 {
		return custerror.FormatInvalidArgument("invalid camera fps")
	}

	if req.DetectionFPS < 0 {
		return custerror.FormatInvalidArgument("invalid detection fps")
	}

	if req.CapturePID < 0 {
		return custerror.FormatInvalidArgument("invalid capture pid")
	}

	if req.ProcessID <= 0 {
		return custerror.FormatInvalidArgument("invalid invalid")
	}

	if req.ProcessFPS < 0 {
		return custerror.FormatInvalidArgument("invalid process fps")
	}

	if req.SkippedFPS < 0 {
		return custerror.FormatInvalidArgument("invalid skipped fps")
	}

	return nil
}

func (s *WebService) UpsertOpenGateCameraStats(ctx context.Context, req *web.UpsertOpenGateCameraStatsRequest) (*web.UpsertOpenGateCameraStatsResponse, error) {
	logger.SDebug("UpsertOpenGateCameraStats: upsert stats",
		zap.Reflect("request", req))
	if err := s.validateOpenGateStatsRequest(req); err != nil {
		logger.SError("UpsertOpenGateCameraStats: validateOpenGateStatsRequest error",
			zap.Error(err))
		return nil, err
	}
	currentStats, err := s.getOpenGateCameraStats(ctx, req.TranscoderId, req.CameraName)
	switch {
	case err == nil:
		patched := s.patchOpenGateCameraStats(currentStats, req)
		if err := s.updateOpenGateCameraStats(ctx, patched); err != nil {
			logger.SError("UpsertOpenGateCameraStats: updateOpenGateCameraStats error",
				zap.Error(err))
			return nil, err
		}
		return &web.UpsertOpenGateCameraStatsResponse{
			CameraStatId: currentStats.CameraStatId,
		}, nil
	case errors.Is(err, custerror.ErrorNotFound):
		stats := db.OpenGateCameraStats{
			CameraStatId: uuid.New(),
			CameraName:   req.CameraName,
			TranscoderId: req.TranscoderId,
			Timestamp:    time.Now(),
			CameraFPS:    req.CameraFPS,
			DetectionFPS: req.DetectionFPS,
			CapturePID:   req.CapturePID,
			ProcessID:    req.ProcessID,
			ProcessFPS:   req.ProcessFPS,
			SkippedFPS:   req.SkippedFPS,
		}
		if err := s.addOpenGateCameraStats(ctx, &stats); err != nil {
			logger.SError("UpsertOpenGateCameraStats: AddOpenGateCameraStats error",
				zap.Error(err))
			return nil, err
		}
		return &web.UpsertOpenGateCameraStatsResponse{
			CameraStatId: stats.CameraStatId,
		}, nil
	default:
		logger.SError("UpsertOpenGateCameraStats: getOpenGateCameraStats error",
			zap.Error(err))
		return nil, err
	}
}

func (s *WebService) patchOpenGateCameraStats(old *db.OpenGateCameraStats, req *web.UpsertOpenGateCameraStatsRequest) *db.OpenGateCameraStats {
	new := old
	if req.CameraName != "" {
		new.CameraName = req.CameraName
	}
	if req.TranscoderId != "" {
		new.TranscoderId = req.TranscoderId
	}
	if req.CameraFPS > 0 {
		new.CameraFPS = req.CameraFPS
	}
	if req.DetectionFPS > 0 {
		new.DetectionFPS = req.DetectionFPS
	}
	if req.CapturePID > 0 {
		new.CapturePID = req.CapturePID
	}
	if req.ProcessID > 0 {
		new.ProcessID = req.ProcessID
	}
	if req.ProcessFPS > 0 {
		new.ProcessFPS = req.ProcessFPS
	}
	if req.SkippedFPS > 0 {
		new.SkippedFPS = req.SkippedFPS
	}
	return new
}

func (s *WebService) updateOpenGateCameraStats(ctx context.Context, m *db.OpenGateCameraStats) error {
	valueMap := map[string]interface{}{}
	fields := m.Fields()
	values := m.Values()
	for i := 0; i < len(fields); i += 1 {
		valueMap[fields[i]] = values[i]
	}

	q := s.builder.Update("open_gate_camera_stats").
		Where("camera_stat_id = ?", m.CameraStatId).
		SetMap(valueMap)
	sql, args, _ := q.ToSql()
	logger.SDebug("updateOpenGateCameraStats: SQL query",
		zap.String("query", sql),
		zap.Reflect("args", args))
	if err := s.db.Update(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) getOpenGateCameraStats(ctx context.Context, transcoderId string, cameraName string) (*db.OpenGateCameraStats, error) {
	q := s.builder.Select("*").
		From("open_gate_camera_stats").
		Where("transcoder_id = ?", transcoderId).
		Where("camera_name = ?", cameraName).
		OrderByClause("? DESC", "timestamp").
		Limit(1)

	sql, args, _ := q.ToSql()
	logger.SDebug("getOpenGateCameraStats: SQL",
		zap.Reflect("q", sql),
		zap.Reflect("args", args))

	var stats []db.OpenGateCameraStats
	if err := s.db.Select(ctx, q, &stats); err != nil {
		return nil, err
	}
	if len(stats) == 0 {
		return nil, custerror.FormatNotFound("camera stats not found")
	}

	return &stats[0], nil
}

func (s *WebService) addOpenGateCameraStats(ctx context.Context, stats *db.OpenGateCameraStats) error {
	q := s.builder.Insert("open_gate_camera_stats").
		Columns(stats.Fields()...).
		Values(stats.Values()...)
	if err := s.db.Insert(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) UpsertOpenGateDetectorStats(ctx context.Context, req *web.UpsertOpenGateDetectorsStatsRequest) (*web.UpsertOpenGateDetectorsStatsResponse, error) {
	logger.SDebug("UpsertOpenGateDetectorStats: add stats", zap.Reflect("request", req))

	err := s.validateAddOpenGateDetectorStatsRequest(req)
	if err != nil {
		return nil, err
	}

	currentStats, err := s.getOpenGateDetectorStats(ctx, req.TranscoderId, req.DetectorName)
	switch {
	case err == nil:
		patched := s.patchOpenGateDetectorStats(currentStats, req)
		if err := s.updateOpenGateDetectorStats(ctx, patched); err != nil {
			logger.SError("UpsertOpenGateDetectorStats: updateOpenGateDetectorStats error", zap.Error(err))
			return nil, err
		}
		return &web.UpsertOpenGateDetectorsStatsResponse{
			DetectorStatId: currentStats.DetectorStatId,
		}, nil
	case errors.Is(err, custerror.ErrorNotFound):
		stats := db.OpenGateDetectorStats{
			DetectorStatId: uuid.New(),
			DetectorName:   req.DetectorName,
			TranscoderId:   req.TranscoderId,
			Timestamp:      time.Now(),
			DetectorStart:  req.DetectorStart,
			InferenceSpeed: req.InferenceSpeed,
			ProcessID:      req.ProcessID,
		}
		if err := s.addOpenGateDetectorStats(ctx, &stats); err != nil {
			logger.SError("UpsertOpenGateDetectorStats: addOpenGateDetectorStats error", zap.Error(err))
			return nil, err
		}
		return &web.UpsertOpenGateDetectorsStatsResponse{
			DetectorStatId: stats.DetectorStatId,
		}, nil
	default:
		logger.SError("UpsertOpenGateDetectorStats: getOpenGateDetectorStats error", zap.Error(err))
		return nil, err
	}
}

func (s *WebService) patchOpenGateDetectorStats(old *db.OpenGateDetectorStats, req *web.UpsertOpenGateDetectorsStatsRequest) *db.OpenGateDetectorStats {
	new := old
	if req.DetectorName != "" {
		new.DetectorName = req.DetectorName
	}
	if req.TranscoderId != "" {
		new.TranscoderId = req.TranscoderId
	}
	if req.DetectorStart > 0 {
		new.DetectorStart = req.DetectorStart
	}
	if req.InferenceSpeed > 0 {
		new.InferenceSpeed = req.InferenceSpeed
	}
	if req.ProcessID > 0 {
		new.ProcessID = req.ProcessID
	}
	return new
}

func (s *WebService) validateAddOpenGateDetectorStatsRequest(req *web.UpsertOpenGateDetectorsStatsRequest) error {
	if req.DetectorName == "" {
		return custerror.FormatInvalidArgument("missing camera name")
	}
	if req.DetectorStart < 0 {
		return custerror.FormatInvalidArgument("invalid detection start time")
	}
	if req.InferenceSpeed < 0 {
		return custerror.FormatInvalidArgument("invalid inference speed")
	}
	if req.ProcessID <= 0 {
		return custerror.FormatInvalidArgument("invalid PID")
	}
	return nil
}

func (s *WebService) getOpenGateDetectorStats(ctx context.Context, transcoderId string, detectorName string) (*db.OpenGateDetectorStats, error) {
	q := s.builder.Select("*").
		From("open_gate_detector_stats").
		Where("transcoder_id = ?", transcoderId).
		Where("detector_name = ?", detectorName).
		OrderByClause("? DESC", "timestamp").
		Limit(1)

	sql, args, _ := q.ToSql()
	logger.SDebug("getOpenGateDetectorStats: SQL",
		zap.Reflect("q", sql),
		zap.Reflect("args", args))

	var stats []db.OpenGateDetectorStats
	if err := s.db.Select(ctx, q, &stats); err != nil {
		return nil, err
	}
	if len(stats) == 0 {
		return nil, custerror.FormatNotFound("detector stats not found")
	}

	return &stats[0], nil
}

func (s *WebService) updateOpenGateDetectorStats(ctx context.Context, m *db.OpenGateDetectorStats) error {
	valueMap := map[string]interface{}{}
	fields := m.Fields()
	values := m.Values()
	for i := 0; i < len(fields); i += 1 {
		valueMap[fields[i]] = values[i]
	}

	q := s.builder.Update("open_gate_detector_stats").
		Where("detector_stat_id = ?", m.DetectorStatId).
		SetMap(valueMap)
	sql, args, _ := q.ToSql()
	logger.SDebug("updateOpenGateDetectorStats: SQL query",
		zap.String("query", sql),
		zap.Reflect("args", args))
	if err := s.db.Update(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) addOpenGateDetectorStats(ctx context.Context, stats *db.OpenGateDetectorStats) error {
	q := s.builder.Insert("open_gate_detector_stats").
		Columns(stats.Fields()...).
		Values(stats.Values()...)

	sql, args, _ := q.ToSql()
	logger.SDebug("addOpenGateDetectorStats: SQL",
		zap.Reflect("q", sql),
		zap.Reflect("args", args))

	if err := s.db.Insert(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) GetLatestOpenGateCameraStats(ctx context.Context, req *web.GetLatestOpenGateCameraStatsRequest) (*web.GetLatestOpenGateStatsResponse, error) {
	logger.SDebug("GetLatestOpenGateCameraStats: request")

	if err := s.validateGetLatestOpenGateCameraStats(req); err != nil {
		logger.SError("GetLatestOpenGateCameraStats: validateGetLatestOpenGateCameraStats error", zap.Error(err))
		return nil, err
	}

	cameraStats, err := s.getLatestOpenGateCameraStats(ctx, req.TranscoderId, req.CameraNames)
	if err != nil {
		logger.SError("GetLatestOpenGateCameraStats: getLatestOpenGateCameraStats error", zap.Error(err))
		return nil, err
	}

	detectorStats, err := s.getLatestOpenGateDetectorStats(ctx, req.TranscoderId)
	if err != nil {
		logger.SError("GetLatestOpenGateCameraStats: getLatestOpenGateDetectorStats error", zap.Error(err))
		return nil, err
	}

	return &web.GetLatestOpenGateStatsResponse{
		CameraStats:   cameraStats,
		DetectorStats: detectorStats,
	}, nil
}

func (s *WebService) validateGetLatestOpenGateCameraStats(req *web.GetLatestOpenGateCameraStatsRequest) error {
	if req.TranscoderId == "" {
		return custerror.FormatInvalidArgument("missing transcoder id")
	}
	if req.CameraNames == nil {
		return custerror.FormatInvalidArgument("missing camera names")
	}
	if len(req.CameraNames) == 0 {
		return custerror.FormatInvalidArgument("missing camera names")
	}
	return nil
}

func (s *WebService) getLatestOpenGateCameraStats(ctx context.Context, transcoderId string, names []string) ([]db.OpenGateCameraStats, error) {
	q := s.builder.Select("*").
		From("open_gate_camera_stats").
		OrderByClause("? DESC", "timestamp").
		Limit(1)
	if transcoderId != "" {
		q = q.Where("transcoder_id = ?", transcoderId)
	}
	if len(names) > 0 {
		or := squirrel.Or{}
		for _, n := range names {
			or = append(or, squirrel.Eq{"camera_name": n})
		}
		q = q.Where(or)
	}

	sql, args, _ := q.ToSql()
	logger.SDebug("getOpenGateCameraStats: SQL",
		zap.Reflect("q", sql),
		zap.Reflect("args", args))

	var stats []db.OpenGateCameraStats
	if err := s.db.Select(ctx, q, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

func (s *WebService) getLatestOpenGateDetectorStats(ctx context.Context, transcoderId string) ([]db.OpenGateDetectorStats, error) {
	q := s.builder.Select("*").
		From("open_gate_detector_stats").
		Where("transcoder_id = ?", transcoderId).
		OrderByClause("? DESC", "timestamp").
		Limit(1)

	sql, args, _ := q.ToSql()
	logger.SDebug("getOpenGateDetectorStats: SQL",
		zap.Reflect("q", sql),
		zap.Reflect("args", args))

	var stats []db.OpenGateDetectorStats
	if err := s.db.Select(ctx, q, &stats); err != nil {
		return nil, err
	}
	return stats, nil
}

func (s *WebService) DeleteOpengateCameraStats(ctx context.Context) error {
	logger.SInfo("DeleteOpengateCameraStats: cron job")
	return s.db.Delete(ctx,
		s.builder.Delete("open_gate_camera_stats"))
}

func (s *WebService) DeleteOpengateDetectorStats(ctx context.Context) error {
	logger.SInfo("DeleteOpengateDetectorStats: cron job")
	return s.db.Delete(ctx,
		s.builder.Delete("open_gate_detector_stats"))
}

func (s *WebService) GetDetectablePeople(ctx context.Context, req *web.GetDetectablePeopleRequest) (*web.GetDetectablePeopleResponse, error) {
	logger.SInfo("GetDetectablePeople: request",
		zap.Reflect("request", req))

	if err := s.validateGetDetectablePeopleRequest(req); err != nil {
		logger.SError("GetDetectablePeople: validateGetDetectablePeopleRequest",
			zap.Error(err))
		return nil, err
	}

	people, err := s.getPersonById(ctx, req.PersonIds)
	if err != nil {
		logger.SError("GetDetectablePeople: getPersonById",
			zap.Error(err))
		return nil, err
	}
	if len(people) == 0 {
		return &web.GetDetectablePeopleResponse{
			People: []db.DetectablePerson{},
		}, nil
	}

	return &web.GetDetectablePeopleResponse{
		People: people,
	}, nil
}

func (s *WebService) validateGetDetectablePeopleRequest(_ *web.GetDetectablePeopleRequest) error {
	return nil
}

func (s *WebService) getPersonById(ctx context.Context, ids []string) ([]db.DetectablePerson, error) {
	q := s.builder.Select("*").
		From("detectable_people")

	if len(ids) > 0 {
		or := squirrel.Or{}
		for _, i := range ids {
			or = append(or, squirrel.Eq{"person_id": i})
		}
		q = q.Where(or)
	}

	sql, args, _ := q.ToSql()
	logger.SDebug("getPersonById: SQL",
		zap.Reflect("q", sql),
		zap.Reflect("args", args))

	var people []db.DetectablePerson
	if err := s.db.Select(ctx, q, &people); err != nil {
		return nil, err
	}

	return people, nil
}

func (s *WebService) GetDetectablePersonImagePresignedUrl(ctx context.Context, req *web.GetDetectablePeopleImagePresignedUrlRequest) (*web.GetDetectablePeopleImagePresignedUrlResponse, error) {
	logger.SInfo("GetPersonImagePresignedUrl: request",
		zap.Reflect("request", req))

	if err := s.validateGetPersonImagePresignedUrlRequest(req); err != nil {
		logger.SError("GetPersonImagePresignedUrl: validateGetPersonImagePresignedUrlRequest",
			zap.Error(err))
		return nil, err
	}

	person, err := s.getPersonById(ctx, []string{req.PersonId})
	if err != nil {
		logger.SError("GetPersonImagePresignedUrl: getPersonById",
			zap.Error(err))
		return nil, err
	}

	if len(person) == 0 {
		logger.SError("GetPersonImagePresignedUrl: person not found")
		return nil, custerror.FormatNotFound("person not found")
	}

	url, err := s.MediaHelper.GetPresignedUrl(
		ctx,
		&media.GetPresignedUrlRequest{
			Path: person[0].
				ImagePath,
			Type: media.AssetsTypePeople,
		})
	if err != nil {
		logger.SError("GetPersonImagePresignedUrl: getPresignedUrl",
			zap.Error(err))
		return nil, err
	}

	return &web.GetDetectablePeopleImagePresignedUrlResponse{
		PresignedUrl: url,
	}, nil
}

func (s *WebService) validateGetPersonImagePresignedUrlRequest(req *web.GetDetectablePeopleImagePresignedUrlRequest) error {
	if req.PersonId == "" {
		return custerror.FormatInvalidArgument("missing person id")
	}
	return nil
}

func (s *WebService) AddDetectablePerson(ctx context.Context, req *web.AddDetectablePersonRequest) (*web.AddDetectablePersonResponse, error) {
	logger.SInfo("AddDetectablePerson: request",
		zap.String("request", req.String()))

	if err := s.validateAddDetectablePersonRequest(req); err != nil {
		logger.SError("AddDetectablePerson: validateAddDetectablePersonRequest",
			zap.Error(err))
		return nil, err
	}

	faces, err := s.cvs.Detect(ctx, &DetectRequest{
		Base64Image: req.Base64Image,
	})
	if err != nil {
		logger.SError("AddDetectablePerson: detect error",
			zap.Error(err))
		return nil, err
	}
	if len(faces) == 0 {
		logger.SError("AddDetectablePerson: no face detected")
		return nil, custerror.FormatInternalError("no face detected")
	}
	if len(faces) > 1 {
		logger.SError("AddDetectablePerson: multiple faces detected")
		return nil, custerror.FormatInternalError("multiple faces detected")
	}

	firstFace := faces[0]
	similarPeople, err := s.cvs.Search(ctx, &SearchRequest{
		Vector:     firstFace.Descriptor,
		TopKResult: 1,
	})
	switch {
	case errors.Is(err, custerror.ErrorNotFound):
	case err == nil:
		if len(similarPeople) > 0 {
			logger.SError("AddDetectablePerson: similar person already exists")
			return nil, custerror.FormatAlreadyExists("similar person already exists")
		}
	default:
		logger.SError("AddDetectablePerson: search error",
			zap.Error(err))
		return nil, err
	}

	id, err := s.recordDetectablePerson(ctx, req, firstFace)
	if err != nil {
		logger.SError("AddDetectablePerson: recordDetectablePerson error",
			zap.Error(err))
		return nil, err
	}
	return &web.AddDetectablePersonResponse{
		PersonId: id,
	}, nil
}

func (s *WebService) validateAddDetectablePersonRequest(req *web.AddDetectablePersonRequest) error {
	if req.Name == "" {
		return custerror.FormatInvalidArgument("missing name")
	}
	if req.Age == "" {
		return custerror.FormatInvalidArgument("missing age")
	}
	if req.Base64Image == "" {
		return custerror.FormatInvalidArgument("missing image path")
	}
	return nil
}

func (s *WebService) recordDetectablePerson(ctx context.Context, req *web.AddDetectablePersonRequest, f face.Face) (string, error) {
	var wg sync.WaitGroup
	wg.Add(2)
	var s3Error error
	var postgresError error
	id := uuid.New().
		String()
	go func() {
		model := db.DetectablePerson{
			PersonId:  id,
			Name:      req.Name,
			Age:       req.Age,
			ImagePath: id,
			Embedding: helper.ToPgvector(f.Descriptor),
		}
		postgresError = s.cvs.Record(ctx, &model)
		wg.Done()
	}()
	go func() {
		fileDesc := media.UploadImageRequest{
			Base64Image: req.Base64Image,
			Path:        id,
			Type:        media.AssetsTypePeople,
		}
		s3Error = s.MediaHelper.UploadImage(ctx, &fileDesc)
		wg.Done()
	}()
	wg.Wait()
	if s3Error != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.cvs.Remove(context.Background(), id)
		}()
		return "", s3Error
	}
	if postgresError != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.MediaHelper.DeleteImage(context.Background(), &media.DeleteImageRequest{
				Path: id,
				Type: media.AssetsTypePeople,
			})
		}()
		return "", postgresError
	}
	wg.Wait()
	return id, nil
}

func (s *WebService) DeleteDetectablePerson(ctx context.Context, req *web.DeleteDetectablePersonRequest) error {
	logger.SInfo("DeleteDetectablePerson: request",
		zap.Reflect("request", req))

	if err := s.validateDeleteDetectablePersonRequest(req); err != nil {
		logger.SError("DeleteDetectablePerson: validateDeleteDetectablePersonRequest",
			zap.Error(err))
		return err
	}

	person, err := s.getPersonById(ctx, []string{req.PersonId})
	if err != nil {
		logger.SError("DeleteDetectablePerson: getPersonById",
			zap.Error(err))
		return err
	}

	if len(person) == 0 {
		logger.SError("DeleteDetectablePerson: person not found")
		return custerror.FormatNotFound("person not found")
	}

	if err := s.deleteDetectablePerson(ctx, req.PersonId); err != nil {
		logger.SError("DeleteDetectablePerson: deleteDetectablePerson",
			zap.Error(err))
		return err
	}

	return nil
}

func (s *WebService) validateDeleteDetectablePersonRequest(req *web.DeleteDetectablePersonRequest) error {
	if req.PersonId == "" {
		return custerror.FormatInvalidArgument("missing person id")
	}
	return nil
}

func (s *WebService) deleteDetectablePerson(ctx context.Context, id string) error {
	var wg sync.WaitGroup
	wg.Add(2)
	var s3Error error
	var postgresError error
	go func() {
		postgresError = s.cvs.Remove(ctx, id)
		wg.Done()
	}()
	go func() {
		s3Error = s.MediaHelper.DeleteImage(ctx, &media.DeleteImageRequest{
			Path: id,
			Type: media.AssetsTypePeople,
		})
		wg.Done()
	}()
	wg.Wait()
	switch {
	case s3Error != nil:
		return s3Error
	case postgresError != nil:
		return postgresError
	default:
		return nil
	}
}

func (s *WebService) GetSnapshotPresignedUrl(ctx context.Context, req *web.GetSnapshotPresignedUrlRequest) (*web.GetSnapshotPresignedUrlResponse, error) {
	logger.SInfo("GetSnapshotPresignedUrl: request",
		zap.Reflect("request", req))

	if err := s.validateGetSnapshotPresignedUrlRequest(req); err != nil {
		logger.SError("GetSnapshotPresignedUrl: validateGetSnapshotPresignedUrlRequest",
			zap.Error(err))
		return nil, err
	}

	snapshots, err := s.getSnapshotById(ctx, req.SnapshotId)
	if err != nil {
		logger.SError("GetSnapshotPresignedUrl: getSnapshotById", zap.Error(err))
		return nil, err
	}

	presignedUrl := make(map[string]string)
	for _, snap := range snapshots {
		url, err := s.MediaHelper.GetPresignedUrl(
			ctx,
			&media.GetPresignedUrlRequest{
				Path: snap.SnapshotId,
				Type: media.AssetsTypeSnapshot,
			})
		if err != nil {
			logger.SError("GetSnapshotPresignedUrl: getPresignedUrl",
				zap.Error(err))
			return nil, err
		}
		presignedUrl[snap.SnapshotId] = url
	}

	return &web.GetSnapshotPresignedUrlResponse{
		PresignedUrl: presignedUrl,
		Snapshots:    snapshots,
	}, nil
}

func (s *WebService) getSnapshotById(ctx context.Context, ids []string) ([]db.Snapshot, error) {
	q := s.builder.Select("*").
		From("snapshots")

	if len(ids) > 0 {
		or := squirrel.Or{}
		for _, i := range ids {
			or = append(or, squirrel.Eq{"snapshot_id": i})
		}
		q = q.Where(or)
	}

	sql, args, _ := q.ToSql()
	logger.SDebug("getSnapshotById: SQL",
		zap.Reflect("q", sql),
		zap.Reflect("args", args))

	var snapshots []db.Snapshot
	if err := s.db.Select(ctx, q, &snapshots); err != nil {
		return nil, err
	}

	return snapshots, nil

}

func (s *WebService) validateGetSnapshotPresignedUrlRequest(req *web.GetSnapshotPresignedUrlRequest) error {
	if len(req.SnapshotId) == 0 {
		return custerror.FormatInvalidArgument("missing snapshot id")
	}
	return nil
}

func (s *WebService) UpsertSnapshot(ctx context.Context, req *web.UpsertSnapshotRequest) error {
	if err := s.validateUpsertSnapshot(req); err != nil {
		logger.SError("UpsertSnapshot: validateUpdateSnapshotRequest",
			zap.Error(err))
		return err
	}
	snapshotId := ""
	needDetection := true
	currentSnapshot, err := s.getCurrentSnapshot(ctx,
		req.OpenGateEventId,
		req.TranscoderId)
	switch {
	case errors.Is(err, custerror.ErrorNotFound):
		snapshot := s.buildSnapshotModel(req)

		snapshotId = snapshot.SnapshotId
		needDetection = true
		detectionRes, err := s.asyncUploadToS3AndDetectPerson(
			ctx,
			needDetection,
			snapshotId,
			req.RawImage)
		if err != nil {
			logger.SError("UpsertSnapshot: updateSnapshotToS3",
				zap.Error(err))
			return err
		}
		if detectionRes.detectedPerson != nil {
			snapshot.DetectedPeopleId = &detectionRes.
				detectedPerson.
				PersonId
		}
		if err := s.addSnapshot(ctx, snapshot); err != nil {
			logger.SError("UpsertSnapshot: addSnapshot",
				zap.Error(err))
			return err
		}
	case err == nil:
		currentSnapshot.Timestamp = time.Now()
		snapshotId = currentSnapshot.SnapshotId
		if currentSnapshot.DetectedPeopleId != nil {
			needDetection = false
		}
		detectionRes, err := s.asyncUploadToS3AndDetectPerson(
			ctx,
			needDetection,
			snapshotId,
			req.RawImage)
		if err != nil {
			logger.SError("UpsertSnapshot: updateSnapshotToS3",
				zap.Error(err))
			return err
		}
		if detectionRes.detectedPerson != nil {
			currentSnapshot.DetectedPeopleId = &detectionRes.
				detectedPerson.
				PersonId
		}
		if err := s.updateSnapshot(ctx, currentSnapshot); err != nil {
			logger.SError("UpsertSnapshot: updateSnapshot",
				zap.Error(err))
			return err
		}
	default:
		logger.SError("UpsertSnapshot: getCurrentSnapshot",
			zap.Error(err))
		return err
	}

	if err := s.updateEventSnapshotReference(ctx, req.OpenGateEventId, snapshotId); err != nil {
		logger.SError("UpsertSnapshot: updateEventSnapshotReference",
			zap.Error(err))
		return err
	}

	return nil
}

func (s *WebService) updateEventSnapshotReference(ctx context.Context, openGateEventId string, snapshotId string) error {
	events, err := s.getObjectTrackingEventById(ctx, nil, []string{openGateEventId})
	if err != nil {
		logger.SError("updateEventSnapshotReference: getEventByOpenGateEventId",
			zap.Error(err))
		if errors.Is(err, custerror.ErrorNotFound) {
			logger.SInfo("updateEventSnapshotReference: event not found")
			return nil
		}
		return err
	}
	if len(events) == 0 {
		logger.SInfo("updateEventSnapshotReference: event not found")
		return nil
	}
	event := events[0]
	if event.SnapshotId == nil {
		event.SnapshotId = &snapshotId
		if err := s.updateObjectTrackingEvent(ctx, &event); err != nil {
			logger.SError("updateEventSnapshotReference: updateEvent", zap.Error(err))
			return err
		}
		logger.SInfo("updateEventSnapshotReference: event updated",
			zap.String("openGateEventId", openGateEventId))
		return nil
	}
	if *event.SnapshotId != snapshotId {
		event.SnapshotId = &snapshotId
		if err := s.updateObjectTrackingEvent(ctx, &event); err != nil {
			logger.SError("updateEventSnapshotReference: updateEvent",
				zap.Error(err))
			return err
		}
		logger.SInfo("updateEventSnapshotReference: event updated",
			zap.String("openGateEventId", openGateEventId))
		return nil
	}
	return nil
}

func (s *WebService) validateUpsertSnapshot(req *web.UpsertSnapshotRequest) error {
	if req.RawImage == "" {
		return custerror.FormatInvalidArgument("missing image raw bytes")
	}
	return nil
}

func (s *WebService) buildSnapshotModel(req *web.UpsertSnapshotRequest) *db.Snapshot {
	return &db.Snapshot{
		SnapshotId:       uuid.NewString(),
		Timestamp:        time.Now(),
		TranscoderId:     req.TranscoderId,
		OpenGateEventId:  req.OpenGateEventId,
		DetectedPeopleId: nil,
	}
}

func (s *WebService) addSnapshot(ctx context.Context, snapshot *db.Snapshot) error {
	q := s.builder.Insert("snapshots").
		Columns(snapshot.Fields()...).
		Values(snapshot.Values()...)
	if err := s.db.Insert(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) getCurrentSnapshot(ctx context.Context, openGateEventId string, transcoderId string) (*db.Snapshot, error) {
	q := s.builder.Select("*").
		From("snapshots").
		Where("open_gate_event_id = ?", openGateEventId).
		Where("transcoder_id = ?", transcoderId).
		OrderByClause("? DESC", "timestamp").
		Limit(1)

	sql, args, _ := q.ToSql()
	logger.SDebug("getCurrentSnapshot: SQL",
		zap.Reflect("q", sql),
		zap.Reflect("args", args))

	var snapshot db.Snapshot
	if err := s.db.Get(ctx, q, &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func (s *WebService) updateSnapshot(ctx context.Context, snapshot *db.Snapshot) error {
	valueMap := map[string]interface{}{}
	fields := snapshot.Fields()
	values := snapshot.Values()
	for i := 0; i < len(fields); i += 1 {
		valueMap[fields[i]] = values[i]
	}

	q := s.builder.Update("snapshots").
		Where("open_gate_event_id = ?", snapshot.OpenGateEventId).
		Where("transcoder_id = ?", snapshot.TranscoderId).
		SetMap(valueMap)
	sql, args, _ := q.ToSql()
	logger.SDebug("updateSnapshot: SQL query",
		zap.String("query", sql),
		zap.Reflect("args", args))
	if err := s.db.Update(ctx, q); err != nil {
		return err
	}
	return nil
}

type asyncUploadToS3AndDetectPersonResult struct {
	detectedPerson *db.DetectablePerson
}

func (s *WebService) asyncUploadToS3AndDetectPerson(
	ctx context.Context,
	needDetection bool,
	snapshotId string,
	base64Image string) (*asyncUploadToS3AndDetectPersonResult, error) {
	var wg sync.WaitGroup
	var s3Error error
	var detectionError error
	wg.Add(1)
	if needDetection {
		wg.Add(1)
	}
	go func() {
		fileDesc := media.UploadImageRequest{
			Base64Image: base64Image,
			Path:        snapshotId,
			Type:        media.AssetsTypeSnapshot,
		}
		s3Error = s.MediaHelper.UploadImage(ctx, &fileDesc)
		if s3Error != nil {
			logger.SError("asyncUploadToS3AndDetectPerson: upload error",
				zap.Error(s3Error))
		}
		wg.Done()
	}()
	var detectedPerson *db.DetectablePerson
	if needDetection {
		go func() {
			defer wg.Done()
			detectionResp, err := s.cvs.Detect(ctx, &DetectRequest{
				Base64Image: base64Image,
			})
			if err != nil {
				detectionError = err
				return
			}
			if len(detectionResp) == 0 {
				logger.SInfo("asyncUploadToS3AndDetectPerson: no face detected",
					zap.String("snapshotId", snapshotId))
				return
			}
			recognizeResp, err := s.cvs.Search(ctx, &SearchRequest{
				Vector: detectionResp[0].
					Descriptor,
				TopKResult: 1,
			})
			if err != nil {
				detectionError = err
				return
			}
			if len(recognizeResp) == 0 {
				logger.SInfo("asyncUploadToS3AndDetectPerson: no similar person found",
					zap.String("snapshotId", snapshotId))
				return
			}
			h := db.PersonHistory{
				HistoryId: uuid.NewString(),
				PersonId:  recognizeResp[0].PersonId,
				Timestamp: time.Now(),
				EventId:   snapshotId,
			}
			if err := s.recordPersonHistory(ctx, &h); err != nil {
				detectionError = err
				return
			}
			detectedPerson = &recognizeResp[0]
		}()
	}
	wg.Wait()
	if detectionError != nil {
		return nil, custerror.FormatInternalError("detection error: %s", detectionError)
	}
	if s3Error != nil {
		return nil, custerror.FormatInternalError("s3 error: %s", s3Error)
	}
	return &asyncUploadToS3AndDetectPersonResult{
		detectedPerson: detectedPerson,
	}, nil
}

func (s *WebService) recordPersonHistory(ctx context.Context, history *db.PersonHistory) error {
	q := s.builder.Insert("person_history").
		Columns(history.Fields()...).
		Values(history.Values()...)
	if err := s.db.Insert(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) getPersonHistory(ctx context.Context, personId []string, historyId []string) ([]db.PersonHistory, error) {
	q := s.builder.Select("*").
		From("person_history").
		OrderBy("timestamp DESC")

	if len(personId) > 0 {
		or := squirrel.Or{}
		for _, i := range personId {
			or = append(or, squirrel.Eq{"person_id": i})
		}
		q = q.Where(or)
	}
	if len(historyId) > 0 {
		or := squirrel.Or{}
		for _, i := range historyId {
			or = append(or, squirrel.Eq{"history_id": i})
		}
		q = q.Where(or)
	}

	sql, args, _ := q.ToSql()
	logger.SDebug("getPersonHistory: SQL",
		zap.Reflect("q", sql),
		zap.Reflect("args", args))

	var history []db.PersonHistory
	if err := s.db.Select(ctx, q, &history); err != nil {
		return nil, err
	}

	return history, nil
}

func (s *WebService) GetPersonHistory(ctx context.Context, req *web.GetPersonHistoryRequest) (*web.GetPersonHistoryResponse, error) {
	logger.SInfo("GetPersonHistory: request",
		zap.Reflect("request", req))

	if err := s.validateGetPersonHistoryRequest(req); err != nil {
		logger.SError("GetPersonHistory: validateGetPersonHistoryRequest",
			zap.Error(err))
		return nil, err
	}

	history, err := s.getPersonHistory(ctx, req.PersonId, req.HistoryId)
	if err != nil {
		logger.SError("GetPersonHistory: getPersonHistory",
			zap.Error(err))
		return nil, err
	}

	return &web.GetPersonHistoryResponse{
		Histories: history,
	}, nil
}

func (s *WebService) validateGetPersonHistoryRequest(req *web.GetPersonHistoryRequest) error {
	if len(req.PersonId) == 0 && len(req.HistoryId) == 0 {
		return custerror.FormatInvalidArgument("missing person id or history id")
	}
	return nil
}

func (s *WebService) GetTranscoderStatus(ctx context.Context, req *web.GetTranscoderStatusRequest) (*web.GetTranscoderStatusResponse, error) {
	logger.SInfo("GetTranscoderStatus: request",
		zap.Reflect("request", req))

	if err := s.validateGetTranscoderStatusRequest(req); err != nil {
		logger.SError("GetTranscoderStatus: validateGetTranscoderStatusRequest",
			zap.Error(err))
		return nil, err
	}

	status, err := s.getTranscoderStatus(ctx, req.TranscoderId, req.CameraId)
	if err != nil {
		logger.SError("GetTranscoderStatus: getTranscoderStatus",
			zap.Error(err))
		return nil, err
	}

	return &web.GetTranscoderStatusResponse{
		Status: status,
	}, nil
}

func (s *WebService) UpdateTranscoderStatus(ctx context.Context, req *web.UpdateTranscoderStatusRequest) error {
	logger.SInfo("UpdateTranscoderStatus: request",
		zap.Reflect("request", req))

	if err := s.validateUpdateTranscoderStatusRequest(req); err != nil {
		logger.SError("UpdateTranscoderStatus: validateUpdateTranscoderStatusRequest",
			zap.Error(err))
		return err
	}
	cameraId := ""
	if req.CameraId != nil {
		cameraId = *req.CameraId
	} else {
		cameraResp, err := s.getCameraByName(ctx, []string{*req.CameraName})
		if err != nil {
			logger.SError("UpdateTranscoderStatus: getCameraByName",
				zap.Error(err))
			return err
		}
		if len(cameraResp) == 0 {
			logger.SError("UpdateTranscoderStatus: camera not found",
				zap.String("cameraName", *req.CameraName))
			return custerror.FormatNotFound("camera not found")
		}
		cameraId = cameraResp[0].CameraId
		req.CameraId = &cameraId
	}

	status, err := s.getTranscoderStatus(ctx,
		[]string{req.TranscoderId},
		[]string{cameraId})
	switch {
	case errors.Is(err, custerror.ErrorNotFound):
		if err := s.addTranscoderStatus(ctx, req); err != nil {
			logger.SError("UpdateTranscoderStatus: addTranscoderStatus",
				zap.Error(err))
			return err
		}
	case err == nil:
		if len(status) == 0 {
			if err := s.addTranscoderStatus(ctx, req); err != nil {
				logger.SError("UpdateTranscoderStatus: addTranscoderStatus",
					zap.Error(err))
				return err
			}
			return nil
		}
		if err := s.updateTranscoderStatus(ctx, &status[0], req); err != nil {
			logger.SError("UpdateTranscoderStatus: updateTranscoderStatus",
				zap.Error(err))
			return err
		}
	default:
		logger.SError("UpdateTranscoderStatus: getTranscoderStatus",
			zap.Error(err))
		return err
	}

	return nil
}

func (s *WebService) validateUpdateTranscoderStatusRequest(req *web.UpdateTranscoderStatusRequest) error {
	if req.TranscoderId == "" {
		return custerror.FormatInvalidArgument("missing transcoder id")
	}
	if req.CameraName == nil {
		if req.CameraId == nil {
			return custerror.FormatInvalidArgument("missing camera name or camera id")
		} else {
			if *req.CameraId == "" {
				return custerror.FormatInvalidArgument("missing camera id")
			}
		}
	} else {
		if *req.CameraName == "" {
			return custerror.FormatInvalidArgument("missing camera name")
		}
	}
	return nil
}

func (s *WebService) addTranscoderStatus(ctx context.Context, req *web.UpdateTranscoderStatusRequest) error {
	status := &db.TranscoderStatus{}
	if err := copier.Copy(status, req); err != nil {
		return err
	}
	status.TranscoderId = req.TranscoderId
	status.StatusId = uuid.NewString()

	q := s.builder.Insert("transcoder_statuses").
		Columns(status.Fields()...).
		Values(status.Values()...)
	if err := s.db.Insert(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) updateTranscoderStatus(ctx context.Context, status *db.TranscoderStatus, req *web.UpdateTranscoderStatusRequest) error {
	newStatus := s.patchTranscoderStatus(status, req)
	valueMap := map[string]interface{}{}
	fields := newStatus.Fields()
	values := newStatus.Values()
	for i := 0; i < len(fields); i += 1 {
		valueMap[fields[i]] = values[i]
	}

	q := s.builder.Update("transcoder_statuses").
		Where("transcoder_id = ?", newStatus.TranscoderId).
		Where("camera_id = ?", newStatus.CameraId).
		SetMap(valueMap)
	sql, args, _ := q.ToSql()
	logger.SDebug("updateTranscoderStatus: SQL query",
		zap.String("query", sql),
		zap.Reflect("args", args))
	if err := s.db.Update(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *WebService) patchTranscoderStatus(old *db.TranscoderStatus, req *web.UpdateTranscoderStatusRequest) *db.TranscoderStatus {
	new := old
	if req.TranscoderId != "" {
		new.TranscoderId = req.TranscoderId
	}
	if req.AudioDetection != nil {
		new.AudioDetection = *req.AudioDetection
	}
	if req.CameraId != nil {
		new.CameraId = *req.CameraId
	}
	if req.Autotracker != nil {
		new.Autotracker = *req.Autotracker
	}
	if req.ObjectDetection != nil {
		new.ObjectDetection = *req.ObjectDetection
	}
	if req.OpenGateRecordings != nil {
		new.OpenGateRecordings = *req.OpenGateRecordings
	}
	if req.Snapshots != nil {
		new.Snapshots = *req.Snapshots
	}
	if req.MotionDetection != nil {
		new.MotionDetection = *req.MotionDetection
	}
	if req.ImproveContrast != nil {
		new.ImproveContrast = *req.ImproveContrast
	}
	if req.BirdseyeView != nil {
		new.BirdseyeView = *req.BirdseyeView
	}
	if req.OpenGateStatus != nil {
		new.OpenGateStatus = *req.OpenGateStatus
	}
	if req.TranscoderStatus != nil {
		new.TranscoderStatus = *req.TranscoderStatus
	}
	return new
}

func (s *WebService) getTranscoderStatus(ctx context.Context, transcoderId []string, cameraId []string) ([]db.TranscoderStatus, error) {
	q := s.builder.Select("*").
		From("transcoder_statuses")

	if transcoderId != nil {
		or := squirrel.Or{}
		for _, i := range transcoderId {
			or = append(or, squirrel.Eq{"transcoder_id": i})
		}
		q = q.Where(or)
	}
	if cameraId != nil {
		or := squirrel.Or{}
		for _, i := range cameraId {
			or = append(or, squirrel.Eq{"camera_id": i})
		}
		q = q.Where(or)
	}

	var transcoderStatus []db.TranscoderStatus
	if err := s.db.Select(ctx, q, &transcoderStatus); err != nil {
		return nil, err
	}
	return transcoderStatus, nil
}

func (s *WebService) validateGetTranscoderStatusRequest(req *web.GetTranscoderStatusRequest) error {
	if len(req.TranscoderId) == 0 {
		return custerror.FormatInvalidArgument("missing transcoder id")
	}
	if len(req.CameraId) == 0 {
		return custerror.FormatInvalidArgument("missing camera id")
	}
	return nil
}
