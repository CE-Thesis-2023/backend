package db

import "time"

type Transcoder struct {
	DeviceId string `json:"deviceId" db:"device_id,primary"`
	Name     string `json:"name" db:"name"`

	OpenGateIntegrationId string `json:"openGateIntegrationId" db:"open_gate_integration_id"`
}

type OpenGateIntegration struct {
	OpenGateId            string `json:"openGateId" db:"open_gate_id,primary"`
	Available             bool   `json:"available" db:"available"`
	IsRestarting          bool   `json:"isRestarting" db:"is_restarting"`
	LogLevel              string `json:"logLevel" db:"log_level"`
	SnapshotRetentionDays int    `json:"snapshotRetentionDays" db:"snapshot_retention_days"`

	MqttId       string `json:"mqttId" db:"mqtt_id"`
	TranscoderId string `json:"transcoderId" db:"transcoder_id"`
}

type OpenGateMqttConfiguration struct {
	ConfigurationId string `json:"configurationId" db:"configuration_id,primary"`
	Enabled         bool   `json:"enabled" db:"enabled"`
	Host            string `json:"host" db:"host"`
	Port            int    `json:"port" db:"port"`
	User            string `json:"user" db:"user"`
	Password        string `json:"password" db:"password"`

	OpenGateId string `json:"openGateId" db:"open_gate_id"`
}

type ObjectTrackingEvent struct {
	EventId   string `json:"eventId" db:"event_id,primary"`
	EventType string `json:"eventType" db:"event_type"`

	CameraId string `json:"cameraId" db:"camera_id"`
}

type Camera struct {
	CameraId string `json:"cameraId" db:"camera_id,primary"`
	Name     string `json:"name" db:"name"`
	Ip       string `json:"ip" db:"ip"`
	Port     int    `json:"port" db:"port"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
	Enabled  bool   `json:"enabled" db:"enabled"`

	GroupId      string `json:"groupId" db:"group_id,omitempty"`
	TranscoderId string `json:"transcoderId" db:"transcoder_id,omitempty"`
	SettingsId   string `json:"settingsId" db:"settings_id,omitempty"`
}

type OpenGateCameraSettings struct {
	SettingsId  string `json:"settingsId" db:"settings_id,primary"`
	Height      int    `json:"height" db:"height"`
	Width       int    `json:"width" db:"width"`
	Fps         int    `json:"fps" db:"fps"`
	MqttEnabled bool   `json:"mqttEnabled" db:"mqtt_enabled"`
	Timestamp   bool   `json:"timestamp" db:"timestamp"`
	BoundingBox bool   `json:"boundingBox" db:"bounding_box"`
	Crop        bool   `json:"crop" db:"crop"`

	OpenGateId string `json:"openGateId" db:"open_gate_id"`
	CameraId   string `json:"cameraId" db:"camera_id"`
}

type CameraGroup struct {
	GroupId     string    `json:"groupId" db:"group_id,primary"`
	Name        string    `json:"name" db:"name"`
	CreatedDate time.Time `json:"createdDate" db:"created_date"`
}

func (t *Transcoder) Fields() []string {
	fs := []string{}
	fs = append(fs,
		"device_id",
		"name",
		"open_gate_integration_id",
	)
	return fs
}

func (t *Transcoder) Values() []interface{} {
	vs := []interface{}{
		t.DeviceId,
		t.Name,
		t.OpenGateIntegrationId,
	}
	return vs
}

func (t *Camera) Fields() []string {
	fs := []string{}
	fs = append(fs,
		"camera_id",
		"name",
		"ip",
		"port",
		"username",
		"password",
		"enabled",
		"transcoder_id",
		"group_id",
		"settings_id",
	)
	return fs
}

func (t *Camera) Values() []interface{} {
	vs := []interface{}{}
	vs = append(vs,
		t.CameraId,
		t.Name,
		t.Ip,
		t.Port,
		t.Username,
		t.Password,
		t.Enabled,
		t.TranscoderId,
		t.GroupId,
		t.SettingsId,
	)
	return vs
}

func (t *CameraGroup) Fields() []string {
	fs := []string{}
	fs = append(fs,
		"group_id",
		"name",
		"created_date",
	)
	return fs
}

func (t *CameraGroup) Values() []interface{} {
	vs := []interface{}{}
	vs = append(vs,
		t.GroupId,
		t.Name,
		t.CreatedDate,
	)
	return vs
}

func (t *OpenGateIntegration) Fields() []string {
	fs := []string{}
	fs = append(fs,
		"open_gate_id",
		"available",
		"is_restarting",
		"log_level",
		"snapshot_retention_days",
		"mqtt_id",
		"transcoder_id",
	)
	return fs
}

func (t *OpenGateIntegration) Values() []interface{} {
	vs := []interface{}{}
	vs = append(vs,
		t.OpenGateId,
		t.Available,
		t.IsRestarting,
		t.LogLevel,
		t.SnapshotRetentionDays,
		t.MqttId,
		t.TranscoderId,
	)
	return vs
}

func (t *OpenGateMqttConfiguration) Fields() []string {
	fs := []string{}
	fs = append(fs,
		"configuration_id",
		"enabled",
		"host",
		"port",
		"user",
		"password",
		"open_gate_id",
	)
	return fs
}

func (t *OpenGateMqttConfiguration) Values() []interface{} {
	vs := []interface{}{}
	vs = append(vs,
		t.ConfigurationId,
		t.Enabled,
		t.Host,
		t.Port,
		t.User,
		t.Password,
		t.OpenGateId,
	)
	return vs
}

func (t *OpenGateCameraSettings) Fields() []string {
	fs := []string{}
	fs = append(fs,
		"settings_id",
		"height",
		"width",
		"fps",
		"mqtt_enabled",
		"timestamp",
		"bounding_box",
		"crop",
		"open_gate_id",
		"camera_id",
	)
	return fs
}

func (t *OpenGateCameraSettings) Values() []interface{} {
	vs := []interface{}{}
	vs = append(vs,
		t.SettingsId,
		t.Height,
		t.Width,
		t.Fps,
		t.MqttEnabled,
		t.Timestamp,
		t.BoundingBox,
		t.Crop,
		t.OpenGateId,
		t.CameraId,
	)
	return vs
}
