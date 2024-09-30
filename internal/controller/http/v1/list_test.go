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

// TestHandler_CreateList tests the CreateList handler
func TestHandler_CreateList(t *testing.T) {
	testCases := []struct {
		name           string
		userId         int
		input          string
		mockBehavior   func(mockList *MockList)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			userId: 1,
			input: `{
				"title": "Groceries",
				"description": "Weekly grocery shopping list"
			}`,
			mockBehavior: func(mockList *MockList) {
				mockList.On("Create", mock.Anything, 1, &entity.List{
					Title:       "Groceries",
					Description: "Weekly grocery shopping list",
				}).Return(1, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: `{
				"status": "ok",
				"message": "List created successfully",
				"data": {"id": 1}
			}`,
		},
		{
			name:   "Invalid input",
			userId: 1,
			input: `{
				"description": "Weekly grocery shopping list"
			}`,
			mockBehavior:   func(mockList *MockList) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Invalid input",
				"errors": {"body": "Invalid or malformed JSON"}
			}`,
		},
		{
			name:   "Unauthorized",
			userId: 0,
			input: `{
				"title": "Groceries",
				"description": "Weekly grocery shopping list"
			}`,
			mockBehavior:   func(mockList *MockList) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: `{
				"status": "error",
				"message": "Unauthorized",
				"errors": {"auth": "User authentication failed or user not logged in"}
			}`,
		},
		{
			name:   "Internal server error",
			userId: 1,
			input: `{
				"title": "Groceries",
				"description": "Weekly grocery shopping list"
			}`,
			mockBehavior: func(mockList *MockList) {
				mockList.On("Create", mock.Anything, 1, &entity.List{
					Title:       "Groceries",
					Description: "Weekly grocery shopping list",
				}).Return(0, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: `{
				"status": "error",
				"message": "Failed to create list",
				"errors": {"database": "Error during list creation"}
			}`,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockList := new(MockList)
			appAuth := new(MockAppAuth)
			noOpLogger := &logger.NoOpLogger{}
			noOpMetrics := &metrics.NoOpMetrics{}

			mockUseCase := &usecase.UseCase{
				List: mockList,
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

			r.POST("/api/lists/", handler.CreateList)

			req := httptest.NewRequest("POST", "/api/lists/", bytes.NewBufferString(testCase.input))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			testCase.mockBehavior(mockList)

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatus, w.Code)
			assert.JSONEq(t, testCase.expectedBody, w.Body.String())

			mockList.AssertExpectations(t)
		})
	}
}

// TestHandler_GetAllLists tests the GetAllLists handler
func TestHandler_GetAllLists(t *testing.T) {
	testCases := []struct {
		name           string
		userId         int
		mockBehavior   func(mockList *MockList)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			userId: 1,
			mockBehavior: func(mockList *MockList) {
				mockList.On("GetAll", mock.Anything, 1).
					Return([]entity.List{
						{Id: 1, Title: "Groceries", Description: "Weekly groceries list"},
						{Id: 2, Title: "Work", Description: "Work tasks for the week"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"status": "ok",
				"message": "Lists retrieved successfully",
				"data": [
					{"id": 1, "title": "Groceries", "description": "Weekly groceries list"},
					{"id": 2, "title": "Work", "description": "Work tasks for the week"}
				]
			}`,
		},
		{
			name:   "No lists found",
			userId: 1,
			mockBehavior: func(mockList *MockList) {
				mockList.On("GetAll", mock.Anything, 1).
					Return([]entity.List{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"status": "ok",
				"message": "Lists retrieved successfully",
				"data": []
			}`,
		},
		{
			name:           "Unauthorized",
			userId:         0,
			mockBehavior:   func(mockList *MockList) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: `{
				"status": "error",
				"message": "Unauthorized",
				"errors": {"auth": "User authentication failed or user not logged in"}
			}`,
		},
		{
			name:   "Internal server error",
			userId: 1,
			mockBehavior: func(mockList *MockList) {
				mockList.On("GetAll", mock.Anything, 1).
					Return([]entity.List{}, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: `{
				"status": "error",
				"message": "Failed to retrieve lists",
				"errors": {"database": "Error during list retrieval"}
			}`,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockList := new(MockList)
			appAuth := new(MockAppAuth)
			noOpLogger := &logger.NoOpLogger{}
			noOpMetrics := &metrics.NoOpMetrics{}

			mockUseCase := &usecase.UseCase{
				List: mockList,
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

			r.GET("/api/lists/", handler.GetAllLists)

			req := httptest.NewRequest("GET", "/api/lists/", nil)
			w := httptest.NewRecorder()

			testCase.mockBehavior(mockList)

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatus, w.Code)
			assert.JSONEq(t, testCase.expectedBody, w.Body.String())

			mockList.AssertExpectations(t)
		})
	}
}

// TestHandler_GetListById tests the GetListById handler
func TestHandler_GetListById(t *testing.T) {
	testCases := []struct {
		name           string
		userId         int
		listId         string
		mockBehavior   func(mockList *MockList)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			userId: 1,
			listId: "2",
			mockBehavior: func(mockList *MockList) {
				mockList.On("GetOneById", mock.Anything, 1, 2).Return(entity.List{
					Id:          2,
					Title:       "Groceries",
					Description: "Weekly groceries list",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"status": "ok",
				"message": "List retrieved successfully",
				"data": {
					"id": 2,
					"title": "Groceries",
					"description": "Weekly groceries list"
				}
			}`,
		},
		{
			name:           "Invalid listId parameter",
			userId:         1,
			listId:         "invalid",
			mockBehavior:   func(mockList *MockList) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Invalid listId param",
				"errors": {"param": "listId must be a valid integer"}
			}`,
		},
		{
			name:   "List not belongs to user",
			userId: 1,
			listId: "2",
			mockBehavior: func(mockList *MockList) {
				mockList.On("GetOneById", mock.Anything, 1, 2).Return(entity.List{}, utils.ErrUserNotOwner)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: `{
				"status": "error",
				"message": "List not found",
				"errors": {"listId": "The requested list does not exist"}
			}`,
		},
		{
			name:   "List not found",
			userId: 1,
			listId: "2",
			mockBehavior: func(mockList *MockList) {
				mockList.On("GetOneById", mock.Anything, 1, 2).Return(entity.List{}, utils.ErrListNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: `{
				"status": "error",
				"message": "List not found",
				"errors": {"listId": "The requested list does not exist"}
			}`,
		},
		{
			name:   "Internal server error",
			userId: 1,
			listId: "2",
			mockBehavior: func(mockList *MockList) {
				mockList.On("GetOneById", mock.Anything, 1, 2).Return(entity.List{}, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: `{
				"status": "error",
				"message": "Failed to retrieve list",
				"errors": {"database": "Error during list retrieval"}
			}`,
		},
		{
			name:           "Unauthorized",
			userId:         0,
			listId:         "2",
			mockBehavior:   func(mockList *MockList) {},
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
			appAuth := new(MockAppAuth)
			noOpLogger := &logger.NoOpLogger{}
			noOpMetrics := &metrics.NoOpMetrics{}

			mockUseCase := &usecase.UseCase{
				List: mockList,
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

			r.GET("/api/lists/:id", handler.GetListById)

			req := httptest.NewRequest("GET", "/api/lists/"+testCase.listId, nil)
			w := httptest.NewRecorder()

			testCase.mockBehavior(mockList)

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatus, w.Code)
			assert.JSONEq(t, testCase.expectedBody, w.Body.String())

			mockList.AssertExpectations(t)
		})
	}
}

// TestHandler_UpdateList tests the UpdateList handler
func TestHandler_UpdateList(t *testing.T) {
	title := "Updated title"
	description := "Updated description"

	testCases := []struct {
		name           string
		userId         int
		listId         string
		input          string
		mockBehavior   func(mockList *MockList)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			userId: 1,
			listId: "2",
			input: `{
				"title": "Updated title",
				"description": "Updated description"
			}`,
			mockBehavior: func(mockList *MockList) {
				mockList.On("UpdateOneById", mock.Anything, 1, 2, entity.UpdateListInput{
					Title:       &title,
					Description: &description,
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"status": "ok",
				"message": "List updated successfully"
			}`,
		},
		{
			name:   "Invalid listId parameter",
			userId: 1,
			listId: "invalid",
			input: `{
				"title": "Updated title"
			}`,
			mockBehavior:   func(mockList *MockList) {},
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
			mockBehavior:   func(mockList *MockList) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Invalid input",
				"errors": {"body": "Invalid or malformed JSON"}
			}`,
		},
		{
			name:           "No update data provided",
			userId:         1,
			listId:         "2",
			input:          `{}`,
			mockBehavior:   func(mockList *MockList) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Validation failed",
				"errors": {"validation": "list update structure has no values"}
			}`,
		},
		{
			name:   "List not belongs to user",
			userId: 1,
			listId: "2",
			input: `{
				"title": "Updated title"
			}`,
			mockBehavior: func(mockList *MockList) {
				mockList.On("UpdateOneById", mock.Anything, 1, 2, entity.UpdateListInput{
					Title: &title,
				}).Return(utils.ErrUserNotOwner)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: `{
				"status": "error",
				"message": "List not found",
				"errors": {"listId": "The requested list does not exist"}
			}`,
		},
		{
			name:   "List not found",
			userId: 1,
			listId: "2",
			input: `{
				"title": "Updated title"
			}`,
			mockBehavior: func(mockList *MockList) {
				mockList.On("UpdateOneById", mock.Anything, 1, 2, entity.UpdateListInput{
					Title: &title,
				}).Return(utils.ErrListNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: `{
				"status": "error",
				"message": "List not found",
				"errors": {"listId": "The requested list does not exist"}
			}`,
		},
		{
			name:   "Internal server error",
			userId: 1,
			listId: "2",
			input: `{
				"title": "Updated title"
			}`,
			mockBehavior: func(mockList *MockList) {
				mockList.On("UpdateOneById", mock.Anything, 1, 2, entity.UpdateListInput{
					Title: &title,
				}).Return(errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: `{
				"status": "error",
				"message": "Failed to update list",
				"errors": {"database": "Error during list update"}
			}`,
		},
		{
			name:   "Unauthorized",
			userId: 0,
			listId: "2",
			input: `{
				"title": "Updated title"
			}`,
			mockBehavior:   func(mockList *MockList) {},
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
			appAuth := new(MockAppAuth)
			noOpLogger := &logger.NoOpLogger{}
			noOpMetrics := &metrics.NoOpMetrics{}

			mockUseCase := &usecase.UseCase{
				List: mockList,
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

			r.PUT("/api/lists/:id", handler.UpdateList)

			req := httptest.NewRequest("PUT", "/api/lists/"+testCase.listId, bytes.NewBufferString(testCase.input))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			testCase.mockBehavior(mockList)

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatus, w.Code)
			assert.JSONEq(t, testCase.expectedBody, w.Body.String())

			mockList.AssertExpectations(t)
		})
	}
}

// TestHandler_DeleteList tests the DeleteList handler
func TestHandler_DeleteList(t *testing.T) {
	testCases := []struct {
		name           string
		userId         int
		listId         string
		mockBehavior   func(mockList *MockList)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			userId: 1,
			listId: "2",
			mockBehavior: func(mockList *MockList) {
				mockList.On("DeleteOneById", mock.Anything, 1, 2).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"status": "ok",
				"message": "List deleted successfully"
			}`,
		},
		{
			name:           "Invalid listId parameter",
			userId:         1,
			listId:         "invalid",
			mockBehavior:   func(mockList *MockList) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Invalid listId param",
				"errors": {"param": "listId must be a valid integer"}
			}`,
		},
		{
			name:   "List not belongs to user",
			userId: 1,
			listId: "2",
			mockBehavior: func(mockList *MockList) {
				mockList.On("DeleteOneById", mock.Anything, 1, 2).Return(utils.ErrUserNotOwner)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: `{
				"status": "error",
				"message": "List not found",
				"errors": {"listId": "The requested list does not exist"}
			}`,
		},
		{
			name:   "List not found",
			userId: 1,
			listId: "2",
			mockBehavior: func(mockList *MockList) {
				mockList.On("DeleteOneById", mock.Anything, 1, 2).Return(utils.ErrListNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: `{
				"status": "error",
				"message": "List not found",
				"errors": {"listId": "The requested list does not exist"}
			}`,
		},
		{
			name:   "Internal server error",
			userId: 1,
			listId: "2",
			mockBehavior: func(mockList *MockList) {
				mockList.On("DeleteOneById", mock.Anything, 1, 2).Return(errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: `{
				"status": "error",
				"message": "Failed to delete list",
				"errors": {"database": "Error during list deletion"}
			}`,
		},
		{
			name:           "Unauthorized",
			userId:         0,
			listId:         "2",
			mockBehavior:   func(mockList *MockList) {},
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
			appAuth := new(MockAppAuth)
			noOpLogger := &logger.NoOpLogger{}
			noOpMetrics := &metrics.NoOpMetrics{}

			mockUseCase := &usecase.UseCase{
				List: mockList,
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

			r.DELETE("/api/lists/:id", handler.DeleteList)

			req := httptest.NewRequest("DELETE", "/api/lists/"+testCase.listId, nil)
			w := httptest.NewRecorder()

			testCase.mockBehavior(mockList)

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatus, w.Code)
			assert.JSONEq(t, testCase.expectedBody, w.Body.String())

			mockList.AssertExpectations(t)
		})
	}
}

// TestHandler_DeleteListByAdmin tests the DeleteListByAdmin handler
func TestHandler_DeleteListByAdmin(t *testing.T) {
	testCases := []struct {
		name           string
		listId         string
		mockBehavior   func(mockList *MockList)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			listId: "2",
			mockBehavior: func(mockList *MockList) {
				mockList.On("DeleteOneByAdmin", mock.Anything, 2).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"status": "ok",
				"message": "List deleted successfully"
			}`,
		},
		{
			name:           "Invalid listId parameter",
			listId:         "invalid",
			mockBehavior:   func(mockList *MockList) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Invalid listId param",
				"errors": {"param": "listId must be a valid integer"}
			}`,
		},
		{
			name:   "List not found",
			listId: "2",
			mockBehavior: func(mockList *MockList) {
				mockList.On("DeleteOneByAdmin", mock.Anything, 2).Return(utils.ErrListNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: `{
				"status": "error",
				"message": "List not found",
				"errors": {"listId": "The requested list does not exist"}
			}`,
		},
		{
			name:   "Internal server error",
			listId: "2",
			mockBehavior: func(mockList *MockList) {
				mockList.On("DeleteOneByAdmin", mock.Anything, 2).Return(errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: `{
				"status": "error",
				"message": "Failed to delete list",
				"errors": {"database": "Error during list deletion"}
			}`,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {

			t.Parallel()

			mockList := new(MockList)
			appAuth := new(MockAppAuth)
			noOpLogger := &logger.NoOpLogger{}
			noOpMetrics := &metrics.NoOpMetrics{}

			mockUseCase := &usecase.UseCase{
				List: mockList,
			}

			handler := v1.NewHandler(mockUseCase, appAuth, noOpLogger, noOpMetrics)

			gin.SetMode(gin.TestMode)
			r := gin.Default()

			r.DELETE("/private/api/lists/:id", handler.DeleteListByAdmin)

			req := httptest.NewRequest("DELETE", "/private/api/lists/"+testCase.listId, nil)
			w := httptest.NewRecorder()

			testCase.mockBehavior(mockList)

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatus, w.Code)
			assert.JSONEq(t, testCase.expectedBody, w.Body.String())

			mockList.AssertExpectations(t)
		})
	}
}

// TestHandler_SearchLists tests the SearchLists handler
func TestHandler_SearchLists(t *testing.T) {
	testCases := []struct {
		name           string
		userId         int
		searchText     string
		mockBehavior   func(mockList *MockList)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "Success",
			userId:     1,
			searchText: "coffee",
			mockBehavior: func(mockList *MockList) {
				mockList.On("Search", mock.Anything, 1, "coffee").
					Return([]entity.List{
						{Id: 1, Title: "Coffee Enthusiast Checklist", Description: "A list of tasks and items for coffee lovers"},
						{Id: 2, Title: "Buy Coffee Beans", Description: "Purchase a fresh batch of Arabica beans from the local roaster"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"status": "ok",
				"message": "Lists searched successfully",
				"data": [
					{"id": 1, "title": "Coffee Enthusiast Checklist", "description": "A list of tasks and items for coffee lovers"},
					{"id": 2, "title": "Buy Coffee Beans", "description": "Purchase a fresh batch of Arabica beans from the local roaster"}
				]
			}`,
		},
		{
			name:           "Unauthorized",
			userId:         0,
			searchText:     "coffee",
			mockBehavior:   func(mockList *MockList) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: `{
				"status": "error",
				"message": "Unauthorized",
				"errors": {"auth": "User authentication failed or user not logged in"}
			}`,
		},
		{
			name:           "Empty searchText param",
			userId:         1,
			searchText:     "",
			mockBehavior:   func(mockList *MockList) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Empty searchText param",
				"errors": {"param": "Query parameter 'search_text' is missing or empty"}
			}`,
		},
		{
			name:       "Internal server error",
			userId:     1,
			searchText: "coffee",
			mockBehavior: func(mockList *MockList) {
				mockList.On("Search", mock.Anything, 1, "coffee").
					Return([]entity.List{}, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: `{
				"status": "error",
				"message": "Failed to search lists",
				"errors": {"database": "Error during list search"}
			}`,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockList := new(MockList)
			appAuth := new(MockAppAuth)
			noOpLogger := &logger.NoOpLogger{}
			noOpMetrics := &metrics.NoOpMetrics{}

			mockUseCase := &usecase.UseCase{
				List: mockList,
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

			r.GET("/api/lists/search", handler.SearchLists)

			req := httptest.NewRequest("GET", "/api/lists/search?search_text="+testCase.searchText, nil)
			w := httptest.NewRecorder()

			testCase.mockBehavior(mockList)

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatus, w.Code)
			assert.JSONEq(t, testCase.expectedBody, w.Body.String())

			mockList.AssertExpectations(t)
		})
	}
}
