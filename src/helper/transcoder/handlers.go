package transcoder

import (
	"context"
	"encoding/json"

	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/events"
	"github.com/CE-Thesis-2023/backend/src/models/web"
	"go.uber.org/zap"
)

func (p *transcoderEventProcessor) OpenGateAvailable(ctx context.Context, openGateId string, message []byte) error {
	payload := string(message)
	logger.SDebug("processor.OpenGateAvailable",
		zap.String("openGateId", openGateId),
		zap.String("message", payload))
	switch payload {
	case "online":
		logger.SInfo("OpenGate is online",
			zap.String("openGateId", openGateId))
	case "offline":
		logger.SInfo("OpenGate is offline",
			zap.String("openGateId", openGateId))
	}
	return nil
}

func (p *transcoderEventProcessor) OpenGateObjectTrackingEvent(ctx context.Context, openGateId string, message []byte) error {
	logger.SDebug("processor.OpenGateEvent",
		zap.String("openGateId", openGateId))

	var detectionEvent events.DetectionEvent
	if err := json.Unmarshal(message, &detectionEvent); err != nil {
		logger.SError("failed to unmarshal detection event",
			zap.Error(err))
		return err
	}

	switch detectionEvent.Type {
	case "new":
		logger.SInfo("new detection",
			zap.String("openGateId", openGateId))
		if err := p.addObjectTrackingEvent(ctx, &detectionEvent); err != nil {
			logger.SError("failed to add event to database",
				zap.Error(err))
			return err
		}
	case "update":
		logger.SInfo("detection update",
			zap.String("openGateId", openGateId))
		if err := p.updateObjectTrackingEventInDatabase(ctx, &detectionEvent); err != nil {
			logger.SError("failed to update event in database",
				zap.Error(err))
			return err
		}
	case "end":
		logger.SInfo("detection end",
			zap.String("openGateId", openGateId))
		if err := p.updateObjectTrackingEventInDatabase(ctx, &detectionEvent); err != nil {
			logger.SError("failed to update event in database",
				zap.Error(err))
			return err
		}
	}

	logger.SDebug("detection event",
		zap.Any("before", detectionEvent.Before),
		zap.Any("after", detectionEvent.After))
	return nil
}

func (p *transcoderEventProcessor) addObjectTrackingEvent(ctx context.Context, req *events.DetectionEvent) error {
	logger.SDebug("processor.addObjectTrackingEvent", zap.Reflect("req", req))

	resp, err := p.webService.GetObjectTrackingEventById(ctx, &web.GetObjectTrackingEventByIdRequest{
		OpenGateEventId: []string{req.Before.ID},
	})
	if err != nil {
		return err
	}

	if len(resp.ObjectTrackingEvents) > 0 {
		return custerror.FormatAlreadyExists("event already exists")
	}

	addResp, err := p.privateService.AddObjectTrackingEvent(ctx, &web.AddObjectTrackingEventRequest{
		Event: req,
	})
	if err != nil {
		return err
	}

	logger.SInfo("event added",
		zap.String("eventId", addResp.EventId))
	return nil
}

func (p *transcoderEventProcessor) updateObjectTrackingEventInDatabase(ctx context.Context, req *events.DetectionEvent) error {
	logger.SDebug("processor.updateEventInDatabase", zap.Reflect("req", req))

	resp, err := p.webService.GetObjectTrackingEventById(ctx, &web.GetObjectTrackingEventByIdRequest{
		OpenGateEventId: []string{req.Before.ID},
	})
	if err != nil {
		return err
	}

	if len(resp.ObjectTrackingEvents) == 0 {
		return custerror.FormatNotFound("event not found")
	}

	err = p.privateService.UpdateObjectTrackingEvent(ctx, &web.UpdateObjectTrackingEventRequest{
		EventId: resp.ObjectTrackingEvents[0].EventId,
		Event:   req,
	})
	if err != nil {
		return err
	}

	logger.SInfo("event updated",
		zap.String("openGateEventId", req.Before.ID))
	return nil
}
