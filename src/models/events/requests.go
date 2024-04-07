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

type DetectionEvent struct {
	Type   string               `json:"type"`
	Before DetectionEventStatus `json:"before"`
	After  DetectionEventStatus `json:"after"`
}

type DetectionEventStatus struct {
	ID                string                    `json:"id"`
	Camera            string                    `json:"camera"`
	FrameTime         float64                   `json:"frame_time"`
	SnapshotTime      float64                   `json:"snapshot_time"`
	Label             string                    `json:"label"`
	SubLabel          []interface{}             `json:"sub_label"`
	TopScore          float64                   `json:"top_score"`
	FalsePositive     bool                      `json:"false_positive"`
	StartTime         float64                   `json:"start_time"`
	EndTime           *float64                  `json:"end_time"`
	Score             float64                   `json:"score"`
	Box               []int64                   `json:"box"`
	Area              int64                     `json:"area"`
	Ratio             float64                   `json:"ratio"`
	Region            []int64                   `json:"region"`
	CurrentZones      []string                  `json:"current_zones"`
	EnteredZones      []string                  `json:"entered_zones"`
	Thumbnail         interface{}               `json:"thumbnail"`
	HasSnapshot       bool                      `json:"has_snapshot"`
	HasClip           bool                      `json:"has_clip"`
	Stationary        bool                      `json:"stationary"`
	MotionlessCount   int64                     `json:"motionless_count"`
	PositionChanges   int64                     `json:"position_changes"`
	Attributes        ObjectFeatures            `json:"attributes"`
	CurrentAttributes []ObjectCurrentAttributes `json:"current_attributes"`
}

type ObjectFeatures map[string]float64

type ObjectCurrentAttributes struct {
	Label string  `json:"label"`
	Box   []int64 `json:"box"`
	Score float64 `json:"score"`
}

type ObjectSubLabel struct {
	Double *float64
	String *string
}
