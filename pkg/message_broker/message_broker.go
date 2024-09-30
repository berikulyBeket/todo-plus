package messagebroker

import "github.com/berikulyBeket/todo-plus/internal/entity"

// Producer defines the methods for publishing events related to lists and items
type Producer interface {
	PublishListCreatedEvent(userId int, list *entity.List)
	PublishListUpdatedEvent(userId, listId int)
	PublishListDeletedEvent(listId int)
	PublishItemCreatedEvent(userId, listId int, item *entity.Item)
	PublishItemUpdatedEvent(userId, listId, itemId int)
	PublishItemDeletedEvent(itemId int)
}

// Consumer defines the method for subscribing to Kafka topics
type Consumer interface {
	Subscribe(topic string, handler func(message []byte) error) error
}
