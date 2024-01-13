package service

import (
	"context"
	"fmt"
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
	"net/url"

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

func (s *WebService) deleteCameraById(ctx context.Context, id string) error {
	return s.db.Delete(ctx,
		squirrel.Delete("cameras").
			Where("camera_id = ?", id))
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
	if err := s.db.Select(ctx, q, &cameras); err != nil {
		return nil, err
	}

	return cameras, nil
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

	transcoders := []db.Transcoder{}
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
	q := squirrel.Update("transcoders").Where("device_id = ?", d.DeviceId).SetMap(map[string]interface{}{
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
	url.Scheme = "ws"
	url.Host = fmt.Sprintf("%s:%d", configs.Host, configs.Ports.WebRTC)
	url = url.JoinPath(configs.ApplicationName, camera.CameraId)
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

	q := squirrel.Update("cameras").Where("camera_id = ?", camera.CameraId).SetMap(valueMap)
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
	logger.SDebug("requestLtdStreamControl: message published succesfully")
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
	q := squirrel.Select("*").From("cameras").Where("transcoder_id = ?", transcoderId)

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
