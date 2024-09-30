package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/berikulyBeket/todo-plus/internal/entity"
	"github.com/berikulyBeket/todo-plus/internal/repository"
	"github.com/berikulyBeket/todo-plus/pkg/database"
	"github.com/berikulyBeket/todo-plus/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

var (
	createUserQuery = fmt.Sprintf(`INSERT INTO %s \(name, username, password_hash\) VALUES \(\$1, \$2, \$3\) RETURNING id`, repository.UsersTable)
	getUserQuery    = fmt.Sprintf(`SELECT id FROM %s WHERE username = \$1 AND password_hash = \$2`, repository.UsersTable)
)

// Helper function to set up the mock database, sqlmock, and repository
func setupAuthRepoTest(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, *repository.AuthRepo) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	database := database.New(sqlxDB)

	dbRepo := repository.NewAuthRepo(database)

	return sqlxDB, mock, dbRepo
}

// Common helper function to validate the sqlmock expectations
func assertExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	err := mock.ExpectationsWereMet()
	assert.NoError(t, err, "All expectations were not met")
}

// TestCreateUser tests the user create
func TestCreateUser(t *testing.T) {
	testCases := []struct {
		name        string
		user        entity.User
		mockQuery   func(sqlmock.Sqlmock)
		expectedId  int
		expectedErr error
	}{
		{
			name: "Success",
			user: entity.User{Name: "JohnDoe", Username: "johndoe", Password: "hashedPassword"},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(createUserQuery).
					WithArgs("JohnDoe", "johndoe", "hashedPassword").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			expectedId:  1,
			expectedErr: nil,
		},
		{
			name: "SQLConnectionFailure",
			user: entity.User{Name: "JohnDoe", Username: "johndoe", Password: "hashedPassword"},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(createUserQuery).
					WithArgs("JohnDoe", "johndoe", "hashedPassword").
					WillReturnError(sql.ErrConnDone)
			},
			expectedId:  0,
			expectedErr: sql.ErrConnDone,
		},
		{
			name: "UsernameAlreadyExists",
			user: entity.User{Name: "JohnDoe", Username: "johndoe", Password: "hashedPassword"},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(createUserQuery).
					WithArgs("JohnDoe", "johndoe", "hashedPassword").
					WillReturnError(sqlmock.ErrCancelled)
			},
			expectedId:  0,
			expectedErr: sqlmock.ErrCancelled,
		},
		{
			name: "EmptyUserFields",
			user: entity.User{Name: "", Username: "", Password: ""},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(createUserQuery).
					WithArgs("", "", "").
					WillReturnError(sql.ErrNoRows)
			},
			expectedId:  0,
			expectedErr: sql.ErrNoRows,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			sqlxDB, mock, dbRepo := setupAuthRepoTest(t)
			defer sqlxDB.Close()

			testCase.mockQuery(mock)

			id, err := dbRepo.CreateUser(context.Background(), testCase.user)

			assert.Equal(t, testCase.expectedId, id)
			assert.Equal(t, testCase.expectedErr, err)

			assertExpectations(t, mock)
		})
	}
}

// TestCreateUser tests the getting user
func TestGetUser(t *testing.T) {
	testCases := []struct {
		name        string
		username    string
		password    string
		mockQuery   func(sqlmock.Sqlmock)
		expectedId  int
		expectedErr error
	}{
		{
			name:     "Success",
			username: "johndoe",
			password: "hashedPassword",
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(getUserQuery).
					WithArgs("johndoe", "hashedPassword").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			expectedId:  1,
			expectedErr: nil,
		},
		{
			name:     "UserNotFound",
			username: "johndoe",
			password: "wrongPassword",
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(getUserQuery).
					WithArgs("johndoe", "wrongPassword").
					WillReturnError(sql.ErrNoRows)
			},
			expectedId:  0,
			expectedErr: utils.ErrUserNotFound,
		},
		{
			name:     "SQLConnectionError",
			username: "johndoe",
			password: "hashedPassword",
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(getUserQuery).
					WithArgs("johndoe", "hashedPassword").
					WillReturnError(sql.ErrConnDone)
			},
			expectedId:  0,
			expectedErr: sql.ErrConnDone,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			sqlxDB, mock, dbRepo := setupAuthRepoTest(t)
			defer sqlxDB.Close()

			testCase.mockQuery(mock)

			user, err := dbRepo.GetUserByCredentials(context.Background(), testCase.username, testCase.password)

			assert.Equal(t, testCase.expectedId, user.Id)
			assert.Equal(t, testCase.expectedErr, err)

			assertExpectations(t, mock)
		})
	}
}
