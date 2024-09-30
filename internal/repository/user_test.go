package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/berikulyBeket/todo-plus/internal/repository"
	"github.com/berikulyBeket/todo-plus/pkg/database"
	"github.com/berikulyBeket/todo-plus/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// setupUserRepoTest initializes the repository and mock database for UserRepo tests
func setupUserRepoTest(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, *repository.UserRepo) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	database := database.New(sqlxDB)
	userRepo := repository.NewUserRepo(database)

	return sqlxDB, mock, userRepo
}

// assertUserRepoExpectations checks if all the SQL mock expectations were met for UserRepo tests
func assertUserRepoExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	err := mock.ExpectationsWereMet()
	assert.NoError(t, err, "All expectations for UserRepo were not met")
}

// TestUserRepo_DeleteUserById tests the DeleteUserById function in UserRepo
func TestUserRepo_DeleteUserById(t *testing.T) {
	tests := []struct {
		name        string
		userId      int
		mockExec    func(mock sqlmock.Sqlmock)
		expectedErr error
	}{
		{
			name:   "Success case",
			userId: 1,
			mockExec: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedErr: nil,
		},
		{
			name:   "User not found",
			userId: 1,
			mockExec: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 0))
			},
			expectedErr: utils.ErrUserNotFound,
		},
		{
			name:   "Database error",
			userId: 1,
			mockExec: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnError(errors.New("db error"))
			},
			expectedErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sqlxDB, mock, userRepo := setupUserRepoTest(t)
			defer sqlxDB.Close()

			tc.mockExec(mock)

			err := userRepo.DeleteOneById(context.Background(), tc.userId)

			assert.Equal(t, tc.expectedErr, err)

			assertUserRepoExpectations(t, mock)
		})
	}
}
