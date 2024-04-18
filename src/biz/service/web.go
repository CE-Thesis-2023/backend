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
	mediaHelper *media.MediaHelper
	reqreply    *custmqtt.MQTTSession
	cvs         *ComputerVisionService
}

func NewWebService(reqreply *custmqtt.MQTTSession, mediaHelper *media.MediaHelper, cvs *ComputerVisionService) *WebService {
	return &WebService{
		db: custdb.Layered(),
		builder: squirrel.
			StatementBuilder.
			PlaceholderFormat(squirrel.Dollar),
		mediaHelper: mediaHelper,
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

func (s *WebService) validateGetCamerasByClientIdRequest(req *web.GetCameraByClientIdRequest) error {
	if len(req.ClientId) == 0 {
		return custerror.FormatInvalidArgument("missing client")
	}
	return nil
}

func (s *WebService) GetCamerasByClientId(ctx context.Context, req *web.GetCameraByClientIdRequest) (*web.GetCameraByOpenGateIdResponse, error) {
	logger.SDebug("GetCamerasByClientId: request",
		zap.Reflect("request", req))

	if err := s.validateGetCamerasByClientIdRequest(req); err != nil {
		logger.SDebug("GetCamerasByClientId: validateGetCamerasByClientIdRequest", zap.Error(err))
		return nil, err
	}

	OpenGateIntegration, err := s.getOpenGateIdbyTranscoderId(ctx, req.ClientId)
	if err != nil {
		logger.SError("GetCamerasByClientId: getOpenGateIntegrationById", zap.Error(err))
		return nil, err
	}

	integration, err := s.getOpenGateIntegrationById(ctx, OpenGateIntegration.OpenGateId)
	if err != nil {
		logger.SError("GetCamerasByClientId: getOpenGateIntegrationById", zap.Error(err))
		return nil, err
	}

	transcoder, err := s.getTranscoderById(ctx, integration.TranscoderId)
	if err != nil {
		logger.SError("GetCamerasByClientId: getTranscoderById", zap.Error(err))
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

func (s *WebService) getOpenGateIdbyTranscoderId(ctx context.Context, id string) (*db.OpenGateIntegration, error) {
	q := s.builder.Select("*").
		From("open_gate_integrations").
		Where("transcoder_id = ?", id)
	sql, args, _ := q.ToSql()
	logger.SDebug("getTranscoderById: SQL",
		zap.Reflect("q", sql),
		zap.Reflect("args", args))

	var openGateIntegration db.OpenGateIntegration
	if err := s.db.Get(ctx, q, &openGateIntegration); err != nil {
		return nil, err
	}
	return &openGateIntegration, nil
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

	streamUrl := s.mediaHelper.BuildWebRTCViewStream(camera[0].CameraId)
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

	results := []db.Camera{}
	if err := s.db.Select(ctx, q, &results); err != nil {
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
	// TODO: Request reply to the device
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

	// TODO: Request-reply to the device
	return &web.GetCameraDeviceInfo{}, nil
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

	events, err := s.getObjectTrackingEventById(ctx, req.EventId, req.OpenGateEventId)
	if err != nil {
		logger.SError("GetObjectTrackingEventById: getObjectTrackingEventById error", zap.Error(err))
		return nil, err
	}

	return &web.GetObjectTrackingEventByIdResponse{
		ObjectTrackingEvents: events,
	}, nil
}

func (s *WebService) getObjectTrackingEventById(ctx context.Context, ids []string, openGateIds []string) ([]db.ObjectTrackingEvent, error) {
	q := s.builder.Select("*").
		From("object_tracking_events")

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

	var events []db.ObjectTrackingEvent
	if err := s.db.Select(ctx, q, &events); err != nil {
		return nil, err
	}

	return events, nil
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

	url, err := s.mediaHelper.GetPresignedUrl(
		ctx,
		person[0].
			ImagePath)
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
		}
		s3Error = s.mediaHelper.UploadImage(ctx, &fileDesc)
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
			_ = s.mediaHelper.DeleteImage(context.Background(), id)
		}()
		return "", postgresError
	}
	wg.Wait()
	return id, nil
}
