package cache

import (
	"context"
	"time"
)

// Interface defines the methods for interacting with a cache system
type Interface interface {
	Get(ctx context.Context, key string, dest interface{}) (bool, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// Cache is a wrapper that holds references to a master and replica cache
type Cache struct {
	Master  Interface
	Replica Interface
}

// New creates a new Cache instance with the provided master and replica caches
func New(masterCache Interface, replicaCache Interface) *Cache {
	return &Cache{
		Master:  masterCache,
		Replica: replicaCache,
	}
}
