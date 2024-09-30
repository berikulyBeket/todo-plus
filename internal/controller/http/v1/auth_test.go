package v1_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	v1 "github.com/berikulyBeket/todo-plus/internal/controller/http/v1"
	"github.com/berikulyBeket/todo-plus/internal/entity"
	"github.com/berikulyBeket/todo-plus/internal/usecase"
	"github.com/berikulyBeket/todo-plus/pkg/logger"
	"github.com/berikulyBeket/todo-plus/pkg/metrics"
	"github.com/berikulyBeket/todo-plus/utils"
)

func TestHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		mockBehavior   func(mockAuth *MockAuth)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			input: `{
				"name": "Test User",
				"username": "testuser",
				"password": "password"
			}`,
			mockBehavior: func(mockAuth *MockAuth) {
				mockAuth.On("CreateUser", mock.Anything, entity.User{
					Name:     "Test User",
					Username: "testuser",
					Password: "password",
				}).Return(1, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"status": "ok",
				"message": "User registered successfully",
				"data": {"id": 1}
			}`,
		},
		{
			name: "Missing required fields",
			input: `{
				"username": "testuser",
				"password": "password"
			}`,
			mockBehavior:   func(mockAuth *MockAuth) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Invalid request body",
				"errors": {"body": "Invalid or malformed JSON"}
			}`,
		},
		{
			name: "Invalid input",
			input: `{
				"name": "testuser",
				"username": "testuser",
				"password": "pass"
			}`,
			mockBehavior:   func(mockAuth *MockAuth) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Invalid request body",
				"errors": {"body": "Invalid or malformed JSON"}
			}`,
		},
		{
			name: "CreateUser failure",
			input: `{
				"name": "Test User",
				"username": "testuser",
				"password": "password"
			}`,
			mockBehavior: func(mockAuth *MockAuth) {
				mockAuth.On("CreateUser", mock.Anything, entity.User{
					Name:     "Test User",
					Username: "testuser",
					Password: "password",
				}).Return(0, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: `{
				"status": "error",
				"message": "Failed to create user",
				"errors": {"database": "db error"}
			}`,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockAuth := new(MockAuth)
			appAuth := new(MockAppAuth)
			noOpLogger := &logger.NoOpLogger{}
			noOpMetrics := &metrics.NoOpMetrics{}

			mockUseCase := &usecase.UseCase{
				Auth: mockAuth,
			}

			handler := v1.NewHandler(mockUseCase, appAuth, noOpLogger, noOpMetrics)

			gin.SetMode(gin.TestMode)
			r := gin.Default()
			r.POST("/auth/sign-up", handler.SignUp)

			req := httptest.NewRequest("POST", "/auth/sign-up", bytes.NewBufferString(testCase.input))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			testCase.mockBehavior(mockAuth)

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatus, w.Code)
			assert.JSONEq(t, testCase.expectedBody, w.Body.String())

			mockAuth.AssertExpectations(t)
		})
	}
}

func TestHandler_SignIn(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		mockBehavior   func(mockAuth *MockAuth)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			input: `{
				"username": "testuser",
				"password": "password"
			}`,
			mockBehavior: func(mockAuth *MockAuth) {
				mockAuth.On("AuthenticateUser", mock.Anything, "testuser", "password").
					Return(entity.User{Id: 1, Username: "testuser"}, nil)
				mockAuth.On("GenerateToken", mock.Anything, 1).
					Return("valid-token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"status": "ok",
				"message": "Signed in successfully",
				"data": {"token": "valid-token"}
			}`,
		},
		{
			name: "Invalid input",
			input: `{
				"username": "ab"
				"password": "qwerty"
			}`,
			mockBehavior:   func(mockAuth *MockAuth) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Invalid request body",
				"errors": {"body": "Invalid or malformed JSON"}
			}`,
		},
		{
			name: "Missing required fields",
			input: `{
				"username": "testuser"
			}`,
			mockBehavior:   func(mockAuth *MockAuth) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"status": "error",
				"message": "Invalid request body",
				"errors": {"body": "Invalid or malformed JSON"}
			}`,
		},
		{
			name: "Invalid credentials",
			input: `{
				"username": "invaliduser",
				"password": "wrongpassword"
			}`,
			mockBehavior: func(mockAuth *MockAuth) {
				mockAuth.On("AuthenticateUser", mock.Anything, "invaliduser", "wrongpassword").
					Return(entity.User{}, utils.ErrUserNotFound)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: `{
				"status": "error",
				"message": "Invalid username or password",
				"errors": {"credentials": "Invalid username or password"}
			}`,
		},
		{
			name: "Failed to generate token",
			input: `{
				"username": "testuser",
				"password": "password"
			}`,
			mockBehavior: func(mockAuth *MockAuth) {
				mockAuth.On("AuthenticateUser", mock.Anything, "testuser", "password").
					Return(entity.User{Id: 1, Username: "testuser"}, nil)
				mockAuth.On("GenerateToken", mock.Anything, 1).
					Return("", errors.New("token generation error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: `{
				"status": "error",
				"message": "Failed to generate token",
				"errors": {"token": "token generation error"}
			}`,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockAuth := new(MockAuth)
			appAuth := new(MockAppAuth)
			noOpLogger := &logger.NoOpLogger{}
			noOpMetrics := &metrics.NoOpMetrics{}

			mockUseCase := &usecase.UseCase{
				Auth: mockAuth,
			}

			handler := v1.NewHandler(mockUseCase, appAuth, noOpLogger, noOpMetrics)

			mockAuth.ExpectedCalls = nil

			gin.SetMode(gin.TestMode)
			r := gin.Default()
			r.POST("/auth/sign-in", handler.SignIn)

			req := httptest.NewRequest("POST", "/auth/sign-in", bytes.NewBufferString(testCase.input))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			testCase.mockBehavior(mockAuth)

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatus, w.Code)
			assert.JSONEq(t, testCase.expectedBody, w.Body.String())

			mockAuth.AssertExpectations(t)
		})
	}
}
