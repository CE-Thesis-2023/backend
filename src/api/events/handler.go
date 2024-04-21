package eventsapi

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/CE-Thesis-2023/backend/src/biz/service"
	"github.com/CE-Thesis-2023/backend/src/helper/transcoder"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
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

// https://docs.frigate.video/integrations/mqtt
func (c *Command) runOpenGate(ctx context.Context, pub *paho.Publish) error {
	switch c.topic.Action {
	case transcoder.OPENGATE_EVENTS:
		return c.runOpenGateEvents(ctx, pub)
	case transcoder.OPENGATE_STATS:
		return c.runOpenGateStats(ctx, pub)
	case transcoder.OPENGATE_SNAPSHOT:
		return c.runOpenGateSnapshot(ctx, pub)
	case transcoder.OPENGATE_AVAILABLE:
		// TODO: Change status of transcoder device
		logger.SInfo("OpenGate available")
	}
	return nil
}

func (c *Command) runOpenGateEvents(ctx context.Context, pub *paho.Publish) error {
	transcoderId := c.topic.SenderId

	return c.actorPool.Send(transcoder.TranscoderEventMessage{
		Type:         transcoder.OPENGATE_EVENTS,
		TranscoderId: transcoderId,
		Payload:      pub.Payload,
	})
}

func (c *Command) runOpenGateSnapshot(ctx context.Context, pub *paho.Publish) error {
	transcoderId := c.topic.SenderId
	eventId := c.topic.ExtraIds[0]

	pl := transcoder.OpenGateSnapshotPayload{
		EventId:  eventId,
		RawImage: pub.Payload,
	}
	msg, _ := json.Marshal(pl)

	return c.actorPool.Send(transcoder.TranscoderEventMessage{
		Type:         transcoder.OPENGATE_SNAPSHOT,
		TranscoderId: transcoderId,
		Payload:      msg,
	})
}

func (c *Command) runOpenGateStats(ctx context.Context, pub *paho.Publish) error {
	transcoderId := c.topic.SenderId
	return c.actorPool.Send(transcoder.TranscoderEventMessage{
		Type:         transcoder.OPENGATE_STATS,
		TranscoderId: transcoderId,
		Payload:      pub.Payload,
	})
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
