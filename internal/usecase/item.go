package usecase

import (
	"context"

	"github.com/berikulyBeket/todo-plus/internal/entity"
	"github.com/berikulyBeket/todo-plus/internal/repository"
	"github.com/berikulyBeket/todo-plus/pkg/logger"
	messagebroker "github.com/berikulyBeket/todo-plus/pkg/message_broker"
	"github.com/berikulyBeket/todo-plus/pkg/search"
)

// ItemUseCase manages the item-related use cases
type ItemUseCase struct {
	repo           repository.Item
	listRepo       repository.List
	search         search.Item
	brokerProducer messagebroker.Producer
	logger         logger.Interface
}

// NewItemUseCase creates a new instance of ItemUseCase
func NewItemUseCase(
	r repository.Item,
	l repository.List,
	s search.Item,
	p messagebroker.Producer,
	logger logger.Interface,
) *ItemUseCase {
	return &ItemUseCase{
		repo:           r,
		listRepo:       l,
		search:         s,
		brokerProducer: p,
		logger:         logger,
	}
}

// Create creates a new item in a list if the user is the owner, and publishes a created event
func (uc *ItemUseCase) Create(ctx context.Context, userId, listId int, item *entity.Item) (int, error) {
	if err := uc.listRepo.IsUserOwnerOfList(ctx, userId, listId); err != nil {
		return 0, err
	}

	itemId, err := uc.repo.CreateListItem(ctx, listId, item)
	if err != nil {
		return 0, err
	}

	go uc.brokerProducer.PublishItemCreatedEvent(userId, listId, item)

	return itemId, nil
}

// GetAll retrieves all items from a list if the user is the owner
func (uc *ItemUseCase) GetAll(ctx context.Context, userId, listId int) ([]entity.Item, error) {
	if err := uc.listRepo.IsUserOwnerOfList(ctx, userId, listId); err != nil {
		return make([]entity.Item, 0), err
	}

	return uc.repo.GetAllListItems(ctx, listId)
}

// GetOneById retrieves a single item by its ID if the user is the owner
func (uc *ItemUseCase) GetOneById(ctx context.Context, userId, itemId int) (entity.Item, error) {
	if err := uc.repo.IsUserOwnerOfItem(ctx, userId, itemId); err != nil {
		return entity.Item{}, err
	}

	return uc.repo.GetOneById(ctx, itemId)
}

// UpdateOneById updates an item by its ID if the user is the owner, and publishes an updated event
func (uc *ItemUseCase) UpdateOneById(ctx context.Context, userId int, listId int, itemId int, input entity.UpdateItemInput) error {
	if err := uc.repo.IsUserOwnerOfItem(ctx, userId, itemId); err != nil {
		return err
	}

	if err := uc.repo.UpdateOneById(ctx, &listId, itemId, input); err != nil {
		return err
	}

	go uc.brokerProducer.PublishItemUpdatedEvent(userId, listId, itemId)

	return nil
}

// DeleteOneById deletes an item by its ID if the user is the owner, and publishes a deleted event
func (uc *ItemUseCase) DeleteOneById(ctx context.Context, userId int, listId int, itemId int) error {
	if err := uc.repo.IsUserOwnerOfItem(ctx, userId, itemId); err != nil {
		return err
	}

	if err := uc.repo.DeleteOneById(ctx, &listId, itemId); err != nil {
		return err
	}

	go uc.brokerProducer.PublishItemDeletedEvent(itemId)

	return nil
}

// DeleteOneByAdmin deletes an item by its ID as an admin, and publishes a deleted event
func (uc *ItemUseCase) DeleteOneByAdmin(ctx context.Context, itemId int) error {
	if err := uc.repo.DeleteOneById(ctx, nil, itemId); err != nil {
		return err
	}

	go uc.brokerProducer.PublishItemDeletedEvent(itemId)

	return nil
}

// Search performs a search for items based on various filters and search text
func (uc *ItemUseCase) Search(ctx context.Context, userId int, listId *int, done *bool, searchText string) ([]entity.Item, error) {
	itemIds, err := uc.search.SearchIds(ctx, userId, listId, done, searchText)
	if err != nil {
		return []entity.Item{}, err
	}

	return uc.repo.GetManyByIds(ctx, itemIds)
}

// HandleCreated handles the event when a new item is created by indexing it in the search service
func (uc *ItemUseCase) HandleCreated(ctx context.Context, message entity.ItemCreatedEvent) error {
	if err := uc.search.Index(ctx, message.UserId, message.ListId, message.Item); err != nil {
		uc.logger.Errorf("failed to index document in Elasticsearch: %v", err)
		return err
	}

	return nil
}

// HandleUpdated handles the event when an item is updated by indexing it in the search service
func (uc *ItemUseCase) HandleUpdated(ctx context.Context, message entity.ItemUpdatedEvent) error {
	item, err := uc.repo.GetOneById(ctx, message.ItemId)
	if err != nil {
		uc.logger.Errorf("error fetching updated item: %v", err)
		return err
	}

	if err := uc.search.Index(ctx, message.UserId, message.ListId, item); err != nil {
		uc.logger.Errorf("failed to index document in Elasticsearch: %v", err)
		return err
	}

	return nil
}

// HandleDeleted handles the event when an item is deleted by removing it from the search service
func (uc *ItemUseCase) HandleDeleted(ctx context.Context, message entity.ItemDeletedEvent) error {
	if err := uc.search.Delete(ctx, message.ItemId); err != nil {
		uc.logger.Errorf("failed to delete document in Elasticsearch: %v", err)
		return err
	}

	return nil
}
