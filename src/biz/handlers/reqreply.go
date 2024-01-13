package handlers

import (
	"context"
	"sync"

	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/events"
	ltdEvent "github.com/CE-Thesis-2023/ltd/src/models/events"
	"github.com/gofiber/contrib/websocket"
	"go.uber.org/zap"
)

type requestReplyContext struct {
	Request  *MessageWithId
	Response *events.CommandResponse
	Done     chan bool
	Error    error
}

type MessageWithId struct {
	MessageId uint64                   `json:"messageId"`
	Request   *ltdEvent.CommandRequest `json:"commandRequest"`
}

type RequestReplyCommunicator struct {
	sync.Mutex
	deviceId string
	counter  uint64
	pending  map[uint64]*requestReplyContext
	conn     *websocket.Conn
}

type RequestReplyResponse struct {
	MessageId uint64                  `json:"messageId"`
	Response  *events.CommandResponse `json:"commandResponse"`
}

func (c *RequestReplyCommunicator) OnResponse(resp *RequestReplyResponse) error {
	c.Lock()
	rrCtx, found := c.pending[resp.MessageId]
	delete(c.pending, resp.MessageId)
	c.Unlock()

	if !found {
		logger.SError("RRCommunicator.OnResponse: no pending request with ID found")
		return custerror.FormatFailedPrecondition("no pending request with ID found")
	}

	rrCtx.Response = resp.Response
	rrCtx.Done <- true
	return nil
}

func (c *RequestReplyCommunicator) Request(ctx context.Context, req *ltdEvent.CommandRequest) (*events.CommandResponse, error) {
	c.Lock()
	id := c.counter
	c.counter += 1
	msg := &MessageWithId{
		MessageId: id,
		Request:   req,
	}
	rrCtx := &requestReplyContext{
		Request:  msg,
		Response: &events.CommandResponse{},
		Done:     make(chan bool, 1),
		Error:    nil,
	}
	c.pending[id] = rrCtx
	if err := c.conn.WriteJSON(&msg); err != nil {
		logger.SError("RRCommunicator.Request: error",
			zap.Error(err),
			zap.String("deviceId", c.deviceId))
		delete(c.pending, id)
		c.Unlock()
		return nil, err
	}
	c.Unlock()

	select {
	case <-rrCtx.Done:
		logger.SDebug("RRCommunicator.Request: response received",
			zap.String("deviceId", c.deviceId),
			zap.Uint64("messageId", id))
	case <-ctx.Done():
		logger.SDebug("RRCommunicator.Request: context canceled")
		rrCtx.Error = ctx.Err()
	}

	if rrCtx.Error != nil {
		logger.SError("RRCommunicator.Request: response error",
			zap.Error(rrCtx.Error))
		return nil, rrCtx.Error
	}

	return rrCtx.Response, nil
}
