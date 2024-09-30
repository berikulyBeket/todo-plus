package entity

import "github.com/berikulyBeket/todo-plus/utils"

// List represents a to-do list
type List struct {
	Id          int    `json:"id"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

// ListCreatedEvent represents the event data when a list is created
type ListCreatedEvent struct {
	UserId int  `json:"user_id"`
	List   List `json:"list"`
}

// ListUpdatedEvent represents the event data when a list is updated
type ListUpdatedEvent struct {
	UserId int `json:"user_id"`
	ListId int `json:"list_id"`
}

// ListDeletedEvent represents the event data when a list is deleted
type ListDeletedEvent struct {
	ListId int `json:"list_id"`
}

// UpdateListInput represents the input for updating a list
type UpdateListInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

// Validate checks if the update input has at least one valid field
func (i UpdateListInput) Validate() error {
	if i.Title == nil && i.Description == nil {
		return utils.ErrListEmptyRequest
	}

	return nil
}
