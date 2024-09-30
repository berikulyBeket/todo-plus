package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/berikulyBeket/todo-plus/internal/entity"

	"github.com/elastic/go-elasticsearch/v8"
)

const (
	listIndexName = "lists"
)

type ListSearch struct {
	client *elasticsearch.Client
}

func NewListSearch(client *elasticsearch.Client) List {
	return &ListSearch{client: client}
}

// SearchIds performs a search for lists owned by a user, with optional filtering by a search term
func (ls *ListSearch) SearchIds(ctx context.Context, userId int, searchText string) ([]int, error) {
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
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
		},
		"_source": []string{"id"},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, fmt.Errorf("error encoding query: %w", err)
	}

	res, err := ls.client.Search(
		ls.client.Search.WithContext(ctx),
		ls.client.Search.WithIndex(listIndexName),
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
		idStr := source["id"]
		switch v := idStr.(type) {
		case float64:
			ids = append(ids, int(v))
		case string:
			id, err := strconv.Atoi(v)
			if err == nil {
				ids = append(ids, id)
			}
		}
	}
	return ids, nil
}

// Index indexes a new list in Elasticsearch with the given userId and list details
func (ls *ListSearch) Index(ctx context.Context, userId int, list *entity.List) error {
	document := map[string]interface{}{
		"id":          list.Id,
		"userId":      userId,
		"title":       list.Title,
		"description": list.Description,
	}

	data, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("error encoding list to JSON: %w", err)
	}

	req := bytes.NewReader(data)

	res, err := ls.client.Index(
		listIndexName,
		req,
		ls.client.Index.WithContext(ctx),
		ls.client.Index.WithDocumentID(fmt.Sprintf("%d", list.Id)),
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

// Delete removes a list from Elasticsearch by its listId
func (ls *ListSearch) Delete(ctx context.Context, listId int) error {
	listIdStr := fmt.Sprintf("%d", listId)

	res, err := ls.client.Delete(
		listIndexName,
		listIdStr,
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
