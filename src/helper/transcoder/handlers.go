package transcoder

import (
	"context"

	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"go.uber.org/zap"
)

func (p *transcoderEventProcessor) OpenGateAvailable(ctx context.Context, openGateId string) error {
	logger.SDebug("processor.OpenGateAvailable",
		zap.String("openGateId", openGateId))
	return nil
}

func (p *transcoderEventProcessor) OpenGateEvent(ctx context.Context, openGateId string) error {
	logger.SDebug("processor.OpenGateEvent",
		zap.String("openGateId", openGateId))
	return nil
}
