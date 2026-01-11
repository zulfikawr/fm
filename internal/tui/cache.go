package tui

import (
	"container/list"
	"encoding/gob"
	"os"
	"path/filepath"
	"time"
)

// MaxCacheEntries is the maximum number of entries to keep in LRU caches
const (
	MaxCacheEntries = 2000
)

// SizeCacheEntry represents a cached directory size with MTime validation
type SizeCacheEntry struct {
	Size      int64
	MTime     time.Time
	Timestamp time.Time
}

// LRUCache implements a persistent LRU cache for directory sizes
type LRUCache struct {
	capacity int
	cache    map[string]*list.Element
	lruList  *list.List
}

// NewLRUCache creates a new LRU cache with the specified capacity
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		lruList:  list.New(),
	}
}

// Get retrieves a value from the cache
func (c *LRUCache) Get(key string) (SizeCacheEntry, bool) {
	if elem, ok := c.cache[key]; ok {
		entry := elem.Value.(*lruEntry)
		c.lruList.MoveToFront(elem)
		return entry.value, true
	}
	return SizeCacheEntry{}, false
}

// Put adds or updates a value in the cache
func (c *LRUCache) Put(key string, value SizeCacheEntry) {
	if elem, ok := c.cache[key]; ok {
		// Update existing entry
		c.lruList.MoveToFront(elem)
		entry := elem.Value.(*lruEntry)
		entry.value = value
		entry.value.Timestamp = time.Now()
		return
	}

	// Add new entry
	value.Timestamp = time.Now()
	entry := &lruEntry{
		key:   key,
		value: value,
	}
	elem := c.lruList.PushFront(entry)
	c.cache[key] = elem

	// Evict oldest if over capacity
	if c.lruList.Len() > c.capacity {
		c.evict()
	}
}

func (c *LRUCache) evict() {
	oldest := c.lruList.Back()
	if oldest != nil {
		c.lruList.Remove(oldest)
		oldEntry := oldest.Value.(*lruEntry)
		delete(c.cache, oldEntry.key)
	}
}

type lruEntry struct {
	key   string
	value SizeCacheEntry
}

// Save persists the cache to disk
func (c *LRUCache) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Convert LRU list to a map for serialization
	data := make(map[string]SizeCacheEntry)
	for k, v := range c.cache {
		data[k] = v.Value.(*lruEntry).value
	}

	encoder := gob.NewEncoder(file)
	return encoder.Encode(data)
}

// Load populates the cache from disk
func (c *LRUCache) Load(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	var data map[string]SizeCacheEntry
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	for k, v := range data {
		c.Put(k, v)
	}
	return nil
}

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
