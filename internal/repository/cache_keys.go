package repository

import "time"

// Cache key patterns for various entities
const (
	patternListById  = "list_by_id:%d"
	patternUserLists = "user_lists:%d"
	patternItemById  = "item_by_id:%d"
	patternListItems = "list_items:%d"
)

// cacheKey represents a cache key with a pattern and TTL (time to live)
type cacheKey struct {
	Pattern string
	TTL     time.Duration
}

// Predefined cache keys with TTLs
var (
	cacheKeyListById = cacheKey{
		Pattern: patternListById,
		TTL:     time.Minute * 5,
	}
	cacheKeyUserLists = cacheKey{
		Pattern: patternUserLists,
		TTL:     time.Minute * 5,
	}
	cacheKeyItemById = cacheKey{
		Pattern: patternItemById,
		TTL:     time.Minute * 5,
	}
	cacheKeyListItems = cacheKey{
		Pattern: patternListItems,
		TTL:     time.Minute * 5,
	}
)
