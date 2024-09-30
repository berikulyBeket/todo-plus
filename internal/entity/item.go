package entity

import "github.com/berikulyBeket/todo-plus/utils"

// Item represents a task or item in a to-do list
type Item struct {
	Id          int    `json:"id"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

// ItemCreatedEvent represents the event data when an item is created
type ItemCreatedEvent struct {
	UserId int  `json:"user_id"`
	ListId int  `json:"list_id"`
	Item   Item `json:"item"`
}

// ItemUpdatedEvent represents the event data when an item is updated
type ItemUpdatedEvent struct {
	UserId int `json:"user_id"`
	ListId int `json:"list_id"`
	ItemId int `json:"item_id"`
}

// ItemDeletedEvent represents the event data when an item is deleted
type ItemDeletedEvent struct {
	ItemId int `json:"item_id"`
}

// UpdateItemInput represents the input for updating an item
type UpdateItemInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Done        *bool   `json:"done"`
}

// Validate checks if the update input has at least one valid field
func (i UpdateItemInput) Validate() error {
	if i.Title == nil && i.Description == nil && i.Done == nil {
		return utils.ErrItemEmptyRequest
	}

	return nil
}
