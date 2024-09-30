package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/berikulyBeket/todo-plus/internal/entity"
	"github.com/berikulyBeket/todo-plus/pkg/cache"
	"github.com/berikulyBeket/todo-plus/pkg/database"
	"github.com/berikulyBeket/todo-plus/pkg/logger"
	"github.com/berikulyBeket/todo-plus/utils"
)

type ListRepo struct {
	db     *database.Database
	cache  *cache.Cache
	logger logger.Interface
}

// NewListRepo creates a new instance of ListRepo
func NewListRepo(
	db *database.Database,
	cache *cache.Cache,
	logger logger.Interface,
) *ListRepo {
	return &ListRepo{db, cache, logger}
}

// CreateUserList creates a new list and links it to the user
func (r *ListRepo) CreateUserList(ctx context.Context, userId int, list *entity.List) (int, error) {
	tx, err := r.db.Transaction.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	var listId int

	query := fmt.Sprintf(`
		INSERT INTO %s (title, description)
		VALUES ($1, $2)
		RETURNING id`, ListsTable)

	err = tx.QueryRow(query, list.Title, list.Description).Scan(&listId)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	query = fmt.Sprintf(`
		INSERT INTO %s (user_id, list_id)
		VALUES ($1, $2)`, UsersListsTable)

	_, err = tx.Exec(query, userId, listId)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	list.Id = listId
	listCacheKey := fmt.Sprintf(cacheKeyListById.Pattern, listId)
	userListsCacheKey := fmt.Sprintf(cacheKeyUserLists.Pattern, userId)

	if err := r.cache.Master.Set(ctx, listCacheKey, list, cacheKeyListById.TTL); err != nil {
		r.logger.Errorf("failed to set cache for key %s: %v", listCacheKey, err)
	}
	if err := r.cache.Master.Delete(ctx, userListsCacheKey); err != nil {
		r.logger.Errorf("failed to invalidate cache for key %s: %v", userListsCacheKey, err)
	}

	return listId, nil
}

// GetAllUserLists retrieves all lists associated with a user
func (r *ListRepo) GetAllUserLists(ctx context.Context, userId int) ([]entity.List, error) {
	lists := []entity.List{}

	userListsCacheKey := fmt.Sprintf(cacheKeyUserLists.Pattern, userId)
	if exists, err := r.cache.Replica.Get(ctx, userListsCacheKey, &lists); err != nil {
		r.logger.Errorf("cache error for key %s: %v", userListsCacheKey, err)
	} else if exists {
		return lists, nil
	}

	query := fmt.Sprintf(`
		SELECT l.id, l.title, l.description
		FROM %s l
		JOIN %s ul ON l.id = ul.list_id
		WHERE ul.user_id = $1`, ListsTable, UsersListsTable)

	rows, err := r.db.Querier.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var list entity.List
		if err := rows.Scan(&list.Id, &list.Title, &list.Description); err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := r.cache.Master.Set(ctx, userListsCacheKey, lists, cacheKeyUserLists.TTL); err != nil {
		r.logger.Errorf("failed to set cache for key %s: %v", userListsCacheKey, err)
	}

	return lists, nil
}

// GetOneById retrieves a single list by its Id
func (r *ListRepo) GetOneById(ctx context.Context, listId int) (entity.List, error) {
	var list entity.List

	listCacheKey := fmt.Sprintf(cacheKeyListById.Pattern, listId)
	if exists, err := r.cache.Replica.Get(ctx, listCacheKey, &list); err != nil {
		r.logger.Errorf("cache error for key %s: %v", listCacheKey, err)
	} else if exists {
		return list, nil
	}

	query := fmt.Sprintf("SELECT id, title, description FROM %s WHERE id = $1", ListsTable)

	err := r.db.Querier.QueryRow(query, listId).Scan(&list.Id, &list.Title, &list.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return list, utils.ErrListNotFound
		}

		return list, err
	}

	if err = r.cache.Master.Set(ctx, listCacheKey, list, cacheKeyListById.TTL); err != nil {
		r.logger.Errorf("failed to set cache for key %s: %v", listCacheKey, err)
	}

	return list, nil
}

// GetManyByIds retrieves multiple lists by their Ids
func (r *ListRepo) GetManyByIds(ctx context.Context, listIds []int) ([]entity.List, error) {
	lists := []entity.List{}

	if len(listIds) == 0 {
		return []entity.List{}, nil
	}

	query := fmt.Sprintf("SELECT id, title, description FROM %s WHERE id IN (%s)", ListsTable, utils.CreatePlaceholders(len(listIds)))
	rows, err := r.db.Querier.Query(query, utils.ConvertToInterfaceSlice(listIds)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var list entity.List
		if err := rows.Scan(&list.Id, &list.Title, &list.Description); err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}

	return lists, nil
}

// UpdateOneById updates a list by its ID based on the provided input
func (r *ListRepo) UpdateOneById(ctx context.Context, userId *int, listId int, newTodoInput entity.UpdateListInput) error {
	query := fmt.Sprintf("UPDATE %s SET ", ListsTable)
	args := []interface{}{}

	setClauses := []string{}
	argIndex := 1
	if newTodoInput.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, *newTodoInput.Title)
		argIndex++
	}
	if newTodoInput.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *newTodoInput.Description)
		argIndex++
	}

	if len(setClauses) == 0 {
		return utils.ErrItemEmptyRequest
	}

	query += strings.Join(setClauses, ", ")

	query += fmt.Sprintf(" WHERE id = $%d", argIndex)

	args = append(args, listId)

	result, err := r.db.Executer.Exec(query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return utils.ErrListNotFound
	}

	listCacheKey := fmt.Sprintf(cacheKeyListById.Pattern, listId)
	userListsCacheKey := fmt.Sprintf(cacheKeyUserLists.Pattern, *userId)

	if err := r.cache.Master.Delete(ctx, listCacheKey); err != nil {
		r.logger.Errorf("failed to invalidate cache for key %s: %v", listCacheKey, err)
	}
	if err := r.cache.Master.Delete(ctx, userListsCacheKey); err != nil {
		r.logger.Errorf("failed to invalidate cache for key %s: %v", userListsCacheKey, err)
	}

	return nil
}

// DeleteOneById deletes a list by its Id
func (r *ListRepo) DeleteOneById(ctx context.Context, userId *int, listId int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", ListsTable)

	result, err := r.db.Executer.Exec(query, listId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return utils.ErrListNotFound
	}

	listCacheKey := fmt.Sprintf(cacheKeyListById.Pattern, listId)
	if err := r.cache.Master.Delete(ctx, listCacheKey); err != nil {
		r.logger.Errorf("failed to invalidate cache for key %s: %v", listCacheKey, err)
	}
	if userId != nil {
		userListsCacheKey := fmt.Sprintf(cacheKeyUserLists.Pattern, *userId)
		if err := r.cache.Master.Delete(ctx, userListsCacheKey); err != nil {
			r.logger.Errorf("failed to invalidate cache for key %s: %v", userListsCacheKey, err)
		}
	}

	return nil
}

// IsUserOwnerOfList checks if a user is the owner of a specific list
func (r *ListRepo) IsUserOwnerOfList(ctx context.Context, userId, listId int) error {
	query := fmt.Sprintf("SELECT 1 FROM %s WHERE user_id = $1 AND list_id = $2", UsersListsTable)

	var exists int
	err := r.db.Querier.QueryRow(query, userId, listId).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrUserNotOwner
		}

		return err
	}

	return nil
}
