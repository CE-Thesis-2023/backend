package eventsapi

import (
	"context"
	"time"

	"github.com/CE-Thesis-2023/backend/src/helper"
	"github.com/CE-Thesis-2023/backend/src/helper/transcoder"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	custmqtt "github.com/CE-Thesis-2023/backend/src/internal/mqtt"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"go.uber.org/zap"
)

func Register(cm *autopaho.ConnectionManager, connack *paho.Connack) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	subs := makeSubscriptions()
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

func makeSubscriptions() []paho.SubscribeOptions {
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
		registerTranscoderTopics(router)
	}
}

func WrapForHandlers(handler func(p *paho.Publish) error) func(p *paho.Publish) {
	return func(p *paho.Publish) {
		if err := handler(p); err != nil {
			helper.EventHandlerErrorHandler(err)
		}
	}
}

func registerTranscoderTopics(router paho.Router) {
	actors := transcoder.NewTranscoderActorsPool()
	transcoderHandler := transcoder.NewTranscoderEventProcessor(actors)

	router.RegisterHandler("opengate/#", func(p *paho.Publish) {
		ctx, cancel := context.WithTimeout(
			context.Background(), time.Second*5)
		defer cancel()
		cmd, err := CommandFromPath(p.Topic, transcoderHandler)
		if err != nil {
			logger.SError("unable to parse command from path",
				zap.Error(err),
				zap.String("topic", p.Topic))
		}
		if err := cmd.Run(ctx, p); err != nil {
			logger.SError("unable to run command",
				zap.Error(err),
				zap.Any("command", cmd))
		}
	})

	router.RegisterHandler("transcoder/#", func(p *paho.Publish) {
		ctx, cancel := context.WithTimeout(
			context.Background(), time.Second*5)
		defer cancel()
		cmd, err := CommandFromPath(p.Topic, transcoderHandler)
		if err != nil {
			logger.SError("unable to parse command from path",
				zap.Error(err),
				zap.String("topic", p.Topic))
		}
		if err := cmd.Run(ctx, p); err != nil {
			logger.SError("unable to run command",
				zap.Error(err),
				zap.Any("command", cmd))
		}
	})
}
