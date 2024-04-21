package transcoder

import (
	"context"
	"time"

	"github.com/CE-Thesis-2023/backend/src/biz/service"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/anthdm/hollywood/actor"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
)

type TranscoderEventProcessor interface {
	OpenGateAvailable(ctx context.Context, transcoderId string, message []byte) error
	OpenGateObjectTrackingEvent(ctx context.Context, transcoderId string, message []byte) error
	OpenGateSnapshot(ctx context.Context, transcoderId string, message []byte) error
	OpenGateStats(ctx context.Context, transcoderId string, message []byte) error
}

type transcoderEventProcessor struct {
	privateService *service.PrivateService
	webService     *service.WebService
}

func NewTranscoderEventProcessor(privateService *service.PrivateService, webService *service.WebService) TranscoderEventProcessor {
	return &transcoderEventProcessor{
		privateService: privateService,
		webService:     webService,
	}
}

type TranscoderActorsPool struct {
	engine         *actor.Engine
	privateService *service.PrivateService
	webService     *service.WebService
}

func NewTranscoderActorsPool(privateService *service.PrivateService, webService *service.WebService) *TranscoderActorsPool {
	engine, err := actor.NewEngine(&actor.EngineConfig{})
	if err != nil {
		logger.SFatal("unable to create actor engine",
			zap.Error(err))
	}
	return &TranscoderActorsPool{
		engine:         engine,
		privateService: privateService,
		webService:     webService,
	}
}

func (p *TranscoderActorsPool) Exists(transcoderId string) bool {
	pid := p.engine.Registry.GetPID("transcoder", transcoderId)
	return pid != nil
}

func (p *TranscoderActorsPool) Allocate(transcoderId string) *actor.PID {
	existingPid := p.engine.Registry.GetPID("transcoder", transcoderId)
	if existingPid != nil {
		return existingPid
	}
	pid := p.engine.Spawn(func() actor.Receiver {
		return NewTranscoderActor(p.privateService, p.webService)
	},
		"transcoder",
		actor.WithID(transcoderId),
		actor.WithInboxSize(10))
	return pid
}

func (p *TranscoderActorsPool) Send(message TranscoderEventMessage) error {
	pid := p.Allocate(message.TranscoderId)
	p.engine.Send(pid, message)
	return nil
}

func (p *TranscoderActorsPool) Deallocate(cameraGroupId string, transcoderId string, finished chan bool) error {
	pid := p.engine.
		Registry.
		GetPID("transcoder", transcoderId)
	if pid == nil {
		return custerror.FormatNotFound("transcoder actor not found")
	}
	wg := p.engine.Poison(pid)
	go func() {
		wg.Wait()
		finished <- true
	}()
	return nil
}

type TranscoderActor struct {
	handler TranscoderEventProcessor
}

func NewTranscoderActor(privateService *service.PrivateService, webService *service.WebService) actor.Receiver {
	return &TranscoderActor{
		handler: NewTranscoderEventProcessor(
			privateService,
			webService),
	}
}

func (a *TranscoderActor) Receive(ctx *actor.Context) {
	logger.SDebug("TranscoderActor received message",
		zap.String("pid", ctx.PID().String()))

	message := ctx.Message()
	var event TranscoderEventMessage
	if err := copier.Copy(&event, message); err != nil {
		logger.SError("unable to copy message",
			zap.Error(err))
		return
	}

	timeOutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	payload := event.Payload

	switch event.Type {
	case OPENGATE_EVENTS:
		logger.SInfo("TranscoderActor received opengate event",
			zap.String("payload", string(payload)),
			zap.String("transcoderId", event.TranscoderId))
		if err := a.handler.OpenGateObjectTrackingEvent(timeOutCtx, event.TranscoderId, payload); err != nil {
			logger.SError("unable to process opengate event",
				zap.Error(err))
		}
	case OPENGATE_SNAPSHOT:
		logger.SInfo("TranscoderActor received opengate snapshot",
			zap.String("transcoderId", event.TranscoderId))
		if err := a.handler.OpenGateSnapshot(timeOutCtx, event.TranscoderId, payload); err != nil {
			logger.SError("unable to process opengate snapshot",
				zap.Error(err))
		}
	case OPENGATE_STATS:
		logger.SInfo("TranscoderActor received opengate stats",
			zap.String("transcoderId", event.TranscoderId))
		if err := a.handler.OpenGateStats(timeOutCtx, event.TranscoderId, payload); err != nil {
			logger.SError("unable to process opengate stats",
				zap.Error(err))
		}
	default:
		logger.SError("unknown event type",
			zap.String("type", event.Type))
	}
}
