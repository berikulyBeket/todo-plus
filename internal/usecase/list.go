package usecase

import (
	"context"

	"github.com/berikulyBeket/todo-plus/internal/entity"
	"github.com/berikulyBeket/todo-plus/internal/repository"
	"github.com/berikulyBeket/todo-plus/pkg/logger"
	messagebroker "github.com/berikulyBeket/todo-plus/pkg/message_broker"
	"github.com/berikulyBeket/todo-plus/pkg/search"
)

// ListUseCase handles the business logic related to lists
type ListUseCase struct {
	repo           repository.List
	search         search.List
	brokerProducer messagebroker.Producer
	logger         logger.Interface
}

// NewListUseCase creates a new instance of ListUseCase
func NewListUseCase(r repository.List, s search.List, p messagebroker.Producer, l logger.Interface) *ListUseCase {
	return &ListUseCase{
		repo:           r,
		search:         s,
		brokerProducer: p,
		logger:         l,
	}
}

// Create creates a new list for a user and publishes a list created event
func (uc *ListUseCase) Create(ctx context.Context, userId int, list *entity.List) (int, error) {
	listId, err := uc.repo.CreateUserList(ctx, userId, list)
	if err != nil {
		return 0, err
	}

	go uc.brokerProducer.PublishListCreatedEvent(userId, list)

	return listId, nil
}

// GetAll retrieves all lists for a given user
func (uc *ListUseCase) GetAll(ctx context.Context, userId int) ([]entity.List, error) {
	return uc.repo.GetAllUserLists(ctx, userId)
}

// GetOneById retrieves a list by its ID if the user is the owner
func (uc *ListUseCase) GetOneById(ctx context.Context, userId, listId int) (entity.List, error) {
	err := uc.repo.IsUserOwnerOfList(ctx, userId, listId)
	if err != nil {
		return entity.List{}, err
	}

	return uc.repo.GetOneById(ctx, listId)
}

// UpdateOneById updates a list's details if the user is the owner and publishes a list updated event
func (uc *ListUseCase) UpdateOneById(ctx context.Context, userId, listId int, newTodoInput entity.UpdateListInput) error {
	if err := uc.repo.IsUserOwnerOfList(ctx, userId, listId); err != nil {
		return err
	}

	err := uc.repo.UpdateOneById(ctx, &userId, listId, newTodoInput)
	if err != nil {
		return err
	}

	go uc.brokerProducer.PublishListUpdatedEvent(userId, listId)

	return nil
}

// DeleteOneById deletes a list if the user is the owner and publishes a list deleted event
func (uc *ListUseCase) DeleteOneById(ctx context.Context, userId, listId int) error {
	if err := uc.repo.IsUserOwnerOfList(ctx, userId, listId); err != nil {
		return err
	}

	err := uc.repo.DeleteOneById(ctx, &userId, listId)
	if err != nil {
		return err
	}

	go uc.brokerProducer.PublishListDeletedEvent(listId)

	return nil
}

// DeleteOneByAdmin deletes a list by an admin and publishes a list deleted event
func (uc *ListUseCase) DeleteOneByAdmin(ctx context.Context, listId int) error {
	err := uc.repo.DeleteOneById(ctx, nil, listId)
	if err != nil {
		return err
	}

	go uc.brokerProducer.PublishListDeletedEvent(listId)

	return nil
}

// Search performs a search for lists based on the search text for a given user
func (uc *ListUseCase) Search(ctx context.Context, userId int, searchText string) ([]entity.List, error) {
	listIds, err := uc.search.SearchIds(ctx, userId, searchText)
	if err != nil {
		return []entity.List{}, err
	}

	return uc.repo.GetManyByIds(ctx, listIds)
}

// HandleCreated handles the list created event by indexing the list in the search service
func (uc *ListUseCase) HandleCreated(ctx context.Context, message entity.ListCreatedEvent) error {
	uc.logger.Info("handling list created event in use case layer")

	if err := uc.search.Index(ctx, message.UserId, &message.List); err != nil {
		uc.logger.Errorf("failed to index document in Elasticsearch: %v", err)
		return err
	}

	return nil
}

// HandleUpdated handles the list updated event by re-indexing the list in the search service
func (uc *ListUseCase) HandleUpdated(ctx context.Context, message entity.ListUpdatedEvent) error {
	uc.logger.Info("handling list updated event in use case layer")

	list, err := uc.repo.GetOneById(ctx, message.ListId)
	if err != nil {
		uc.logger.Errorf("error fetching updated list: %v", err)
		return err
	}

	if err := uc.search.Index(ctx, message.UserId, &list); err != nil {
		uc.logger.Errorf("failed to index document in Elasticsearch: %v", err)
		return err
	}

	return nil
}

// HandleDeleted handles the list deleted event by removing the list from the search service
func (uc *ListUseCase) HandleDeleted(ctx context.Context, message entity.ListDeletedEvent) error {
	uc.logger.Info("handling list deleted event in use case layer")

	if err := uc.search.Delete(ctx, message.ListId); err != nil {
		uc.logger.Errorf("failed to delete document in Elasticsearch: %v", err)
		return err
	}

	return nil
}
