package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/berikulyBeket/todo-plus/internal/entity"
	"github.com/berikulyBeket/todo-plus/internal/middleware"
	"github.com/berikulyBeket/todo-plus/internal/usecase"
	"github.com/berikulyBeket/todo-plus/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockAuthUseCase mocks the Auth use case
type mockAuthUseCase struct {
	mock.Mock
}

// CreateUser mocks the creation of a user in the authentication use case
func (m *mockAuthUseCase) CreateUser(ctx context.Context, user entity.User) (int, error) {
	args := m.Called(ctx, user)
	return args.Int(0), args.Error(1)
}

// AuthenticateUser mocks the authentication of a user with username and password
func (m *mockAuthUseCase) AuthenticateUser(ctx context.Context, username, password string) (entity.User, error) {
	args := m.Called(ctx, username, password)
	return args.Get(0).(entity.User), args.Error(1)
}

// GenerateToken mocks the generation of a JWT token for a user
func (m *mockAuthUseCase) GenerateToken(ctx context.Context, userId int) (string, error) {
	args := m.Called(ctx, userId)
	return args.String(0), args.Error(1)
}

// ParseToken mocks the parsing of a JWT token to extract the user ID
func (m *mockAuthUseCase) ParseToken(ctx context.Context, token string) (int, error) {
	args := m.Called(ctx, token)
	return args.Int(0), args.Error(1)
}

// setupAuthRouter sets up the router with the authentication middleware
func setupAuthRouter(authUseCase usecase.Auth, logger logger.Interface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Authentication(authUseCase, logger))

	router.GET("/test", func(c *gin.Context) {
		userId, err := middleware.GetUserId(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"userId": userId})
	})

	return router
}

// TestAuthentication tests the Authentication middleware
func TestAuthentication(t *testing.T) {
	testCases := []struct {
		name           string
		authHeader     string
		parseTokenResp int
		parseTokenErr  error
		expectedCode   int
		expectedBody   string
		mockAuth       func(*mockAuthUseCase)
	}{
		{
			name:           "Valid Token",
			authHeader:     "Bearer validToken",
			parseTokenResp: 1,
			parseTokenErr:  nil,
			expectedCode:   http.StatusOK,
			expectedBody:   `{"userId":1}`,
			mockAuth: func(mockAuth *mockAuthUseCase) {
				mockAuth.On("ParseToken", mock.Anything, "validToken").Return(1, nil)
			},
		},
		{
			name:         "Empty Auth Header",
			authHeader:   "",
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"errors":{"token":"empty auth header"}, "message":"Unauthorized", "status":"error"}`,
			mockAuth:     func(mockAuth *mockAuthUseCase) {},
		},
		{
			name:         "Invalid Auth Header Format",
			authHeader:   "invalidHeader",
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"errors":{"token":"invalid auth header"}, "message":"Unauthorized", "status":"error"}`,
			mockAuth:     func(mockAuth *mockAuthUseCase) {},
		},
		{
			name:           "Invalid Token",
			authHeader:     "Bearer invalidToken",
			parseTokenResp: 0,
			parseTokenErr:  errors.New("invalid token"),
			expectedCode:   http.StatusUnauthorized,
			expectedBody:   `{"errors":{"token":"invalid token"}, "message":"Unauthorized", "status":"error"}`,
			mockAuth: func(mockAuth *mockAuthUseCase) {
				mockAuth.On("ParseToken", mock.Anything, "invalidToken").Return(0, errors.New("invalid token"))
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockAuth := new(mockAuthUseCase)
			testCase.mockAuth(mockAuth)

			router := setupAuthRouter(mockAuth, &logger.NoOpLogger{})
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)

			if testCase.authHeader != "" {
				req.Header.Set(middleware.AuthorizationHeader, testCase.authHeader)
			}

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			assert.Equal(t, testCase.expectedCode, resp.Code)
			assert.JSONEq(t, testCase.expectedBody, resp.Body.String())

			mockAuth.AssertExpectations(t)
		})
	}
}
