package events

import "github.com/CE-Thesis-2023/backend/src/models/db"

type DeviceRegistrationRequest struct {
	DeviceId string `json:"deviceId"`
}

type UpdateCameraListRequest struct {
	DeviceId string `json:"deviceId"`
}

type UpdateCameraListResponse struct {
	Cameras []db.Camera `json:"cameras"`
}
