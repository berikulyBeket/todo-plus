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
	queryInsertList                = fmt.Sprintf("INSERT INTO %s \\(title, description\\) VALUES \\(\\$1, \\$2\\) RETURNING id", repository.ListsTable)
	queryLinkUser                  = fmt.Sprintf("INSERT INTO %s \\(user_id, list_id\\) VALUES \\(\\$1, \\$2\\)", repository.UsersListsTable)
	queryGetAllLists               = fmt.Sprintf("SELECT l.id, l.title, l.description FROM %s l JOIN %s ul ON l.id = ul.list_id WHERE ul.user_id = \\$1", repository.ListsTable, repository.UsersListsTable)
	queryGetListById               = fmt.Sprintf("SELECT id, title, description FROM %s WHERE id = \\$1", repository.ListsTable)
	queryGetManyListsByIds         = fmt.Sprintf("SELECT id, title, description FROM %s WHERE id IN \\(\\$1, \\$2\\)", repository.ListsTable)
	queryUpdateTitleListById       = fmt.Sprintf("UPDATE %s SET title = \\$1 WHERE id = \\$2", repository.ListsTable)
	queryUpdateDescriptionListById = fmt.Sprintf("UPDATE %s SET description = \\$1 WHERE id = \\$2", repository.ListsTable)
	queryUpdateListById            = fmt.Sprintf("UPDATE %s SET title = \\$1, description = \\$2 WHERE id = \\$3", repository.ListsTable)
	queryDeleteListById            = fmt.Sprintf("DELETE FROM %s WHERE id = \\$1", repository.ListsTable)
	queryIsUserOwnerOfList         = fmt.Sprintf("SELECT 1 FROM %s WHERE user_id = \\$1 AND list_id = \\$2", repository.UsersListsTable)
)

// Helper function to set up the mock database, sqlmock, and repository
func setupListRepoTest(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, *repository.ListRepo, *MockCache) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	database := database.New(sqlxDB)
	noOpLogger := &logger.NoOpLogger{}
	mockCache := new(MockCache)
	caches := cache.New(mockCache, mockCache)

	listRepo := repository.NewListRepo(database, caches, noOpLogger)

	return sqlxDB, mock, listRepo, mockCache
}

// Helper function to validate the sqlmock expectations for ListRepo tests
func assertListRepoExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	err := mock.ExpectationsWereMet()
	assert.NoError(t, err, "All expectations for ListRepo were not met")
}

// TestCreateList tests the creation of a new list
func TestCreateList(t *testing.T) {
	testCases := []struct {
		name        string
		list        entity.List
		mockQuery   func(sqlmock.Sqlmock)
		mockCache   func(*MockCache)
		expectedId  int
		expectedErr error
	}{
		{
			name: "Success",
			list: entity.List{Title: "Grocery List", Description: "A list of groceries"},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(queryInsertList).
					WithArgs("Grocery List", "A list of groceries").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				mock.ExpectExec(queryLinkUser).
					WithArgs(123, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))

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
			name: "SQLConnectionFailure",
			list: entity.List{Title: "Grocery List", Description: "A list of groceries"},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(queryInsertList).
					WithArgs("Grocery List", "A list of groceries").
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			mockCache:   func(mockCache *MockCache) {},
			expectedId:  0,
			expectedErr: sql.ErrConnDone,
		},
		{
			name: "UserListLinkFailure",
			list: entity.List{Title: "Grocery List", Description: "A list of groceries"},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(queryInsertList).
					WithArgs("Grocery List", "A list of groceries").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				mock.ExpectExec(queryLinkUser).
					WithArgs(123, 1).
					WillReturnError(sql.ErrTxDone)
				mock.ExpectRollback()
			},
			mockCache:   func(mockCache *MockCache) {},
			expectedId:  0,
			expectedErr: sql.ErrTxDone,
		},
		{
			name: "TransactionBeginFailure",
			list: entity.List{Title: "Grocery List", Description: "A list of groceries"},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(sql.ErrConnDone)
			},
			mockCache:   func(mockCache *MockCache) {},
			expectedId:  0,
			expectedErr: sql.ErrConnDone,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			sqlxDB, mock, listRepo, mockCache := setupListRepoTest(t)
			defer sqlxDB.Close()

			testCase.mockQuery(mock)
			testCase.mockCache(mockCache)

			listId, err := listRepo.CreateUserList(context.Background(), 123, &testCase.list)

			assert.Equal(t, testCase.expectedId, listId)
			assert.Equal(t, testCase.expectedErr, err)

			assertListRepoExpectations(t, mock)
			mockCache.AssertExpectations(t)
		})
	}
}

// TestGetAll tests retrieving all user lists
func TestGetAll(t *testing.T) {
	testCases := []struct {
		name          string
		mockQuery     func(sqlmock.Sqlmock)
		mockCache     func(*MockCache)
		expectedLists []entity.List
		expectedErr   error
	}{
		{
			name: "Success",
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryGetAllLists).
					WithArgs(123).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description"}).
						AddRow(1, "Grocery List", "A list of groceries").
						AddRow(2, "Work Tasks", "Tasks to complete at work"))
			},
			mockCache: func(mockCache *MockCache) {
				mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil)
				mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
			},
			expectedLists: []entity.List{
				{Id: 1, Title: "Grocery List", Description: "A list of groceries"},
				{Id: 2, Title: "Work Tasks", Description: "Tasks to complete at work"},
			},
			expectedErr: nil,
		},
		{
			name: "NoLists",
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryGetAllLists).
					WithArgs(123).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description"}))
			},
			mockCache: func(mockCache *MockCache) {
				mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil)
				mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
			},
			expectedLists: []entity.List{},
			expectedErr:   nil,
		},
		{
			name: "QueryError",
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryGetAllLists).
					WithArgs(123).
					WillReturnError(sql.ErrConnDone)
			},
			mockCache: func(mockCache *MockCache) {
				mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil)
			},
			expectedLists: nil,
			expectedErr:   sql.ErrConnDone,
		},
		{
			name: "ScanError",
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryGetAllLists).
					WithArgs(123).
					WillReturnRows(sqlmock.NewRows([]string{"title", "description"}).
						AddRow("Grocery List", "A list of groceries")) // Simulate scan error
			},
			mockCache: func(mockCache *MockCache) {
				mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil)
			},
			expectedLists: nil,
			expectedErr:   fmt.Errorf("sql: expected 2 destination arguments in Scan, not 3"),
		},
		{
			name: "RowsError",
			mockQuery: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "title", "description"}).
					RowError(0, sql.ErrConnDone).
					AddRow(1, "Grocery List", "A list of groceries")

				mock.ExpectQuery(queryGetAllLists).
					WithArgs(123).
					WillReturnRows(rows)
			},
			mockCache: func(mockCache *MockCache) {
				mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil)
			},
			expectedLists: nil,
			expectedErr:   sql.ErrConnDone,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			sqlxDB, mock, listRepo, mockCache := setupListRepoTest(t)
			defer sqlxDB.Close()

			testCase.mockQuery(mock)
			testCase.mockCache(mockCache)

			lists, err := listRepo.GetAllUserLists(context.Background(), 123)

			assert.Equal(t, testCase.expectedLists, lists)
			assert.Equal(t, testCase.expectedErr, err)

			assertListRepoExpectations(t, mock)
			mockCache.AssertExpectations(t)
		})
	}
}

// TestGetOneById tests retrieving a single list by its Id
func TestGetOneById(t *testing.T) {
	testCases := []struct {
		name         string
		listId       int
		mockQuery    func(sqlmock.Sqlmock)
		mockCache    func(*MockCache)
		expectedList entity.List
		expectedErr  error
	}{
		{
			name:   "Success",
			listId: 1,
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryGetListById
				mock.ExpectQuery(query).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description"}).AddRow(1, "Grocery List", "A list of groceries"))
			},
			mockCache: func(mockCache *MockCache) {
				mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil)
				mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
			},
			expectedList: entity.List{Id: 1, Title: "Grocery List", Description: "A list of groceries"},
			expectedErr:  nil,
		},
		{
			name:   "ListNotFound",
			listId: 999,
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryGetListById
				mock.ExpectQuery(query).WithArgs(999).WillReturnError(sql.ErrNoRows)
			},
			mockCache: func(mockCache *MockCache) {
				mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil)
			},
			expectedList: entity.List{},
			expectedErr:  utils.ErrListNotFound,
		},
		{
			name:   "QueryError",
			listId: 1,
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryGetListById
				mock.ExpectQuery(query).WithArgs(1).WillReturnError(sql.ErrConnDone)
			},
			mockCache: func(mockCache *MockCache) {
				mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil)
			},
			expectedList: entity.List{},
			expectedErr:  sql.ErrConnDone,
		},
		{
			name:   "ScanError",
			listId: 1,
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryGetListById
				mock.ExpectQuery(query).WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"title", "description"}).
						AddRow("Grocery List", "A list of groceries"))
			},
			mockCache: func(mockCache *MockCache) {
				mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil)
			},
			expectedList: entity.List{},
			expectedErr:  fmt.Errorf("sql: expected 2 destination arguments in Scan, not 3"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			sqlxDB, mock, listRepo, mockCache := setupListRepoTest(t)
			defer sqlxDB.Close()

			testCase.mockQuery(mock)
			testCase.mockCache(mockCache)

			list, err := listRepo.GetOneById(context.Background(), testCase.listId)

			assert.Equal(t, testCase.expectedList, list)
			assert.Equal(t, testCase.expectedErr, err)

			assertListRepoExpectations(t, mock)
			mockCache.AssertExpectations(t)
		})
	}
}

// TestGetManyByIds tests retrieving multiple lists by their Ids
func TestGetManyByIds(t *testing.T) {
	testCases := []struct {
		name          string
		listIds       []int
		mockQuery     func(sqlmock.Sqlmock)
		expectedLists []entity.List
		expectedErr   error
	}{
		{
			name:    "Success",
			listIds: []int{1, 2},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryGetManyListsByIds).
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description"}).
						AddRow(1, "List 1", "Description 1").
						AddRow(2, "List 2", "Description 2"))
			},
			expectedLists: []entity.List{
				{Id: 1, Title: "List 1", Description: "Description 1"},
				{Id: 2, Title: "List 2", Description: "Description 2"},
			},
			expectedErr: nil,
		},
		{
			name:          "Empty list ids",
			listIds:       []int{},
			mockQuery:     func(mock sqlmock.Sqlmock) {},
			expectedLists: []entity.List{},
			expectedErr:   nil,
		},
		{
			name:    "Query error",
			listIds: []int{1, 2},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryGetManyListsByIds).
					WithArgs(1, 2).
					WillReturnError(fmt.Errorf("query error"))
			},
			expectedLists: nil,
			expectedErr:   fmt.Errorf("query error"),
		},
		{
			name:    "Scan error",
			listIds: []int{1, 2},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryGetManyListsByIds).
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description"}).
						AddRow("invalid", "List 1", "Description 1"))
			},
			expectedLists: nil,
			expectedErr:   fmt.Errorf("sql: Scan error"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {

			t.Parallel()

			sqlxDB, mock, listRepo, _ := setupListRepoTest(t)
			defer sqlxDB.Close()

			testCase.mockQuery(mock)

			lists, err := listRepo.GetManyByIds(context.Background(), testCase.listIds)

			assert.Equal(t, testCase.expectedLists, lists)

			if testCase.name == "Scan error" && err != nil {
				assert.Contains(t, err.Error(), "converting driver.Value type string (\"invalid\") to a int")
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}

			assertListRepoExpectations(t, mock)
		})
	}
}

// TestUpdateOneById tests updating a list by its Id
func TestUpdateOneById(t *testing.T) {
	title := "Updated Title"
	description := "Updated Description"
	testCases := []struct {
		name        string
		userId      int
		listId      int
		updateInput entity.UpdateListInput
		mockQuery   func(sqlmock.Sqlmock)
		mockCache   func(*MockCache)
		expectedErr error
	}{
		{
			name:   "Success",
			userId: 2,
			listId: 1,
			updateInput: entity.UpdateListInput{
				Title:       &title,
				Description: &description,
			},
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryUpdateListById
				mock.ExpectExec(query).
					WithArgs("Updated Title", "Updated Description", 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			mockCache: func(mockCache *MockCache) {
				mockCache.On("Delete", mock.Anything, mock.Anything).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "UpdateOnlyTitle",
			userId: 2,
			listId: 1,
			updateInput: entity.UpdateListInput{
				Title: &title,
			},
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryUpdateTitleListById
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
			name:   "UpdateOnlyDescription",
			userId: 2,
			listId: 1,
			updateInput: entity.UpdateListInput{
				Description: &description,
			},
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryUpdateDescriptionListById
				mock.ExpectExec(query).
					WithArgs("Updated Description", 1).
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
			userId:      2,
			listId:      1,
			updateInput: entity.UpdateListInput{},
			mockQuery:   func(mock sqlmock.Sqlmock) {},
			mockCache:   func(mockCache *MockCache) {},
			expectedErr: utils.ErrItemEmptyRequest,
		},
		{
			name:   "DatabaseError",
			userId: 2,
			listId: 1,
			updateInput: entity.UpdateListInput{
				Title:       &title,
				Description: &description,
			},
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryUpdateListById
				mock.ExpectExec(query).
					WithArgs("Updated Title", "Updated Description", 1).
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

			sqlxDB, mock, listRepo, mockCache := setupListRepoTest(t)
			defer sqlxDB.Close()

			testCase.mockQuery(mock)
			testCase.mockCache(mockCache)

			err := listRepo.UpdateOneById(context.Background(), &testCase.userId, testCase.listId, testCase.updateInput)

			assert.Equal(t, testCase.expectedErr, err)

			assertListRepoExpectations(t, mock)
			mockCache.AssertExpectations(t)
		})
	}
}

// TestDeleteListById tests deleting a list by its Id
func TestDeleteListById(t *testing.T) {
	testCases := []struct {
		name        string
		userId      int
		listId      int
		mockQuery   func(sqlmock.Sqlmock)
		mockCache   func(*MockCache)
		expectedErr error
	}{
		{
			name:   "Success",
			userId: 2,
			listId: 1,
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryDeleteListById
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
			name:   "ListNotFound",
			userId: 2,
			listId: 999,
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryDeleteListById
				mock.ExpectExec(query).
					WithArgs(999).
					WillReturnResult(sqlmock.NewResult(1, 0))
			},
			mockCache:   func(mockCache *MockCache) {},
			expectedErr: utils.ErrListNotFound,
		},
		{
			name:   "DatabaseError",
			userId: 2,
			listId: 1,
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryDeleteListById
				mock.ExpectExec(query).
					WithArgs(1).
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

			sqlxDB, mock, listRepo, mockCache := setupListRepoTest(t)
			defer sqlxDB.Close()

			testCase.mockQuery(mock)
			testCase.mockCache(mockCache)

			err := listRepo.DeleteOneById(context.Background(), &testCase.userId, testCase.listId)

			assert.Equal(t, testCase.expectedErr, err)

			assertListRepoExpectations(t, mock)
			mockCache.AssertExpectations(t)
		})
	}
}

// TestIsUserOwnerOfList tests if a user is the owner of a specific list
func TestIsUserOwnerOfList(t *testing.T) {
	testCases := []struct {
		name        string
		userId      int
		listId      int
		mockQuery   func(sqlmock.Sqlmock)
		expectedErr error
	}{
		{
			name:   "Success",
			userId: 1,
			listId: 1,
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryIsUserOwnerOfList
				mock.ExpectQuery(query).WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
			},
			expectedErr: nil,
		},
		{
			name:   "List not belongs to user",
			userId: 1,
			listId: 1,
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryIsUserOwnerOfList
				mock.ExpectQuery(query).WithArgs(1, 1).
					WillReturnError(utils.ErrUserNotOwner)
			},
			expectedErr: utils.ErrUserNotOwner,
		},
		{
			name:   "DatabaseError",
			userId: 1,
			listId: 1,
			mockQuery: func(mock sqlmock.Sqlmock) {
				query := queryIsUserOwnerOfList
				mock.ExpectQuery(query).WithArgs(1, 1).
					WillReturnError(sql.ErrConnDone)
			},
			expectedErr: sql.ErrConnDone,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			sqlxDB, mock, listRepo, _ := setupListRepoTest(t)
			defer sqlxDB.Close()

			testCase.mockQuery(mock)

			err := listRepo.IsUserOwnerOfList(context.Background(), testCase.userId, testCase.listId)

			assert.Equal(t, testCase.expectedErr, err)

			assertListRepoExpectations(t, mock)
		})
	}
}
