package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/berikulyBeket/todo-plus/internal/entity"
	"github.com/berikulyBeket/todo-plus/internal/usecase"
	"github.com/berikulyBeket/todo-plus/pkg/logger"
	"github.com/berikulyBeket/todo-plus/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestCreateList tests the CreateList function in the ListUseCas
func TestCreateList(t *testing.T) {
	testCases := []struct {
		name        string
		userId      int
		list        entity.List
		expectedId  int
		expectedErr error
	}{
		{
			name:        "Successful creation",
			userId:      1,
			list:        entity.List{Title: "New List"},
			expectedId:  101,
			expectedErr: nil,
		},
		{
			name:        "Failed creation",
			userId:      1,
			list:        entity.List{Title: "New List"},
			expectedId:  0,
			expectedErr: errors.New("failed to create list"),
		},
	}

	for _, testCase := range testCases {
		mockRepo := new(MockListRepo)
		mockListSearch := new(MockListSearch)
		mockBrokerProducer := new(MockBrokerProducer)
		listUseCase := usecase.NewListUseCase(mockRepo, mockListSearch, mockBrokerProducer, &logger.NoOpLogger{})

		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockRepo.On("CreateUserList", mock.Anything, testCase.userId, &testCase.list).Return(testCase.expectedId, testCase.expectedErr)

			actualId, err := listUseCase.Create(context.Background(), testCase.userId, &testCase.list)

			assert.Equal(t, testCase.expectedId, actualId)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGetAllLists tests the GetAllLists function in the ListUseCase
func TestGetAllLists(t *testing.T) {
	testCases := []struct {
		name          string
		userId        int
		expectedLists []entity.List
		expectedErr   error
	}{
		{
			name:          "Successful retrieval",
			userId:        1,
			expectedLists: []entity.List{{Title: "List1"}, {Title: "List2"}},
			expectedErr:   nil,
		},
		{
			name:          "No lists found",
			userId:        1,
			expectedLists: nil,
			expectedErr:   errors.New("no lists found"),
		},
	}

	for _, testCase := range testCases {
		mockRepo := new(MockListRepo)
		mockListSearch := new(MockListSearch)
		mockBrokerProducer := new(MockBrokerProducer)
		listUseCase := usecase.NewListUseCase(mockRepo, mockListSearch, mockBrokerProducer, &logger.NoOpLogger{})

		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockRepo.On("GetAllUserLists", mock.Anything, testCase.userId).Return(testCase.expectedLists, testCase.expectedErr)

			actualLists, err := listUseCase.GetAll(context.Background(), testCase.userId)

			assert.Equal(t, testCase.expectedLists, actualLists)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGetListById tests the GetListById function in the ListUseCase
func TestGetListById(t *testing.T) {
	testCases := []struct {
		name             string
		userId           int
		listId           int
		expectedList     entity.List
		expectedOwnerErr error
		expectedErr      error
	}{
		{
			name:             "Successful retrieval",
			userId:           1,
			listId:           1,
			expectedList:     entity.List{Title: "List1"},
			expectedOwnerErr: nil,
			expectedErr:      nil,
		},
		{
			name:             "List not belongs to user",
			userId:           1,
			listId:           99,
			expectedList:     entity.List{},
			expectedOwnerErr: utils.ErrUserNotOwner,
			expectedErr:      nil,
		},
		{
			name:             "List not found",
			userId:           1,
			listId:           99,
			expectedList:     entity.List{},
			expectedOwnerErr: nil,
			expectedErr:      utils.ErrListNotFound,
		},
	}

	for _, testCase := range testCases {
		mockRepo := new(MockListRepo)
		mockListSearch := new(MockListSearch)
		mockBrokerProducer := new(MockBrokerProducer)
		listUseCase := usecase.NewListUseCase(mockRepo, mockListSearch, mockBrokerProducer, &logger.NoOpLogger{})

		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockRepo.On("IsUserOwnerOfList", mock.Anything, testCase.userId, testCase.listId).Return(testCase.expectedOwnerErr)
			mockRepo.On("GetOneById", mock.Anything, testCase.listId).Return(testCase.expectedList, testCase.expectedErr)

			actualList, err := listUseCase.GetOneById(context.Background(), testCase.userId, testCase.listId)

			assert.Equal(t, testCase.expectedList, actualList)
			if testCase.expectedOwnerErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedOwnerErr, err)
			} else if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUpdateListById tests the UpdateListById function in the ListUseCase
func TestUpdateListById(t *testing.T) {
	title := "Updated List"

	testCases := []struct {
		name             string
		userId           int
		listId           int
		newTodoInput     entity.UpdateListInput
		expectedOwnerErr error
		expectedErr      error
	}{
		{
			name:             "Successful update",
			userId:           1,
			listId:           1,
			newTodoInput:     entity.UpdateListInput{Title: &title},
			expectedOwnerErr: nil,
			expectedErr:      nil,
		},
		{
			name:             "List not belongs to user",
			userId:           1,
			listId:           99,
			newTodoInput:     entity.UpdateListInput{Title: &title},
			expectedOwnerErr: utils.ErrUserNotOwner,
			expectedErr:      utils.ErrListNotFound,
		},
		{
			name:             "List not found",
			userId:           1,
			listId:           99,
			newTodoInput:     entity.UpdateListInput{Title: &title},
			expectedOwnerErr: nil,
			expectedErr:      utils.ErrListNotFound,
		},
	}

	for _, testCase := range testCases {
		mockRepo := new(MockListRepo)
		mockListSearch := new(MockListSearch)
		mockBrokerProducer := new(MockBrokerProducer)
		listUseCase := usecase.NewListUseCase(mockRepo, mockListSearch, mockBrokerProducer, &logger.NoOpLogger{})

		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockRepo.On("IsUserOwnerOfList", mock.Anything, testCase.userId, testCase.listId).Return(testCase.expectedOwnerErr)
			mockRepo.On("UpdateOneById", mock.Anything, &testCase.userId, testCase.listId, testCase.newTodoInput).Return(testCase.expectedErr)

			err := listUseCase.UpdateOneById(context.Background(), testCase.userId, testCase.listId, testCase.newTodoInput)

			if testCase.expectedOwnerErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedOwnerErr, err)
			} else if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDeleteListById tests the DeleteListById function in the ListUseCase
func TestDeleteListById(t *testing.T) {
	testCases := []struct {
		name             string
		userId           int
		listId           int
		expectedOwnerErr error
		expectedErr      error
	}{
		{
			name:             "Successful deletion",
			userId:           1,
			listId:           1,
			expectedOwnerErr: nil,
			expectedErr:      nil,
		},
		{
			name:             "List not belongs to user",
			userId:           1,
			listId:           99,
			expectedOwnerErr: utils.ErrUserNotFound,
			expectedErr:      utils.ErrListNotFound,
		},
		{
			name:             "List not found",
			userId:           1,
			listId:           99,
			expectedOwnerErr: nil,
			expectedErr:      utils.ErrListNotFound,
		},
	}

	for _, testCase := range testCases {
		mockRepo := new(MockListRepo)
		mockListSearch := new(MockListSearch)
		mockBrokerProducer := new(MockBrokerProducer)
		listUseCase := usecase.NewListUseCase(mockRepo, mockListSearch, mockBrokerProducer, &logger.NoOpLogger{})

		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockRepo.On("IsUserOwnerOfList", mock.Anything, testCase.userId, testCase.listId).Return(testCase.expectedOwnerErr)
			mockRepo.On("DeleteOneById", mock.Anything, &testCase.userId, testCase.listId).Return(testCase.expectedErr)

			err := listUseCase.DeleteOneById(context.Background(), testCase.userId, testCase.listId)

			if testCase.expectedOwnerErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedOwnerErr, err)
			} else if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDeleteOneByAdmin tests the DeleteOneByAdmin function in the ListUseCase
func TestDeleteOneByAdmin(t *testing.T) {
	testCases := []struct {
		name        string
		listId      int
		expectedErr error
	}{
		{
			name:        "Successful deletion",
			listId:      1,
			expectedErr: nil,
		},
		{
			name:        "List not found",
			listId:      99,
			expectedErr: utils.ErrListNotFound,
		},
	}

	for _, testCase := range testCases {
		mockRepo := new(MockListRepo)
		mockListSearch := new(MockListSearch)
		mockBrokerProducer := new(MockBrokerProducer)
		listUseCase := usecase.NewListUseCase(mockRepo, mockListSearch, mockBrokerProducer, &logger.NoOpLogger{})

		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var userId *int
			mockRepo.On("DeleteOneById", mock.Anything, userId, testCase.listId).Return(testCase.expectedErr)

			err := listUseCase.DeleteOneByAdmin(context.Background(), testCase.listId)

			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
