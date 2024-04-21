package eventsapi

import (
	"context"
	"encoding/json"
	"github.com/CE-Thesis-2023/backend/src/biz/service"
	"github.com/CE-Thesis-2023/backend/src/helper/transcoder"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/models/events"
	"github.com/CE-Thesis-2023/backend/src/models/web"
	"github.com/eclipse/paho.golang/paho"
	"strings"
)

var (
	TYPE_OPENGATE   = "opengate"
	TYPE_TRANSCODER = "transcoder"
)

type Command struct {
	Type     string `json:"type"`
	ClientId string `json:"clientId"`
	Action   string `json:"action"`

	actorPool  *transcoder.TranscoderActorsPool
	webService *service.WebService
}

type OpenGateStats struct {
	Cameras      map[string]events.OpenGateCameraStats    `json:"cameras"`
	DetectionFPS float64                                  `json:"detection_fps"`
	Detectors    map[string]events.OpenGateDetectorsStats `json:"detectors"`
	Service      map[string]interface{}                   `json:"service"`
}

func CommandFromPath(pub *paho.Publish, pool *transcoder.TranscoderActorsPool, webService *service.WebService) (*Command, error) {
	parts := strings.Split(pub.Topic, "/")
	if len(parts) == 0 {
		return nil, custerror.FormatInvalidArgument("path is empty")
	}
	if len(parts) < 3 {
		return nil, custerror.FormatInvalidArgument("path is too short")
	}
	return &Command{
		Type:       parts[0],
		ClientId:   parts[1],
		Action:     strings.Join(parts[2:], "/"),
		actorPool:  pool,
		webService: webService,
	}, nil
}

func (c *Command) Run(ctx context.Context, pub *paho.Publish) error {
	switch c.Type {
	case TYPE_OPENGATE:
		return c.runOpenGate(ctx, pub)
	case TYPE_TRANSCODER:
		return c.runTranscoder(ctx, pub)
	default:
		return custerror.FormatInvalidArgument("unknown type: %s", c.Type)
	}
}

const (
	OPENGATE_EVENTS = "events"
	OPENGATE_STATS  = "stats"
)

// https://docs.frigate.video/integrations/mqtt
func (c *Command) runOpenGate(ctx context.Context, pub *paho.Publish) error {
	switch c.Action {
	case OPENGATE_EVENTS:
		return c.runOpenGateEvents(ctx, pub)
	case OPENGATE_STATS:
		return c.runOpenGateStats(ctx, pub)
	default:
		{
			parts := strings.Split(c.Action, "/")
			if len(parts) < 3 {
				// If not, return an error indicating an invalid action format
				return custerror.FormatInvalidArgument("unknown action: %s", c.Action)
			}
			if parts[2] == "snapshot" {
				return c.runOpenGateSnapshot(ctx, pub)
			}
			return custerror.FormatInvalidArgument("unknown action: %s", c.Action)
		}
	}
}

func (c *Command) runOpenGateEvents(ctx context.Context, pub *paho.Publish) error {
	names := []string{c.extractCameraNameFromEvent(pub.Payload)}
	resp, err := c.webService.GetCamerasByTranscoderId(
		ctx,
		&web.GetCameraByTranscoderId{
			TranscoderId:        c.ClientId,
			OpenGateCameraNames: names,
		})
	if err != nil {
		return err
	}
	cameras := resp.Cameras

	for _, cam := range cameras {
		err = c.actorPool.Send(transcoder.TranscoderEventMessage{
			CameraId:     cam.CameraId,
			TranscoderId: cam.TranscoderId,
			OpenGateId:   c.ClientId,
			Type:         c.Type,
			Action:       c.Action,
			Payload:      pub.Payload,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) runOpenGateSnapshot(ctx context.Context, pub *paho.Publish) error {
	eventId := strings.Split(c.Action, "/")[1]

	if eventId == "" {
		return custerror.FormatInvalidArgument("unknown action: %s", c.Action)
	}

	result, err := c.webService.AddSnapshot(
		ctx,
		&web.AddSnapshotRequest{
			Base64Image: string(pub.Payload[:]),
		})

	if err != nil {
		return custerror.FormatInvalidArgument("failed to get event by id")
	}

	_, err = c.webService.UpdateSnapshotToEvent(ctx, &web.UpdateSnapshotToEventRequest{
		SnapshotId: result.SnapshotId,
		EventId:    eventId,
	})
	return nil
}

func (c *Command) runOpenGateStats(ctx context.Context, pub *paho.Publish) error {
	stats := c.extractStatFromStatsResponse(pub.Payload)

	if stats == nil {
		return custerror.FormatInvalidArgument("failed to extract stats from payload")
	}

	for cameraName, cameraStats := range stats.Cameras {
		_, err := c.webService.AddOpenGateCameraStats(ctx, &web.AddOpenGateCameraStatsRequest{
			CameraName:   cameraName,
			CameraFPS:    cameraStats.CameraFPS,
			DetectionFPS: cameraStats.DetectionFPS,
			CapturePID:   cameraStats.CapturePID,
			ProcessID:    cameraStats.PID,
			ProcessFPS:   cameraStats.ProcessFPS,
			SkippedFPS:   cameraStats.SkippedFPS,
		})

		if err != nil {
			return err
		}
	}

	for detectorName, detectorStats := range stats.Detectors {
		_, err := c.webService.AddOpenGateDetectorStats(ctx, &web.AddOpenGateDetectorsStatsRequest{
			DetectorName:   detectorName,
			DetectorStart:  detectorStats.DetectionStart,
			InferenceSpeed: detectorStats.InferenceSpeed,
			ProcessID:      detectorStats.PID,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) extractStatFromStatsResponse(stats []byte) *OpenGateStats {
	var statStruct OpenGateStats
	// Unmarshal the JSON into a map[string]interface{}
	if err := json.Unmarshal(stats, &statStruct); err != nil {
		return nil
	}

	return &statStruct
}

func (c *Command) extractCameraNameFromEvent(event []byte) string {
	var names []string
	var eventStruct map[string]interface{}
	if err := json.Unmarshal(event, &eventStruct); err != nil {
		return ""
	}
	before := eventStruct["before"]
	if before != nil {
		beforeStruct := before.(map[string]interface{})
		if beforeStruct["camera"] != nil {
			names = append(names,
				beforeStruct["camera"].(string))
		}
	}
	return names[0]
}

func (c *Command) runTranscoder(_ context.Context, _ *paho.Publish) error {
	switch c.Action {
	default:
		return custerror.FormatInvalidArgument("unknown action: %s", c.Action)
	}
}
