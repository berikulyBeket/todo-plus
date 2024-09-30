package consumer

import (
	"context"
	"encoding/json"

	"github.com/berikulyBeket/todo-plus/internal/entity"
)

// handleListCreated processes the list created event
func (c *Consumer) handleListCreated(msgBytes []byte) error {
	var message entity.ListCreatedEvent
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		c.logger.Errorf("Failed to unmarshal message: %v", err)
		return err
	}

	if err := c.usecases.List.HandleCreated(context.Background(), message); err != nil {
		c.logger.Errorf("Error handling list created event: %v", err)
		return err
	}

	return nil
}

// handleListUpdated processes the list updated event
func (c *Consumer) handleListUpdated(msgBytes []byte) error {
	var message entity.ListUpdatedEvent
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		c.logger.Errorf("Failed to unmarshal message: %v", err)
		return err
	}

	if err := c.usecases.List.HandleUpdated(context.Background(), message); err != nil {
		c.logger.Errorf("Error handling list updated event: %v", err)
		return err
	}

	return nil
}

// handleListDeleted processes the list deleted event
func (c *Consumer) handleListDeleted(msgBytes []byte) error {
	var message entity.ListDeletedEvent
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		c.logger.Errorf("Failed to unmarshal message: %v", err)
		return err
	}

	if err := c.usecases.List.HandleDeleted(context.Background(), message); err != nil {
		c.logger.Errorf("Error handling list deleted event: %v", err)
		return err
	}

	return nil
}
