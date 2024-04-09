package custmqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/events"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"go.uber.org/zap"
)

type MQTTSession struct {
	client  *autopaho.ConnectionManager
	configs *configs.EventStoreConfigs

	// maps request id with the reply channel
	currReqReplySession map[uint64]chan []byte

	incrementalId uint64
	mu            sync.Mutex
}

func NewMQTTSession(
	ctx context.Context,
	configs *configs.EventStoreConfigs) (*MQTTSession, error) {
	s := &MQTTSession{
		client:              nil,
		configs:             configs,
		currReqReplySession: make(map[uint64]chan []byte),
		incrementalId:       0,
	}
	if err := s.init(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *MQTTSession) init(ctx context.Context) error {
	client, err := NewClient(ctx,
		WithClientGlobalConfigs(s.configs),
		WithHandlerRegister(s.registerReplyTopics),
		WithOnReconnection(s.handleTopicSubscribe))
	if err != nil {
		return err
	}
	s.client = client
	return nil
}

func (s *MQTTSession) registerReplyTopics(r *paho.StandardRouter) {
	r.RegisterHandler("reply/#", s.handleReplyMessages)
}

func (s *MQTTSession) handleReplyMessages(p *paho.Publish) {
	id := p.Properties.CorrelationData
	strId := string(id)
	intId, err := strconv.ParseUint(strId, 10, 64)
	if err != nil {
		logger.SError("failed to parse correlation data",
			zap.String("correlation_data", strId),
			zap.Error(err))
		return
	}
	s.mu.Lock()
	channel, found := s.currReqReplySession[intId]
	if !found {
		logger.SWarn("reply channel not found",
			zap.Uint64("id", intId))
		return
	}
	channel <- p.Payload
	s.mu.Unlock()
}

func (s *MQTTSession) handleTopicSubscribe(cm *autopaho.ConnectionManager, _ *paho.Connack) {
	topics := []paho.SubscribeOptions{
		{Topic: "reply/#", QoS: 1},
	}
	logger.SInfo("subscribing to reply topics",
		zap.Reflect("topics", topics))
	if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
		Subscriptions: topics,
	}); err != nil {
		logger.SFatal("failed to subscribe to reply topic",
			zap.Error(err))
	}
}

type RequestReplyRequest struct {
	Topic      events.Event
	Request    interface{}
	Reply      interface{}
	MaxTimeout time.Duration
}

func (s *MQTTSession) Request(ctx context.Context, r *RequestReplyRequest) error {
	payload, err := json.Marshal(r.Request)
	if err != nil {
		return err
	}
	reply := s.toReplyTopic(r.Topic)
	replyTopic := reply.Topic()
	requestTopic := r.Topic.Topic()

	id := s.allocateChannel()
	defer s.deallocateChannel(id)

	publishment := &paho.Publish{
		Topic:   requestTopic,
		QoS:     1,
		Payload: payload,
		Properties: &paho.PublishProperties{
			ContentType:     "application/json",
			ResponseTopic:   replyTopic,
			CorrelationData: []byte(fmt.Sprintf("%d", id)),
		},
	}

	if _, err := s.client.Publish(ctx, publishment); err != nil {
		return err
	}

	c, found := s.currReqReplySession[id]
	if !found {
		return custerror.FormatInternalError("request reply session not found")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(r.MaxTimeout):
		return custerror.FormatTimeout("request timeout")
	case msg := <-c:
		if err := json.Unmarshal(msg, r.Reply); err != nil {
			return err
		}
	}

	return nil
}

func (s *MQTTSession) deallocateChannel(id uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	channel, ok := s.currReqReplySession[id]
	if !ok {
		return
	}
	close(channel)
	delete(s.currReqReplySession, id)
}

func (s *MQTTSession) allocateChannel() (id uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id = s.incrementalId
	s.incrementalId += 1
	s.currReqReplySession[id] = make(chan []byte)
	return id
}

func (s *MQTTSession) toReplyTopic(topic events.Event) events.Event {
	topic.Prefix = fmt.Sprintf("reply/%s", topic.Prefix)
	return topic
}
