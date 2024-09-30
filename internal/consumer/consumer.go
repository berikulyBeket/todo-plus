package consumer

import (
	"fmt"

	"github.com/berikulyBeket/todo-plus/internal/usecase"
	"github.com/berikulyBeket/todo-plus/pkg/logger"
	messagebroker "github.com/berikulyBeket/todo-plus/pkg/message_broker"
)

// Consumer handles subscriptions to message topics
type Consumer struct {
	usecases       *usecase.UseCase
	brokerConsumer messagebroker.Consumer
	logger         logger.Interface
}

// New creates a new Consumer instance
func New(
	brokerConsumer messagebroker.Consumer,
	usecases *usecase.UseCase,
	logger logger.Interface,
) *Consumer {
	return &Consumer{
		usecases:       usecases,
		brokerConsumer: brokerConsumer,
		logger:         logger,
	}
}

// Init subscribes the consumer to various topics
func (c *Consumer) Init() error {
	if err := c.brokerConsumer.Subscribe(messagebroker.ListCreatedTopic, c.handleListCreated); err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", messagebroker.ListCreatedTopic, err)
	}

	if err := c.brokerConsumer.Subscribe(messagebroker.ListUpdatedTopic, c.handleListUpdated); err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", messagebroker.ListUpdatedTopic, err)
	}

	if err := c.brokerConsumer.Subscribe(messagebroker.ListDeletedTopic, c.handleListDeleted); err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", messagebroker.ListDeletedTopic, err)
	}

	if err := c.brokerConsumer.Subscribe(messagebroker.ItemCreatedTopic, c.handleItemCreated); err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", messagebroker.ItemCreatedTopic, err)
	}

	if err := c.brokerConsumer.Subscribe(messagebroker.ItemUpdatedTopic, c.handleItemUpdated); err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", messagebroker.ItemUpdatedTopic, err)
	}

	if err := c.brokerConsumer.Subscribe(messagebroker.ItemDeletedTopic, c.handleItemDeleted); err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", messagebroker.ItemDeletedTopic, err)
	}

	return nil
}
