package opengate

import (
	"testing"

	"github.com/CE-Thesis-2023/backend/src/helper/media"
	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	"github.com/CE-Thesis-2023/backend/src/models/db"
)

func TestOpenGate_Configuration(t *testing.T) {
	integration := &db.OpenGateIntegration{
		OpenGateId:            "d97f08a0-4a97-4bde-96ea-b10cf4760ac2",
		Available:             false,
		IsRestarting:          false,
		LogLevel:              "info",
		SnapshotRetentionDays: 7,
		MqttId:                "ef57dc1e-8cba-4330-8f06-814f644e53e5",
		TranscoderId:          "test-device-01",
	}
	cameras := []db.Camera{
		{
			CameraId:           "f9a6e7d2-5133-4512-bb47-8ab56c01483e",
			Name:               "Front Door",
			Ip:                 "10.40.8.100",
			Port:               80,
			Username:           "admin",
			Password:           "bkcamera2023",
			Enabled:            false,
			OpenGateCameraName: "front_door",
			GroupId:            "",
			TranscoderId:       "test-device-01",
			SettingsId:         "5434f7cb-1e8c-47a6-a3d7-322303f8202e",
		},
	}
	openGateCameraSettings := []db.OpenGateCameraSettings{
		{
			SettingsId:  "5434f7cb-1e8c-47a6-a3d7-322303f8202e",
			Height:      480,
			Width:       640,
			Fps:         5,
			MqttEnabled: true,
			Timestamp:   true,
			BoundingBox: true,
			Crop:        true,
			OpenGateId:  "d97f08a0-4a97-4bde-96ea-b10cf4760ac2",
			CameraId:    "f9a6e7d2-5133-4512-bb47-8ab56c01483e",
		},
	}
	mqtt := &db.OpenGateMqttConfiguration{
		ConfigurationId: "ef57dc1e-8cba-4330-8f06-814f644e53e5",
		Enabled:         true,
		Host:            "mosquitto.mqtt.ntranlab.com",
		Username:        "admin",
		Password:        "ctportal2024",
		Port:            8883,
		OpenGateId:      "d97f08a0-4a97-4bde-96ea-b10cf4760ac2",
	}

	configuration := NewConfiguration(
		integration,
		mqtt,
		openGateCameraSettings,
		cameras,
		media.NewMediaHelper(&configs.MediaMtxConfigs{
			Host:     "api.mediamtx.ntranlab.com",
			MediaUrl: "103.165.142.15",
			PublishPorts: configs.MtxPorts{
				WebRTC: 8889,
			},
			ProviderPorts: configs.MtxPorts{
				Srt: 8890,
			},
			Api: 9997,
		}),
	)

	if err := configuration.build(); err != nil {
		t.Errorf("Error while building configuration: %v", err)
	}

	json, err := configuration.JSON()
	if err != nil {
		t.Errorf("Error while getting JSON configuration: %v", err)
	}

	yaml, err := configuration.YAML()
	if err != nil {
		t.Errorf("Error while getting YAML configuration: %v", err)
	}

	t.Logf("JSON configuration: %v", string(json))
	t.Logf("YAML configuration: %v", string(yaml))
}
