package eventsapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/CE-Thesis-2023/backend/src/biz/service"
	"github.com/CE-Thesis-2023/backend/src/helper/transcoder"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/events"
	"github.com/CE-Thesis-2023/backend/src/models/web"
	"github.com/eclipse/paho.golang/paho"
)

var (
	TYPE_OPENGATE   = "opengate"
	TYPE_TRANSCODER = "transcoder"
)

type Topic struct {
	Type     string   `json:"type"`
	SenderId string   `json:"senderId"`
	ExtraIds []string `json:"extraIds"`
	Action   string   `json:"action"`
}

type Command struct {
	topic *Topic

	actorPool  *transcoder.TranscoderActorsPool
	webService *service.WebService
}

type OpenGateStats struct {
	Cameras      map[string]events.OpenGateCameraStats    `json:"cameras"`
	DetectionFPS float64                                  `json:"detection_fps"`
	Detectors    map[string]events.OpenGateDetectorsStats `json:"detectors"`
	Service      map[string]interface{}                   `json:"service"`
}

func ToTopic(topic string) (*Topic, error) {
	parts := strings.Split(topic, "/")
	if len(parts) == 0 {
		return nil, custerror.FormatInvalidArgument("path is empty")
	}
	if len(parts) < 3 {
		return nil, custerror.FormatInvalidArgument("path is too short")
	}
	switch len(parts) {
	case 3:
		return &Topic{
			Type:     parts[0],
			SenderId: parts[1],
			Action:   parts[2],
		}, nil
	case 5:
		return &Topic{
			Type:     parts[0],
			SenderId: parts[1],
			ExtraIds: []string{parts[2]},
			Action:   strings.Join([]string{parts[3], parts[4]}, "/"),
		}, nil
	case 4:
		return &Topic{
			Type:     parts[0],
			SenderId: parts[1],
			ExtraIds: []string{parts[2]},
			Action:   parts[3],
		}, nil
	default:
		return &Topic{
			Type:     parts[0],
			SenderId: parts[1],
			ExtraIds: []string{parts[2]},
			Action:   strings.Join(parts[3:], "/"),
		}, nil
	}
}

func CommandFromPath(pub *paho.Publish, pool *transcoder.TranscoderActorsPool, webService *service.WebService) (*Command, error) {
	topic, err := ToTopic(pub.Topic)
	if err != nil {
		return nil, err
	}
	return &Command{
		topic:      topic,
		actorPool:  pool,
		webService: webService,
	}, nil
}

func (c *Command) Run(ctx context.Context, pub *paho.Publish) error {
	switch c.topic.Type {
	case TYPE_OPENGATE:
		return c.runOpenGate(ctx, pub)
	case TYPE_TRANSCODER:
		return c.runTranscoder(ctx, pub)
	default:
		return custerror.FormatInvalidArgument("unknown type: %s", c.topic)
	}
}

const (
	OPENGATE_EVENTS    = "events"
	OPENGATE_STATS     = "stats"
	OPENGATE_AVAILABLE = "available"
	OPENGATE_SNAPSHOT  = "snapshot"
)

// https://docs.frigate.video/integrations/mqtt
func (c *Command) runOpenGate(ctx context.Context, pub *paho.Publish) error {
	switch c.topic.Action {
	case OPENGATE_EVENTS:
		return c.runOpenGateEvents(ctx, pub)
	case OPENGATE_STATS:
		return c.runOpenGateStats(ctx, pub)
	case OPENGATE_SNAPSHOT:
		return c.runOpenGateSnapshot(ctx, pub)
	case OPENGATE_AVAILABLE:
		// TODO: Change status of transcoder device
		logger.SInfo("OpenGate available")
	}
	return nil
}

func (c *Command) runOpenGateEvents(ctx context.Context, pub *paho.Publish) error {
	names := []string{c.extractCameraNameFromEvent(pub.Payload)}
	resp, err := c.webService.GetCamerasByTranscoderId(
		ctx,
		&web.GetCameraByTranscoderId{
			TranscoderId:        c.topic.SenderId,
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
			OpenGateId:   c.topic.SenderId,
			Type:         c.topic.Type,
			Action:       c.topic.Action,
			Payload:      pub.Payload,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) snapshotToBase64(snapshot []byte) string {
	return base64.
		StdEncoding.
		EncodeToString(snapshot)
}

func (c *Command) runOpenGateSnapshot(ctx context.Context, pub *paho.Publish) error {
	transcoderId := c.topic.SenderId
	eventId := c.topic.ExtraIds[0]

	if err := c.webService.UpsertSnapshot(ctx, &web.UpsertSnapshotRequest{
		TranscoderId:    transcoderId,
		OpenGateEventId: eventId,
		RawImage:        string(pub.Payload),
	}); err != nil {
		return err
	}
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
	switch c.topic.Action {
	default:
		return custerror.FormatInvalidArgument("unknown action: %s", c.topic.Action)
	}
}
