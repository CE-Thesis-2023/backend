package handlers

import (
	"sync"
	"time"

	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var once sync.Once

var communicator *WebSocketCommunicator

func WebsocketInit(options ...CommunicatorOptioner) {
	once.Do(func() {
		communicator = NewWebSocketCommunicator(options...)
	})
}

func GetWebsocketCommunicator() *WebSocketCommunicator {
	return communicator
}

type WebSocketCommunicator struct {
	options                   *CommunicatorOptions
	requestReplyCommunicators map[string]*RequestReplyCommunicator
}

func NewWebSocketCommunicator(options ...CommunicatorOptioner) *WebSocketCommunicator {
	opts := &CommunicatorOptions{}
	for _, o := range options {
		o(opts)
	}
	return &WebSocketCommunicator{
		options:                   opts,
		requestReplyCommunicators: map[string]*RequestReplyCommunicator{},
	}
}

type CommunicatorOptions struct {
	handler     ConnectionHandler
	authorizer  AuthorizerHandler
	channelSize int
}

type CommunicatorOptioner func(o *CommunicatorOptions)

func WithConnectionHandler(handler ConnectionHandler) CommunicatorOptioner {
	return func(o *CommunicatorOptions) {
		o.handler = handler
	}
}

func WithAuthorizer(handler AuthorizerHandler) CommunicatorOptioner {
	return func(o *CommunicatorOptions) {
		o.authorizer = handler
	}
}

func WithChannelSize(size int) CommunicatorOptioner {
	return func(o *CommunicatorOptions) {
		o.channelSize = size
	}
}

func (c *WebSocketCommunicator) HandleRegisterRequest(ctx *fiber.Ctx) error {
	logger.SDebug("RegisterWebSocketHandlers: register request")
	if !websocket.IsWebSocketUpgrade(ctx) {
		logger.SError("RegisterWebSocketHandlers: upgrade rejected")
		return fiber.ErrUpgradeRequired
	}
	if c.options.authorizer != nil {
		if err := c.options.authorizer(ctx); err != nil {
			logger.SError("RegisterWebSocketHandlers: authorizer", zap.Error(err))
			return err
		}
	}
	ctx.Locals("allowed", true)
	return ctx.Next()
}

func (comm *WebSocketCommunicator) CreateWebsocketHandler() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		id := c.Params("id")
		version := c.Params("v")
		defer c.Close()

		c.EnableWriteCompression(true)

		logger.SDebug("WebsocketHandler: connection started",
			zap.String("ip", c.IP()))

		logger.SInfo("WebsocketHandler: connection ready",
			zap.String("id", id))

		if comm.options.handler != nil {
			rr := &RequestReplyCommunicator{
				deviceId: id,
				counter:  1, // messageId 0 is to signal that response or request is not request-reply
				pending:  map[uint64]*requestReplyContext{},
				conn:     c,
			}
			comm.requestReplyCommunicators[id] = rr
			ci := &ConnectionInformation{
				Id:           id,
				Version:      version,
				Connection:   c,
				RequestReply: rr,
			}

			c.SetCloseHandler(comm.createCloseHandler(ci))
			if err := comm.options.handler(ci); err != nil {
				logger.SError("CreateWebsocketHandler.handler: error", zap.Error(err))
				return
			}

			delete(comm.requestReplyCommunicators, id)
			logger.SInfo("WebsocketHandler: closed", zap.String("id", id))
		}
	}, websocket.Config{
		HandshakeTimeout:  2 * time.Second,
		EnableCompression: true,
	})
}

func (comm *WebSocketCommunicator) createCloseHandler(i *ConnectionInformation) func(code int, text string) error {
	return func(code int, text string) error {
		delete(comm.requestReplyCommunicators, i.Id)
		logger.SInfo("CloseHandler: closed request reply channel")

		switch code {
		case websocket.CloseNoStatusReceived:
			logger.SDebug("CloseHandler: closed by peer",
				zap.String("text", text),
				zap.String("id", i.Id))
			return nil
		default:
			logger.SDebug("CloseHandler: closed by other reasons",
				zap.Int("code", code),
				zap.String("text", text))
			return nil
		}
	}
}

func (comm *WebSocketCommunicator) RequestReply(deviceId string) (*RequestReplyCommunicator, error) {
	c, found := comm.requestReplyCommunicators[deviceId]
	if !found {
		return nil, custerror.FormatFailedPrecondition("transcoder device not connected")
	}
	return c, nil
}

type ConnectionInformation struct {
	Id           string `json:"id"`
	Version      string `json:"version"`
	Connection   *websocket.Conn
	RequestReply *RequestReplyCommunicator
}

type ConnectionHandler func(i *ConnectionInformation) error

type AuthorizerHandler func(ctx *fiber.Ctx) error
