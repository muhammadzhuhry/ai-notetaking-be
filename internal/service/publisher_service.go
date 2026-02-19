package service

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

type IPublisherService interface {
	Publish(ctx context.Context, payload []byte) error
}

// Concrete implementation of the IPublisherService interface
type publisherService struct {
	pubSub    *gochannel.GoChannel
	topicName string
}

// Factory function to create a new instance of PublisherService
func NewPublisherService(topicName string, pubSub *gochannel.GoChannel) IPublisherService {
	return &publisherService{
		pubSub:    pubSub,
		topicName: topicName,
	}
}

func (ps *publisherService) Publish(ctx context.Context, payload []byte) error {
	err := ps.pubSub.Publish(ps.topicName, message.NewMessage(watermill.NewUUID(), payload))
	if err != nil {
		return err
	}

	return nil
}
