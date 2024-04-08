package async

import (
	"context"
	"time"

	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/eclipse/paho.golang/paho"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type asyncRequest struct {
	client        *paho.Client
	publishTopic  string
	responseTopic string
	msg           []byte
	timeout       time.Duration
}

func NewAsyncRequest(
	client *paho.Client,
	topic string,
	msg []byte,
	timeout time.Duration,
) *asyncRequest {
	return &asyncRequest{}
}

func (a *asyncRequest) Do(ctx context.Context, resp interface{}) error {
	correlationId := uuid.
		New().
		String()
	publish := paho.Publish{
		Topic:   a.publishTopic,
		QoS:     1,
		Payload: a.msg,
		Properties: &paho.PublishProperties{
			CorrelationData: []byte(correlationId),
			ContentType:     "application/json",
			ResponseTopic:   a.responseTopic,
		},
	}
	subscribe := paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{
				Topic: a.responseTopic,
				QoS:   1,
			},
		},
		Properties: &paho.SubscribeProperties{},
	}
	if _, err := a.client.Subscribe(ctx, &subscribe); err != nil {
		return err
	}
	defer func() {
		if _, err := a.client.Unsubscribe(ctx, &paho.Unsubscribe{
			Topics: []string{a.responseTopic},
		}); err != nil {
			logger.SError("unsubscribe error", zap.Error(err))
		}
	}()

	if _, err := a.client.Publish(ctx, &publish); err != nil {
		return err
	}

	// TODO: Finish

	return nil
}
