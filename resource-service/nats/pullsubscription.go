package nats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/keptn/keptn/resource-service/models"
	"github.com/nats-io/nats.go"
	logger "github.com/sirupsen/logrus"
)

type PullSubscription struct {
	queueGroup     string
	topic          string
	subscription   *nats.Subscription
	ctx            context.Context
	jetStream      nats.JetStreamContext
	messageHandler func(event models.Event, sync bool) error
	isActive       bool
}

func NewPullSubscription(ctx context.Context, queueGroup, topic string, js nats.JetStreamContext, messageHandler func(event models.Event, sync bool) error) *PullSubscription {
	return &PullSubscription{
		queueGroup:     queueGroup,
		topic:          topic,
		jetStream:      js,
		ctx:            ctx,
		messageHandler: messageHandler,
	}
}

func (ps *PullSubscription) Activate() error {
	ps.isActive = true
	consumerInfo, _ := ps.jetStream.ConsumerInfo(streamName, consumerName)
	if consumerInfo == nil {
		_, err := ps.jetStream.AddConsumer(streamName, &nats.ConsumerConfig{
			AckPolicy:     nats.AckExplicitPolicy,
			FilterSubject: ps.topic,
		})
		if err != nil {
			return fmt.Errorf("failed to create nats consumer: %s", err.Error())
		}
	}

	sub, err := ps.jetStream.PullSubscribe(ps.topic, consumerName, nats.ManualAck())

	if err != nil {
		return fmt.Errorf("failed to subscribe to topic: %s", err.Error())
	}
	ps.subscription = sub
	go ps.pullMessages()
	return nil
}

func (ps *PullSubscription) pullMessages() {
	for {
		select {
		case <-ps.ctx.Done():
			ps.isActive = false
			return
		default:
		}
		msgs, err := ps.subscription.Fetch(10)
		if err != nil {
			// timeout is not a problem since that simply means that no event for that topic has been sent
			if !errors.Is(err, nats.ErrTimeout) {
				logger.WithError(err).Errorf("could not fetch messages for topic %s", ps.subscription.Subject)
			}
		}
		for _, msg := range msgs {
			if ps.processMessage(msg) {
				return
			}
		}
	}
}

func (ps *PullSubscription) processMessage(msg *nats.Msg) bool {
	event := &models.Event{}
	if err := json.Unmarshal(msg.Data, event); err != nil {
		logger.WithError(err).Error("could not unmarshal message")
		// ACK the message to avoid re-sending it
		if err := msg.Ack(); err != nil {
			logger.WithError(err).Error("could not ack message")
		}
		return true
	}
	if err := ps.messageHandler(*event, false); err != nil {
		logger.WithError(err).Error("could not process message")
	}
	if err := msg.Ack(); err != nil {
		logger.WithError(err).Error("could not ack message")
	}
	return false
}

func (ps *PullSubscription) Unsubscribe() error {
	return ps.subscription.Unsubscribe()
}
