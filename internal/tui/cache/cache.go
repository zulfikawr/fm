package cache

import (
	"container/list"
)

// SimpleCache implements a simple size-limited cache for cursor/offset memory
type SimpleCache struct {
	capacity int
	cache    map[string]int
	order    *list.List
}

// NewSimpleCache creates a new simple cache with the specified capacity
func NewSimpleCache(capacity int) *SimpleCache {
	return &SimpleCache{
		capacity: capacity,
		cache:    make(map[string]int),
		order:    list.New(),
	}
}

// Get retrieves a value from the cache
func (c *SimpleCache) Get(key string) (int, bool) {
	val, ok := c.cache[key]
	return val, ok
}

// Put adds or updates a value in the cache
func (c *SimpleCache) Put(key string, value int) {
	// If already exists, just update
	if _, ok := c.cache[key]; ok {
		c.cache[key] = value
		return
	}

	// Add new entry
	c.cache[key] = value
	c.order.PushFront(key)

	// Evict oldest if over capacity
	if c.order.Len() > c.capacity {
		oldest := c.order.Back()
		if oldest != nil {
			c.order.Remove(oldest)
			delete(c.cache, oldest.Value.(string))
		}
	}
}
