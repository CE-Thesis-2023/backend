package transcoder

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/CE-Thesis-2023/backend/src/biz/service"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/web"
	"github.com/anthdm/hollywood/actor"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
)

type TranscoderEventProcessor interface {
	OpenGateAvailable(ctx context.Context, transcoderId string, message []byte) error
	OpenGateObjectTrackingEvent(ctx context.Context, transcoderId string, message []byte) error
	OpenGateSnapshot(ctx context.Context, transcoderId string, message []byte) error
	OpenGateStats(ctx context.Context, transcoderId string, message []byte) error
	TranscoderStatus(ctx context.Context, transcoderId string, status *web.UpdateTranscoderStatusRequest) error
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
	handler     TranscoderEventProcessor
	mu          sync.RWMutex
	backoff     *time.Timer
	statusModel *web.UpdateTranscoderStatusRequest
}

func NewTranscoderActor(privateService *service.PrivateService, webService *service.WebService) actor.Receiver {
	return &TranscoderActor{
		handler: NewTranscoderEventProcessor(
			privateService,
			webService),
		statusModel: nil,
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
	case OPENGATE_AVAILABLE:
		logger.SInfo("TranscoderActor received opengate available",
			zap.String("transcoderId", event.TranscoderId))
		cn := ""
		if event.CameraName != nil {
			cn = *event.CameraName
		}
		a.updateTranscoderStatusModel(OPENGATE_AVAILABLE, event.TranscoderId, cn, payload)
	default:
		if strings.HasSuffix(event.Type, "/state") {
			logger.SInfo("TranscoderActor received opengate state",
				zap.String("transcoderId", event.TranscoderId),
				zap.String("type", event.Type))
			cn := ""
			if event.CameraName != nil {
				cn = *event.CameraName
			}
			a.updateTranscoderStatusModel(event.Type, event.TranscoderId, cn, payload)
			return
		}
		logger.SError("unknown event type",
			zap.String("type", event.Type))
	}
}

func (a *TranscoderActor) waitForStatusUpdate() {
	defer func() {
		a.mu.Lock()
		defer a.mu.Unlock()
		a.statusModel = nil
		a.backoff.Stop()
		a.backoff = nil
	}()
	<-a.backoff.C
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.statusModel != nil {
		if err := a.handler.TranscoderStatus(context.Background(), a.statusModel.TranscoderId, a.statusModel); err != nil {
			logger.SError("flush transcoder status failed",
				zap.Error(err))
		}
	}
}

func (a *TranscoderActor) updateTranscoderStatusModel(kind string, transcoderId string, cameraName string, msg []byte) {
	e := msgToEnabled(msg)
	a.mu.Lock()
	if a.statusModel == nil {
		a.statusModel = &web.UpdateTranscoderStatusRequest{
			TranscoderId: transcoderId,
			CameraName:   &cameraName,
		}
		// flush the statusModel to the database
		// after 10 seconds, if no new status update
		// on status update, reset this timer
		a.backoff = time.NewTimer(10 * time.Second)
		go a.waitForStatusUpdate()
	} else {
		a.backoff.Reset(10 * time.Second)
	}
	switch kind {
	case OPENGATE_STATE_DETECT:
		a.statusModel.ObjectDetection = &e
	case OPENGATE_STATE_AUDIO:
		a.statusModel.AudioDetection = &e
	case OPENGATE_STATE_RECORDINGS:
		a.statusModel.OpenGateRecordings = &e
	case OPENGATE_STATE_SNAPSHOTS:
		a.statusModel.Snapshots = &e
	case OPENGATE_STATE_MOTION:
		a.statusModel.MotionDetection = &e
	case OPENGATE_STATE_IMPROVE_CONTRAST:
		a.statusModel.ImproveContrast = &e
	case OPENGATE_STATE_PTZ_AUTOTRACKER:
		a.statusModel.Autotracker = &e
	case OPENGATE_STATE_BIRDSEYE:
		a.statusModel.BirdseyeView = &e
	case OPENGATE_AVAILABLE:
		a.statusModel.OpenGateStatus = &e
	}
	a.mu.Unlock()
}

func msgToEnabled(msg []byte) bool {
	str := string(msg)
	switch str {
	case "ON":
		return true
	case "OFF":
		return false
	default:
		return false
	}
}
