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

// TestCreateItem tests the CreateItem function in the ItemUseCase
func TestCreateItem(t *testing.T) {
	testCases := []struct {
		name             string
		userId           int
		listId           int
		item             entity.Item
		expectedId       int
		expectedOwnerErr error
		expectedErr      error
	}{
		{
			name:             "Successful creation",
			userId:           1,
			listId:           1,
			item:             entity.Item{Title: "New Item"},
			expectedId:       101,
			expectedOwnerErr: nil,
			expectedErr:      nil,
		},
		{
			name:             "List not belongs to user",
			userId:           1,
			listId:           1,
			item:             entity.Item{Title: "New Item"},
			expectedId:       0,
			expectedOwnerErr: utils.ErrUserNotOwner,
			expectedErr:      nil,
		},
		{
			name:             "Failed creation",
			userId:           1,
			listId:           1,
			item:             entity.Item{Title: "New Item"},
			expectedId:       0,
			expectedOwnerErr: nil,
			expectedErr:      errors.New("failed to create item"),
		},
	}

	for _, testCase := range testCases {
		mockItemRepo := new(MockItemRepo)
		mockListRepo := new(MockListRepo)
		mockItemSearch := new(MockItemSearch)
		mockBrokerProducer := new(MockBrokerProducer)
		itemUseCase := usecase.NewItemUseCase(mockItemRepo, mockListRepo, mockItemSearch, mockBrokerProducer, &logger.NoOpLogger{})

		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockListRepo.On("IsUserOwnerOfList", mock.Anything, testCase.userId, testCase.listId).Return(testCase.expectedOwnerErr)
			mockItemRepo.On("CreateListItem", mock.Anything, testCase.listId, &testCase.item).Return(testCase.expectedId, testCase.expectedErr)

			actualId, err := itemUseCase.Create(context.Background(), testCase.userId, testCase.listId, &testCase.item)

			assert.Equal(t, testCase.expectedId, actualId)
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

// TestGetAllItems tests the GetAllItems function in the ItemUseCase
func TestGetAllItems(t *testing.T) {
	testCases := []struct {
		name             string
		userId           int
		listId           int
		expectedItems    []entity.Item
		expectedOwnerErr error
		expectedErr      error
	}{
		{
			name:             "Successful retrieval",
			userId:           1,
			listId:           1,
			expectedItems:    []entity.Item{{Title: "Item 1"}, {Title: "Item 2"}},
			expectedOwnerErr: nil,
			expectedErr:      nil,
		},
		{
			name:             "List not belongs to user",
			userId:           1,
			listId:           1,
			expectedItems:    make([]entity.Item, 0),
			expectedOwnerErr: utils.ErrUserNotOwner,
			expectedErr:      nil,
		},
		{
			name:             "Failed retrieval",
			userId:           1,
			listId:           1,
			expectedItems:    make([]entity.Item, 0),
			expectedOwnerErr: nil,
			expectedErr:      errors.New("failed to retrieve items"),
		},
	}

	for _, testCase := range testCases {
		mockItemRepo := new(MockItemRepo)
		mockListRepo := new(MockListRepo)
		mockItemSearch := new(MockItemSearch)
		mockBrokerProducer := new(MockBrokerProducer)
		itemUseCase := usecase.NewItemUseCase(mockItemRepo, mockListRepo, mockItemSearch, mockBrokerProducer, &logger.NoOpLogger{})

		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockListRepo.On("IsUserOwnerOfList", mock.Anything, testCase.userId, testCase.listId).Return(testCase.expectedOwnerErr)
			mockItemRepo.On("GetAllListItems", mock.Anything, testCase.listId).Return(testCase.expectedItems, testCase.expectedErr)

			actualItems, err := itemUseCase.GetAll(context.Background(), testCase.userId, testCase.listId)

			assert.Equal(t, testCase.expectedItems, actualItems)
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

// TestGetItemById tests the GetItemById function in the ItemUseCase
func TestGetItemById(t *testing.T) {
	testCases := []struct {
		name             string
		userId           int
		itemId           int
		expectedItem     entity.Item
		expectedOwnerErr error
		expectedErr      error
	}{
		{
			name:             "Successful retrieval",
			userId:           1,
			itemId:           1,
			expectedItem:     entity.Item{Title: "Item 1"},
			expectedOwnerErr: nil,
			expectedErr:      nil,
		},
		{
			name:             "Item not belongs to user",
			userId:           1,
			itemId:           99,
			expectedItem:     entity.Item{},
			expectedOwnerErr: utils.ErrUserNotOwner,
			expectedErr:      nil,
		},
		{
			name:             "Item not found",
			userId:           1,
			itemId:           99,
			expectedItem:     entity.Item{},
			expectedOwnerErr: nil,
			expectedErr:      utils.ErrItemNotFound,
		},
	}

	for _, testCase := range testCases {
		mockItemRepo := new(MockItemRepo)
		mockListRepo := new(MockListRepo)
		mockItemSearch := new(MockItemSearch)
		mockBrokerProducer := new(MockBrokerProducer)
		itemUseCase := usecase.NewItemUseCase(mockItemRepo, mockListRepo, mockItemSearch, mockBrokerProducer, &logger.NoOpLogger{})

		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockItemRepo.On("IsUserOwnerOfItem", mock.Anything, testCase.userId, testCase.itemId).Return(testCase.expectedOwnerErr)
			mockItemRepo.On("GetOneById", mock.Anything, testCase.itemId).Return(testCase.expectedItem, testCase.expectedErr)

			actualItem, err := itemUseCase.GetOneById(context.Background(), testCase.userId, testCase.itemId)

			assert.Equal(t, testCase.expectedItem, actualItem)
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

// TestUpdateItemById tests the UpdateItemById function in the ItemUseCase
func TestUpdateItemById(t *testing.T) {
	title := "Updated Item"

	testCases := []struct {
		name             string
		userId           int
		listId           int
		itemId           int
		input            entity.UpdateItemInput
		expectedOwnerErr error
		expectedErr      error
	}{
		{
			name:             "Successful update",
			userId:           1,
			listId:           3,
			itemId:           1,
			input:            entity.UpdateItemInput{Title: &title},
			expectedOwnerErr: nil,
			expectedErr:      nil,
		},
		{
			name:             "Item not belongs to user",
			userId:           1,
			listId:           3,
			itemId:           99,
			input:            entity.UpdateItemInput{Title: &title},
			expectedOwnerErr: utils.ErrUserNotOwner,
			expectedErr:      nil,
		},
		{
			name:             "Item not found",
			userId:           1,
			listId:           3,
			itemId:           99,
			input:            entity.UpdateItemInput{Title: &title},
			expectedOwnerErr: nil,
			expectedErr:      utils.ErrItemNotFound,
		},
	}

	for _, testCase := range testCases {
		mockItemRepo := new(MockItemRepo)
		mockListRepo := new(MockListRepo)
		mockItemSearch := new(MockItemSearch)
		mockBrokerProducer := new(MockBrokerProducer)
		itemUseCase := usecase.NewItemUseCase(mockItemRepo, mockListRepo, mockItemSearch, mockBrokerProducer, &logger.NoOpLogger{})

		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockItemRepo.On("IsUserOwnerOfItem", mock.Anything, testCase.userId, testCase.itemId).Return(testCase.expectedOwnerErr)
			mockItemRepo.On("UpdateOneById", mock.Anything, &testCase.listId, testCase.itemId, testCase.input).Return(testCase.expectedErr)

			err := itemUseCase.UpdateOneById(context.Background(), testCase.userId, testCase.listId, testCase.itemId, testCase.input)

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

// TestDeleteItemById tests the DeleteItemById function in the ItemUseCase
func TestDeleteItemById(t *testing.T) {
	testCases := []struct {
		name             string
		userId           int
		listId           int
		itemId           int
		expectedOwnerErr error
		expectedErr      error
	}{
		{
			name:             "Successful deletion",
			userId:           1,
			listId:           3,
			itemId:           1,
			expectedOwnerErr: nil,
			expectedErr:      nil,
		},
		{
			name:             "Item not belongs to user",
			userId:           1,
			listId:           3,
			itemId:           99,
			expectedOwnerErr: utils.ErrUserNotOwner,
			expectedErr:      nil,
		},
		{
			name:             "Item not found",
			userId:           1,
			listId:           3,
			itemId:           99,
			expectedOwnerErr: nil,
			expectedErr:      utils.ErrItemNotFound,
		},
	}

	for _, testCase := range testCases {
		mockItemRepo := new(MockItemRepo)
		mockListRepo := new(MockListRepo)
		mockItemSearch := new(MockItemSearch)
		mockBrokerProducer := new(MockBrokerProducer)
		itemUseCase := usecase.NewItemUseCase(mockItemRepo, mockListRepo, mockItemSearch, mockBrokerProducer, &logger.NoOpLogger{})

		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockItemRepo.On("IsUserOwnerOfItem", mock.Anything, testCase.userId, testCase.itemId).Return(testCase.expectedOwnerErr)
			mockItemRepo.On("DeleteOneById", mock.Anything, &testCase.listId, testCase.itemId).Return(testCase.expectedErr)

			err := itemUseCase.DeleteOneById(context.Background(), testCase.userId, testCase.listId, testCase.itemId)

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

// TestDeleteItemByAdmin tests the DeleteItemByAdmin function in the ItemUseCase
func TestDeleteItemByAdmin(t *testing.T) {
	testCases := []struct {
		name        string
		itemId      int
		expectedErr error
	}{
		{
			name:        "Successful deletion",
			itemId:      1,
			expectedErr: nil,
		},
		{
			name:        "Item not found",
			itemId:      99,
			expectedErr: utils.ErrItemNotFound,
		},
	}

	for _, testCase := range testCases {
		mockItemRepo := new(MockItemRepo)
		mockListRepo := new(MockListRepo)
		mockItemSearch := new(MockItemSearch)
		mockBrokerProducer := new(MockBrokerProducer)
		itemUseCase := usecase.NewItemUseCase(mockItemRepo, mockListRepo, mockItemSearch, mockBrokerProducer, &logger.NoOpLogger{})

		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var listId *int
			mockItemRepo.On("DeleteOneById", mock.Anything, listId, testCase.itemId).Return(testCase.expectedErr)

			err := itemUseCase.DeleteOneByAdmin(context.Background(), testCase.itemId)

			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
