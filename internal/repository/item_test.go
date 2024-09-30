package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/berikulyBeket/todo-plus/internal/entity"
	"github.com/berikulyBeket/todo-plus/internal/repository"
	"github.com/berikulyBeket/todo-plus/pkg/cache"
	"github.com/berikulyBeket/todo-plus/pkg/database"
	"github.com/berikulyBeket/todo-plus/pkg/logger"
	"github.com/berikulyBeket/todo-plus/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	queryCreateItem          = fmt.Sprintf("INSERT INTO %s \\(title, description\\) VALUES \\(\\$1, \\$2\\) RETURNING id", repository.ItemsTable)
	qyeryCreateListsItems    = fmt.Sprintf("INSERT INTO %s \\(list_id, item_id\\) VALUES \\(\\$1, \\$2\\)", repository.ListsItemsTable)
	queryGetAllItems         = fmt.Sprintf("SELECT i.id, i.title, i.description, i.done FROM %s i JOIN %s li ON i.id = li.item_id WHERE li.list_id = \\$1", repository.ItemsTable, repository.ListsItemsTable)
	queryGetItemById         = fmt.Sprintf("SELECT id, title, description, done FROM %s WHERE id = \\$1", repository.ItemsTable)
	queryGetManyItemsByIds   = fmt.Sprintf("SELECT id, title, description, done FROM %s WHERE id IN \\(\\$1, \\$2\\)", repository.ItemsTable)
	queryUpdateItemById      = fmt.Sprintf("UPDATE %s SET title = \\$1, description = \\$2, done = \\$3 WHERE id = \\$4", repository.ItemsTable)
	queryUpdateTitleItemById = fmt.Sprintf("UPDATE %s SET title = \\$1 WHERE id = \\$2", repository.ItemsTable)
	queryDeleteItemById      = fmt.Sprintf("DELETE FROM %s WHERE id = \\$1", repository.ItemsTable)
	queryIsUserOwnerOfItem   = fmt.Sprintf("SELECT 1 FROM %s li JOIN %s ul ON li.list_id = ul.list_id WHERE ul.user_id = \\$1 AND li.item_id = \\$2", repository.ListsItemsTable, repository.UsersListsTable)
)

// setupItemRepoTest initializes the database and repository for ItemRepo tests
func setupItemRepoTest(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, *repository.ItemRepo, *MockCache) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	database := database.New(sqlxDB)
	noOpLogger := &logger.NoOpLogger{}
	mockCache := new(MockCache)
	caches := cache.New(mockCache, mockCache)

	itemRepo := repository.NewItemRepo(database, caches, noOpLogger)

	return sqlxDB, mock, itemRepo, mockCache
}

// Helper function to validate the sqlmock expectations for ItemRepo tests
func assertItemRepoExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	err := mock.ExpectationsWereMet()
	assert.NoError(t, err, "All expectations for ItemRepo were not met")
}

// TestCreateItem tests the creation of an item in the repository
func TestCreateItem(t *testing.T) {
	testCases := []struct {
		name        string
		listId      int
		item        entity.Item
		mockQuery   func(sqlmock.Sqlmock)
		mockCache   func(*MockCache)
		expectedId  int
		expectedErr error
	}{
		{
			name:   "Success",
			listId: 1,
			item:   entity.Item{Title: "New Item", Description: "Item description"},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				createItemQuery := queryCreateItem
				mock.ExpectQuery(createItemQuery).WithArgs("New Item", "Item description").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				createListItemsQuery := qyeryCreateListsItems
				mock.ExpectExec(createListItemsQuery).WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
			mockCache: func(mockCache *MockCache) {
				mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				mockCache.On("Delete", mock.Anything, mock.Anything).
					Return(nil)
			},
			expectedId:  1,
			expectedErr: nil,
		},
		{
			name:   "TransactionBeginFailure",
			listId: 1,
			item:   entity.Item{Title: "New Item", Description: "Item description"},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(sql.ErrConnDone)
			},
			mockCache:   func(mockCache *MockCache) {},
			expectedId:  0,
			expectedErr: sql.ErrConnDone,
		},
		{
			name:   "ItemCreationFailure",
			listId: 1,
			item:   entity.Item{Title: "New Item", Description: "Item description"},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				createItemQuery := queryCreateItem
				mock.ExpectQuery(createItemQuery).WithArgs("New Item", "Item description").
					WillReturnError(sql.ErrConnDone)

				mock.ExpectRollback()
			},
			mockCache:   func(mockCache *MockCache) {},
			expectedId:  0,
			expectedErr: sql.ErrConnDone,
		},
		{
			name:   "ListLinkFailure",
			listId: 1,
			item:   entity.Item{Title: "New Item", Description: "Item description"},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				createItemQuery := queryCreateItem
				mock.ExpectQuery(createItemQuery).WithArgs("New Item", "Item description").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				createListItemsQuery := qyeryCreateListsItems
				mock.ExpectExec(createListItemsQuery).WithArgs(1, 1).WillReturnError(sql.ErrConnDone)

				mock.ExpectRollback()
			},
			mockCache:   func(mockCache *MockCache) {},
			expectedId:  0,
			expectedErr: sql.ErrConnDone,
		},
		{
			name:   "CommitFailure",
			listId: 1,
			item:   entity.Item{Title: "New Item", Description: "Item description"},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				createItemQuery := queryCreateItem
				mock.ExpectQuery(createItemQuery).WithArgs("New Item", "Item description").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				createListItemsQuery := qyeryCreateListsItems
				mock.ExpectExec(createListItemsQuery).WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit().WillReturnError(sql.ErrTxDone)
			},
			mockCache:   func(mockCache *MockCache) {},
			expectedId:  0,
			expectedErr: sql.ErrTxDone,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			sqlxDB, mock, itemRepo, mockCache := setupItemRepoTest(t)
			defer sqlxDB.Close()

			testCase.mockQuery(mock)
			testCase.mockCache(mockCache)

			itemId, err := itemRepo.CreateListItem(context.Background(), testCase.listId, &testCase.item)

			assert.Equal(t, testCase.expectedId, itemId)
			assert.Equal(t, testCase.expectedErr, err)

			assertItemRepoExpectations(t, mock)
			mockCache.AssertExpectations(t)
		})
	}
}

// TestGetAllItems tests retrieving all items from the repository
func TestGetAllItems(t *testing.T) {
	testCases := []struct {
		name          string
		listId        int
		mockQuery     func(mock sqlmock.Sqlmock)
		mockCache     func(*MockCache)
		expectedItems []entity.Item
		expectedErr   error
	}{
		{
			name:   "Success",
			listId: 1,
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryGetAllItems
				mock.ExpectQuery(query).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "done"}).
						AddRow(1, "Item 1", "Description 1", false).
						AddRow(2, "Item 2", "Description 2", true))
			},
			mockCache: func(mockCache *MockCache) {
				mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil)
				mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
			},
			expectedItems: []entity.Item{
				{Id: 1, Title: "Item 1", Description: "Description 1", Done: false},
				{Id: 2, Title: "Item 2", Description: "Description 2", Done: true},
			},
			expectedErr: nil,
		},
		{
			name:   "QueryError",
			listId: 1,
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryGetAllItems
				mock.ExpectQuery(query).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			mockCache: func(mockCache *MockCache) {
				mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil)
			},
			expectedItems: nil,
			expectedErr:   sql.ErrConnDone,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			sqlxDB, mock, itemRepo, mockCache := setupItemRepoTest(t)
			defer sqlxDB.Close()

			testCase.mockQuery(mock)
			testCase.mockCache(mockCache)

			items, err := itemRepo.GetAllListItems(context.Background(), testCase.listId)

			assert.Equal(t, testCase.expectedItems, items)
			assert.Equal(t, testCase.expectedErr, err)

			assertItemRepoExpectations(t, mock)
			mockCache.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

// TestGetItemById tests retrieving an item by its ID from the repository
func TestGetItemById(t *testing.T) {
	testCases := []struct {
		name         string
		itemId       int
		mockQuery    func(mock sqlmock.Sqlmock)
		mockCache    func(*MockCache)
		expectedItem entity.Item
		expectedErr  error
	}{
		{
			name:   "Success",
			itemId: 1,
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryGetItemById
				mock.ExpectQuery(query).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "done"}).
						AddRow(1, "Item 1", "Description 1", false))
			},
			mockCache: func(mockCache *MockCache) {
				mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil)
				mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
			},
			expectedItem: entity.Item{Id: 1, Title: "Item 1", Description: "Description 1", Done: false},
			expectedErr:  nil,
		},
		{
			name:   "ItemNotFound",
			itemId: 999,
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryGetItemById
				mock.ExpectQuery(query).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			mockCache: func(mockCache *MockCache) {
				mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil)
			},
			expectedItem: entity.Item{Id: 0},
			expectedErr:  utils.ErrItemNotFound,
		},
		{
			name:   "QueryError",
			itemId: 1,
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryGetItemById
				mock.ExpectQuery(query).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			mockCache: func(mockCache *MockCache) {
				mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil)
			},
			expectedItem: entity.Item{Id: 0},
			expectedErr:  sql.ErrConnDone,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			sqlxDB, mock, itemRepo, mockCache := setupItemRepoTest(t)
			defer sqlxDB.Close()

			testCase.mockQuery(mock)
			testCase.mockCache(mockCache)

			item, err := itemRepo.GetOneById(context.Background(), testCase.itemId)

			assert.Equal(t, testCase.expectedItem, item)
			assert.Equal(t, testCase.expectedErr, err)

			assertItemRepoExpectations(t, mock)
			mockCache.AssertExpectations(t)
		})
	}
}

// TestItemRepo_GetManyByIds tests retrieving multiple items by their IDs from the repository
func TestItemRepo_GetManyByIds(t *testing.T) {
	testCases := []struct {
		name          string
		itemIds       []int
		mockQuery     func(sqlmock.Sqlmock)
		expectedItems []entity.Item
		expectedErr   error
	}{
		{
			name:    "Success",
			itemIds: []int{1, 2},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryGetManyItemsByIds).
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "done"}).
						AddRow(1, "Item 1", "Description 1", false).
						AddRow(2, "Item 2", "Description 2", true))
			},
			expectedItems: []entity.Item{
				{Id: 1, Title: "Item 1", Description: "Description 1", Done: false},
				{Id: 2, Title: "Item 2", Description: "Description 2", Done: true},
			},
			expectedErr: nil,
		},
		{
			name:          "Empty item ids",
			itemIds:       []int{},
			mockQuery:     func(mock sqlmock.Sqlmock) {},
			expectedItems: []entity.Item{},
			expectedErr:   nil,
		},
		{
			name:    "Query error",
			itemIds: []int{1, 2},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryGetManyItemsByIds).
					WithArgs(1, 2).
					WillReturnError(fmt.Errorf("query error"))
			},
			expectedItems: nil,
			expectedErr:   fmt.Errorf("query error"),
		},
		{
			name:    "Scan error",
			itemIds: []int{1, 2},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryGetManyItemsByIds).
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "done"}).
						AddRow("invalid", "Item 1", "Description 1", false))
			},
			expectedItems: nil,
			expectedErr:   fmt.Errorf("sql: Scan error"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {

			t.Parallel()

			sqlxDB, mock, itemRepo, _ := setupItemRepoTest(t)
			defer sqlxDB.Close()

			testCase.mockQuery(mock)

			items, err := itemRepo.GetManyByIds(context.Background(), testCase.itemIds)

			assert.Equal(t, testCase.expectedItems, items)

			if testCase.name == "Scan error" && err != nil {
				assert.Contains(t, err.Error(), "converting driver.Value type string (\"invalid\") to a int")
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}

			assertItemRepoExpectations(t, mock)
		})
	}
}

// TestUpdateItemById tests updating an item by its ID in the repository
func TestUpdateItemById(t *testing.T) {
	title := "Updated Title"
	description := "Updated Description"
	doneFlag := true

	testCases := []struct {
		name        string
		listId      int
		itemId      int
		updateInput entity.UpdateItemInput
		mockQuery   func(mock sqlmock.Sqlmock)
		mockCache   func(*MockCache)
		expectedErr error
	}{
		{
			name:   "Success_AllFieldsUpdated",
			listId: 3,
			itemId: 1,
			updateInput: entity.UpdateItemInput{
				Title:       &title,
				Description: &description,
				Done:        &doneFlag,
			},
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryUpdateItemById
				mock.ExpectExec(query).
					WithArgs("Updated Title", "Updated Description", true, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			mockCache: func(mockCache *MockCache) {
				mockCache.On("Delete", mock.Anything, mock.Anything).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "Success_TitleOnly",
			listId: 3,
			itemId: 1,
			updateInput: entity.UpdateItemInput{
				Title: &title,
			},
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryUpdateTitleItemById
				mock.ExpectExec(query).
					WithArgs("Updated Title", 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			mockCache: func(mockCache *MockCache) {
				mockCache.On("Delete", mock.Anything, mock.Anything).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "NoFieldsToUpdate",
			listId:      3,
			itemId:      1,
			updateInput: entity.UpdateItemInput{},
			mockQuery:   func(mock sqlmock.Sqlmock) {},
			mockCache:   func(mockCache *MockCache) {},
			expectedErr: utils.ErrItemEmptyRequest,
		},
		{
			name:   "DatabaseError",
			listId: 3,
			itemId: 1,
			updateInput: entity.UpdateItemInput{
				Title: &title,
			},
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryUpdateTitleItemById
				mock.ExpectExec(query).
					WithArgs("Updated Title", 1).
					WillReturnError(sql.ErrConnDone)
			},
			mockCache:   func(mockCache *MockCache) {},
			expectedErr: sql.ErrConnDone,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			sqlxDB, mock, itemRepo, mockCache := setupItemRepoTest(t)
			defer sqlxDB.Close()

			testCase.mockQuery(mock)
			testCase.mockCache(mockCache)

			err := itemRepo.UpdateOneById(context.Background(), &testCase.listId, testCase.itemId, testCase.updateInput)

			assert.Equal(t, testCase.expectedErr, err)

			assertItemRepoExpectations(t, mock)
			mockCache.AssertExpectations(t)
		})
	}
}

// TestDeleteOneById tests deleting an item by its ID in the repository
func TestDeleteOneById(t *testing.T) {
	testCases := []struct {
		name        string
		itemId      int
		listId      int
		mockQuery   func(mock sqlmock.Sqlmock)
		mockCache   func(*MockCache)
		expectedErr error
	}{
		{
			name:   "Success",
			listId: 3,
			itemId: 1,
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryDeleteItemById
				mock.ExpectExec(query).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			mockCache: func(mockCache *MockCache) {
				mockCache.On("Delete", mock.Anything, mock.Anything).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "Item not found",
			listId: 3,
			itemId: 1,
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryDeleteItemById
				mock.ExpectExec(query).
					WithArgs(1).
					WillReturnError(utils.ErrItemNotFound)
			},
			mockCache: func(mockCache *MockCache) {
				mockCache.On("Delete", mock.Anything, mock.Anything).
					Return(nil)
			},
			expectedErr: utils.ErrItemNotFound,
		},
		{
			name:   "QueryError",
			listId: 3,
			itemId: 1,
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryDeleteItemById
				mock.ExpectExec(query).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			mockCache: func(mockCache *MockCache) {
				mockCache.On("Delete", mock.Anything, mock.Anything).
					Return(nil)
			},
			expectedErr: sql.ErrConnDone,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			sqlxDB, mock, itemRepo, mockCache := setupItemRepoTest(t)
			defer sqlxDB.Close()

			testCase.mockQuery(mock)
			testCase.mockCache(mockCache)

			err := itemRepo.DeleteOneById(context.Background(), &testCase.listId, testCase.itemId)

			assert.Equal(t, testCase.expectedErr, err)

			assertItemRepoExpectations(t, mock)
		})
	}
}

// TestIsUserOwnerOfItem tests whether a user is the owner of an item in the repository
func TestIsUserOwnerOfItem(t *testing.T) {
	testCases := []struct {
		name        string
		userId      int
		itemId      int
		mockQuery   func(sqlmock.Sqlmock)
		mockCache   func(*MockCache)
		expectedErr error
	}{
		{
			name:   "Success",
			userId: 1,
			itemId: 1,
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryIsUserOwnerOfItem
				mock.ExpectQuery(query).WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
			},
			mockCache:   func(mockCache *MockCache) {},
			expectedErr: nil,
		},
		{
			name:   "Item not belongs to user",
			userId: 1,
			itemId: 1,
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryIsUserOwnerOfItem
				mock.ExpectQuery(query).WithArgs(1, 1).
					WillReturnError(utils.ErrUserNotOwner)
			},
			mockCache:   func(mockCache *MockCache) {},
			expectedErr: utils.ErrUserNotOwner,
		},
		{
			name:   "DatabaseError",
			userId: 1,
			itemId: 1,
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryIsUserOwnerOfItem
				mock.ExpectQuery(query).WithArgs(1, 1).
					WillReturnError(sql.ErrConnDone)
			},
			mockCache:   func(mockCache *MockCache) {},
			expectedErr: sql.ErrConnDone,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			sqlxDB, mock, itemRepo, mockCache := setupItemRepoTest(t)
			defer sqlxDB.Close()

			testCase.mockQuery(mock)
			testCase.mockCache(mockCache)

			err := itemRepo.IsUserOwnerOfItem(context.Background(), testCase.userId, testCase.itemId)

			assert.Equal(t, testCase.expectedErr, err)

			assertItemRepoExpectations(t, mock)
		})
	}
}
