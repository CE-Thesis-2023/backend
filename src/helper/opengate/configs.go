package opengate

import (
	"bytes"
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
	if err := c.buildBirdseye(); err != nil {
		return err
	}
	if err := c.buildLogger(); err != nil {
		return err
	}
	if err := c.buildSnapshots(); err != nil {
		return err
	}
	if err := c.buildFFmpeg(); err != nil {
		return err
	}
	if err := c.buildCameras(); err != nil {
		return err
	}
	if err := c.buildDetectors(); err != nil {
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
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	defer encoder.Close()
	if err := encoder.Encode(c.c); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *Configuration) buildMQTTConfiguration() error {
	mqtt := make(map[string]interface{})
	mqtt["enabled"] = true
	mqtt["host"] = c.mqtt.Host
	mqtt["port"] = 1883
	mqtt["user"] = c.mqtt.Username
	mqtt["password"] = c.mqtt.Password
	mqtt["topic_prefix"] = fmt.Sprintf("opengate/%s", c.
		integration.
		TranscoderId)

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

func (c *Configuration) buildBirdseye() error {
	birdseye := make(map[string]interface{})
	birdseye["enabled"] = false
	c.c["birdseye"] = birdseye
	return nil
}

func (c *Configuration) buildLogger() error {
	logger := make(map[string]interface{})
	logger["default"] = c.integration.LogLevel
	logs := make(map[string]interface{})
	logs["opengate.ptz.autotrack"] = "debug"
	logs["opengate.ptz.onvif"] = "debug"
	logs["opengate.ptz.sidecar"] = "debug"
	logger["logs"] = logs
	c.c["logger"] = logger

	return nil
}

func (c *Configuration) buildFFmpeg() error {
	ffmpeg := make(map[string]interface{})
	t := ""
	switch c.integration.HardwareAccelerationType {
	case "vaapi":
		t = "preset-vaapi"
	case "quicksync":
		t = "preset-intel-qsv-h264"
	}
	if len(t) > 0 {
		ffmpeg["hwaccel_args"] = t
	}
	ffmpeg["retry_interval"] = 10
	c.c["ffmpeg"] = ffmpeg
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
		m["enabled"] = camera.Enabled

		ffmpeg := make(map[string]interface{})

		inputs := make([]map[string]interface{}, 0, 1)
		input := make(map[string]interface{})
		input["path"] = c.mediaHelper.BuildRTSPSourceUrl(camera)

		input["roles"] = []string{"detect"}
		inputs = append(inputs, input)
		ffmpeg["inputs"] = inputs

		ffmpeg["input_args"] = "preset-rtsp-generic"
		m["ffmpeg"] = ffmpeg

		zones := make(map[string]interface{})
		defaultZone := make(map[string]interface{})
		defaultZone["coordinates"] = fmt.Sprintf("%d,%d,%d,0,0,0,0,%d",
			configs.Width,
			configs.Height,
			configs.Width,
			configs.Height)
		defaultZone["objects"] = []string{
			"person",
		}
		zones["all"] = defaultZone
		m["zones"] = zones

		onvif := make(map[string]interface{})
		onvif["host"] = camera.Ip
		onvif["port"] = camera.Port
		onvif["user"] = camera.Username
		onvif["password"] = camera.Password
		onvif["isapi_fallback"] = true

		isapiSidecar := make(map[string]interface{})
		isapiSidecar["host"] = "localhost"
		isapiSidecar["port"] = 5600
		onvif["isapi_sidecar"] = isapiSidecar

		autotracking := make(map[string]interface{})
		autotracking["enabled"] = true
		autotracking["zooming"] = "disabled"
		autotracking["track"] = []string{
			"person",
		}
		autotracking["required_zones"] = []string{
			"all",
		}
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
		mqtt["required_zones"] = []string{"all"}
		m["mqtt"] = mqtt

		cameras[camera.OpenGateCameraName] = m
	}

	c.c["cameras"] = cameras
	return nil
}

func (c *Configuration) buildDetectors() error {
	detectors := make(map[string]interface{})
	defaultDetector := make(map[string]interface{})
	det := "cpu"
	deviceType := ""
	switch c.integration.WithEdgeTpu {
	case true:
		det = "edgetpu"
		deviceType = "usb"
	case false:
	}
	if len(det) > 0 {
		defaultDetector["type"] = det
	}
	if len(deviceType) > 0 {
		defaultDetector["device"] = "usb"
	}
	detectors["default"] = defaultDetector
	c.c["detectors"] = detectors
	return nil
}
