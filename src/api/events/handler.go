package eventsapi

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/CE-Thesis-2023/backend/src/biz/service"
	"github.com/CE-Thesis-2023/backend/src/helper/transcoder"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
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
		return c.runOpenGateUpdateStatus(ctx, transcoder.OPENGATE_AVAILABLE, pub)
	case transcoder.OPENGATE_STATE_AUDIO:
		return c.runOpenGateUpdateStatus(ctx, transcoder.OPENGATE_STATE_AUDIO, pub)
	case transcoder.OPENGATE_STATE_SNAPSHOTS:
		return c.runOpenGateUpdateStatus(ctx, transcoder.OPENGATE_STATE_SNAPSHOTS, pub)
	case transcoder.OPENGATE_STATE_MOTION:
		return c.runOpenGateUpdateStatus(ctx, transcoder.OPENGATE_STATE_MOTION, pub)
	case transcoder.OPENGATE_STATE_DETECT:
		return c.runOpenGateUpdateStatus(ctx, transcoder.OPENGATE_STATE_DETECT, pub)
	case transcoder.OPENGATE_STATE_RECORDINGS:
		return c.runOpenGateUpdateStatus(ctx, transcoder.OPENGATE_STATE_RECORDINGS, pub)
	case transcoder.OPENGATE_STATE_PTZ_AUTOTRACKER:
		return c.runOpenGateUpdateStatus(ctx, transcoder.OPENGATE_STATE_PTZ_AUTOTRACKER, pub)
	case transcoder.OPENGATE_STATE_IMPROVE_CONTRAST:
		return c.runOpenGateUpdateStatus(ctx, transcoder.OPENGATE_STATE_IMPROVE_CONTRAST, pub)
	case transcoder.OPENGATE_STATE_BIRDSEYE:
		return c.runOpenGateUpdateStatus(ctx, transcoder.OPENGATE_STATE_BIRDSEYE, pub)
	}
	return nil
}

func (c *Command) runOpenGateEvents(_ context.Context, pub *paho.Publish) error {
	transcoderId := c.topic.SenderId

	return c.actorPool.Send(transcoder.TranscoderEventMessage{
		Type:         transcoder.OPENGATE_EVENTS,
		TranscoderId: transcoderId,
		Payload:      pub.Payload,
	})
}

func (c *Command) runOpenGateSnapshot(_ context.Context, pub *paho.Publish) error {
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

func (c *Command) runOpenGateStats(_ context.Context, pub *paho.Publish) error {
	transcoderId := c.topic.SenderId
	return c.actorPool.Send(transcoder.TranscoderEventMessage{
		Type:         transcoder.OPENGATE_STATS,
		TranscoderId: transcoderId,
		Payload:      pub.Payload,
	})
}

func (c *Command) runOpenGateUpdateStatus(_ context.Context, t string, pub *paho.Publish) error {
	transcoderId := c.topic.SenderId
	return c.actorPool.Send(transcoder.TranscoderEventMessage{
		Type:         t,
		TranscoderId: transcoderId,
		Payload:      pub.Payload,
	})
}
