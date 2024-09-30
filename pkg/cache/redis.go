package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements the Interface using Redis as the caching backend
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new RedisCache instance with the provided Redis client
func NewRedisCache(client *redis.Client) Interface {
	return &RedisCache{client}
}

// Get retrieves a value from Redis by key and unmarshals it into the provided destination
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}

		return false, err
	}

	err = json.Unmarshal([]byte(val), dest)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Set stores a value in Redis as a JSON string with a specified time-to-live (ttl)
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, jsonValue, ttl).Err()
}

// Delete removes a value from Redis by key
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Exists checks if a key exists in Redis
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}
