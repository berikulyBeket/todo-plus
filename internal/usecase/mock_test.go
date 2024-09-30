package usecase_test

import (
	"context"

	"github.com/berikulyBeket/todo-plus/internal/entity"

	"github.com/stretchr/testify/mock"
)

// Mocking the repository
type MockAuthRepo struct {
	mock.Mock
}

// CreateUser mocks the creation of a user in the repository
func (m *MockAuthRepo) CreateUser(ctx context.Context, user entity.User) (int, error) {
	args := m.Called(ctx, user)
	return args.Int(0), args.Error(1)
}

// GetUserByCredentials mocks retrieving a user by credentials
func (m *MockAuthRepo) GetUserByCredentials(ctx context.Context, username string, password string) (entity.User, error) {
	args := m.Called(ctx, username, password)
	return args.Get(0).(entity.User), args.Error(1)
}

// Mocking the hasher
type MockHasher struct {
	mock.Mock
}

// HashPassword mocks hashing a password
func (m *MockHasher) HashPassword(password string) string {
	args := m.Called(password)
	return args.String(0)
}

// Mocking the token maker
type MockTokenMaker struct {
	mock.Mock
}

// CreateToken mocks creating a token for a user
func (m *MockTokenMaker) CreateToken(userId int) (string, error) {
	args := m.Called(userId)
	return args.String(0), args.Error(1)
}

// VerifyToken mocks verifying a token and returning the associated user Id
func (m *MockTokenMaker) VerifyToken(tokenString string) (int, error) {
	args := m.Called(tokenString)
	return args.Int(0), args.Error(1)
}

// Mocking the repository.List interface
type MockListRepo struct {
	mock.Mock
}

// Implementing the Create method
func (m *MockListRepo) CreateUserList(ctx context.Context, userId int, list *entity.List) (int, error) {
	args := m.Called(ctx, userId, list)
	return args.Int(0), args.Error(1)
}

// Implementing the GetAll method
func (m *MockListRepo) GetAllUserLists(ctx context.Context, userId int) ([]entity.List, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).([]entity.List), args.Error(1)
}

func (m *MockListRepo) IsUserOwnerOfList(ctx context.Context, userId, listId int) error {
	args := m.Called(ctx, userId, listId)
	return args.Error(0)
}

// Implementing the GetOneById method
func (m *MockListRepo) GetOneById(ctx context.Context, listId int) (entity.List, error) {
	args := m.Called(ctx, listId)
	return args.Get(0).(entity.List), args.Error(1)
}

// Implementing the UpdateOneById method
func (m *MockListRepo) UpdateOneById(ctx context.Context, userId *int, listId int, newTodoInput entity.UpdateListInput) error {
	args := m.Called(ctx, userId, listId, newTodoInput)
	return args.Error(0)
}

// Implementing the DeleteOneById method
func (m *MockListRepo) DeleteOneById(ctx context.Context, userId *int, listId int) error {
	args := m.Called(ctx, userId, listId)
	return args.Error(0)
}

// GetManyByIds mocks retrieving multiple lists by their Ids
func (m *MockListRepo) GetManyByIds(ctx context.Context, ids []int) ([]entity.List, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]entity.List), args.Error(1)
}

// Mocking the repository.Item interface
type MockItemRepo struct {
	mock.Mock
}

// CreateListItem mocks creating an item in a list
func (m *MockItemRepo) CreateListItem(ctx context.Context, listId int, item *entity.Item) (int, error) {
	args := m.Called(ctx, listId, item)
	return args.Int(0), args.Error(1)
}

// GetAllListItems mocks retrieving all items in a list
func (m *MockItemRepo) GetAllListItems(ctx context.Context, listId int) ([]entity.Item, error) {
	args := m.Called(ctx, listId)
	return args.Get(0).([]entity.Item), args.Error(1)
}

// IsUserOwnerOfItem mocks checking if a user owns a specific item
func (m *MockItemRepo) IsUserOwnerOfItem(ctx context.Context, userId, itemId int) error {
	args := m.Called(ctx, userId, itemId)
	return args.Error(0)
}

// GetOneById mocks retrieving an item by its Id
func (m *MockItemRepo) GetOneById(ctx context.Context, itemId int) (entity.Item, error) {
	args := m.Called(ctx, itemId)
	return args.Get(0).(entity.Item), args.Error(1)
}

// UpdateOneById mocks updating an item by its Id
func (m *MockItemRepo) UpdateOneById(ctx context.Context, listId *int, itemId int, input entity.UpdateItemInput) error {
	args := m.Called(ctx, listId, itemId, input)
	return args.Error(0)
}

// DeleteOneById mocks deleting an item by its Id
func (m *MockItemRepo) DeleteOneById(ctx context.Context, listId *int, itemId int) error {
	args := m.Called(ctx, listId, itemId)
	return args.Error(0)
}

// GetManyByIds mocks retrieving multiple items by their Ids
func (m *MockItemRepo) GetManyByIds(ctx context.Context, itemIds []int) ([]entity.Item, error) {
	args := m.Called(ctx, itemIds)
	return args.Get(0).([]entity.Item), args.Error(1)
}

// MockUserRepo mocks the repository.User interface for user operations
type MockUserRepo struct {
	mock.Mock
}

// GetOneById mocks retrieving a user by their Id
func (m *MockUserRepo) GetOneById(ctx context.Context, userId int) (entity.User, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).(entity.User), args.Error(1)
}

// DeleteOneById mocks deleting a user by their Id
func (m *MockUserRepo) DeleteOneById(ctx context.Context, userId int) error {
	args := m.Called(ctx, userId)
	return args.Error(0)
}

// MockListSearch mocks the search.List interface for searching lists
type MockListSearch struct {
	mock.Mock
}

// SearchIds mocks searching for list Ids based on user Id and search text
func (m *MockListSearch) SearchIds(ctx context.Context, userId int, searchText string) ([]int, error) {
	args := m.Called(ctx, userId, searchText)
	return args.Get(0).([]int), args.Error(1)
}

// Index mocks indexing a list in the search service
func (m *MockListSearch) Index(ctx context.Context, userId int, list *entity.List) error {
	args := m.Called(ctx, userId, list)
	return args.Error(0)
}

// Delete mocks deleting a list from the search service
func (m *MockListSearch) Delete(ctx context.Context, listId int) error {
	args := m.Called(ctx, listId)
	return args.Error(0)
}

// MockBrokerProducer mocks the message broker producer for publishing events
type MockBrokerProducer struct{}

// PublishListCreatedEvent mocks publishing a list created event
func (m *MockBrokerProducer) PublishListCreatedEvent(userId int, list *entity.List) {}

// PublishListUpdatedEvent mocks publishing a list updated event
func (m *MockBrokerProducer) PublishListUpdatedEvent(userId, listId int) {}

// PublishListDeletedEvent mocks publishing a list deleted event
func (m *MockBrokerProducer) PublishListDeletedEvent(listId int) {}

// PublishItemCreatedEvent mocks publishing an item created event
func (m *MockBrokerProducer) PublishItemCreatedEvent(userId, listId int, item *entity.Item) {}

// PublishItemUpdatedEvent mocks publishing an item updated event
func (m *MockBrokerProducer) PublishItemUpdatedEvent(userId, listId, itemId int) {}

// PublishItemDeletedEvent mocks publishing an item deleted event
func (m *MockBrokerProducer) PublishItemDeletedEvent(itemId int) {}

// MockItemSearch mocks the search.Item interface for searching and indexing items
type MockItemSearch struct {
	mock.Mock
}

// SearchIds mocks searching for item IDs based on user Id, list Id, done status, and search text
func (m *MockItemSearch) SearchIds(ctx context.Context, userId int, listId *int, done *bool, searchText string) ([]int, error) {
	args := m.Called(ctx, userId, listId, done, searchText)
	return args.Get(0).([]int), args.Error(1)
}

// Index mocks indexing an item in the search service
func (m *MockItemSearch) Index(ctx context.Context, userId int, listId int, item entity.Item) error {
	args := m.Called(ctx, userId, listId, item)
	return args.Error(0)
}

// Delete mocks deleting an item from the search service
func (m *MockItemSearch) Delete(ctx context.Context, itemId int) error {
	args := m.Called(ctx, itemId)
	return args.Error(0)
}
