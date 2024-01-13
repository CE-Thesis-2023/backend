package handlers

import (
	"sync"
	"time"

	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
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
	options    *CommunicatorOptions
	inputChan  map[string]chan interface{}
	outputChan map[string]chan interface{}
}

func NewWebSocketCommunicator(options ...CommunicatorOptioner) *WebSocketCommunicator {
	opts := &CommunicatorOptions{}
	for _, o := range options {
		o(opts)
	}
	return &WebSocketCommunicator{
		options:    opts,
		inputChan:  map[string]chan interface{}{},
		outputChan: map[string]chan interface{}{},
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

		inputChannel := make(chan interface{}, comm.options.channelSize)
		outputChannel := make(chan interface{}, comm.options.channelSize)

		comm.inputChan[id] = inputChannel
		comm.outputChan[id] = outputChannel
		defer func() {
			delete(comm.inputChan, id)
			close(inputChannel)
			delete(comm.outputChan, id)
			close(outputChannel)
			logger.SDebug("WebsocketHandler: closed input and output channels",
				zap.String("id", id))
		}()

		logger.SInfo("WebsocketHandler: connection ready",
			zap.String("id", id))

		if comm.options.handler != nil {
			if err := comm.options.handler(&ConnectionInformation{
				Id:         id,
				Version:    version,
				Connection: c,
				// message that comes is the output to other functions
				IncommingMessageChan: outputChannel,

				// message that needs to be sent comes from other functions
				OutgoingMessageChan: inputChannel,
			}); err != nil {
				logger.SError("CreateWebsocketHandler.handler: error", zap.Error(err))
				return
			}
		}
	})
}

func (c *WebSocketCommunicator) Input(id string, msg interface{}) error {
	channel, found := c.inputChan[id]
	if !found {
		return custerror.FormatNotFound("channel for given ID not found")
	}
	channel <- msg
	return nil
}

func (c *WebSocketCommunicator) WaitForOutput(id string, dest interface{}, timeout time.Duration) error {
	channel, found := c.outputChan[id]
	if !found {
		return custerror.FormatNotFound("channel for given ID not found")
	}
	select {
	case msg := <-channel:
		if err := copier.Copy(dest, msg); err != nil {
			logger.SError("WaitForOutput: copy message error", zap.Error(err))
			return err
		}
	case <-time.After(timeout):
		logger.SDebug("WaitForOutput: timeout exceeded")
		return custerror.ErrorTimeout
	}
	return nil
}

type ConnectionInformation struct {
	Id                   string `json:"id"`
	Version              string `json:"version"`
	IncommingMessageChan chan interface{}
	OutgoingMessageChan  chan interface{}
	Connection           *websocket.Conn
}

type ConnectionHandler func(i *ConnectionInformation) error

type AuthorizerHandler func(ctx *fiber.Ctx) error
