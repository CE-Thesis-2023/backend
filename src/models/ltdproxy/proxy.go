package ltdproxy

import "github.com/CE-Thesis-2023/backend/src/models/events"

type EventSnapshot struct {
	Base64Image string `json:"base64Image,omitempty"`
}

type UploadEventRequest struct {
	CameraName   string                 `json:"cameraName"`
	TranscoderId string                 `json:"transcoderId"`
	Event        *events.DetectionEvent `json:"event"`
	Snapshot     *EventSnapshot         `json:"snapshot,omitempty"`
}
