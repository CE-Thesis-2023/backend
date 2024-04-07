package transcoder

import (
	"context"

	"encoding/json"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"go.uber.org/zap"
)

func (p *transcoderEventProcessor) OpenGateAvailable(ctx context.Context, openGateId string, message []byte) error {
	payload := string(message)
	logger.SDebug("processor.OpenGateAvailable",
		zap.String("openGateId", openGateId),
		zap.String("message", payload))
	switch payload {
	case "online":
		logger.SInfo("OpenGate is online",
			zap.String("openGateId", openGateId))
	case "offline":
		logger.SInfo("OpenGate is offline",
			zap.String("openGateId", openGateId))
	}
	return nil
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
	SubLabel          []ObjectSubLabel          `json:"sub_label"`
	TopScore          float64                   `json:"top_score"`
	FalsePositive     bool                      `json:"false_positive"`
	StartTime         float64                   `json:"start_time"`
	EndTime           interface{}               `json:"end_time"`
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

func (p *transcoderEventProcessor) OpenGateEvent(ctx context.Context, openGateId string, message []byte) error {
	logger.SDebug("processor.OpenGateEvent",
		zap.String("openGateId", openGateId))

	var detectionEvent DetectionEvent
	if err := json.Unmarshal(message, &detectionEvent); err != nil {
		logger.SError("failed to unmarshal detection event",
			zap.Error(err))
		return err
	}

	switch detectionEvent.Type {
	case "new":
		logger.SInfo("new detection",
			zap.String("openGateId", openGateId))
	case "update":
		logger.SInfo("detection update",
			zap.String("openGateId", openGateId))
	case "end":
		logger.SInfo("detection end",
			zap.String("openGateId", openGateId))
	}

	logger.SDebug("detection event",
		zap.Any("before", detectionEvent.Before),
		zap.Any("after", detectionEvent.After))
	return nil
}
func (p *transcoderEventProcessor) addEventToDatabase(ctx context.Context, req *DetectionEvent) error {

}
