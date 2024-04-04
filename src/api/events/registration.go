package eventsapi

import (
	"context"
	"time"

	"github.com/CE-Thesis-2023/backend/src/helper"
	custactors "github.com/CE-Thesis-2023/backend/src/internal/actor"
	custcon "github.com/CE-Thesis-2023/backend/src/internal/concurrent"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	custmqtt "github.com/CE-Thesis-2023/backend/src/internal/mqtt"
	"github.com/anthdm/hollywood/actor"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"go.uber.org/zap"
)

func Register(cm *autopaho.ConnectionManager, connack *paho.Connack) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	subs := makeSubscriptions(ctx, cm, connack)
	if _, err := cm.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: subs,
	}); err != nil {
		logger.SError("unable to make MQTT subscriptions",
			zap.String("where", "api.events.Register"),
			zap.Any("subs", subs),
		)
		return
	}

	logger.SInfo("MQTT subscriptions made success", zap.Any("subs", subs))
}

func makeSubscriptions(ctx context.Context, cm *autopaho.ConnectionManager, connack *paho.Connack) []paho.SubscribeOptions {
	return []paho.SubscribeOptions{
		{Topic: "updates/#", QoS: 1},
		{Topic: "opengate/#", QoS: 1},
	}
}

func ClientErrorHandler(err error) {
	logger := logger.Logger()

	logger.Error("MQTT Client", zap.Error(err))
}

func DisconnectHandler(d *paho.Disconnect) {
	logger := logger.Logger()

	logger.Error("MQTT Server Disconnect", zap.String("reason", d.Properties.ReasonString))
}

func RouterHandler() custmqtt.RouterRegister {
	return func(router *paho.StandardRouter) {
		eh := GetStandardEventsHandler()

		router.RegisterHandler("updates/#",
			WrapForHandlers(eh.UpdateEventsHandler))

		registerOpenGateHandlers(router)
	}
}

func WrapForHandlers(handler func(p *paho.Publish) error) func(p *paho.Publish) {
	return func(p *paho.Publish) {
		if err := handler(p); err != nil {
			helper.EventHandlerErrorHandler(err)
		}
	}
}

func registerOpenGateHandlers(router paho.Router) {
	// https://docs.frigate.video/integrations/mqtt
	router.RegisterHandler("opengate/#", func(p *paho.Publish) {
	})
}
