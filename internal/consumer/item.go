package consumer

import (
	"context"
	"encoding/json"

	"github.com/berikulyBeket/todo-plus/internal/entity"
)

// handleItemCreated processes the item created event
func (c *Consumer) handleItemCreated(msgBytes []byte) error {
	var message entity.ItemCreatedEvent
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		c.logger.Errorf("Failed to unmarshal message: %v", err)
		return err
	}

	if err := c.usecases.Item.HandleCreated(context.Background(), message); err != nil {
		c.logger.Errorf("Error handling item created event: %v", err)
		return err
	}

	return nil
}

// handleItemUpdated processes the item updated event
func (c *Consumer) handleItemUpdated(msgBytes []byte) error {
	var message entity.ItemUpdatedEvent
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		c.logger.Errorf("Failed to unmarshal message: %v", err)
		return err
	}

	if err := c.usecases.Item.HandleUpdated(context.Background(), message); err != nil {
		c.logger.Errorf("Error handling item updated event: %v", err)
		return err
	}

	return nil
}

// handleItemDeleted processes the item deleted event
func (c *Consumer) handleItemDeleted(msgBytes []byte) error {
	var message entity.ItemDeletedEvent
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		c.logger.Errorf("Failed to unmarshal message: %v", err)
		return err
	}

	if err := c.usecases.Item.HandleDeleted(context.Background(), message); err != nil {
		c.logger.Errorf("Error handling item deleted event: %v", err)
		return err
	}

	return nil
}
