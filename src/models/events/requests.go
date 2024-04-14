package events

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/CE-Thesis-2023/backend/src/models/db"
)

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

type PTZCtrlRequest struct {
	CameraId string `json:"cameraId"`
	Pan      int    `json:"pan"`
	Tilt     int    `json:"tilt"`
	Duration int    `json:"duration"`
}

type Event struct {
	Prefix    string
	ID        string
	Type      string
	Arguments []string
}

func (e *Event) Topic() string {
	p := e.Prefix
	if e.ID != "" {
		p = filepath.Join(p, e.ID)
	}
	if e.Type != "" {
		p = filepath.Join(p, e.Type)
	}
	return p
}

func (e *Event) String() string {
	return e.Topic()
}

func (e *Event) Parse(topic string) {
	parts := strings.Split(topic, "/")
	if len(parts) == 0 {
		return
	}
	e.Prefix = parts[0]
	if len(parts) > 1 {
		_ = parts[1]
	}
	if len(parts) > 2 {
		e.Type = parts[2]
	}
	if len(parts) > 3 {
		e.ID = parts[3]
	}
	if len(parts) > 4 {
		e.Arguments = parts[4:]
	}
}

const (
	EventReply_OK = `{ "status": "ok" }`
)

type EventReply struct {
	Status string `json:"status"`
	Err    error  `json:"error"`
}

func (r *EventReply) JSON() ([]byte, error) {
	return json.Marshal(r)
}
