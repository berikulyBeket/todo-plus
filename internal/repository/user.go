package repository

import (
	"context"
	"fmt"

	"github.com/berikulyBeket/todo-plus/pkg/database"
	"github.com/berikulyBeket/todo-plus/utils"
)

// UserRepo handles database operations for users
type UserRepo struct {
	db *database.Database
}

// NewUserRepo creates a new instance of UserRepo
func NewUserRepo(db *database.Database) *UserRepo {
	return &UserRepo{db}
}

// DeleteOneById deletes a user by their ID and returns an error if the user is not found or there's a database error
func (r *UserRepo) DeleteOneById(ctx context.Context, userId int) error {
	deleteUserQuery := fmt.Sprintf("DELETE FROM %s WHERE id = $1", UsersTable)

	result, err := r.db.Executer.Exec(deleteUserQuery, userId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return utils.ErrUserNotFound
	}

	return nil
}
