package eventsapi

import (
	"context"
	"strings"

	"github.com/CE-Thesis-2023/backend/src/helper/transcoder"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
)

var (
	TYPE_OPENGATE   = "opengate"
	TYPE_TRANSCODER = "transcoder"
)

type Command struct {
	Type     string `json:"type"`
	ClientId string `json:"clientId"`
	Action   string `json:"action"`

	transcoderHandler transcoder.TranscoderEventProcessor
}

func CommandFromPath(path string, transcoderHandler transcoder.TranscoderEventProcessor) (*Command, error) {
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return nil, custerror.FormatInvalidArgument("path is empty")
	}
	if len(parts) > 3 {
		return nil, custerror.FormatInvalidArgument("path is too long")
	}
	if len(parts) < 3 {
		return nil, custerror.FormatInvalidArgument("path is too short")
	}
	return &Command{
		Type:              parts[0],
		ClientId:          parts[1],
		Action:            parts[2],
		transcoderHandler: transcoderHandler,
	}, nil
}

func (c *Command) Run(ctx context.Context) error {
	switch c.Type {
	case TYPE_OPENGATE:
		return c.runOpenGate(ctx)
	case TYPE_TRANSCODER:
		return c.runTranscoder(ctx)
	default:
		return custerror.FormatInvalidArgument("unknown type: %s", c.Type)
	}
}

const (
	OPENGATE_AVAILABLE = "available"
	OPENGATE_EVENTS    = "events"
)

// https://docs.frigate.video/integrations/mqtt
func (c *Command) runOpenGate(ctx context.Context) error {
	switch c.Action {
	case OPENGATE_AVAILABLE:
		return c.transcoderHandler.OpenGateAvailable(ctx, c.ClientId)
	case OPENGATE_EVENTS:
		return c.transcoderHandler.OpenGateEvent(ctx, c.ClientId)
	}
	return nil
}

func (c *Command) runTranscoder(ctx context.Context) error {
	return nil
}
