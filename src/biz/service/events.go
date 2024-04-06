package service

import (
	"context"
	"errors"

	custdb "github.com/CE-Thesis-2023/backend/src/internal/db"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/db"
	"github.com/CE-Thesis-2023/backend/src/models/events"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CommandService struct {
	db         *custdb.LayeredDb
	webService *WebService
}

func NewCommandService() *CommandService {
	return &CommandService{
		db:         custdb.Layered(),
		webService: GetWebService(),
	}
}

func (s *CommandService) RegisterDevice(ctx context.Context, req *events.DeviceRegistrationRequest) error {
	logger.SDebug("RegisterDevice: request",
		zap.Any("request", req))

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

func (s *CommandService) initializeOpenGateDefaultConfigurations(ctx context.Context, device *db.Transcoder) error {
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

func (s *CommandService) UpdateCameraList(ctx context.Context, req *events.UpdateCameraListRequest) (*events.UpdateCameraListResponse, error) {
	logger.SInfo("commandService.UpdateCameraList: request", zap.Any("request", req))

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
