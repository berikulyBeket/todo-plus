package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/berikulyBeket/todo-plus/internal/entity"

	"github.com/elastic/go-elasticsearch/v8"
)

const (
	itemIndexName = "items"
)

type ItemSearch struct {
	client *elasticsearch.Client
}

func NewItemSearch(client *elasticsearch.Client) Item {
	return &ItemSearch{client: client}
}

// SearchIds performs a search with optional filtering by listId and done status, and a full-text search on items
func (ls *ItemSearch) SearchIds(ctx context.Context, userId int, listId *int, done *bool, searchText string) ([]int, error) {
	query := map[string]interface{}{
		"bool": map[string]interface{}{
			"must": []interface{}{
				map[string]interface{}{
					"multi_match": map[string]interface{}{
						"query":     searchText,
						"fields":    []string{"title^2", "description"},
						"fuzziness": "AUTO",
					},
				},
			},
			"filter": []interface{}{
				map[string]interface{}{
					"term": map[string]interface{}{
						"userId": userId,
					},
				},
			},
		},
	}

	if listId != nil {
		query["bool"].(map[string]interface{})["filter"] = append(query["bool"].(map[string]interface{})["filter"].([]interface{}), map[string]interface{}{
			"term": map[string]interface{}{
				"listId": *listId,
			},
		})
	}

	if done != nil {
		query["bool"].(map[string]interface{})["filter"] = append(query["bool"].(map[string]interface{})["filter"].([]interface{}), map[string]interface{}{
			"term": map[string]interface{}{
				"done": *done,
			},
		})
	}

	body := map[string]interface{}{
		"query":   query,
		"_source": []string{"id"},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("error encoding query: %w", err)
	}

	res, err := ls.client.Search(
		ls.client.Search.WithContext(ctx),
		ls.client.Search.WithIndex(itemIndexName),
		ls.client.Search.WithBody(&buf),
		ls.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("error executing search query: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error response from Elasticsearch: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing the search response: %w", err)
	}

	hits, ok := result["hits"].(map[string]interface{})["hits"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("error parsing hits from search response")
	}

	var ids []int
	for _, hit := range hits {
		source, ok := hit.(map[string]interface{})["_source"].(map[string]interface{})
		if !ok {
			continue
		}

		idFloat, ok := source["id"].(float64)
		if !ok {
			continue
		}

		id := int(idFloat)
		ids = append(ids, id)
	}

	return ids, nil
}

// Index indexes a new item in Elasticsearch with the given userId, listId, and item details
func (ls *ItemSearch) Index(ctx context.Context, userId int, listId int, item entity.Item) error {
	document := map[string]interface{}{
		"id":          item.Id,
		"userId":      userId,
		"listId":      listId,
		"title":       item.Title,
		"description": item.Description,
		"done":        item.Done,
	}

	data, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("error encoding item to JSON: %w", err)
	}

	req := bytes.NewReader(data)
	res, err := ls.client.Index(
		itemIndexName,
		req,
		ls.client.Index.WithContext(ctx),
		ls.client.Index.WithDocumentID(fmt.Sprintf("%d", item.Id)),
		ls.client.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("error indexing document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error response from Elasticsearch: %s", res.String())
	}

	return nil
}

// Delete removes an item from Elasticsearch by its itemId
func (ls *ItemSearch) Delete(ctx context.Context, itemId int) error {
	itemIdStr := fmt.Sprintf("%d", itemId)

	res, err := ls.client.Delete(
		itemIndexName,
		itemIdStr,
		ls.client.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("error deleting document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error response from Elasticsearch: %s", res.String())
	}

	return nil
}
