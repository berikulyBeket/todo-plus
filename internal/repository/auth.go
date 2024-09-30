package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/berikulyBeket/todo-plus/internal/entity"
	"github.com/berikulyBeket/todo-plus/pkg/database"
	"github.com/berikulyBeket/todo-plus/utils"
)

// AuthRepo represents a repository for authentication operations
type AuthRepo struct {
	db *database.Database
}

// NewAuthRepo creates a new instance of AuthRepo with a given database connection
func NewAuthRepo(db *database.Database) *AuthRepo {
	return &AuthRepo{db}
}

// CreateUser inserts a new user into the database
func (r *AuthRepo) CreateUser(ctx context.Context, user entity.User) (int, error) {
	var id int

	query := fmt.Sprintf("INSERT INTO %s (name, username, password_hash) VALUES ($1, $2, $3) RETURNING id", UsersTable)

	err := r.db.Querier.QueryRow(query, user.Name, user.Username, user.Password).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetUser fetches the user by username and password
func (r *AuthRepo) GetUserByCredentials(ctx context.Context, username, password string) (entity.User, error) {
	var user entity.User

	query := fmt.Sprintf(`
		SELECT id
		FROM %s
		WHERE username = $1 AND password_hash = $2`, UsersTable)

	err := r.db.Querier.QueryRow(query, username, password).Scan(&user.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, utils.ErrUserNotFound
		}

		return user, err
	}

	return user, nil
}
