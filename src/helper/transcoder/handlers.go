package transcoder

import (
	"context"
	"encoding/json"
	"sync"

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

func (p *transcoderEventProcessor) OpenGateObjectTrackingEvent(ctx context.Context, transcoderId string, message []byte) error {
	var detectionEvent events.DetectionEvent
	if err := json.Unmarshal(message, &detectionEvent); err != nil {
		logger.SError("failed to unmarshal detection event",
			zap.Error(err))
		return err
	}

	switch detectionEvent.Type {
	case "new":
		logger.SInfo("new detection",
			zap.String("transcoderId", transcoderId))
		if err := p.addObjectTrackingEvent(ctx, transcoderId, &detectionEvent); err != nil {
			logger.SError("failed to add event to database",
				zap.Error(err))
			return err
		}
	case "update":
		logger.SInfo("detection update",
			zap.String("transcoderId", transcoderId))
		if err := p.updateObjectTrackingEventInDatabase(ctx, transcoderId, &detectionEvent); err != nil {
			logger.SError("failed to update event in database",
				zap.Error(err))
			return err
		}
	case "end":
		logger.SInfo("detection end",
			zap.String("transcoderId", transcoderId))
		if err := p.updateObjectTrackingEventInDatabase(ctx, transcoderId, &detectionEvent); err != nil {
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

func (p *transcoderEventProcessor) addObjectTrackingEvent(ctx context.Context, transcoderId string, req *events.DetectionEvent) error {
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
		TranscoderId: transcoderId,
		Event:        req,
	})
	if err != nil {
		return err
	}

	logger.SInfo("event added",
		zap.String("eventId", addResp.EventId))
	return nil
}

func (p *transcoderEventProcessor) updateObjectTrackingEventInDatabase(ctx context.Context, transcoderId string, req *events.DetectionEvent) error {
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
		EventId:      resp.ObjectTrackingEvents[0].EventId,
		TranscoderId: transcoderId,
		Event:        req,
	})
	if err != nil {
		return err
	}

	logger.SInfo("event updated",
		zap.String("openGateEventId", req.Before.ID))
	return nil
}

func (p *transcoderEventProcessor) OpenGateSnapshot(ctx context.Context, transcoderId string, message []byte) error {
	var snapshotPayload OpenGateSnapshotPayload
	if err := json.Unmarshal(message, &snapshotPayload); err != nil {
		logger.SError("failed to unmarshal snapshot payload",
			zap.Error(err))
		return err
	}

	logger.SDebug("processor.OpenGateSnapshot",
		zap.String("transcoderId", transcoderId),
		zap.String("eventId", snapshotPayload.EventId))

	err := p.webService.UpsertSnapshot(ctx, &web.UpsertSnapshotRequest{
		OpenGateEventId: snapshotPayload.EventId,
		RawImage:        string(snapshotPayload.RawImage),
		TranscoderId:    transcoderId,
	})
	if err != nil {
		return err
	}

	logger.SInfo("snapshot added",
		zap.String("eventId", snapshotPayload.EventId))
	return nil
}

func (p *transcoderEventProcessor) OpenGateStats(ctx context.Context, transcoderId string, message []byte) error {
	var statStruct events.OpenGateStats
	if err := json.Unmarshal(message, &statStruct); err != nil {
		return nil
	}

	var cameraStatsError error
	var detectorStatsError error

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for cameraName, cameraStats := range statStruct.Cameras {
			_, err := p.webService.UpsertOpenGateCameraStats(ctx, &web.UpsertOpenGateCameraStatsRequest{
				TranscoderId: transcoderId,
				CameraName:   cameraName,
				CameraFPS:    cameraStats.CameraFPS,
				DetectionFPS: cameraStats.DetectionFPS,
				CapturePID:   cameraStats.CapturePID,
				ProcessID:    cameraStats.PID,
				ProcessFPS:   cameraStats.ProcessFPS,
				SkippedFPS:   cameraStats.SkippedFPS,
			})
			if err != nil {
				cameraStatsError = err
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		for detectorName, detectorStats := range statStruct.Detectors {
			_, err := p.webService.UpsertOpenGateDetectorStats(ctx, &web.UpsertOpenGateDetectorsStatsRequest{
				TranscoderId:   transcoderId,
				DetectorName:   detectorName,
				DetectorStart:  detectorStats.DetectionStart,
				InferenceSpeed: detectorStats.InferenceSpeed,
				ProcessID:      detectorStats.PID,
			})

			if err != nil {
				detectorStatsError = err
				return
			}
		}
	}()

	wg.Wait()
	if cameraStatsError != nil {
		return custerror.FormatInternalError("failed to add camera stats: %s", cameraStatsError)
	}
	if detectorStatsError != nil {
		return custerror.FormatInternalError("failed to add detector stats: %s", detectorStatsError)
	}

	logger.SInfo("OpenGate stats added",
		zap.String("transcoderId", transcoderId))
	return nil
}
