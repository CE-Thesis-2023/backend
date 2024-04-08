package opengate

import (
	"encoding/json"
	"fmt"

	"github.com/CE-Thesis-2023/backend/src/helper/media"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/db"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type Configuration struct {
	c               map[string]interface{}
	integration     *db.OpenGateIntegration
	mqtt            *db.OpenGateMqttConfiguration
	openGateCameras []db.OpenGateCameraSettings
	cameras         []db.Camera
	mediaHelper     *media.MediaHelper
}

func NewConfiguration(
	integration *db.OpenGateIntegration,
	mqtt *db.OpenGateMqttConfiguration,
	openGateCameras []db.OpenGateCameraSettings,
	cameras []db.Camera,
	mediaHelper *media.MediaHelper,
) *Configuration {
	return &Configuration{
		c:               make(map[string]interface{}),
		integration:     integration,
		mqtt:            mqtt,
		openGateCameras: openGateCameras,
		cameras:         cameras,
		mediaHelper:     mediaHelper,
	}
}

func (c *Configuration) build() error {
	if err := c.buildMQTTConfiguration(); err != nil {
		return err
	}
	if err := c.buildAudio(); err != nil {
		return err
	}
	if err := c.buildLogger(); err != nil {
		return err
	}
	if err := c.buildSnapshots(); err != nil {
		return err
	}
	if err := c.buildCameras(); err != nil {
		return err
	}
	return nil
}

func (c *Configuration) JSON() ([]byte, error) {
	if err := c.build(); err != nil {
		return nil, err
	}
	return json.Marshal(c.c)
}

func (c *Configuration) YAML() ([]byte, error) {
	if err := c.build(); err != nil {
		return nil, err
	}
	return yaml.Marshal(c.c)
}

func (c *Configuration) buildMQTTConfiguration() error {
	mqtt := make(map[string]interface{})
	mqtt["enabled"] = true
	mqtt["host"] = c.mqtt.Host
	mqtt["port"] = c.mqtt.Port
	mqtt["user"] = c.mqtt.Username
	mqtt["password"] = c.mqtt.Password
	mqtt["topic_prefix"] = fmt.Sprintf("opengate/%s", c.integration.OpenGateId)

	c.c["mqtt"] = mqtt
	return nil
}

func (c *Configuration) buildSnapshots() error {
	snapshots := make(map[string]interface{})
	snapshots["enabled"] = true

	retain := make(map[string]interface{})
	retain["default"] = c.integration.SnapshotRetentionDays
	snapshots["retain"] = retain

	c.c["snapshots"] = snapshots
	return nil
}

func (c *Configuration) buildAudio() error {
	audio := make(map[string]interface{})
	audio["enabled"] = false
	c.c["audio"] = audio
	return nil
}

func (c *Configuration) buildLogger() error {
	logger := make(map[string]interface{})
	logger["default"] = c.integration.LogLevel
	c.c["logger"] = logger
	return nil
}

func (c *Configuration) buildCameras() error {
	cameras := make(map[string]interface{})

	cameraConfigsToMap := make(map[string]db.OpenGateCameraSettings)
	for _, camera := range c.openGateCameras {
		cameraConfigsToMap[camera.CameraId] = camera
	}

	for _, camera := range c.cameras {
		configs, found := cameraConfigsToMap[camera.CameraId]
		if !found {
			logger.SError("OpenGate camera configuration not found",
				zap.String("camera", camera.Name))
			continue
		}

		m := make(map[string]interface{})

		ffmpeg := make(map[string]interface{})
		inputs := make([]map[string]interface{}, 0, 1)
		input := make(map[string]interface{})
		input["path"] = c.mediaHelper.BuildRTSPSourceUrl(camera)
		input["input_args"] = "preset-rtsp-generic"
		input["output_args"] = "preset-rtsp-generic"
		input["hwaccel_args"] = []string{"preset-vaapi"}
		input["retry_interval"] = 10
		input["roles"] = []string{"detect"}
		inputs = append(inputs, input)
		ffmpeg["inputs"] = inputs
		m["ffmpeg"] = ffmpeg

		onvif := make(map[string]interface{})
		onvif["host"] = camera.Ip
		onvif["port"] = camera.Port
		onvif["username"] = camera.Username
		onvif["password"] = camera.Password
		autotracking := make(map[string]interface{})
		autotracking["enabled"] = true
		autotracking["zooming"] = "disabled"
		onvif["autotracking"] = autotracking
		m["onvif"] = onvif

		detect := make(map[string]interface{})
		detect["height"] = configs.Height
		detect["width"] = configs.Width
		detect["fps"] = configs.Fps
		m["detect"] = detect

		mqtt := make(map[string]interface{})
		mqtt["enabled"] = configs.MqttEnabled
		mqtt["timestamp"] = configs.Timestamp
		mqtt["bounding_box"] = configs.BoundingBox
		mqtt["crop"] = configs.Crop
		m["mqtt"] = mqtt

		cameras[camera.OpenGateCameraName] = m
	}

	c.c["cameras"] = cameras
	return nil
}
