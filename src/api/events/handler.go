package eventsapi

import (
	"context"
	"strings"

	"github.com/CE-Thesis-2023/backend/src/biz/service"
	"github.com/CE-Thesis-2023/backend/src/helper/transcoder"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/models/web"
	"github.com/eclipse/paho.golang/paho"
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
	OPENGATE_AVAILABLE = "available"
	OPENGATE_EVENTS    = "events"
)

// https://docs.frigate.video/integrations/mqtt
func (c *Command) runOpenGate(ctx context.Context, pub *paho.Publish) error {
	resp, err := c.webService.GetCamerasByOpenGateId(ctx, &web.GetCameraByOpenGateIdRequest{
		OpenGateId: c.ClientId,
	})
	if err != nil {
		return err
	}
	camera := resp.Camera
	err = c.actorPool.Send(transcoder.TranscoderEventMessage{
		CameraId:     camera.CameraId,
		GroupId:      camera.GroupId,
		TranscoderId: camera.TranscoderId,
		OpenGateId:   camera.OpenGateId,
		Type:         c.Type,
		Action:       c.Action,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Command) runTranscoder(ctx context.Context, pub *paho.Publish) error {
	switch c.Action {
	default:
		return custerror.FormatInvalidArgument("unknown action: %s", c.Action)
	}
}
