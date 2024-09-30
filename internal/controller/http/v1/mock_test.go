package v1_test

import (
	"context"

	"github.com/berikulyBeket/todo-plus/internal/entity"

	"github.com/stretchr/testify/mock"
)

// MockAuth is a mock implementation of the Auth interface
type MockAuth struct {
	mock.Mock
}

// CreateUser mocks the creation of a new user
func (m *MockAuth) CreateUser(ctx context.Context, user entity.User) (int, error) {
	args := m.Called(ctx, user)
	return args.Int(0), args.Error(1)
}

// AuthenticateUser mocks the authentication of a user
func (m *MockAuth) AuthenticateUser(ctx context.Context, username, password string) (entity.User, error) {
	args := m.Called(ctx, username, password)
	return args.Get(0).(entity.User), args.Error(1)
}

// GenerateToken mocks the generation of a token for a user
func (m *MockAuth) GenerateToken(ctx context.Context, userId int) (string, error) {
	args := m.Called(ctx, userId)
	return args.String(0), args.Error(1)
}

// ParseToken mocks the parsing of a token to extract user ID
func (m *MockAuth) ParseToken(ctx context.Context, token string) (int, error) {
	args := m.Called(ctx, token)
	return args.Int(0), args.Error(1)
}

// MockAppAuth is a mock implementation of the AppAuth interface
type MockAppAuth struct {
	mock.Mock
}

// Validate mocks the validation of app credentials
func (m *MockAppAuth) Validate(appId, appKey, accessType string) bool {
	args := m.Called(appId, appKey, accessType)
	return args.Bool(0)
}

// MockList is a mock implementation of the List interface
type MockList struct {
	mock.Mock
}

// Create mocks the creation of a new list
func (m *MockList) Create(ctx context.Context, userId int, list *entity.List) (int, error) {
	args := m.Called(ctx, userId, list)
	return args.Int(0), args.Error(1)
}

// GetAll mocks retrieving all lists for a user
func (m *MockList) GetAll(ctx context.Context, userId int) ([]entity.List, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).([]entity.List), args.Error(1)
}

// GetOneById mocks retrieving a specific list by ID
func (m *MockList) GetOneById(ctx context.Context, userId, listId int) (entity.List, error) {
	args := m.Called(ctx, userId, listId)
	return args.Get(0).(entity.List), args.Error(1)
}

// UpdateOneById mocks updating a specific list by ID
func (m *MockList) UpdateOneById(ctx context.Context, userId, listId int, input entity.UpdateListInput) error {
	args := m.Called(ctx, userId, listId, input)
	return args.Error(0)
}

// DeleteOneById mocks deleting a specific list by ID
func (m *MockList) DeleteOneById(ctx context.Context, userId, listId int) error {
	args := m.Called(ctx, userId, listId)
	return args.Error(0)
}

// DeleteOneByAdmin mocks the deletion of a list by an admin
func (m *MockList) DeleteOneByAdmin(ctx context.Context, listId int) error {
	args := m.Called(ctx, listId)
	return args.Error(0)
}

// Search mocks searching lists by a given text
func (m *MockList) Search(ctx context.Context, userId int, searchText string) ([]entity.List, error) {
	args := m.Called(ctx, userId, searchText)
	return args.Get(0).([]entity.List), args.Error(1)
}

// HandleCreated mocks handling a "list created" event
func (m *MockList) HandleCreated(ctx context.Context, message entity.ListCreatedEvent) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

// HandleUpdated mocks handling a "list updated" event
func (m *MockList) HandleUpdated(ctx context.Context, message entity.ListUpdatedEvent) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

// HandleDeleted mocks handling a "list deleted" event
func (m *MockList) HandleDeleted(ctx context.Context, message entity.ListDeletedEvent) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

// MockItem is a mock implementation of the Item interface
type MockItem struct {
	mock.Mock
}

// Create mocks the creation of a new item in a list
func (m *MockItem) Create(ctx context.Context, userId, listId int, item *entity.Item) (int, error) {
	args := m.Called(ctx, userId, listId, item)
	return args.Int(0), args.Error(1)
}

// GetAll mocks retrieving all items for a list
func (m *MockItem) GetAll(ctx context.Context, userId, listId int) ([]entity.Item, error) {
	args := m.Called(ctx, userId, listId)
	return args.Get(0).([]entity.Item), args.Error(1)
}

// GetOneById mocks retrieving a specific item by ID
func (m *MockItem) GetOneById(ctx context.Context, userId, itemId int) (entity.Item, error) {
	args := m.Called(ctx, userId, itemId)
	return args.Get(0).(entity.Item), args.Error(1)
}

// UpdateOneById mocks updating a specific item by ID
func (m *MockItem) UpdateOneById(ctx context.Context, userId, listId, itemId int, input entity.UpdateItemInput) error {
	args := m.Called(ctx, userId, listId, itemId, input)
	return args.Error(0)
}

// DeleteOneById mocks deleting a specific item by ID
func (m *MockItem) DeleteOneById(ctx context.Context, userId, listId, itemId int) error {
	args := m.Called(ctx, userId, listId, itemId)
	return args.Error(0)
}

// DeleteOneByAdmin mocks deleting an item by admin
func (m *MockItem) DeleteOneByAdmin(ctx context.Context, itemId int) error {
	args := m.Called(ctx, itemId)
	return args.Error(0)
}

// Search mocks searching items in a list by given parameters
func (m *MockItem) Search(ctx context.Context, userId int, listId *int, done *bool, searchText string) ([]entity.Item, error) {
	args := m.Called(ctx, userId, listId, done, searchText)
	return args.Get(0).([]entity.Item), args.Error(1)
}

// HandleCreated mocks handling an "item created" event
func (m *MockItem) HandleCreated(ctx context.Context, message entity.ItemCreatedEvent) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

// HandleUpdated mocks handling an "item updated" event
func (m *MockItem) HandleUpdated(ctx context.Context, message entity.ItemUpdatedEvent) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

// HandleDeleted mocks handling an "item deleted" event
func (m *MockItem) HandleDeleted(ctx context.Context, message entity.ItemDeletedEvent) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

// MockUser is a mock implementation of the User interface
type MockUser struct {
	mock.Mock
}

// DeleteOneByAdmin mocks the deletion of a user by an admin
func (m *MockUser) DeleteOneByAdmin(ctx context.Context, userId int) error {
	args := m.Called(ctx, userId)
	return args.Error(0)
}
