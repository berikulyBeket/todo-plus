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

// ItemRepo handles item-related operations with database and cache management
type ItemRepo struct {
	db     *database.Database
	cache  *cache.Cache
	logger logger.Interface
}

// NewItemRepo creates a new ItemRepo instance
func NewItemRepo(db *database.Database, cache *cache.Cache, logger logger.Interface) *ItemRepo {
	return &ItemRepo{
		db,
		cache,
		logger,
	}
}

// CreateListItem creates a new item in the list and stores it in the database
func (r *ItemRepo) CreateListItem(ctx context.Context, listId int, item *entity.Item) (int, error) {
	tx, err := r.db.Transaction.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	var itemId int
	createItemQuery := fmt.Sprintf(`INSERT INTO %s (title, description) VALUES ($1, $2) RETURNING id`, ItemsTable)

	err = tx.QueryRow(createItemQuery, item.Title, item.Description).Scan(&itemId)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	createListItemsQuery := fmt.Sprintf(`INSERT INTO %s (list_id, item_id) VALUES ($1, $2)`, ListsItemsTable)

	if _, err := tx.Exec(createListItemsQuery, listId, itemId); err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	item.Id = itemId
	itemCacheKey := fmt.Sprintf(cacheKeyItemById.Pattern, itemId)
	listItemsCacheKey := fmt.Sprintf(cacheKeyListItems.Pattern, listId)

	if err := r.cache.Master.Set(ctx, itemCacheKey, item, cacheKeyItemById.TTL); err != nil {
		r.logger.Errorf("failed to set cache for key %s: %v", itemCacheKey, err)
	}
	if err := r.cache.Master.Delete(ctx, listItemsCacheKey); err != nil {
		r.logger.Errorf("failed to invalidate cache for key %s: %v", listItemsCacheKey, err)
	}

	return itemId, nil
}

// GetAllListItems retrieves all items for a specific list from the database
func (r *ItemRepo) GetAllListItems(ctx context.Context, listId int) ([]entity.Item, error) {
	items := []entity.Item{}

	cacheKey := fmt.Sprintf(cacheKeyListItems.Pattern, listId)
	if exists, err := r.cache.Replica.Get(ctx, cacheKey, &items); err != nil {
		r.logger.Errorf("cache error for key %s: %v", cacheKey, err)
	} else if exists {
		return items, nil
	}

	query := fmt.Sprintf(`
		SELECT i.id, i.title, i.description, i.done
		FROM %s i
		JOIN %s li ON i.id = li.item_id
		WHERE li.list_id = $1`, ItemsTable, ListsItemsTable)

	rows, err := r.db.Querier.Query(query, listId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Item
		if err := rows.Scan(&item.Id, &item.Title, &item.Description, &item.Done); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := r.cache.Master.Set(ctx, cacheKey, items, cacheKeyListItems.TTL); err != nil {
		r.logger.Errorf("failed to set cache for key %s: %v", cacheKey, err)
	}

	return items, nil
}

// GetOneById retrieves a specific item by its ID
func (r *ItemRepo) GetOneById(ctx context.Context, itemId int) (entity.Item, error) {
	var item entity.Item

	cacheKey := fmt.Sprintf(cacheKeyItemById.Pattern, itemId)
	if exists, err := r.cache.Replica.Get(ctx, cacheKey, &item); err != nil {
		r.logger.Errorf("cache error for key %s: %v", cacheKey, err)
	} else if exists {
		return item, nil
	}

	query := fmt.Sprintf("SELECT id, title, description, done FROM %s WHERE id = $1", ItemsTable)

	err := r.db.Querier.QueryRow(query, itemId).Scan(&item.Id, &item.Title, &item.Description, &item.Done)
	if err != nil {
		if err == sql.ErrNoRows {
			return item, utils.ErrItemNotFound
		}
		return item, err
	}

	if err := r.cache.Master.Set(ctx, cacheKey, item, cacheKeyItemById.TTL); err != nil {
		r.logger.Errorf("failed to set cache for key %s: %v", cacheKey, err)
	}

	return item, nil
}

// GetManyByIds retrieves multiple items based on a list of item IDs
func (r *ItemRepo) GetManyByIds(ctx context.Context, itemIds []int) ([]entity.Item, error) {
	items := []entity.Item{}

	if len(itemIds) == 0 {
		return []entity.Item{}, nil
	}

	query := fmt.Sprintf("SELECT id, title, description, done FROM %s WHERE id IN (%s)", ItemsTable, utils.CreatePlaceholders(len(itemIds)))

	rows, err := r.db.Querier.Query(query, utils.ConvertToInterfaceSlice(itemIds)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Item
		if err := rows.Scan(&item.Id, &item.Title, &item.Description, &item.Done); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// UpdateOneById updates an item's details by its ID
func (r *ItemRepo) UpdateOneById(ctx context.Context, listId *int, itemId int, input entity.UpdateItemInput) error {
	query := fmt.Sprintf("UPDATE %s SET ", ItemsTable)
	args := []interface{}{}

	setClauses := []string{}
	argIndex := 1

	if input.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, *input.Title)
		argIndex++
	}
	if input.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *input.Description)
		argIndex++
	}
	if input.Done != nil {
		setClauses = append(setClauses, fmt.Sprintf("done = $%d", argIndex))
		args = append(args, *input.Done)
		argIndex++
	}

	if len(setClauses) == 0 {
		return utils.ErrItemEmptyRequest
	}

	query += strings.Join(setClauses, ", ")

	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, itemId)

	result, err := r.db.Executer.Exec(query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return utils.ErrItemNotFound
	}

	itemCacheKey := fmt.Sprintf(cacheKeyItemById.Pattern, itemId)
	listItemsCacheKey := fmt.Sprintf(cacheKeyListItems.Pattern, *listId)

	if err := r.cache.Master.Delete(ctx, itemCacheKey); err != nil {
		r.logger.Errorf("failed to invalidate cache for key %s: %v", itemCacheKey, err)
	}
	if err := r.cache.Master.Delete(ctx, listItemsCacheKey); err != nil {
		r.logger.Errorf("failed to invalidate cache for key %s: %v", listItemsCacheKey, err)
	}

	return nil
}

// DeleteOneById deletes an item by its ID from the database
func (r *ItemRepo) DeleteOneById(ctx context.Context, listId *int, itemId int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", ItemsTable)

	result, err := r.db.Executer.Exec(query, itemId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return utils.ErrItemNotFound
	}

	itemCacheKey := fmt.Sprintf(cacheKeyItemById.Pattern, itemId)
	if err := r.cache.Master.Delete(ctx, itemCacheKey); err != nil {
		r.logger.Errorf("failed to invalidate cache for key %s: %v", itemCacheKey, err)
	}
	if listId != nil {
		listItemsCacheKey := fmt.Sprintf(cacheKeyListItems.Pattern, *listId)
		if err := r.cache.Master.Delete(ctx, listItemsCacheKey); err != nil {
			r.logger.Errorf("failed to invalidate cache for key %s: %v", listItemsCacheKey, err)
		}
	}

	return nil
}

// IsUserOwnerOfItem checks if a user is the owner of a specific item
func (r *ItemRepo) IsUserOwnerOfItem(ctx context.Context, userId, itemId int) error {
	query := fmt.Sprintf(`
		SELECT 1
		FROM %s li
		JOIN %s ul 
		ON li.list_id = ul.list_id 
		WHERE ul.user_id = $1 AND li.item_id = $2`, ListsItemsTable, UsersListsTable)

	var exists int

	err := r.db.Querier.QueryRow(query, userId, itemId).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrUserNotOwner
		}

		return err
	}

	return nil
}
