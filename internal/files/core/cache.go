package core

import (
	"os"
	"sync"
	"time"
)

// CachedDir represents a directory's contents and its expiration time
type CachedDir struct {
	Entries []os.FileInfo
	Expiry  time.Time
}

// MetadataCache implements a short-lived TTL cache for directory listings
type MetadataCache struct {
	mu    sync.RWMutex
	cache map[string]CachedDir
	ttl   time.Duration
}

// NewMetadataCache creates a new metadata cache with the specified TTL
func NewMetadataCache(ttl time.Duration) *MetadataCache {
	return &MetadataCache{
		cache: make(map[string]CachedDir),
		ttl:   ttl,
	}
}

// Get retrieves a cached directory listing if it hasn't expired
func (c *MetadataCache) Get(path string) ([]os.FileInfo, bool) {
	if c == nil {
		return nil, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.cache[path]
	if !ok || time.Now().After(entry.Expiry) {
		return nil, false
	}
	return entry.Entries, true
}

// Put stores a directory listing in the cache
func (c *MetadataCache) Put(path string, entries []os.FileInfo) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[path] = CachedDir{
		Entries: entries,
		Expiry:  time.Now().Add(c.ttl),
	}
}

// Invalidate removes a specific path from the cache
func (c *MetadataCache) Invalidate(path string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, path)
}

// Clear empties the entire cache
func (c *MetadataCache) Clear() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]CachedDir)
}
