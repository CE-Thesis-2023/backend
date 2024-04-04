package opengate

import (
	"sync"

	custcon "github.com/CE-Thesis-2023/backend/src/internal/concurrent"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/anthdm/hollywood/actor"
	"go.uber.org/zap"
)

type OpenGateEventProcessor interface {
}

type OpenGateActorsPool struct {
	mu             sync.Mutex
	groupEngineMap map[string]*actor.Engine
}

func (p *OpenGateActorsPool) NewOpenGateActorsPool() *OpenGateActorsPool {
	return &OpenGateActorsPool{
		groupEngineMap: map[string]*actor.Engine{},
	}
}

func (p *OpenGateActorsPool) Exists(cameraGroupId string, openGateId string) bool {
	engine, found := p.groupEngineMap[cameraGroupId]
	if found {
		pid := engine.Registry.
			GetPID("opengate", openGateId)
		return pid != nil
	}
	return found
}

func (p *OpenGateActorsPool) Allocate(cameraGroupId string, openGateId string) (*actor.PID, error) {
	var engine *actor.Engine
	var found bool
	var err error
	engine, found = p.groupEngineMap[cameraGroupId]
	if found {
		pid := engine.Registry.GetPID("opengate", openGateId)
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
	pid := engine.Spawn(newOpenGateActor,
		"opengate",
		actor.WithID(openGateId),
		actor.WithInboxSize(10))
	return pid, nil
}

func (p *OpenGateActorsPool) Deallocate(cameraGroupId string, openGateId string, finished chan bool) error {
	engine, found := p.groupEngineMap[cameraGroupId]
	if !found {
		return custerror.FormatNotFound("camera group engine not found")
	}
	pid := engine.
		Registry.
		GetPID("opengate", openGateId)
	if pid == nil {
		return custerror.FormatNotFound("opengate actor not found")
	}
	wg := engine.Poison(pid)
	custcon.Do(func() error {
		wg.Wait()
		finished <- true
		return nil
	})
	return nil
}

type openGateActor struct {
}

func newOpenGateActor() actor.Receiver {
	return &openGateActor{}
}

func (a *openGateActor) Receive(ctx *actor.Context) {
	logger.SDebug("OpenGateActor received message",
		zap.String("pid", ctx.PID().String()),
		zap.Any("message", ctx.Message()))

}
