package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/berikulyBeket/todo-plus/internal/entity"
	"github.com/berikulyBeket/todo-plus/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helper function to initialize mocks and usecase
func setupAuthUseCase(mockRepo *MockAuthRepo, mockHasher *MockHasher) *usecase.AuthUseCase {
	return usecase.NewAuthUseCase(mockRepo, mockHasher, nil)
}

// Helper function to assert error and user equality
func assertUserResult(t *testing.T, expectedUser entity.User, actualUser entity.User, expectedErr, actualErr error) {
	assert.Equal(t, expectedUser, actualUser)
	if expectedErr != nil {
		assert.Error(t, actualErr)
		assert.EqualError(t, actualErr, expectedErr.Error())
	} else {
		assert.NoError(t, actualErr)
	}
}

// TestCreateUser tests the CreateUser function in the AuthUseCase
func TestCreateUser(t *testing.T) {
	testCases := []struct {
		name       string
		user       entity.User
		mockHash   string
		mockResult int
		mockError  error
		expectErr  bool
	}{
		{
			name:       "Successful user creation",
			user:       entity.User{Username: "testuser", Password: "password"},
			mockHash:   "hashedPassword",
			mockResult: 1,
			mockError:  nil,
			expectErr:  false,
		},
		{
			name:       "Repo failure",
			user:       entity.User{Username: "testuser", Password: "password"},
			mockHash:   "hashedPassword",
			mockResult: 0,
			mockError:  errors.New("repo error"),
			expectErr:  true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockAuthRepo)
			mockHasher := new(MockHasher)
			authUseCase := setupAuthUseCase(mockRepo, mockHasher)

			mockHasher.On("HashPassword", testCase.user.Password).Return(testCase.mockHash)
			mockRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(user entity.User) bool {
				return user.Password == testCase.mockHash
			})).Return(testCase.mockResult, testCase.mockError)

			_, err := authUseCase.CreateUser(context.Background(), testCase.user)

			if testCase.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockHasher.AssertCalled(t, "HashPassword", testCase.user.Password)
			mockRepo.AssertCalled(t, "CreateUser", mock.Anything, mock.MatchedBy(func(user entity.User) bool {
				return user.Password == testCase.mockHash
			}))
		})
	}
}

// TestGetUser tests the GetUser function in the AuthUseCase
func TestGetUser(t *testing.T) {
	testCases := []struct {
		name         string
		username     string
		password     string
		mockHash     string
		mockUser     entity.User
		mockError    error
		expectedUser entity.User
		expectedErr  error
	}{
		{
			name:         "Successful user retrieval",
			username:     "testuser",
			password:     "password",
			mockHash:     "hashedPassword",
			mockUser:     entity.User{Username: "testuser", Password: "hashedPassword"},
			mockError:    nil,
			expectedUser: entity.User{Username: "testuser", Password: "hashedPassword"},
			expectedErr:  nil,
		},
		{
			name:         "Failed user retrieval",
			username:     "testuser",
			password:     "password",
			mockHash:     "hashedPassword",
			mockUser:     entity.User{},
			mockError:    errors.New("repository error"),
			expectedUser: entity.User{},
			expectedErr:  errors.New("repository error"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockAuthRepo)
			mockHasher := new(MockHasher)
			authUseCase := setupAuthUseCase(mockRepo, mockHasher)

			mockHasher.On("HashPassword", testCase.password).Return(testCase.mockHash)
			mockRepo.On("GetUserByCredentials", mock.Anything, testCase.username, testCase.mockHash).Return(testCase.mockUser, testCase.mockError)

			actualUser, actualErr := authUseCase.AuthenticateUser(context.Background(), testCase.username, testCase.password)

			assertUserResult(t, testCase.expectedUser, actualUser, testCase.expectedErr, actualErr)

			mockHasher.AssertCalled(t, "HashPassword", testCase.password)
			mockRepo.AssertCalled(t, "GetUserByCredentials", mock.Anything, testCase.username, testCase.mockHash)
		})
	}
}

// TestGenerateToken tests the GenerateToken function in the AuthUseCase
func TestGenerateToken(t *testing.T) {
	testCases := []struct {
		name           string
		userId         int
		mockError      error
		expectedToken  string
		expectedErrMsg string
	}{
		{
			name:           "Successful token generation",
			userId:         123,
			mockError:      nil,
			expectedToken:  "generatedToken123",
			expectedErrMsg: "",
		},
		{
			name:           "Failed token generation",
			userId:         456,
			mockError:      errors.New("token generation error"),
			expectedToken:  "",
			expectedErrMsg: "token generation error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockTokenMaker := new(MockTokenMaker)
			authUseCase := usecase.NewAuthUseCase(nil, nil, mockTokenMaker)

			mockTokenMaker.On("CreateToken", testCase.userId).Return(testCase.expectedToken, testCase.mockError)

			actualToken, actualErr := authUseCase.GenerateToken(context.Background(), testCase.userId)

			assert.Equal(t, testCase.expectedToken, actualToken)
			if testCase.expectedErrMsg != "" {
				assert.Error(t, actualErr)
				assert.EqualError(t, actualErr, testCase.expectedErrMsg)
			} else {
				assert.NoError(t, actualErr)
			}

			mockTokenMaker.AssertCalled(t, "CreateToken", testCase.userId)
		})
	}
}

// TestParseToken tests the ParseToken function in the AuthUseCase
func TestParseToken(t *testing.T) {
	testCases := []struct {
		name           string
		accessToken    string
		mockError      error
		expectedUserId int
		expectedErrMsg string
	}{
		{
			name:           "Successful token parsing",
			accessToken:    "validToken123",
			mockError:      nil,
			expectedUserId: 123,
			expectedErrMsg: "",
		},
		{
			name:           "Failed token parsing",
			accessToken:    "invalidToken456",
			mockError:      errors.New("invalid token"),
			expectedUserId: 0,
			expectedErrMsg: "invalid token",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockTokenMaker := new(MockTokenMaker)
			authUseCase := usecase.NewAuthUseCase(nil, nil, mockTokenMaker)

			mockTokenMaker.On("VerifyToken", testCase.accessToken).Return(testCase.expectedUserId, testCase.mockError)

			actualUserId, actualErr := authUseCase.ParseToken(context.Background(), testCase.accessToken)

			assert.Equal(t, testCase.expectedUserId, actualUserId)
			if testCase.expectedErrMsg != "" {
				assert.Error(t, actualErr)
				assert.EqualError(t, actualErr, testCase.expectedErrMsg)
			} else {
				assert.NoError(t, actualErr)
			}

			mockTokenMaker.AssertCalled(t, "VerifyToken", testCase.accessToken)
		})
	}
}
