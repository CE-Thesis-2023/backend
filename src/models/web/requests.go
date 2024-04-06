package web

import (
	"github.com/CE-Thesis-2023/backend/src/models/db"
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
	Id   string `json:"id"`
	Name string `json:"name"`
}

type GetCamerasRequest struct {
	Ids []string `json:"ids"`
}

type GetCamerasResponse struct {
	Cameras []db.Camera `json:"cameras"`
}

type AddCameraRequest struct {
	Name     string `json:"name"`
	Ip       string `json:"ip"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`

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

type GetCameraByOpenGateIdRequest struct {
	OpenGateId  string   `json:"openGateId"`
	CameraNames []string `json:"cameraNames"`
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
	DeviceId string `json:"device_id"`
}
