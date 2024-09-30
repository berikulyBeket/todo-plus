package v1_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "github.com/berikulyBeket/todo-plus/internal/controller/http/v1"
	"github.com/berikulyBeket/todo-plus/internal/usecase"
	"github.com/berikulyBeket/todo-plus/pkg/logger"
	"github.com/berikulyBeket/todo-plus/pkg/metrics"
	"github.com/berikulyBeket/todo-plus/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestHandler_DeleteUser tests the DeleteUserByAdmin handler
func TestHandler_DeleteUser(t *testing.T) {
	testCases := []struct {
		name           string
		userId         string
		mockBehavior   func(mockUser *MockUser)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			userId: "1",
			mockBehavior: func(mockUser *MockUser) {
				mockUser.On("DeleteOneByAdmin", mock.Anything, 1).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"status": "ok",
				"message": "User deleted successfully"
			}`,
		},
		{
			name:           "Invalid userId parameter",
			userId:         "invalid",
			mockBehavior:   func(mockUser *MockUser) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Invalid userId param",
				"errors": {"param": "userId must be a valid integer"}
			}`,
		},
		{
			name:   "User not found",
			userId: "1",
			mockBehavior: func(mockUser *MockUser) {
				mockUser.On("DeleteOneByAdmin", mock.Anything, 1).Return(utils.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: `{
				"status": "error",
				"message": "User not found",
				"errors": {"userId": "The requested user does not exist"}
			}`,
		},
		{
			name:   "Internal server error",
			userId: "1",
			mockBehavior: func(mockUser *MockUser) {
				mockUser.On("DeleteOneByAdmin", mock.Anything, 1).Return(errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: `{
				"status": "error",
				"message": "Failed to delete user",
				"errors": {"database": "Error during user deletion"}
			}`,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockUser := new(MockUser)
			appAuth := new(MockAppAuth)
			noOpLogger := &logger.NoOpLogger{}
			noOpMetrics := &metrics.NoOpMetrics{}

			mockUseCase := &usecase.UseCase{
				User: mockUser,
			}

			handler := v1.NewHandler(mockUseCase, appAuth, noOpLogger, noOpMetrics)

			gin.SetMode(gin.TestMode)
			r := gin.Default()

			r.DELETE("/private/api/users/:id", handler.DeleteUserByAdmin)

			req := httptest.NewRequest("DELETE", fmt.Sprintf("/private/api/users/%s", testCase.userId), nil)
			w := httptest.NewRecorder()

			testCase.mockBehavior(mockUser)

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatus, w.Code)
			assert.JSONEq(t, testCase.expectedBody, w.Body.String())

			mockUser.AssertExpectations(t)
		})
	}
}
