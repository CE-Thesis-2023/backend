package opengate

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/CE-Thesis-2023/backend/src/helper/media"
	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/db"
)

func prepareTestBiz() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	os.Setenv("CONFIG_FILE_PATH", "../../../configs.json")
	configs.Init(ctx)
	logger.Init(ctx,
		logger.WithGlobalConfigs(&configs.Get().Logger))
}

func TestOpenGateConfiguration(t *testing.T) {
	prepareTestBiz()
	t.Run("Test NewConfiguration", func(t *testing.T) {
		integration := &db.OpenGateIntegration{
			OpenGateId:               "1",
			LogLevel:                 "info",
			SnapshotRetentionDays:    30,
			MqttId:                   "1",
			HardwareAccelerationType: "quicksync",
			WithEdgeTpu:              true,
			TranscoderId:             "test-device-01",
		}
		mqtt := &db.OpenGateMqttConfiguration{
			ConfigurationId: "1",
			Enabled:         true,
			Host:            "mosquitto.mqtt.ntranlab.com",
			Port:            1883,
			Username:        "admin",
			Password:        "admin",
			OpenGateId:      "1",
		}
		openGateCameras := []db.OpenGateCameraSettings{
			{
				OpenGateId:  "1",
				CameraId:    "1",
				Height:      480,
				Width:       640,
				Fps:         10,
				MqttEnabled: true,
				Timestamp:   true,
				BoundingBox: true,
				Crop:        true,
			},
		}
		cameras := []db.Camera{
			{
				CameraId:           "1",
				Name:               "camera_name",
				Ip:                 "103.165.142.85",
				Port:               80,
				Username:           "admin",
				Password:           "admin",
				Enabled:            false,
				OpenGateCameraName: "camera_name",
				GroupId:            "1",
				TranscoderId:       "test-device-01",
				SettingsId:         "1",
			},
		}
		mediaHelper := media.NewMediaHelper(
			&configs.Get().MediaEngine,
			&configs.Get().S3)

		config := NewConfiguration(integration, mqtt, openGateCameras, cameras, mediaHelper)
		bytesContent, err := config.YAML()
		if err != nil {
			t.Errorf("Expected NewConfiguration to return nil, but got %v", err)
		}
		if config == nil {
			t.Errorf("Expected NewConfiguration to return a Configuration object, but got nil")
		}

		fmt.Println(string(bytesContent))
	})
}
