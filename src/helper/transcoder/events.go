package transcoder

import (
	"context"
	"sync"

	"github.com/CE-Thesis-2023/backend/src/biz/service"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/anthdm/hollywood/actor"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
)

type TranscoderEventProcessor interface {
	OpenGateAvailable(ctx context.Context, openGateId string, message []byte) error
	OpenGateEvent(ctx context.Context, openGateId string, message []byte) error
}

type transcoderEventProcessor struct {
	commandService *service.PrivateService
}

func NewTranscoderEventProcessor(commandService *service.PrivateService) TranscoderEventProcessor {
	return &transcoderEventProcessor{
		commandService: commandService,
	}
}

type TranscoderActorsPool struct {
	mu     sync.Mutex
	engine *actor.Engine
}

func NewTranscoderActorsPool() *TranscoderActorsPool {
	return &TranscoderActorsPool{
		engine: &actor.Engine{},
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
	pid := p.engine.Spawn(newTranscoderActor,
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

func newTranscoderActor() actor.Receiver {
	return &TranscoderActor{
		handler: NewTranscoderEventProcessor(service.GetPrivateService()),
	}
}

func (a *TranscoderActor) Receive(ctx *actor.Context) {
	logger.SDebug("TranscoderActor received message",
		zap.String("pid", ctx.PID().String()),
		zap.Any("message", ctx.Message()))
	message := ctx.Message()
	var event TranscoderEventMessage
	if err := copier.Copy(&event, message); err != nil {
		logger.SError("unable to copy message",
			zap.Error(err))
		return
	}
	payload := event.Payload
	switch event.Type {
	case "opengate":
		logger.SInfo("TranscoderActor received opengate event",
			zap.String("openGateId", event.OpenGateId),
			zap.Any("payload", payload))
	case "transcoder":
		logger.SInfo("TranscoderActor received transcoder event",
			zap.String("transcoderId", event.TranscoderId),
			zap.Any("payload", payload))
	}
}
