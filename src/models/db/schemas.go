package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/pgvector/pgvector-go"
)

type Transcoder struct {
	DeviceId string `json:"deviceId" db:"device_id,primary" gorm:"index"`
	Name     string `json:"name" db:"name"`

	OpenGateIntegrationId string `json:"openGateIntegrationId" db:"open_gate_integration_id"`
}

type OpenGateIntegration struct {
	OpenGateId            string `json:"openGateId" db:"open_gate_id,primary" gorm:"index"`
	Available             bool   `json:"available" db:"available"`
	IsRestarting          bool   `json:"isRestarting" db:"is_restarting"`
	LogLevel              string `json:"logLevel" db:"log_level"`
	SnapshotRetentionDays int    `json:"snapshotRetentionDays" db:"snapshot_retention_days"`

	MqttId       string `json:"mqttId" db:"mqtt_id"`
	TranscoderId string `json:"transcoderId" db:"transcoder_id" gorm:"index"`
}

type OpenGateMqttConfiguration struct {
	ConfigurationId string `json:"configurationId" db:"configuration_id,primary" gorm:"index"`
	Enabled         bool   `json:"enabled" db:"enabled"`
	Host            string `json:"host" db:"host"`
	Port            int    `json:"port" db:"port"`
	Username        string `json:"username" db:"username"`
	Password        string `json:"password" db:"password"`
	OpenGateId      string `json:"openGateId" db:"open_gate_id"`
}

type ObjectTrackingEvent struct {
	EventId         string `json:"eventId" db:"event_id,primary" gorm:"index"`
	OpenGateEventId string `json:"openGateEventId" db:"open_gate_event_id" gorm:"index"`
	OpenGateId      string `json:"openGateId" db:"open_gate_id"`
	EventType       string `json:"eventType" db:"event_type"`

	CameraId      string     `json:"cameraId" db:"camera_id"`
	CameraName    string     `json:"CameraName" db:"camera_name"`
	FrameTime     *time.Time `json:"frameTime" db:"frame_time"`
	Label         string     `json:"label" db:"label"`
	TopScore      float64    `json:"topScore" db:"top_score"`
	Score         float64    `json:"score" db:"score"`
	HasSnapshot   bool       `json:"hasSnapshot" db:"has_snapshot"`
	HasClip       bool       `json:"hasClip" db:"has_clip"`
	Stationary    bool       `json:"stationary" db:"stationary"`
	FalsePositive bool       `json:"falsePositive" db:"false_positive"`
	StartTime     *time.Time `json:"startTime" db:"start_time"`
	EndTime       *time.Time `json:"endTime" db:"end_time"`
	SnapshotId    *string    `json:"snapshotId" db:"snapshot_id"`
}

func (s *ObjectTrackingEvent) Fields() []string {
	fs := []string{}
	fs = append(fs,
		"event_id",
		"open_gate_event_id",
		"event_type",
		"camera_id",
		"camera_name",
		"frame_time",
		"label",
		"top_score",
		"score",
		"has_snapshot",
		"has_clip",
		"stationary",
		"false_positive",
		"start_time",
		"end_time",
		"snapshot_id",
	)
	return fs
}

func (s *ObjectTrackingEvent) Values() []interface{} {
	vs := []interface{}{}
	vs = append(vs,
		s.EventId,
		s.OpenGateEventId,
		s.EventType,
		s.CameraId,
		s.CameraName,
		s.FrameTime,
		s.Label,
		s.TopScore,
		s.Score,
		s.HasSnapshot,
		s.HasClip,
		s.Stationary,
		s.FalsePositive,
		s.StartTime,
		s.EndTime,
		s.SnapshotId,
	)
	return vs
}

type Camera struct {
	CameraId string `json:"cameraId" db:"camera_id,primary" gorm:"index"`
	Name     string `json:"name" db:"name"`

	Ip       string `json:"ip" db:"ip"`
	Port     int    `json:"port" db:"port"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
	Enabled  bool   `json:"enabled omitempty" db:"enabled"`

	OpenGateCameraName string `json:"openGateCameraName" db:"open_gate_camera_name"`
	GroupId            string `json:"groupId" db:"group_id,omitempty"`
	TranscoderId       string `json:"transcoderId" db:"transcoder_id,omitempty" gorm:"index"`
	SettingsId         string `json:"settingsId omitempty" db:"settings_id,omitempty"`
}

type OpenGateCameraSettings struct {
	SettingsId  string `json:"settingsId" db:"settings_id,primary" gorm:"index"`
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
	GroupId     string    `json:"groupId" db:"group_id,primary" gorm:"index"`
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
		"open_gate_camera_name",
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
		t.OpenGateCameraName,
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
		"username",
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
		t.Username,
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

type DetectablePerson struct {
	PersonId  string          `json:"personId" db:"person_id,primary" gorm:"index"`
	Name      string          `json:"name" db:"name"`
	Age       string          `json:"age" db:"age"`
	ImagePath string          `json:"-" db:"image_path"`
	Embedding pgvector.Vector `json:"-" db:"embedding" gorm:"type:vector(128)"`
}

func (t *DetectablePerson) Index(d *gorm.DB) {
	d.Exec("CREATE INDEX ON detectable_people USING hnsw (embedding vector_l2_ops)")
}

func (t *DetectablePerson) Fields() []string {
	fs := []string{}
	fs = append(fs,
		"person_id",
		"name",
		"age",
		"image_path",
		"embedding",
	)
	return fs
}

func (t *DetectablePerson) Values() []interface{} {
	vs := []interface{}{}
	vs = append(vs,
		t.PersonId,
		t.Name,
		t.Age,
		t.ImagePath,
		t.Embedding,
	)
	return vs
}

type OpenGateCameraStats struct {
	CameraStatId uuid.UUID `json:"camera_stat_id" db:"camera_stat_id,primary" gorm:"index"`
	TranscoderId string    `json:"transcoder_id" db:"transcoder_id" gorm:"index"`
	CameraName   string    `json:"camera_name" db:"camera_name"`
	CameraFPS    float64   `json:"camera_fps" db:"camera_fps"`
	DetectionFPS float64   `json:"detection_fps" db:"detection_fps"`
	CapturePID   int       `json:"capture_pid" db:"capture_p_id"`
	ProcessID    int       `json:"process_id" db:"process_id"`
	ProcessFPS   float64   `json:"process_fps" db:"process_fps"`
	SkippedFPS   float64   `json:"skipped_fps" db:"skipped_fps"`
	Timestamp    time.Time `json:"timestamp" db:"timestamp"`
}

func (t *OpenGateCameraStats) Fields() []string {
	fs := []string{}
	fs = append(fs,
		"camera_stat_id",
		"camera_name",
		"camera_fps",
		"detection_fps",
		"capture_p_id",
		"process_id",
		"process_fps",
		"skipped_fps",
		"timestamp",
	)
	return fs
}

func (t *OpenGateCameraStats) Values() []interface{} {
	vs := []interface{}{}
	vs = append(vs,
		t.CameraStatId,
		t.CameraName,
		t.CameraFPS,
		t.DetectionFPS,
		t.CapturePID,
		t.ProcessID,
		t.ProcessFPS,
		t.SkippedFPS,
		t.Timestamp,
	)
	return vs
}

type OpenGateDetectorStats struct {
	DetectorStatId uuid.UUID `json:"detector_stat_id" db:"detector_stat_id,primary" gorm:"index"`
	DetectorName   string    `json:"detector_name" db:"detector_name"`
	TranscoderId   string    `json:"transcoder_id" db:"transcoder_id" gorm:"index"`
	DetectorStart  float64   `json:"detector_start" db:"detector_start"`
	InferenceSpeed float64   `json:"inference_speed" db:"inference_speed"`
	ProcessID      int       `json:"process_id" db:"process_id"`
	Timestamp      time.Time `json:"timestamp" db:"timestamp"`
}

func (t *OpenGateDetectorStats) Fields() []string {
	fs := []string{}
	fs = append(fs,
		"detector_stat_id",
		"detector_name",
		"detector_start",
		"inference_speed",
		"process_id",
		"timestamp",
	)
	return fs
}

func (t *OpenGateDetectorStats) Values() []interface{} {
	vs := []interface{}{}
	vs = append(vs,
		t.DetectorStatId,
		t.DetectorName,
		t.DetectorStart,
		t.InferenceSpeed,
		t.ProcessID,
		t.Timestamp,
	)
	return vs
}

type Snapshot struct {
	SnapshotId       string    `json:"snapshotId" db:"snapshot_id,primary" gorm:"index"`
	Timestamp        time.Time `json:"timestamp" db:"timestamp"`
	TranscoderId     string    `json:"transcoderId" db:"transcoder_id"`
	OpenGateEventId  string    `json:"openGateEventId" db:"open_gate_event_id" gorm:"index"`
	DetectedPeopleId *string   `json:"detectedPeopleId,omitempty" db:"detected_people_id"`
}

func (t *Snapshot) Fields() []string {
	fs := []string{}
	fs = append(fs,
		"snapshot_id",
		"timestamp",
		"detected_people_id",
		"event_id",
		"transcoder_id",
		"open_gate_event_id",
		"detected_people_id",
	)
	return fs
}

func (t *Snapshot) Values() []interface{} {
	vs := []interface{}{}
	vs = append(vs,
		t.SnapshotId,
		t.Timestamp,
		t.DetectedPeopleId,
		t.TranscoderId,
		t.OpenGateEventId,
		t.DetectedPeopleId,
	)
	return vs
}
