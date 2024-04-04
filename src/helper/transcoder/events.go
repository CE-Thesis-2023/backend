package transcoder

import (
	"context"
	"sync"

	custcon "github.com/CE-Thesis-2023/backend/src/internal/concurrent"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/anthdm/hollywood/actor"
	"go.uber.org/zap"
)

type TranscoderEventProcessor interface {
	OpenGateAvailable(ctx context.Context, openGateId string) error
	OpenGateEvent(ctx context.Context, openGateId string) error
}

type transcoderEventProcessor struct {
	actors *TranscoderActorsPool
}

func NewTranscoderEventProcessor(actors *TranscoderActorsPool) TranscoderEventProcessor {
	return &transcoderEventProcessor{
		actors: actors,
	}
}

type TranscoderActorsPool struct {
	mu             sync.Mutex
	groupEngineMap map[string]*actor.Engine
}

func NewTranscoderActorsPool() *TranscoderActorsPool {
	return &TranscoderActorsPool{
		groupEngineMap: map[string]*actor.Engine{},
	}
}

func (p *TranscoderActorsPool) Exists(cameraGroupId string, TranscoderId string) bool {
	engine, found := p.groupEngineMap[cameraGroupId]
	if found {
		pid := engine.Registry.
			GetPID("transcoder", TranscoderId)
		return pid != nil
	}
	return found
}

func (p *TranscoderActorsPool) Allocate(cameraGroupId string, TranscoderId string) (*actor.PID, error) {
	var engine *actor.Engine
	var found bool
	var err error
	engine, found = p.groupEngineMap[cameraGroupId]
	if found {
		pid := engine.Registry.GetPID("transcoder", TranscoderId)
		if pid != nil {
			return pid, nil
		}
	} else {
		engine, err = actor.NewEngine(&actor.EngineConfig{})
		if err != nil {
			return nil, custerror.FormatInternalError("unable to allocation camera group engine: %s", err)
		}
		p.mu.Lock()
		p.groupEngineMap[cameraGroupId] = engine
		p.mu.Unlock()
	}
	pid := engine.Spawn(newTranscoderActor,
		"transcoder",
		actor.WithID(TranscoderId),
		actor.WithInboxSize(10))
	return pid, nil
}

func (p *TranscoderActorsPool) Deallocate(cameraGroupId string, TranscoderId string, finished chan bool) error {
	engine, found := p.groupEngineMap[cameraGroupId]
	if !found {
		return custerror.FormatNotFound("camera group engine not found")
	}
	pid := engine.
		Registry.
		GetPID("transcoder", TranscoderId)
	if pid == nil {
		return custerror.FormatNotFound("Transcoder actor not found")
	}
	wg := engine.Poison(pid)
	custcon.Do(func() error {
		wg.Wait()
		finished <- true
		return nil
	})
	return nil
}

type TranscoderActor struct {
}

func newTranscoderActor() actor.Receiver {
	return &TranscoderActor{}
}

func (a *TranscoderActor) Receive(ctx *actor.Context) {
	logger.SDebug("TranscoderActor received message",
		zap.String("pid", ctx.PID().String()),
		zap.Any("message", ctx.Message()))

}
