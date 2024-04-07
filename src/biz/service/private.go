package service

import (
	"context"
	"errors"

	custdb "github.com/CE-Thesis-2023/backend/src/internal/db"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/db"
	"github.com/CE-Thesis-2023/backend/src/models/events"
	"github.com/CE-Thesis-2023/backend/src/models/web"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
)

type PrivateService struct {
	db         *custdb.LayeredDb
	webService *WebService
}

func NewPrivateService() *PrivateService {
	return &PrivateService{
		db:         custdb.Layered(),
		webService: GetWebService(),
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

func (s *PrivateService) AddEvent(ctx context.Context, req *web.AddObjectTrackingEventRequest) (*web.AddObjectTrackingEventResponse, error) {
	logger.SInfo("commandService.AddEvent: request", zap.Any("request", req))

	if req.Event == nil {
		logger.SError("AddEvent: missing event")
		return nil, custerror.FormatInvalidArgument("missing event")
	}
	before := req.Event.Before

	var event db.ObjectTrackingEvent
	if err := copier.Copy(&event, before); err != nil {
		logger.SError("AddEvent: unable to copy event", zap.Error(err))
		return nil, err
	}
	event.EventId = uuid.NewString()

	if err := s.webService.addEvent(ctx, &event); err != nil {
		logger.SDebug("AddEvent: addEventToDatabase", zap.Error(err))
		return nil, err
	}

	return &web.AddObjectTrackingEventResponse{
		EventId: event.EventId,
	}, nil
}

func (s *PrivateService) UpdateEvent(ctx context.Context, req *web.UpdateObjectTrackingEventRequest) error {
	logger.SInfo("commandService.UpdateEvent: request", zap.Any("request", req))

	if req.Event == nil {
		logger.SError("UpdateEvent: missing event")
		return custerror.FormatInvalidArgument("missing event")
	}
	after := req.Event.After

	var event db.ObjectTrackingEvent
	if err := copier.Copy(&event, after); err != nil {
		logger.SError("UpdateEvent: unable to copy event", zap.Error(err))
		return err
	}
	event.EventId = req.EventId

	if err := s.webService.updateEvent(ctx, &event); err != nil {
		logger.SDebug("UpdateEvent: updateEventInDatabase", zap.Error(err))
		return err
	}

	return nil
}

func (s *PrivateService) DeleteEvent(ctx context.Context, req *web.DeleteObjectTrackingEventRequest) error {
	logger.SInfo("commandService.DeleteEvent: request", zap.Any("request", req))

	if req.EventId == "" {
		logger.SError("DeleteEvent: missing event_id")
		return custerror.FormatInvalidArgument("missing event_id")
	}

	if err := s.webService.deleteEvent(ctx, req.EventId); err != nil {
		logger.SDebug("DeleteEvent: deleteEvent", zap.Error(err))
		return err
	}

	return nil
}
