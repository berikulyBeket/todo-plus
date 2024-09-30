package usecase

import (
	"context"

	"github.com/berikulyBeket/todo-plus/internal/entity"
	"github.com/berikulyBeket/todo-plus/internal/repository"
	"github.com/berikulyBeket/todo-plus/pkg/hash"
	"github.com/berikulyBeket/todo-plus/pkg/logger"
	messagebroker "github.com/berikulyBeket/todo-plus/pkg/message_broker"
	"github.com/berikulyBeket/todo-plus/pkg/search"
	"github.com/berikulyBeket/todo-plus/pkg/token"
)

type Auth interface {
	CreateUser(ctx context.Context, user entity.User) (int, error)
	AuthenticateUser(ctx context.Context, username, password string) (entity.User, error)
	GenerateToken(ctx context.Context, userId int) (string, error)
	ParseToken(ctx context.Context, token string) (int, error)
}

type User interface {
	DeleteOneByAdmin(ctx context.Context, userId int) error
}

type List interface {
	Create(ctx context.Context, userId int, list *entity.List) (int, error)
	GetAll(ctx context.Context, userId int) ([]entity.List, error)
	GetOneById(ctx context.Context, userId, listId int) (entity.List, error)
	UpdateOneById(ctx context.Context, userId, listId int, newTodoInput entity.UpdateListInput) error
	DeleteOneById(ctx context.Context, userId, listId int) error
	DeleteOneByAdmin(ctx context.Context, listId int) error
	Search(ctx context.Context, userId int, searchText string) ([]entity.List, error)
	HandleCreated(ctx context.Context, message entity.ListCreatedEvent) error
	HandleUpdated(ctx context.Context, message entity.ListUpdatedEvent) error
	HandleDeleted(ctx context.Context, message entity.ListDeletedEvent) error
}

type Item interface {
	Create(ctx context.Context, userId, listId int, item *entity.Item) (int, error)
	GetAll(ctx context.Context, userId, listId int) ([]entity.Item, error)
	GetOneById(ctx context.Context, userId, itemId int) (entity.Item, error)
	UpdateOneById(ctx context.Context, userId, listId, itemId int, input entity.UpdateItemInput) error
	DeleteOneById(ctx context.Context, userId, listId, itemId int) error
	DeleteOneByAdmin(ctx context.Context, itemId int) error
	Search(ctx context.Context, userId int, listId *int, done *bool, searchText string) ([]entity.Item, error)
	HandleCreated(ctx context.Context, message entity.ItemCreatedEvent) error
	HandleUpdated(ctx context.Context, message entity.ItemUpdatedEvent) error
	HandleDeleted(ctx context.Context, message entity.ItemDeletedEvent) error
}

type UseCase struct {
	Auth
	User
	List
	Item
}

// NewUseCase creates a new instance of UseCase with initialized use cases
func NewUseCase(
	repos *repository.Repository,
	searchService *search.SearchServices,
	brokerProducer messagebroker.Producer,
	hasher hash.Hasher,
	tokenMaker token.TokenMaker,
	logger logger.Interface,
) *UseCase {
	return &UseCase{
		Auth: NewAuthUseCase(repos.Auth, hasher, tokenMaker),
		User: NewUserUseCase(repos.User),
		List: NewListUseCase(repos.List, searchService.List, brokerProducer, logger),
		Item: NewItemUseCase(repos.Item, repos.List, searchService.Item, brokerProducer, logger),
	}
}
