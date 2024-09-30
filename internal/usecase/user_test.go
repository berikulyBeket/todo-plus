package usecase_test

import (
	"context"
	"testing"

	"github.com/berikulyBeket/todo-plus/internal/usecase"
	"github.com/berikulyBeket/todo-plus/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestDeleteUserById tests the DeleteOneByAdmin function in the UserUseCase
func TestDeleteUserById(t *testing.T) {
	testCases := []struct {
		name        string
		userId      int
		expectedErr error
	}{
		{
			name:        "Successful deletion",
			userId:      1,
			expectedErr: nil,
		},
		{
			name:        "User not found",
			userId:      1,
			expectedErr: utils.ErrUserNotFound,
		},
	}

	for _, testCase := range testCases {
		mockRepo := new(MockUserRepo)
		userUseCase := usecase.NewUserUseCase(mockRepo)

		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockRepo.On("DeleteOneById", mock.Anything, testCase.userId).Return(testCase.expectedErr)

			err := userUseCase.DeleteOneByAdmin(context.Background(), testCase.userId)

			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
