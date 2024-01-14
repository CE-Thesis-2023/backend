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

type DeleteTranscoderRequest struct {
	Id string `json:"id"`
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

type GetStreamInfoResponse struct {
	StreamUrl      string `json:"streamUrl"`
	Protocol       string `json:"protocol"`
	TranscoderId   string `json:"transcoderId"`
	TranscoderName string `json:"transcoderName"`
	Started        bool   `json:"started"`
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
