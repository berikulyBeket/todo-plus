package search

import (
	"context"

	"github.com/berikulyBeket/todo-plus/internal/entity"

	"github.com/elastic/go-elasticsearch/v8"
)

// List defines the interface for searching, indexing, and deleting lists in Elasticsearch
type List interface {
	SearchIds(ctx context.Context, userId int, searchText string) ([]int, error)
	Index(ctx context.Context, userId int, list *entity.List) error
	Delete(ctx context.Context, listId int) error
}

// Item defines the interface for searching, indexing, and deleting items in Elasticsearch
type Item interface {
	SearchIds(ctx context.Context, userId int, listId *int, done *bool, searchText string) ([]int, error)
	Index(ctx context.Context, userId int, listId int, item entity.Item) error
	Delete(ctx context.Context, itemId int) error
}

// SearchServices aggregates the List and Item search services into one struct
type SearchServices struct {
	List
	Item
}

// New initializes and returns a SearchServices struct, providing List and Item search functionalities
func New(client *elasticsearch.Client) *SearchServices {
	return &SearchServices{
		List: NewListSearch(client),
		Item: NewItemSearch(client),
	}
}
