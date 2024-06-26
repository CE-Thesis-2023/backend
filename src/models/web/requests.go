package web

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"

	"github.com/CE-Thesis-2023/backend/src/models/db"
	"github.com/CE-Thesis-2023/backend/src/models/events"
)

type GetTranscodersRequest struct {
	Ids []string `json:"ids"`
}

type GetTranscodersResponse struct {
	Transcoders []db.Transcoder `json:"transcoders"`
}

type RegisterTranscoderRequest struct {
	Name string `json:"name"`
}
type UpdateTranscoderRequest struct {
	Id                       string `form:"id" json:"id"`
	Name                     string `form:"name" json:"name"`
	LogLevel                 string `form:"logLevel" json:"logLevel" validate:"oneof=debug info warning"`
	HardwareAccelerationType string `form:"hardwareAccelerationType" json:"hardwareAccelerationType" validate:"oneof=cpu vaapi quicksync"`
	EdgeTpuEnabled           bool   `form:"edgeTpuEnabled" json:"edgeTpuEnabled"`
}

func (r UpdateTranscoderRequest) Patch(m *db.Transcoder, o *db.OpenGateIntegration) {
	if r.Name != "" {
		m.Name = r.Name
	}
	if o != nil {
		if r.LogLevel != "" {
			o.LogLevel = r.LogLevel
		}
		if r.HardwareAccelerationType != "" {
			o.HardwareAccelerationType = r.HardwareAccelerationType
		}
		o.WithEdgeTpu = r.EdgeTpuEnabled
	}
}

type GetCamerasRequest struct {
	Ids []string `json:"ids"`
}

type GetCamerasResponse struct {
	Cameras []db.Camera `json:"cameras"`
}

type AddCameraRequest struct {
	Name         string `json:"name"`
	Ip           string `json:"ip"`
	Port         int    `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Autotracking bool   `json:"autotracking"`

	TranscoderId string `json:"transcoderId"`
}

type AddCameraResponse struct {
	CameraId string `json:"cameraId"`
}

type DeleteCameraRequest struct {
	CameraId string `json:"cameraId"`
}

type GetStreamInfoRequest struct {
	CameraId string `json:"cameraId"`
}

type AddCamerasToGroupRequest struct {
	CameraIds []string `json:"cameraIds"`
	GroupId   string   `json:"groupId"`
}

type RemoveCamerasFromGroupRequest struct {
	CameraIds []string `json:"cameraIds"`
	GroupId   string   `json:"groupId"`
}

type AddCamerasToGroupResponse struct {
	GroupId string `json:"groupId"`
}

type RemoveCamerasFromGroupResponse struct {
	GroupId string `json:"groupId"`
}

type GetCameraGroupsRequest struct {
	Ids []string `json:"ids"`
}

type GetCameraGroupsResponse struct {
	CameraGroups []db.CameraGroup `json:"cameraGroups"`
}

type AddCameraGroupRequest struct {
	Name       string   `json:"name"`
	CamerasIds []string `json:"camerasIds"`
}

type AddCameraGroupResponse struct {
	GroupId string `json:"groupId"`
}

type DeleteCameraGroupRequest struct {
	GroupId string `json:"groupId"`
}

type DeleteCameraGroupResponse struct {
	GroupId string `json:"groupId"`
}

type UpdateCameraGroupRequest struct {
	GroupId   string   `json:"groupId"`
	CameraIds []string `json:"cameraIds"`
	Name      string   `json:"name"`
}

type GetStreamInfoResponse struct {
	StreamUrl      string `json:"streamUrl"`
	Protocol       string `json:"protocol"`
	TranscoderId   string `json:"transcoderId"`
	TranscoderName string `json:"transcoderName"`
	Started        bool   `json:"enabled"`
}

type ToggleStreamRequest struct {
	CameraId string `json:"-"`
	Start    bool   `json:"-"`
}

type RemoteControlRequest struct {
	CameraId string `json:"cameraId"`
	Pan      int    `json:"pan"`
	Tilt     int    `json:"tilt"`
}

type GetCameraDeviceInfoRequest struct {
	CameraId string `json:"cameraId"`
}

type GetCameraDeviceInfo struct {
	CameraId             string             `json:"cameraId"`
	DeviceName           string             `json:"deviceName"`
	DeviceLocation       string             `json:"deviceLocation"`
	Status               CameraDeviceStatus `json:"deviceStatus"`
	Model                string             `json:"model"`
	SerialNumber         string             `json:"serialNumber"`
	FirmwareVersion      string             `json:"firmwareVersion"`
	FirmwareReleasedDate string             `json:"firmwareReleasedDate"`
	Capacity             int                `json:"capacity"`
	UsedCapacity         int                `json:"usedCapacity"`
}

type CameraDeviceStatus struct {
	Status               string                     `json:"status"`
	DetailAbnormalStatus CameraDeviceAbnormalStatus `json:"detailAbnormalStatus"`
}

type CameraDeviceAbnormalStatus struct {
	HardDiskFull         bool `json:"hardDiskFull"`
	HardDiskError        bool `json:"hardDiskError"`
	EthernetBroken       bool `json:"ethernetBroken"`
	IPAddrConflict       bool `json:"ipaddrConflict"`
	IllegalAccess        bool `json:"illegalAccess"`
	RecordError          bool `json:"recordError"`
	RAIDLogicDiskError   bool `json:"raidLogicDiskError"`
	SpareWorkDeviceError bool `json:"spareWorkDeviceError"`
}

type SendEventToMqttRequest struct {
	CameraId string `json:"cameraId"`
	Event    string `json:"event"`
}

type EventRequest struct {
	Event string `json:"event"`
}

type PublicEventToOtherCamerasInGroupRequest struct {
	CameraId string `json:"cameraId"`
	Event    string `json:"event"`
}

type GetCamerasByGroupIdRequest struct {
	GroupId string `json:"groupId"`
}

type GetCamerasByGroupIdResponse struct {
	Cameras []db.Camera `json:"cameras"`
}

type GetCameraByTranscoderId struct {
	TranscoderId        string   `json:"openGateId"`
	OpenGateCameraNames []string `json:"openGateCameraNames"`
}

type GetCameraByOpenGateIdResponse struct {
	Cameras []db.Camera `json:"camera"`
}

type GetOpenGateIntegrationByIdRequest struct {
	OpenGateId string `json:"openGateId"`
}

type GetOpenGateIntegrationByIdResponse struct {
	OpenGateIntegration *db.OpenGateIntegration `json:"openGateIntegration"`
}

type UpdateOpenGateIntegrationRequest struct {
	OpenGateId            string                                `json:"-"`
	LogLevel              string                                `json:"logLevel,omitempty"`
	SnapshotRetentionDays int                                   `json:"snapshotRetentionDays,omitempty"`
	Mqtt                  *UpdateOpenGateIntegrationMqttRequest `json:"mqtt,omitempty"`
}

type UpdateOpenGateIntegrationMqttRequest struct {
	Enabled  bool   `json:"enabled"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type GetOpenGateCameraSettingsRequest struct {
	CameraId []string `json:"cameraId"`
}

type GetOpenGateCameraSettingsResponse struct {
	OpenGateCameraSettings []db.OpenGateCameraSettings `json:"openGateCameraSettings"`
}

type GetOpenGateMqttSettingsResponse struct {
	OpenGateMqttConfiguration *db.OpenGateMqttConfiguration `json:"openGateMqttConfiguration"`
}

type GetOpenGateMqttSettingsRequest struct {
	ConfigurationId string `json:"configurationId"`
}

type DeleteTranscoderRequest struct {
	DeviceId string `json:"deviceId"`
}

type GetObjectTrackingEventByIdRequest struct {
	EventId         []string      `json:"eventId"`
	OpenGateEventId []string      `json:"openGateEventId"`
	CameraId        string        `json:"cameraId"`
	Limit           int           `json:"limit"`
	Within          time.Duration `json:"within"`
	IsLastestFirst  bool          `json:"isLastestFirst"`
}

type GetObjectTrackingEventByIdResponse struct {
	ObjectTrackingEvents []db.ObjectTrackingEvent `json:"objectTrackingEvents"`
}

type AddObjectTrackingEventRequest struct {
	TranscoderId string                 `json:"transcoderId"`
	Event        *events.DetectionEvent `json:"event"`
}

type AddObjectTrackingEventResponse struct {
	EventId string `json:"eventId"`
}

type UpdateObjectTrackingEventRequest struct {
	TranscoderId string                 `json:"transcoderId"`
	EventId      string                 `json:"eventId"`
	Event        *events.DetectionEvent `json:"event"`
}

type UpdateObjectTrackingEventResponse struct {
	EventId string `json:"eventId"`
}

type DeleteObjectTrackingEventRequest struct {
	EventId string `json:"eventId"`
}

type GetTranscoderOpenGateConfigurationRequest struct {
	TranscoderId string `json:"transcoderId"`
}

type GetTranscoderOpenGateConfigurationResponse struct {
	Base64 string `json:"base64"`
}

type GetStreamConfigurationsRequest struct {
	CameraId []string `json:"cameraId"`
}

type GetStreamConfigurationsResponse struct {
	StreamConfigurations []TranscoderStreamConfiguration `json:"streamConfigurations"`
}

type GetMQTTEventEndpointRequest struct {
	TranscoderId string `json:"transcoderId"`
}

type GetMQTTEventEndpointResponse struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	TlsEnabled  bool   `json:"tlsEnabled"`
	SubscribeOn string `json:"subscribeOn"`
	PublishOn   string `json:"publishOn"`
}

type DeviceHealthcheckRequest struct {
	TranscoderId string `json:"transcoderId"`
}

type DeviceHealthcheckResponse struct {
	Status string `json:"status"`
}

type UpsertOpenGateCameraStatsRequest struct {
	TranscoderId string  `json:"transcoderId"`
	CameraName   string  `json:"cameraName"`
	CameraFPS    float64 `json:"cameraFps"`
	DetectionFPS float64 `json:"detectionFps"`
	CapturePID   int     `json:"capturePid"`
	ProcessID    int     `json:"processId"`
	ProcessFPS   float64 `json:"processFps"`
	SkippedFPS   float64 `json:"skippedFps"`
}

type UpsertOpenGateCameraStatsResponse struct {
	CameraStatId uuid.UUID `json:"cameraStatId"`
}

type UpsertOpenGateDetectorsStatsResponse struct {
	DetectorStatId uuid.UUID `json:"detectorStatId"`
}

type UpsertOpenGateDetectorsStatsRequest struct {
	TranscoderId   string  `json:"transcoderId"`
	DetectorName   string  `json:"detectName"`
	DetectorStart  float64 `json:"detectorStart"`
	InferenceSpeed float64 `json:"inferenceSpeed"`
	ProcessID      int     `json:"processId"`
}

type GetDetectablePeopleRequest struct {
	PersonIds []string `json:"personIds"`
}

type GetDetectablePeopleResponse struct {
	People []db.DetectablePerson `json:"people"`
}

type GetDetectablePeopleImagePresignedUrlRequest struct {
	PersonId string `json:"personId"`
}

type GetDetectablePeopleImagePresignedUrlResponse struct {
	PresignedUrl string        `json:"presignedUrl"`
	Expires      time.Duration `json:"expires"`
}

type AddDetectablePersonRequest struct {
	Name        string `json:"name"`
	Age         string `json:"age"`
	Base64Image string `json:"base64Image"`
}

func (r AddDetectablePersonRequest) String() string {
	type req AddDetectablePersonRequest
	copied := req(r)
	copied.Base64Image = "LONG_STRING_OMITTED"
	b, _ := json.Marshal(copied)
	return string(b)
}

type AddDetectablePersonResponse struct {
	PersonId string `json:"personId"`
}

type DeleteDetectablePersonRequest struct {
	PersonId string `json:"personId"`
}

type GetLatestOpenGateStatsResponse struct {
	CameraStats   db.OpenGateCameraStats   `json:"cameraStats"`
	DetectorStats db.OpenGateDetectorStats `json:"detectorStats"`
}

type GetSnapshotPresignedUrlRequest struct {
	SnapshotId []string `json:"snapshotId"`
}

type GetSnapshotPresignedUrlResponse struct {
	Snapshots    []db.Snapshot     `json:"snapshot"`
	PresignedUrl map[string]string `json:"presignedUrl"`
}

type UpsertSnapshotRequest struct {
	RawImage        string `json:"rawImage"`
	OpenGateEventId string `json:"openGateEventId"`
	TranscoderId    string `json:"transcoderId"`
}

type GetImageBasePathRequest struct {
	AssetType string `json:"assetType"`
	Id        string `json:"id"`
}

type GetImageBasePathResponse struct {
	BasePath string `json:"basePath"`
}

type GetLatestOpenGateCameraStatsRequest struct {
	TranscoderId string   `json:"trancoderId"`
	CameraNames  []string `json:"cameraNames"`
}

type GetPersonHistoryRequest struct {
	PersonId  []string `json:"personId"`
	HistoryId []string `json:"historyId"`
}

type GetPersonHistoryResponse struct {
	Histories []db.PersonHistory `json:"histories"`
}

type GetTranscoderStatusRequest struct {
	TranscoderId []string `json:"transcoderId"`
	CameraId     []string `json:"cameraId"`
}

type GetTranscoderStatusResponse struct {
	Status []db.TranscoderStatus `json:"status"`
}

type UpdateTranscoderStatusRequest struct {
	TranscoderId       string  `json:"transcoderId" form:"transcoder_id"`
	CameraName         *string `json:"cameraName" form:"camera_name"`
	CameraId           *string `json:"cameraId" form:"camera_id"`
	ObjectDetection    *bool   `json:"objectDetection,omitempty" form:"object_detection"`
	AudioDetection     *bool   `json:"audioDetection,omitempty" form:"audio_detection"`
	OpenGateRecordings *bool   `json:"openGateRecordings,omitempty" form:"open_gate_recordings"`
	Snapshots          *bool   `json:"snapshots,omitempty" form:"snapshots"`
	MotionDetection    *bool   `json:"motionDetection,omitempty" form:"motion_detection"`
	ImproveContrast    *bool   `json:"improveContrast,omitempty" form:"improve_contrast"`
	Autotracker        *bool   `json:"autotracker,omitempty" form:"autotracker"`
	BirdseyeView       *bool   `json:"birdseyeView,omitempty" form:"birdseye_view"`
	OpenGateStatus     *bool   `json:"openGateStatus,omitempty" form:"open_gate_status"`
	TranscoderStatus   *bool   `json:"transcoderStatus,omitempty" form:"transcoder_status"`
}

type UpdateTranscoderStatusResponse struct {
	TranscoderId string `json:"transcoderId"`
}
