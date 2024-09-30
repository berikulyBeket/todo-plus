package repository

import (
	"context"

	"github.com/berikulyBeket/todo-plus/internal/entity"
	"github.com/berikulyBeket/todo-plus/pkg/cache"
	"github.com/berikulyBeket/todo-plus/pkg/database"
	"github.com/berikulyBeket/todo-plus/pkg/logger"
)

type Auth interface {
	CreateUser(ctx context.Context, user entity.User) (int, error)
	GetUserByCredentials(ctx context.Context, username string, password string) (entity.User, error)
}

type User interface {
	DeleteOneById(ctx context.Context, userId int) error
}

type List interface {
	CreateUserList(ctx context.Context, userId int, list *entity.List) (int, error)
	GetAllUserLists(ctx context.Context, userId int) ([]entity.List, error)
	IsUserOwnerOfList(ctx context.Context, userId, listId int) error
	GetOneById(ctx context.Context, listId int) (entity.List, error)
	GetManyByIds(ctx context.Context, ids []int) ([]entity.List, error)
	UpdateOneById(ctx context.Context, userId *int, listId int, newTodoInput entity.UpdateListInput) error
	DeleteOneById(ctx context.Context, userId *int, listId int) error
}

type Item interface {
	CreateListItem(ctx context.Context, listId int, item *entity.Item) (int, error)
	GetAllListItems(ctx context.Context, listId int) ([]entity.Item, error)
	IsUserOwnerOfItem(ctx context.Context, userId, itemId int) error
	GetOneById(ctx context.Context, itemId int) (entity.Item, error)
	GetManyByIds(ctx context.Context, itemIds []int) ([]entity.Item, error)
	UpdateOneById(ctx context.Context, listId *int, itemId int, input entity.UpdateItemInput) error
	DeleteOneById(ctx context.Context, listId *int, itemId int) error
}

type Repository struct {
	Auth
	User
	List
	Item
}

// NewRepository creates a new instance of Repository with initialized repositories
func NewRepository(
	db *database.Database,
	cache *cache.Cache,
	logger logger.Interface,
) *Repository {
	return &Repository{
		Auth: NewAuthRepo(db),
		User: NewUserRepo(db),
		List: NewListRepo(db, cache, logger),
		Item: NewItemRepo(db, cache, logger),
	}
}
