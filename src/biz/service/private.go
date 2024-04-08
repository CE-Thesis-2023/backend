package service

import (
	"context"
	"encoding/base64"
	"errors"

	"github.com/CE-Thesis-2023/backend/src/helper/media"
	"github.com/CE-Thesis-2023/backend/src/helper/opengate"
	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	custdb "github.com/CE-Thesis-2023/backend/src/internal/db"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/db"
	"github.com/CE-Thesis-2023/backend/src/models/events"
	"github.com/CE-Thesis-2023/backend/src/models/web"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PrivateService struct {
	db          *custdb.LayeredDb
	webService  *WebService
	mediaHelper *media.MediaHelper
}

func NewPrivateService() *PrivateService {
	return &PrivateService{
		db:          custdb.Layered(),
		webService:  GetWebService(),
		mediaHelper: media.NewMediaHelper(&configs.Get().MediaEngine),
	}
}

func (s *PrivateService) validateRegisterDeviceRequest(req *events.DeviceRegistrationRequest) error {
	if req.DeviceId == "" {
		return custerror.FormatInvalidArgument("missing device_id")
	}
	return nil
}

func (s *PrivateService) RegisterDevice(ctx context.Context, req *events.DeviceRegistrationRequest) error {
	logger.SDebug("RegisterDevice: request",
		zap.Any("request", req))

	if err := s.validateRegisterDeviceRequest(req); err != nil {
		logger.SDebug("RegisterDevice: validateRegisterDeviceRequest",
			zap.Error(err))
		return err
	}

	device, err := s.webService.getDeviceById(ctx, []string{
		req.DeviceId,
	})
	if err != nil {
		if !errors.Is(err, custerror.ErrorNotFound) {
			logger.SDebug("RegisterDevice: getDeviceById",
				zap.Error(err))
			return err
		}
	}

	if len(device) > 0 {
		logger.SDebug("RegisterDevice: device already registered")
		return custerror.ErrorAlreadyExists
	}

	transcoder := &db.Transcoder{
		DeviceId: req.DeviceId,
		Name:     "Unknown Transcoder",
	}

	if err := s.initializeOpenGateDefaultConfigurations(ctx, transcoder); err != nil {
		logger.SDebug("RegisterDevice: initializeOpenGateDefaultConfigurations",
			zap.Error(err))
		return err
	}

	logger.SInfo("RegisterDevice: device not found",
		zap.String("id", req.DeviceId))
	if err := s.webService.addDevice(ctx, transcoder); err != nil {
		logger.SDebug("RegisterDevice: addDevice",
			zap.Error(err))
		return err
	}

	logger.SInfo("RegisterDevice: device",
		zap.Any("device", device))

	return nil
}

func (s *PrivateService) initializeOpenGateDefaultConfigurations(ctx context.Context, device *db.Transcoder) error {
	logger.SDebug("initializeOpenGateDefaultConfigurations: request",
		zap.Any("device", device))

	openGateIntegration := &db.OpenGateIntegration{
		OpenGateId:            uuid.NewString(),
		TranscoderId:          device.DeviceId,
		Available:             false,
		IsRestarting:          false,
		LogLevel:              "info",
		SnapshotRetentionDays: 7,
	}

	mqttConfigs := db.OpenGateMqttConfiguration{
		ConfigurationId: uuid.NewString(),
		Enabled:         true,
		Host:            "mosquitto.mqtt.ntranlab.com",
		Username:        "admin",
		Password:        "ctportal2024",
		Port:            8883,
		OpenGateId:      openGateIntegration.OpenGateId,
	}
	openGateIntegration.MqttId = mqttConfigs.ConfigurationId
	device.OpenGateIntegrationId = openGateIntegration.OpenGateId

	if err := s.webService.addOpenGateIntegration(ctx, openGateIntegration); err != nil {
		logger.SDebug("initializeOpenGateDefaultConfigurations: addOpenGateIntegration",
			zap.Error(err))
		return err
	}

	if err := s.webService.addOpenGateMqttConfigurations(ctx, &mqttConfigs); err != nil {
		logger.SDebug("initializeOpenGateDefaultConfigurations: addOpenGateMqttConfiguration",
			zap.Error(err))
		return err
	}

	device.OpenGateIntegrationId = openGateIntegration.OpenGateId

	logger.SInfo("initializeOpenGateDefaultConfigurations: success")
	return nil

}

func (s *PrivateService) validateUpdateCameraListRequest(req *events.UpdateCameraListRequest) error {
	if req.DeviceId == "" {
		return custerror.FormatInvalidArgument("missing device_id")
	}
	return nil
}

func (s *PrivateService) UpdateCameraList(ctx context.Context, req *events.UpdateCameraListRequest) (*events.UpdateCameraListResponse, error) {
	logger.SInfo("commandService.UpdateCameraList: request", zap.Any("request", req))

	if err := s.validateUpdateCameraListRequest(req); err != nil {
		logger.SDebug("UpdateCameraList: validateUpdateCameraListRequest", zap.Error(err))
		return nil, err
	}

	transcoders, err := s.webService.getDeviceById(ctx, []string{req.DeviceId})
	if err != nil {
		logger.SDebug("UpdateCameraList: request", zap.Error(err))
		return nil, err
	}

	if len(transcoders) == 0 {
		logger.SError("UpdateCameraList: transcoder not found")
		return nil, custerror.ErrorNotFound
	}

	cameras, err := s.webService.getCamerasByTranscoderId(
		ctx,
		req.DeviceId,
		nil)
	if err != nil {
		logger.SError("UpdateCameraList: getCamerasByTranscoderId", zap.Error(err))
		return nil, err
	}

	logger.SInfo("UpdateCameraList: cameras", zap.Any("cameras", cameras))

	return &events.UpdateCameraListResponse{
		Cameras: cameras,
	}, nil
}

func (c *PrivateService) DeleteTranscoder(ctx context.Context, req *web.DeleteTranscoderRequest) error {
	logger.SInfo("DeleteTranscoder: request", zap.Any("request", req))

	transcoders, err := c.webService.getDeviceById(ctx, []string{req.DeviceId})
	if err != nil {
		logger.SDebug("DeleteTranscoder: getDeviceById", zap.Error(err))
		return err
	}

	if len(transcoders) == 0 {
		logger.SError("DeleteTranscoder: transcoder not found", zap.Reflect("request", req))
		return custerror.ErrorNotFound
	}
	transcoder := transcoders[0]

	openGateIntegration, err := c.webService.getOpenGateIntegrationById(ctx, transcoder.OpenGateIntegrationId)
	if err != nil {
		logger.SDebug("DeleteTranscoder: getOpenGateIntegrationById", zap.Error(err))
		return err
	}

	mqtt := openGateIntegration.MqttId
	if mqtt != "" {
		if err := c.webService.deleteOpenGateMqttConfiguration(ctx, mqtt); err != nil {
			logger.SDebug("DeleteTranscoder: deleteOpenGateMqttConfigurations", zap.Error(err))
			return err
		}
	}

	if openGateIntegration != nil {
		if err := c.webService.deleteOpenGateIntegration(ctx, transcoder.OpenGateIntegrationId); err != nil {
			logger.SDebug("DeleteTranscoder: deleteOpenGateIntegration", zap.Error(err))
			return err
		}
	}

	if err := c.webService.deleteDeviceById(ctx, req.DeviceId); err != nil {
		logger.SDebug("DeleteTranscoder: deleteDevice", zap.Error(err))
		return err
	}

	return nil
}

func (s *PrivateService) validateAddObjectTrackingEventRequest(req *web.AddObjectTrackingEventRequest) error {
	if req.Event == nil {
		return custerror.FormatInvalidArgument("missing event")
	}
	return nil
}

func (s *PrivateService) AddObjectTrackingEvent(ctx context.Context, req *web.AddObjectTrackingEventRequest) (*web.AddObjectTrackingEventResponse, error) {
	logger.SInfo("commandService.AddEvent: request", zap.Any("request", req))

	if err := s.validateAddObjectTrackingEventRequest(req); err != nil {
		logger.SDebug("AddEvent: validateAddObjectTrackingEventRequest", zap.Error(err))
		return nil, err
	}

	before := req.Event.Before
	cameras, err := s.webService.getCamerasByTranscoderId(ctx, "", []string{before.Camera})
	if err != nil {
		logger.SDebug("AddEvent: getCamerasByTranscoderId", zap.Error(err))
		return nil, err
	}
	if len(cameras) == 0 {
		logger.SError("AddEvent: camera not found")
		return nil, custerror.ErrorNotFound
	}

	dbEvent := s.webService.fromObjectTrackingEventToDto(&before)
	dbEvent.EventId = uuid.NewString()
	dbEvent.EventType = req.Event.Type
	dbEvent.CameraId = cameras[0].CameraId

	if err := s.webService.addObjectTrackingEvent(ctx, dbEvent); err != nil {
		logger.SDebug("AddEvent: addEventToDatabase", zap.Error(err))
		return nil, err
	}

	return &web.AddObjectTrackingEventResponse{
		EventId: dbEvent.EventId,
	}, nil
}

func (s *PrivateService) UpdateObjectTrackingEvent(ctx context.Context, req *web.UpdateObjectTrackingEventRequest) error {
	logger.SInfo("commandService.UpdateEvent: request", zap.Any("request", req))

	objectTrackingEvent, err := s.webService.getObjectTrackingEventById(
		ctx, []string{req.EventId}, nil)
	if err != nil {
		return err
	}

	if len(objectTrackingEvent) == 0 {
		logger.SError("UpdateEvent: event not found")
		return custerror.ErrorNotFound
	}

	if req.Event == nil {
		logger.SError("UpdateEvent: missing event")
		return custerror.FormatInvalidArgument("missing event")
	}
	after := req.Event.After

	event := s.webService.fromObjectTrackingEventToDto(&after)
	event.CameraId = objectTrackingEvent[0].CameraId
	event.EventId = objectTrackingEvent[0].EventId
	event.EventType = req.Event.Type

	if err := s.webService.updateObjectTrackingEvent(ctx, event); err != nil {
		logger.SDebug("UpdateEvent: updateEventInDatabase", zap.Error(err))
		return err
	}

	return nil
}

func (s *PrivateService) GetTranscoderOpenGateConfiguration(ctx context.Context, req *web.GetTranscoderOpenGateConfigurationRequest) (*web.GetTranscoderOpenGateConfigurationResponse, error) {
	logger.SInfo("GetTranscoderOpenGateConfiguration: request",
		zap.Any("request", req))

	transcoder, err := s.webService.getDeviceById(ctx, []string{req.TranscoderId})
	if err != nil {
		logger.SDebug("GetTranscoderOpenGateConfiguration: getDeviceById", zap.Error(err))
		return nil, err
	}
	if len(transcoder) == 0 {
		logger.SError("GetTranscoderOpenGateConfiguration: transcoder not found")
		return nil, custerror.ErrorNotFound
	}

	cameras, err := s.webService.getCamerasByTranscoderId(ctx, transcoder[0].DeviceId, nil)
	if err != nil {
		logger.SDebug("GetTranscoderOpenGateConfiguration: getCamerasByTranscoderId", zap.Error(err))
		return nil, err
	}

	integration, err := s.webService.getOpenGateIntegrationById(ctx,
		transcoder[0].
			OpenGateIntegrationId)
	if err != nil {
		logger.SDebug("GetTranscoderOpenGateConfiguration: getOpenGateIntegrationById", zap.Error(err))
		return nil, err
	}

	mqtt, err := s.webService.getOpenGateMqttConfigurationById(ctx, integration.MqttId)
	if err != nil {
		logger.SDebug("GetTranscoderOpenGateConfiguration: getOpenGateMqttConfigurationById", zap.Error(err))
		return nil, err
	}

	cameraIds := make([]string, 0, len(cameras))
	for _, camera := range cameras {
		cameraIds = append(cameraIds, camera.CameraId)
	}
	logger.SDebug("GetTranscoderOpenGateConfiguration: cameraIds",
		zap.Reflect("cameraIds", cameraIds))

	openGateCameras, err := s.webService.getOpenGateCameraSettings(ctx, cameraIds)
	if err != nil {
		logger.SDebug("GetTranscoderOpenGateConfiguration: getOpenGateCameraSettings", zap.Error(err))
		return nil, err
	}

	configs := opengate.NewConfiguration(
		integration,
		mqtt,
		openGateCameras,
		cameras,
	)

	yamlConfigs, err := configs.YAML()
	if err != nil {
		logger.SDebug("GetTranscoderOpenGateConfiguration: YAML", zap.Error(err))
		return nil, err
	}

	encoded := base64.StdEncoding.EncodeToString(yamlConfigs)

	return &web.GetTranscoderOpenGateConfigurationResponse{
		Base64: encoded,
	}, nil
}

func (s *PrivateService) GetStreamConfigurations(ctx context.Context, req *web.GetStreamConfigurationsRequest) (*web.GetStreamConfigurationsResponse, error) {
	logger.SInfo("GetStreamConfigurations: request",
		zap.Any("request", req))

	cameras, err := s.webService.getCameraById(ctx, req.CameraId)
	if err != nil {
		logger.SDebug("GetStreamConfigurations: getCameraGroupById", zap.Error(err))
		return nil, err
	}

	configurations := make([]web.TranscoderStreamConfiguration, 0, len(cameras))
	for _, camera := range cameras {
		rtspSourceUrl := s.mediaHelper.BuildRTSPSourceUrl(camera)
		srtPublishUrl, err := s.mediaHelper.BuildSRTPublishUrl(camera.CameraId)
		if err != nil {
			logger.SDebug("GetStreamConfigurations: BuildSRTPublishUrl", zap.Error(err))
			return nil, err
		}
		height, width, fps := 720, 1280, 30
		c := web.TranscoderStreamConfiguration{
			CameraId:   camera.CameraId,
			SourceUrl:  rtspSourceUrl,
			PublishUrl: srtPublishUrl,
			Height:     height,
			Width:      width,
			Fps:        fps,
		}
		configurations = append(configurations, c)
	}

	return &web.GetStreamConfigurationsResponse{
		StreamConfigurations: configurations,
	}, nil
}