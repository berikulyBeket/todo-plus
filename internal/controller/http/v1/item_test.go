package v1_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "github.com/berikulyBeket/todo-plus/internal/controller/http/v1"
	"github.com/berikulyBeket/todo-plus/internal/entity"
	"github.com/berikulyBeket/todo-plus/internal/middleware"
	"github.com/berikulyBeket/todo-plus/internal/usecase"
	"github.com/berikulyBeket/todo-plus/pkg/logger"
	"github.com/berikulyBeket/todo-plus/pkg/metrics"
	"github.com/berikulyBeket/todo-plus/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestHandler_CreateItem tests the CreateItem handler
func TestHandler_CreateItem(t *testing.T) {
	testCases := []struct {
		name           string
		userId         int
		listId         string
		input          string
		mockBehavior   func(mockList *MockList, mockItem *MockItem)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			userId: 1,
			listId: "2",
			input: `{
				"title": "New Item",
				"description": "Item description"
			}`,
			mockBehavior: func(mockList *MockList, mockItem *MockItem) {
				mockList.On("GetOneById", mock.Anything, 1, 2).Return(entity.List{Id: 2}, nil)
				mockItem.On("Create", mock.Anything, 1, 2, &entity.Item{
					Title:       "New Item",
					Description: "Item description",
				}).Return(1, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: `{
				"status": "ok",
				"message": "Item created successfully",
				"data": {"id": 1}
			}`,
		},
		{
			name:   "Invalid listId parameter",
			userId: 1,
			listId: "invalid", // Invalid listId
			input: `{
				"title": "New Item",
				"description": "Item description"
			}`,
			mockBehavior:   func(mockList *MockList, mockItem *MockItem) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Invalid listId param",
				"errors": {"param": "listId must be a valid integer"}
			}`,
		},
		{
			name:           "Invalid JSON input",
			userId:         1,
			listId:         "2",
			input:          `invalid json`,
			mockBehavior:   func(mockList *MockList, mockItem *MockItem) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Invalid input",
				"errors": {"body": "Invalid or malformed JSON"}
			}`,
		},
		{
			name:   "List not found",
			userId: 1,
			listId: "2",
			input: `{
				"title": "New Item",
				"description": "Item description"
			}`,
			mockBehavior: func(mockList *MockList, mockItem *MockItem) {
				mockList.On("GetOneById", mock.Anything, 1, 2).Return(entity.List{}, utils.ErrListNotFound)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Validation failed",
				"errors": {"validation": "list not found"}
			}`,
		},
		{
			name:   "Internal server error on item creation",
			userId: 1,
			listId: "2",
			input: `{
				"title": "New Item",
				"description": "Item description"
			}`,
			mockBehavior: func(mockList *MockList, mockItem *MockItem) {
				mockList.On("GetOneById", mock.Anything, 1, 2).Return(entity.List{Id: 2}, nil)
				mockItem.On("Create", mock.Anything, 1, 2, &entity.Item{
					Title:       "New Item",
					Description: "Item description",
				}).Return(0, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: `{
				"status": "error",
				"message": "Failed to create item",
				"errors": {"database": "Error during item creation"}
			}`,
		},
		{
			name:   "Unauthorized",
			userId: 0,
			listId: "2",
			input: `{
				"title": "New Item",
				"description": "Item description"
			}`,
			mockBehavior:   func(mockList *MockList, mockItem *MockItem) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: `{
				"status": "error",
				"message": "Unauthorized",
				"errors": {"auth": "User authentication failed or user not logged in"}
			}`,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockList := new(MockList)
			mockItem := new(MockItem)
			appAuth := new(MockAppAuth)
			noOpLogger := &logger.NoOpLogger{}
			noOpMetrics := &metrics.NoOpMetrics{}

			mockUseCase := &usecase.UseCase{
				List: mockList,
				Item: mockItem,
			}

			handler := v1.NewHandler(mockUseCase, appAuth, noOpLogger, noOpMetrics)

			gin.SetMode(gin.TestMode)
			r := gin.Default()

			r.Use(func(c *gin.Context) {
				if testCase.userId != 0 {
					c.Set(middleware.UserIdCtx, testCase.userId)
				}
				c.Next()
			})

			r.POST("/api/lists/:id/items", handler.CreateItem)

			req := httptest.NewRequest("POST", "/api/lists/"+testCase.listId+"/items", bytes.NewBufferString(testCase.input))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			testCase.mockBehavior(mockList, mockItem)

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatus, w.Code)
			assert.JSONEq(t, testCase.expectedBody, w.Body.String())

			mockList.AssertExpectations(t)
			mockItem.AssertExpectations(t)
		})
	}
}

// TestHandler_GetAllItems tests the GetAllItems handler
func TestHandler_GetAllItems(t *testing.T) {
	testCases := []struct {
		name           string
		userId         int
		listId         string
		mockBehavior   func(mockList *MockList, mockItem *MockItem)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			userId: 1,
			listId: "2",
			mockBehavior: func(mockList *MockList, mockItem *MockItem) {
				mockList.On("GetOneById", mock.Anything, 1, 2).Return(entity.List{Id: 2}, nil)
				mockItem.On("GetAll", mock.Anything, 1, 2).Return([]entity.Item{
					{
						Id:          1,
						Title:       "Item 1",
						Description: "Description 1",
						Done:        false,
					},
					{
						Id:          2,
						Title:       "Item 2",
						Description: "Description 2",
						Done:        true,
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"status": "ok",
				"message": "Items retrieved successfully",
				"data": [
					{
						"id": 1,
						"title": "Item 1",
						"description": "Description 1",
						"done": false
					},
					{
						"id": 2,
						"title": "Item 2",
						"description": "Description 2",
						"done": true
					}
				]
			}`,
		},
		{
			name:           "Invalid listId parameter",
			userId:         1,
			listId:         "invalid",
			mockBehavior:   func(mockList *MockList, mockItem *MockItem) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Invalid listId param",
				"errors": {"param": "listId must be a valid integer"}
			}`,
		},
		{
			name:   "List not found",
			userId: 1,
			listId: "2",
			mockBehavior: func(mockList *MockList, mockItem *MockItem) {
				mockList.On("GetOneById", mock.Anything, 1, 2).Return(entity.List{}, utils.ErrListNotFound)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Validation failed",
				"errors": {"validation": "list not found"}
			}`,
		},
		{
			name:   "Internal server error during item retrieval",
			userId: 1,
			listId: "2",
			mockBehavior: func(mockList *MockList, mockItem *MockItem) {
				mockList.On("GetOneById", mock.Anything, 1, 2).Return(entity.List{Id: 2}, nil)
				mockItem.On("GetAll", mock.Anything, 1, 2).Return([]entity.Item{}, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: `{
				"status": "error",
				"message": "Failed to retrieve items",
				"errors": {"database": "Error during item retrieval"}
			}`,
		},

		{
			name:           "Unauthorized",
			userId:         0,
			listId:         "2",
			mockBehavior:   func(mockList *MockList, mockItem *MockItem) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: `{
				"status": "error",
				"message": "Unauthorized",
				"errors": {"auth": "User authentication failed or user not logged in"}
			}`,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockList := new(MockList)
			mockItem := new(MockItem)
			appAuth := new(MockAppAuth)
			noOpLogger := &logger.NoOpLogger{}
			noOpMetrics := &metrics.NoOpMetrics{}

			mockUseCase := &usecase.UseCase{
				List: mockList,
				Item: mockItem,
			}

			handler := v1.NewHandler(mockUseCase, appAuth, noOpLogger, noOpMetrics)

			gin.SetMode(gin.TestMode)
			r := gin.Default()

			r.Use(func(c *gin.Context) {
				if testCase.userId != 0 {
					c.Set(middleware.UserIdCtx, testCase.userId)
				}
				c.Next()
			})

			r.GET("/api/lists/:id/items", handler.GetAllItems)

			req := httptest.NewRequest("GET", "/api/lists/"+testCase.listId+"/items", nil)
			w := httptest.NewRecorder()

			testCase.mockBehavior(mockList, mockItem)

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatus, w.Code)
			assert.JSONEq(t, testCase.expectedBody, w.Body.String())

			mockList.AssertExpectations(t)
			mockItem.AssertExpectations(t)
		})
	}
}

// TestHandler_GetItemById tests the GetItemById handler
func TestHandler_GetItemById(t *testing.T) {
	testCases := []struct {
		name           string
		userId         int
		itemId         string
		mockBehavior   func(mockItem *MockItem)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			userId: 1,
			itemId: "2",
			mockBehavior: func(mockItem *MockItem) {
				mockItem.On("GetOneById", mock.Anything, 1, 2).Return(entity.Item{
					Id:          2,
					Title:       "Item 1",
					Description: "Description 1",
					Done:        false,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"status": "ok",
				"message": "Item retrieved successfully",
				"data": {
					"id": 2,
					"title": "Item 1",
					"description": "Description 1",
					"done": false
				}
			}`,
		},
		{
			name:           "Invalid itemId parameter",
			userId:         1,
			itemId:         "invalid",
			mockBehavior:   func(mockItem *MockItem) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Invalid itemId param",
				"errors": {"param": "itemId must be a valid integer"}
			}`,
		},
		{
			name:   "Item not belongs to user",
			userId: 1,
			itemId: "2",
			mockBehavior: func(mockItem *MockItem) {
				mockItem.On("GetOneById", mock.Anything, 1, 2).Return(entity.Item{}, utils.ErrUserNotOwner)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: `{
				"status": "error",
				"message": "Item not found",
				"errors": {"itemId": "The requested item does not exist"}
			}`,
		},
		{
			name:   "Item not found",
			userId: 1,
			itemId: "2",
			mockBehavior: func(mockItem *MockItem) {
				mockItem.On("GetOneById", mock.Anything, 1, 2).Return(entity.Item{}, utils.ErrItemNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: `{
				"status": "error",
				"message": "Item not found",
				"errors": {"itemId": "The requested item does not exist"}
			}`,
		},
		{
			name:   "Internal server error",
			userId: 1,
			itemId: "2",
			mockBehavior: func(mockItem *MockItem) {
				mockItem.On("GetOneById", mock.Anything, 1, 2).Return(entity.Item{}, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: `{
				"status": "error",
				"message": "Failed to retrieve item",
				"errors": {"database": "Error during item retrieval"}
			}`,
		},
		{
			name:           "Unauthorized",
			userId:         0,
			itemId:         "2",
			mockBehavior:   func(mockItem *MockItem) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: `{
				"status": "error",
				"message": "Unauthorized",
				"errors": {"auth": "User authentication failed or user not logged in"}
			}`,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockItem := new(MockItem)
			appAuth := new(MockAppAuth)
			noOpLogger := &logger.NoOpLogger{}
			noOpMetrics := &metrics.NoOpMetrics{}

			mockUseCase := &usecase.UseCase{
				Item: mockItem,
			}

			handler := v1.NewHandler(mockUseCase, appAuth, noOpLogger, noOpMetrics)

			gin.SetMode(gin.TestMode)
			r := gin.Default()

			r.Use(func(c *gin.Context) {
				if testCase.userId != 0 {
					c.Set(middleware.UserIdCtx, testCase.userId)
				}
				c.Next()
			})

			r.GET("/api/items/:id", handler.GetItemById)

			req := httptest.NewRequest("GET", "/api/items/"+testCase.itemId, nil)
			w := httptest.NewRecorder()

			testCase.mockBehavior(mockItem)

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatus, w.Code)
			assert.JSONEq(t, testCase.expectedBody, w.Body.String())

			mockItem.AssertExpectations(t)
		})
	}
}

// TestHandler_UpdateItem tests the UpdateItem handler
func TestHandler_UpdateItem(t *testing.T) {
	title := "Updated Title"
	description := "Updated Description"

	testCases := []struct {
		name           string
		userId         int
		listId         string
		itemId         string
		input          string
		mockBehavior   func(mockItem *MockItem)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			userId: 1,
			listId: "3",
			itemId: "2",
			input:  `{"title": "Updated Title", "description": "Updated Description"}`,
			mockBehavior: func(mockItem *MockItem) {
				mockItem.On("UpdateOneById", mock.Anything, 1, 3, 2, entity.UpdateItemInput{
					Title:       &title,
					Description: &description,
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"status": "ok",
				"message": "Item updated successfully"
			}`,
		},
		{
			name:           "Invalid itemId parameter",
			userId:         1,
			listId:         "3",
			itemId:         "invalid",
			input:          `{"title": "Updated Title", "description": "Updated Description"}`,
			mockBehavior:   func(mockItem *MockItem) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Invalid itemId param",
				"errors": {"param": "itemId must be a valid integer"}
			}`,
		},
		{
			name:           "Invalid itemId parameter",
			userId:         1,
			listId:         "",
			itemId:         "2",
			input:          `{"title": "Updated Title", "description": "Updated Description"}`,
			mockBehavior:   func(mockItem *MockItem) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Invalid listId param",
				"errors": {"param": "listId must be a valid integer"}
			}`,
		},
		{
			name:           "Invalid JSON input",
			userId:         1,
			listId:         "3",
			itemId:         "2",
			input:          `invalid-json`,
			mockBehavior:   func(mockItem *MockItem) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Invalid input",
				"errors": {"body": "Invalid or malformed JSON"}
			}`,
		},
		{
			name:   "Item not belongs to user",
			userId: 1,
			listId: "3",
			itemId: "2",
			input:  `{"title": "Updated Title"}`,
			mockBehavior: func(mockItem *MockItem) {
				mockItem.On("UpdateOneById", mock.Anything, 1, 3, 2, entity.UpdateItemInput{
					Title: &title,
				}).Return(utils.ErrUserNotOwner)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: `{
				"status": "error",
				"message": "Item not found",
				"errors": {"itemId": "The requested item does not exist"}
			}`,
		},
		{
			name:   "Item not found",
			userId: 1,
			listId: "3",
			itemId: "2",
			input:  `{"title": "Updated Title"}`,
			mockBehavior: func(mockItem *MockItem) {
				mockItem.On("UpdateOneById", mock.Anything, 1, 3, 2, entity.UpdateItemInput{
					Title: &title,
				}).Return(utils.ErrItemNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: `{
				"status": "error",
				"message": "Item not found",
				"errors": {"itemId": "The requested item does not exist"}
			}`,
		},
		{
			name:   "Internal server error",
			userId: 1,
			listId: "3",
			itemId: "2",
			input:  `{"title": "Updated Title"}`,
			mockBehavior: func(mockItem *MockItem) {
				mockItem.On("UpdateOneById", mock.Anything, 1, 3, 2, entity.UpdateItemInput{
					Title: &title,
				}).Return(errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: `{
				"status": "error",
				"message": "Failed to update item",
				"errors": {"database": "Error during item update"}
			}`,
		},
		{
			name:           "Unauthorized",
			userId:         0,
			listId:         "3",
			itemId:         "2",
			input:          `{"title": "Updated Title"}`,
			mockBehavior:   func(mockItem *MockItem) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: `{
				"status": "error",
				"message": "Unauthorized",
				"errors": {"auth": "User authentication failed or user not logged in"}
			}`,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockItem := new(MockItem)
			appAuth := new(MockAppAuth)
			noOpLogger := &logger.NoOpLogger{}
			noOpMetrics := &metrics.NoOpMetrics{}

			mockUseCase := &usecase.UseCase{
				Item: mockItem,
			}

			handler := v1.NewHandler(mockUseCase, appAuth, noOpLogger, noOpMetrics)

			gin.SetMode(gin.TestMode)
			r := gin.Default()

			r.Use(func(c *gin.Context) {
				if testCase.userId != 0 {
					c.Set(middleware.UserIdCtx, testCase.userId)
				}
				c.Next()
			})

			r.PUT("/api/items/:id", handler.UpdateItem)

			req := httptest.NewRequest("PUT", "/api/items/"+testCase.itemId+"?list_id="+testCase.listId, bytes.NewBufferString(testCase.input))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			testCase.mockBehavior(mockItem)

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatus, w.Code)
			assert.JSONEq(t, testCase.expectedBody, w.Body.String())

			mockItem.AssertExpectations(t)
		})
	}
}

// TestHandler_DeleteItem tests the DeleteItem handler
func TestHandler_DeleteItem(t *testing.T) {
	testCases := []struct {
		name           string
		userId         int
		listId         string
		itemId         string
		mockBehavior   func(mockItem *MockItem)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			userId: 1,
			listId: "3",
			itemId: "2",
			mockBehavior: func(mockItem *MockItem) {
				mockItem.On("DeleteOneById", mock.Anything, 1, 3, 2).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"status": "ok",
				"message": "Item deleted successfully"
			}`,
		},
		{
			name:           "Invalid itemId parameter",
			userId:         1,
			listId:         "3",
			itemId:         "invalid",
			mockBehavior:   func(mockItem *MockItem) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Invalid itemId param",
				"errors": {"param": "itemId must be a valid integer"}
			}`,
		},
		{
			name:           "Invalid listId parameter",
			userId:         1,
			listId:         "",
			itemId:         "2",
			mockBehavior:   func(mockItem *MockItem) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Invalid listId param",
				"errors": {"param": "listId must be a valid integer"}
			}`,
		},
		{
			name:   "Item not belongs to user",
			userId: 1,
			listId: "3",
			itemId: "2",
			mockBehavior: func(mockItem *MockItem) {
				mockItem.On("DeleteOneById", mock.Anything, 1, 3, 2).Return(utils.ErrUserNotOwner)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: `{
				"status": "error",
				"message": "Item not found",
				"errors": {"itemId": "The requested item does not exist"}
			}`,
		},
		{
			name:   "Item not found",
			userId: 1,
			listId: "3",
			itemId: "2",
			mockBehavior: func(mockItem *MockItem) {
				mockItem.On("DeleteOneById", mock.Anything, 1, 3, 2).Return(utils.ErrItemNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: `{
				"status": "error",
				"message": "Item not found",
				"errors": {"itemId": "The requested item does not exist"}
			}`,
		},
		{
			name:   "Internal server error",
			userId: 1,
			listId: "3",
			itemId: "2",
			mockBehavior: func(mockItem *MockItem) {
				mockItem.On("DeleteOneById", mock.Anything, 1, 3, 2).Return(errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: `{
				"status": "error",
				"message": "Failed to delete item",
				"errors": {"database": "Error during item deletion"}
			}`,
		},
		{
			name:           "Unauthorized",
			userId:         0,
			listId:         "3",
			itemId:         "2",
			mockBehavior:   func(mockItem *MockItem) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: `{
				"status": "error",
				"message": "Unauthorized",
				"errors": {"auth": "User authentication failed or user not logged in"}
			}`,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockItem := new(MockItem)
			appAuth := new(MockAppAuth)
			noOpLogger := &logger.NoOpLogger{}
			noOpMetrics := &metrics.NoOpMetrics{}

			mockUseCase := &usecase.UseCase{
				Item: mockItem,
			}

			handler := v1.NewHandler(mockUseCase, appAuth, noOpLogger, noOpMetrics)

			gin.SetMode(gin.TestMode)
			r := gin.Default()

			r.Use(func(c *gin.Context) {
				if testCase.userId != 0 {
					c.Set(middleware.UserIdCtx, testCase.userId)
				}
				c.Next()
			})

			r.DELETE("/api/items/:id", handler.DeleteItem)

			req := httptest.NewRequest("DELETE", "/api/items/"+testCase.itemId+"?list_id="+testCase.listId, nil)
			w := httptest.NewRecorder()

			testCase.mockBehavior(mockItem)

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatus, w.Code)
			assert.JSONEq(t, testCase.expectedBody, w.Body.String())

			mockItem.AssertExpectations(t)
		})
	}
}

// TestHandler_DeleteItemByAdmin tests the DeleteItemByAdmin handler
func TestHandler_DeleteItemByAdmin(t *testing.T) {
	testCases := []struct {
		name           string
		itemId         string
		mockBehavior   func(mockItem *MockItem)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			itemId: "2",
			mockBehavior: func(mockItem *MockItem) {
				mockItem.On("DeleteOneByAdmin", mock.Anything, 2).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"status": "ok",
				"message": "Item deleted successfully"
			}`,
		},
		{
			name:           "Invalid itemId parameter",
			itemId:         "invalid",
			mockBehavior:   func(mockItem *MockItem) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Invalid itemId param",
				"errors": {"param": "itemId must be a valid integer"}
			}`,
		},
		{
			name:   "Item not found",
			itemId: "2",
			mockBehavior: func(mockItem *MockItem) {
				mockItem.On("DeleteOneByAdmin", mock.Anything, 2).Return(utils.ErrItemNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: `{
				"status": "error",
				"message": "Item not found",
				"errors": {"itemId": "The requested item does not exist"}
			}`,
		},
		{
			name:   "Internal server error",
			itemId: "2",
			mockBehavior: func(mockItem *MockItem) {
				mockItem.On("DeleteOneByAdmin", mock.Anything, 2).Return(errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: `{
				"status": "error",
				"message": "Failed to delete item",
				"errors": {"database": "Error during item deletion"}
			}`,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockItem := new(MockItem)
			appAuth := new(MockAppAuth)
			noOpLogger := &logger.NoOpLogger{}
			noOpMetrics := &metrics.NoOpMetrics{}

			mockUseCase := &usecase.UseCase{
				Item: mockItem,
			}

			handler := v1.NewHandler(mockUseCase, appAuth, noOpLogger, noOpMetrics)

			gin.SetMode(gin.TestMode)
			r := gin.Default()

			r.DELETE("/private/api/items/:id", handler.DeleteItemByAdmin)

			req := httptest.NewRequest("DELETE", "/private/api/items/"+testCase.itemId, nil)
			w := httptest.NewRecorder()

			testCase.mockBehavior(mockItem)

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatus, w.Code)
			assert.JSONEq(t, testCase.expectedBody, w.Body.String())

			mockItem.AssertExpectations(t)
		})
	}
}
