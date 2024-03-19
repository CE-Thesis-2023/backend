package eventsapi

import (
	"github.com/CE-Thesis-2023/backend/src/internal/cache"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"go.uber.org/zap"

	"github.com/CE-Thesis-2023/backend/src/internal/concurrent"
	"github.com/dgraph-io/ristretto"
	"github.com/eclipse/paho.golang/paho"
	"github.com/panjf2000/ants/v2"
)

type StandardEventHandler struct {
	pool  *ants.Pool
	cache *ristretto.Cache
}

func NewStandardEventHandler() *StandardEventHandler {
	return &StandardEventHandler{
		pool:  custcon.New(100),
		cache: cache.Cache(),
	}
}

func (h *StandardEventHandler) UpdateEventsHandler(p *paho.Publish) error {
	logger.SDebug("eventsapi.StandardEventHandler.UpdateEventsHandler",
		zap.String("topic", p.Topic),
		zap.Any("message", string(p.Payload)),
	)
	return nil
}
