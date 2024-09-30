package repository_test

import (
	"bytes"
	"context"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/stretchr/testify/mock"
)

// MockCache is a mock implementation of the cache interface
type MockCache struct {
	mock.Mock
}

// Get simulates retrieving a value from the cache
func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	args := m.Called(ctx, key, dest)
	return args.Get(0).(bool), args.Error(1)
}

// Set simulates setting a value in the cache
func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

// Delete simulates deleting a key from the cache
func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// Exists simulates checking if a key exists in the cache
func (m *MockCache) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(bool), args.Error(1)
}

// MockElasticsearchClient is a mock implementation of an Elasticsearch client
type MockElasticsearchClient struct {
	mock.Mock
}

// Search simulates performing a search request in Elasticsearch
func (m *MockElasticsearchClient) Search(ctx context.Context, index string, body *bytes.Buffer) (*esapi.Response, error) {
	args := m.Called(ctx, index, body)
	return args.Get(0).(*esapi.Response), args.Error(1)
}

// Index simulates indexing a document in Elasticsearch
func (m *MockElasticsearchClient) Index(ctx context.Context, index string, documentID string, body *bytes.Reader) (*esapi.Response, error) {
	args := m.Called(ctx, index, documentID, body)
	return args.Get(0).(*esapi.Response), args.Error(1)
}

// Delete simulates deleting a document from Elasticsearch
func (m *MockElasticsearchClient) Delete(ctx context.Context, index string, documentID string) (*esapi.Response, error) {
	args := m.Called(ctx, index, documentID)
	return args.Get(0).(*esapi.Response), args.Error(1)
}
