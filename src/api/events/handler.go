package eventsapi

import (
	"github.com/CE-Thesis-2023/backend/src/internal/cache"

	"github.com/dgraph-io/ristretto"
	"github.com/eclipse/paho.golang/paho"
	"github.com/panjf2000/ants/v2"
	"github.com/CE-Thesis-2023/backend/src/internal/concurrent"
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
	return nil
}
